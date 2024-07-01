#!/bin/sh

mqttui -b mqtt://localhost:1883 p -r homeassistant/climate/0x18fc2600000d7ae2/climate/config "`cat broken_config.json`"
