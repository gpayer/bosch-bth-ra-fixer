# HA addon: Bosch BTH-RA fixer

_This addon is only useful for a very limited amount of time!_

This addon automatically fixed broken discovery configs created by Zigbee2MQTT. But only for Bosch BTH-RA climate devices. Nothing else.

This is only useful if you have Zigbee2MQTT versions 1.38.0 or 1.39.0. In 1.39.1 this bug will be fixed.

## Configuration

### With Mosquitto addon and MQTT integration

Leave all fields in the configuration form empty, the necessary values are sent over by HA supervisor.

### Custom MQTT server

Fill in the MQTT URI, e.g. `mqtt://mqtt-server.local:1883` and optionally username and password. Check the logs after starting the addon
to see if everything worked.
