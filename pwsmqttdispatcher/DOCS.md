# PWS to MQTT dispatcher addon

## How to use

Once installed and started, the addon periodically connects to the PWS and fetches the live weather information.  The data will be published in the configured MQTT topic.


## Configuration
| Option | |
|-----------| ----------------------------------------|
| `pws_ip`  | The IP address of the PWS observer unit |
| `mqtt_topic`     |   MQTT topic the data is to be pushed to
| `mqtt_client_id` |   Optional MQTT client ID
| `fetch_interval` |   Seconds between updates
| `debug_enabled`  |   More debug logging

The addon connects to the default MQTT broker configured in HASS, there is no way to configure this at this time.


<small>Images: weather station by Knut M. Synstad from <a href="https://thenounproject.com/browse/icons/term/weather-station/" target="_blank" title="weather station Icons">Noun Project</a></small>