package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	"github.com/gpayer/bosch-bth-ra-fixer/mock"
)

func randomString(n int) string {
	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	b := make([]byte, n)
	for i := range b {
		b[i] = alphanum[rand.Intn(len(alphanum))]
	}
	return string(b)
}

var clientID = "fake-thermostat-" + randomString(8) // Change this to something random if using a public test server
const modeCommandTopic = "zigbee2mqtt/thermostat_arbeitszimmer/set"
const temperatureCommandTopic = "zigbee2mqtt/thermostat_arbeitszimmer/set/occupied_heating_setpoint"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	u, err := url.Parse("mqtt://localhost:1883")
	if err != nil {
		panic(err)
	}

	router := paho.NewStandardRouter()
	router.DefaultHandler(func(p *paho.Publish) { fmt.Printf("DEBUG: defaulthandler received message with topic: %s\n", p.Topic) })

	cliCfg := autopaho.ClientConfig{
		ServerUrls:                    []*url.URL{u},
		KeepAlive:                     20,
		CleanStartOnInitialConnection: false,
		SessionExpiryInterval:         60,
		OnConnectionUp: func(cm *autopaho.ConnectionManager, connAck *paho.Connack) {
			fmt.Println("DEBUG: mqtt connection up")
			if _, err := cm.Subscribe(context.Background(), &paho.Subscribe{
				Subscriptions: []paho.SubscribeOptions{
					{Topic: modeCommandTopic, QoS: 1},
					{Topic: temperatureCommandTopic, QoS: 1},
				},
			}); err != nil {
				fmt.Printf("ERROR: failed to subscribe: %s", err)
			}
			fmt.Println("INFO: mqtt subscription made")
		},
		OnConnectError: func(err error) { fmt.Printf("ERROR: error whilst attempting connection: %s\n", err) },
		ClientConfig: paho.ClientConfig{
			ClientID: clientID,
			OnPublishReceived: []func(paho.PublishReceived) (bool, error){
				func(pr paho.PublishReceived) (bool, error) {
					router.Route(pr.Packet.Packet())
					return true, nil
				},
			},
			OnClientError: func(err error) { fmt.Printf("client error: %s\n", err) },
			OnServerDisconnect: func(d *paho.Disconnect) {
				if d.Properties != nil {
					fmt.Printf("INFO: server requested disconnect: %s\n", d.Properties.ReasonString)
				} else {
					fmt.Printf("INFO: server requested disconnect; reason code: %d\n", d.ReasonCode)
				}
			},
		},
	}

	c, err := autopaho.NewConnection(ctx, cliCfg)
	if err != nil {
		panic(err)
	}
	if err = c.AwaitConnection(ctx); err != nil {
		panic(err)
	}

	mockClimate := mock.NewClimate(c, "zigbee2mqtt/thermostat_arbeitszimmer", router)

	t := time.NewTicker(5 * time.Second)
	defer t.Stop()

	done := false
	for !done {
		select {
		case <-t.C:
			mockClimate.Run()
		case <-ctx.Done():
			done = true
		}
	}

	fmt.Println("INFO: signal caught - exiting")
	<-c.Done()
}
