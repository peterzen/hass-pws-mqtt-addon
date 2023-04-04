package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type WeatherData struct {
	Id             string  `json:"id"`
	Baromin        float64 `json:"baromin"`
	Indoorhumidity float64 `json:"indoorhumidity"`
	Indoortempf    float64 `json:"indoortempf"`
	Rtfreq         float64 `json:"rtfreq"`
	Dewptf         float64 `json:"dewptf"`
	Windchillf     float64 `json:"windchillf"`
	Dailyrainin    float64 `json:"dailyrainin"`
	Weeklyrainin   float64 `json:"weeklyrainin"`
	Tempf          float64 `json:"tempf"`
	Windgustmph    float64 `json:"windgustmph"`
	Rainin         float64 `json:"rainin"`
	Solarradiation float64 `json:"solarradiation"`
	UV             float64 `json:"UV"`
	Humidity       float64 `json:"humidity"`
	Winddir        float64 `json:"winddir"`
	Windspeedmph   float64 `json:"windspeedmph"`
	Dateutc        string  `json:"dateutc"`
	Monthlyrainin  float64 `json:"monthlyrainin"`
	Yearlyrainin   float64 `json:"yearlyrainin"`
}

func extractWeatherData(params url.Values) (WeatherData, error) {
	wd := WeatherData{}

	// Get values from query parameters
	id := params.Get("ID")
	baromin, err := strconv.ParseFloat(params.Get("baromin"), 64)
	if err != nil {
		return wd, err
	}
	indoorhumidity, err := strconv.ParseFloat(params.Get("indoorhumidity"), 64)
	if err != nil {
		return wd, err
	}
	indoortempf, err := strconv.ParseFloat(params.Get("indoortempf"), 64)
	if err != nil {
		return wd, err
	}
	rtfreq, err := strconv.ParseFloat(params.Get("rtfreq"), 64)
	if err != nil {
		return wd, err
	}
	dewptf, err := strconv.ParseFloat(params.Get("dewptf"), 64)
	if err != nil {
		return wd, err
	}
	windchillf, err := strconv.ParseFloat(params.Get("windchillf"), 64)
	if err != nil {
		return wd, err
	}
	dailyrainin, err := strconv.ParseFloat(params.Get("dailyrainin"), 64)
	if err != nil {
		return wd, err
	}
	weeklyrainin, err := strconv.ParseFloat(params.Get("weeklyrainin"), 64)
	if err != nil {
		return wd, err
	}
	tempf, err := strconv.ParseFloat(params.Get("tempf"), 64)
	if err != nil {
		return wd, err
	}
	windgustmph, err := strconv.ParseFloat(params.Get("windgustmph"), 64)
	if err != nil {
		return wd, err
	}
	rainin, err := strconv.ParseFloat(params.Get("rainin"), 64)
	if err != nil {
		return wd, err
	}
	solarradiation, err := strconv.ParseFloat(params.Get("solarradiation"), 64)
	if err != nil {
		return wd, err
	}
	UV, err := strconv.ParseFloat(params.Get("UV"), 64)
	if err != nil {
		return wd, err
	}
	humidity, err := strconv.ParseFloat(params.Get("humidity"), 64)
	if err != nil {
		return wd, err
	}
	winddir, err := strconv.ParseFloat(params.Get("winddir"), 64)
	if err != nil {
		return wd, err
	}
	windspeedmph, err := strconv.ParseFloat(params.Get("windspeedmph"), 64)
	if err != nil {
		return wd, err
	}
	dateutc := params.Get("dateutc")
	monthlyrainin, err := strconv.ParseFloat(params.Get("monthlyrainin"), 64)
	if err != nil {
		return wd, err
	}
	yearlyrainin, err := strconv.ParseFloat(params.Get("yearlyrainin"), 64)
	if err != nil {
		return wd, err
	}

	// Set values in WeatherData struct
	wd.Id = id
	wd.Baromin = baromin
	wd.Indoorhumidity = indoorhumidity
	wd.Indoortempf = indoortempf
	wd.Rtfreq = rtfreq
	wd.Dewptf = dewptf
	wd.Windchillf = windchillf
	wd.Dailyrainin = dailyrainin
	wd.Weeklyrainin = weeklyrainin
	wd.Tempf = tempf
	wd.Windgustmph = windgustmph
	wd.Rainin = rainin
	wd.Solarradiation = solarradiation
	wd.UV = UV
	wd.Humidity = humidity
	wd.Winddir = winddir
	wd.Windspeedmph = windspeedmph
	wd.Dateutc = dateutc
	wd.Monthlyrainin = monthlyrainin
	wd.Yearlyrainin = yearlyrainin
	return wd, nil
}

var wuSubmitUrl = "http://rtupdate.wunderground.com/weatherstation/updateweatherstation.php"

func submitUpdateToWunderground(r *http.Request) {
	// Submit the original HTTP request to rtupdate.wunderground.com
	originalPayload := r.URL.Query().Encode()
	resp, err := http.Post(wuSubmitUrl, "application/x-www-form-urlencoded", strings.NewReader(originalPayload))
	if err != nil {
		log.Printf("Failed to submit original HTTP request: %v", err)
		return
	}
	defer resp.Body.Close()
	log.Printf("Submitted data to %s\n", wuSubmitUrl)
}

func main() {
	// Get MQTT connection parameters from environment variables
	mqttBroker := os.Getenv("MQTT_HOST")
	mqttPort := os.Getenv("MQTT_PORT")
	if mqttPort == "" {
		mqttPort = "1883"
	}
	mqttUser := os.Getenv("MQTT_USER")
	mqttPassword := os.Getenv("MQTT_PASSWORD")
	mqttTopic := os.Getenv(("MQTT_TOPIC"))
	if mqttTopic == "" {
		mqttTopic = "personal_weather_station"
	}

	// Set up MQTT client options
	mqttConnUri := fmt.Sprintf("tcp://%s:%s", mqttBroker, mqttPort)
	opts := mqtt.NewClientOptions().AddBroker(mqttConnUri)
	opts.SetClientID("pwsmqttdispatcher")
	opts.SetUsername(mqttUser)
	opts.SetPassword(mqttPassword)

	// Create MQTT client
	client := mqtt.NewClient(opts)

	log.Printf("Connecting to MQTT broker %s\n", mqttConnUri)

	// Connect to MQTT broker
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		// panic(token.Error())
		log.Fatalf("Cannot connect to MQTT broker: %s\n", token.Error())
	}
	defer client.Disconnect(250)
	log.Printf("Connected to MQTT broker %s\n", opts.Servers[0])

	// Set up HTTP server to listen for incoming requests
	http.HandleFunc("/weatherstation/updateweatherstation.php", func(w http.ResponseWriter, r *http.Request) {

		log.Printf("Update received from %s\n", r.RemoteAddr)

		// Extract weather data from request
		weatherData, err := extractWeatherData(r.URL.Query())
		if err != nil {
			http.Error(w, "Invalid weather data: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Publish weather data to MQTT topic in JSON format
		jsonData, err := json.Marshal(weatherData)
		if err != nil {
			http.Error(w, "Error marshalling weather data to JSON: "+err.Error(), http.StatusInternalServerError)
			return
		}
		token := client.Publish(mqttTopic, 0, false, jsonData)
		token.Wait()

		log.Printf("Published weather data to topic %s\n", mqttTopic)

		// Return success status
		w.WriteHeader(http.StatusOK)

		submitUpdateToWunderground(r)
	})

	// Start HTTP server
	log.Println("Starting server on :8765")
	if err := http.ListenAndServe(":8765", nil); err != nil {
		log.Fatal(err)
	}
}
