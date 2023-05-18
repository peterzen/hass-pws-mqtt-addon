
# PWS to MQTT dispatcher addon for Home Assistant

[![Lint](https://github.com/peterzen/hass-pws-mqtt-addon/actions/workflows/lint.yaml/badge.svg)](https://github.com/peterzen/hass-pws-mqtt-addon/actions/workflows/lint.yaml)
[![Builder](https://github.com/peterzen/hass-pws-mqtt-addon/actions/workflows/builder.yaml/badge.svg)](https://github.com/peterzen/hass-pws-mqtt-addon/actions/workflows/builder.yaml)

This Home Assistant add-on retrieves live weather data from a WH2600 personal weather station (PWS) and publishes it to an MQTT topic where HAAS can access it as sensor information.

Supported weather stations: tested on Renkforce WH2600, probably also supports Froggit units but I can't verify it.

## Installation

Add this URL to your HAAS add-on repositories:

https://github.com/peterzen/hass-pws-mqtt-addon

