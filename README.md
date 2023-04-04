
# PWS to MQTT dispatcher addon for Home Assistant



This addon was inspired by the [Wunderground Update Receiver Binding](https://www.openhab.org/addons/bindings/wundergroundupdatereceiver/)
 for OpenHAB.

The addon acts as a receiver for Personal Weather Stations (WH9000) and publishes the received updates to an MQTT topic, as well as on to [Wunderground.com](https://Wunderground.com).

## Installation

Add this URL to your HAAS add-on repositories:

https://github.com/peterzen/haas-pws-mqtt-addon


## Configuration

There is really only one configuration option, `mqtt_topic` that specifies the MQTT topic the data is to be pushed to.  The addon talks to the default MQTT service configured in HASS.
