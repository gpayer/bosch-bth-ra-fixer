package rewriter

import (
	"encoding/json"
	"fmt"

	"github.com/eclipse/paho.golang/paho"
)

func HandleClimateConfigMessage(pr paho.PublishReceived) (bool, error) {
	var config Config
	err := json.Unmarshal(pr.Packet.Payload, &config)
	if err != nil {
		return false, err
	}

	// TODO: check if config needs to be corrected, if yes, correct it
	if config.Device.Manufacturer == "Bosch" && config.Device.Model == "Radiator thermostat II (BTH-RA)" {
		fmt.Println("DEBUG: Found a Bosch Radiator thermostat II (BTH-RA) device")
	}
	return true, nil
}
