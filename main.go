package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	"github.com/gpayer/bosch-bth-ra-fixer/rewriter"
)

func randomString(n int) string {
	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	b := make([]byte, n)
	for i := range b {
		b[i] = alphanum[rand.Intn(len(alphanum))]
	}
	return string(b)
}

var clientID = "bosch-bth-ra-fixer-" + randomString(8) // Change this to something random if using a public test server
const topic = "homeassistant/climate/#"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	mqttURI := os.Getenv("MQTT_URI")
	if mqttURI == "" {
		mqttURI = "mqtt://localhost:1883"
	}

	u, err := url.Parse(mqttURI)
	if err != nil {
		panic(err)
	}

	cliCfg := autopaho.ClientConfig{
		ServerUrls:                    []*url.URL{u},
		KeepAlive:                     20,
		CleanStartOnInitialConnection: false,
		SessionExpiryInterval:         60,
		OnConnectionUp: func(cm *autopaho.ConnectionManager, connAck *paho.Connack) {
			fmt.Println("DEBUG: mqtt connection up")
			if _, err := cm.Subscribe(context.Background(), &paho.Subscribe{
				Subscriptions: []paho.SubscribeOptions{
					{Topic: topic, QoS: 1},
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
				rewriter.HandleClimateConfigMessage,
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

	<-ctx.Done()

	fmt.Println("INFO: signal caught - exiting")
	<-c.Done()
}
