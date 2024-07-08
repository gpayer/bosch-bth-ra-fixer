package rewriter

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/eclipse/paho.golang/paho"
)

func HandleClimateConfigMessage(pr paho.PublishReceived) (bool, error) {
	if len(pr.Packet.Payload) == 0 {
		return false, nil
	}

	var config Config
	err := json.Unmarshal(pr.Packet.Payload, &config)
	if err != nil {
		fmt.Printf("ERROR: unmarshaling payload failed: %v\n", err)
		return false, err
	}

	if config.Device.Manufacturer == "Bosch" && config.Device.Model == "Radiator thermostat II (BTH-RA)" {
		fmt.Println("DEBUG: Found a Bosch Radiator thermostat II (BTH-RA) device")
		if config.ModeCommandTemplate == "" && len(config.Modes) == 1 {
			fmt.Println("DEBUG: Found an unfixed Bosch Radiator thermostat II (BTH-RA) device")

			config.ModeCommandTopic = strings.TrimSuffix(config.ModeCommandTopic, "/system_mode")
			config.ModeCommandTemplate = "{% set values = { 'auto':'schedule','heat':'manual','off':'pause'} %}{\"operating_mode\": \"{{ values[value] if value in values.keys() else 'pause' }}\"}"
			config.ModeStateTemplate = "{% set values = {'schedule':'auto','manual':'heat','pause':'off'} %}{% set value = value_json.operating_mode %}{{ values[value] if value in values.keys() else 'off' }}"
			config.Modes = []string{"off", "heat", "auto"}

			payload, err := json.Marshal(config)
			if err != nil {
				fmt.Printf("ERROR: marshaling payload failed: %v\n", err)
				return false, err
			}

			pr.Client.Publish(context.Background(), (&paho.Publish{
				QoS:     1,
				Retain:  true,
				Topic:   pr.Packet.Topic,
				Payload: payload,
			}))
		}
	}
	return true, nil
}
