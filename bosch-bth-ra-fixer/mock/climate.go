package mock

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
)

type Climate struct {
	mtx                     sync.Mutex
	conn                    *autopaho.ConnectionManager
	topic                   string
	deviceID                string
	RunningState            string  `json:"running_state"` // idle, heat
	LocalTemperature        float64 `json:"local_temperature"`
	OperatingMode           string  `json:"operating_mode"` // pause, manual, schedule
	OccupiedHeatingSetpoint float64 `json:"occupied_heating_setpoint"`
}

func NewClimate(conn *autopaho.ConnectionManager, deviceID string, router paho.Router) *Climate {
	c := &Climate{
		conn:                    conn,
		topic:                   "zigbee2mqtt/" + deviceID,
		deviceID:                deviceID,
		RunningState:            "idle",
		LocalTemperature:        21.0,
		OperatingMode:           "pause",
		OccupiedHeatingSetpoint: 5.0,
	}

	c.mtx.Lock()
	defer c.mtx.Unlock()

	router.RegisterHandler(c.topic+"/set", c.handleOperatingModeSet)
	router.RegisterHandler(c.topic+"/set/occupied_heating_setpoint", c.handleOccupiedHeatingSetpointSet)

	c.conn.Publish(context.Background(), &paho.Publish{
		QoS:     1,
		Retain:  true,
		Topic:   "zigbee2mqtt/bridge/state",
		Payload: []byte(`{"state":"online"}`),
	})

	c.conn.Publish(context.Background(), &paho.Publish{
		QoS:     1,
		Retain:  true,
		Topic:   "homeassistant/climate/" + deviceID + "/config",
		Payload: []byte(createHAConfigPayload(deviceID)),
	})

	c.publishState()

	return c
}

func (c *Climate) handleOperatingModeSet(p *paho.Publish) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	fmt.Printf("DEBUG: handleOperatingModeSet: %s\n", p.Payload)

	var payload Climate
	if err := json.Unmarshal(p.Payload, &payload); err == nil {
		c.OperatingMode = payload.OperatingMode
	} else {
		fmt.Printf("ERROR: unmarshaling payload failed: %v\n", err)
	}

	c.publishState()
}

func (c *Climate) handleOccupiedHeatingSetpointSet(p *paho.Publish) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	fmt.Printf("DEBUG: handleOccupiedHeatingSetpointSet: %s\n", p.Payload)

	var payload float64
	if err := json.Unmarshal(p.Payload, &payload); err == nil {
		c.OccupiedHeatingSetpoint = payload
	} else {
		fmt.Printf("ERROR: unmarshaling payload failed: %v\n", err)
	}

	c.publishState()
}

func (c *Climate) publishState() {
	payload, err := json.Marshal(c)
	if err != nil {
		fmt.Printf("ERROR: marshaling payload failed: %v\n", err)
		return
	}

	if _, err := c.conn.Publish(context.Background(), &paho.Publish{
		QoS:     1,
		Retain:  true,
		Topic:   c.topic,
		Payload: payload,
	}); err != nil {
		fmt.Printf("ERROR: publishing failed: %v\n", err)
	}
}

func (c *Climate) Run() {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	doPublish := false

	if c.OperatingMode == "pause" {
		if c.RunningState != "idle" {
			c.RunningState = "idle"
			doPublish = true
		}
		if c.OccupiedHeatingSetpoint != 5.0 {
			c.OccupiedHeatingSetpoint = 5.0
			doPublish = true
		}
		if c.LocalTemperature > 20.0 {
			c.LocalTemperature -= 0.1
			doPublish = true
		}
	} else if c.OperatingMode == "manual" {
		oldState := Climate{
			RunningState:            c.RunningState,
			LocalTemperature:        c.LocalTemperature,
			OperatingMode:           c.OperatingMode,
			OccupiedHeatingSetpoint: c.OccupiedHeatingSetpoint,
		}

		if c.OccupiedHeatingSetpoint > c.LocalTemperature+.5 {
			c.RunningState = "heat"
		} else if c.OccupiedHeatingSetpoint < c.LocalTemperature-.5 {
			c.RunningState = "idle"
		}

		if c.RunningState == "heat" {
			c.LocalTemperature += 0.2
		} else if c.LocalTemperature > 20.0 {
			c.LocalTemperature -= 0.1
		}

		// compare oldState with c
		if oldState.RunningState != c.RunningState || oldState.LocalTemperature != c.LocalTemperature {
			doPublish = true
		}
	}

	if doPublish {
		c.publishState()
	}
}

func createHAConfigPayload(deviceID string) string {
	return `
{
  "action_template": "{% set values = {None:None,'idle':'idle','heat':'heating','cool':'cooling','fan_only':'fan'} %}{{ values[value_json.running_state] }}",
  "action_topic": "zigbee2mqtt/` + deviceID + `",
  "availability": [
    {
      "topic": "zigbee2mqtt/bridge/state",
      "value_template": "{{ value_json.state }}"
    }
  ],
  "current_temperature_template": "{{ value_json.local_temperature }}",
  "current_temperature_topic": "zigbee2mqtt/` + deviceID + `",
  "device": {
    "identifiers": [
      "zigbee2mqtt_` + deviceID + `"
    ],
    "manufacturer": "Bosch",
    "model": "Radiator thermostat II (BTH-RA)",
    "name": "Thermostat Arbeitszimmer",
    "sw_version": "3.05.09",
    "via_device": "zigbee2mqtt_bridge_0x00124b002b4866eb"
  },
  "max_temp": "30",
  "min_temp": "5",
  "mode_command_topic": "zigbee2mqtt/` + deviceID + `/set/system_mode",
  "mode_state_template": "{{ value_json.system_mode }}",
  "mode_state_topic": "zigbee2mqtt/` + deviceID + `",
  "modes": [
    "heat"
  ],
  "object_id": "` + deviceID + `",
  "origin": {
    "name": "Zigbee2MQTT",
    "sw": "1.38.0",
    "url": "https://www.zigbee2mqtt.io"
  },
  "temp_step": 0.5,
  "temperature_command_topic": "zigbee2mqtt/` + deviceID + `/set/occupied_heating_setpoint",
  "temperature_state_template": "{{ value_json.occupied_heating_setpoint }}",
  "temperature_state_topic": "zigbee2mqtt/` + deviceID + `",
  "temperature_unit": "C",
  "unique_id": "` + deviceID + `_climate_zigbee2mqtt"
}
  `
}
