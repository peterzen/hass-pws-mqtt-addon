package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type WeatherData struct {
	ReceiverTime      string  `json:"receiverTime"`
	TemperatureIndoor float64 `json:"temperatureIndoor"`
	HumidityIndoor    float64 `json:"humidityIndoor"`
	PressureAbsolute  float64 `json:"pressureAbsolute"`
	PressureRelative  float64 `json:"pressureRelative"`
	Temperature       float64 `json:"temperature"`
	Humidity          float64 `json:"humidity"`
	WindDir           float64 `json:"windDir"`
	WindSpeed         float64 `json:"windSpeed"`
	WindGust          float64 `json:"windGust"`
	SolarRadiation    float64 `json:"solarRadiation"`
	Uv                float64 `json:"uv"`
	Uvi               float64 `json:"uvi"`
	PrecipHourlyRate  float64 `json:"precipHourlyRate"`
	PrecipDaily       float64 `json:"precipDaily"`
	PrecipWeekly      float64 `json:"precipWeekly"`
	PrecipMonthly     float64 `json:"precipMonthly"`
	PrecipYearly      float64 `json:"precipYearly"`
}

var pwsIp string
var fetchInterval int
var debugEnabled bool

func fetchDocumentFromPws() *goquery.Document {

	pwsUrl := fmt.Sprintf("http://%s/livedata.htm", pwsIp)
	// Read the HTML file
	resp, err := http.Get(pwsUrl)
	if err != nil {
		log.Printf("Failed to connect to PWS: %s\n", err)
		return nil
	}
	defer resp.Body.Close()

	htmlData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to fetch data from PWS: %s\n", err)
		return nil
	}

	// Parse the HTML file with goquery
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(htmlData)))

	if err != nil {
		log.Printf("Failed to process HTML: %s\n", err)
		return nil
	}
	if debugEnabled {
		log.Printf("Fetched data from PWS\n")
	}
	return doc
}

func parseFloat(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}
func parseHtml(doc *goquery.Document) []byte {

	var weatherData WeatherData

	doc.Find("table").Each(func(i int, s *goquery.Selection) {

		rows := s.Find("tr")

		// parse the table rows and extract the data
		rows.Each(func(i int, s *goquery.Selection) {
			value := s.Find("input").AttrOr("value", "")

			switch i {
			case 8:
				weatherData.ReceiverTime = value
			case 12:
				weatherData.TemperatureIndoor = parseFloat(value)
			case 13:
				weatherData.HumidityIndoor = parseFloat(value)
			case 14:
				weatherData.PressureAbsolute = parseFloat(value)
			case 15:
				weatherData.PressureRelative = parseFloat(value)
			case 16:
				weatherData.Temperature = parseFloat(value)
			case 17:
				weatherData.Humidity = parseFloat(value)
			case 18:
				weatherData.WindDir = parseFloat(value)
			case 19:
				weatherData.WindSpeed = parseFloat(value)
			case 20:
				weatherData.WindGust = parseFloat(value)
			case 21:
				weatherData.SolarRadiation = parseFloat(value)
			case 22:
				weatherData.Uv = parseFloat(value)
			case 23:
				weatherData.Uvi = parseFloat(value)
			case 24:
				weatherData.PrecipHourlyRate = parseFloat(value)
			case 25:
				weatherData.PrecipDaily = parseFloat(value)
			case 26:
				weatherData.PrecipWeekly = parseFloat(value)
			case 27:
				weatherData.PrecipMonthly = parseFloat(value)
			case 28:
				weatherData.PrecipYearly = parseFloat(value)
			default:
				return
			}
		})
	})

	// Convert the variables to JSON and print the result
	jsonData, err := json.Marshal(weatherData)
	if err != nil {
		log.Printf("Unable to marshal JSON: %s\n", err)
		return nil
	}
	return jsonData
}

func newMqttClient() mqtt.Client {

	pwsIp = os.Getenv("PWS_IP")
	if pwsIp == "" {
		log.Fatalf("PWS_IP env var undefined")
	}
	fetchIntervalStr := os.Getenv("FETCH_INTERVAL")
	if fetchIntervalStr == "" {
		log.Fatalf("FETCH_INTERVAL env var undefined")
	}
	var err error
	fetchInterval, err = strconv.Atoi(fetchIntervalStr)
	if err != nil {
		log.Fatalf("Invalid FETCH_INTERVAL value: %s\n", fetchIntervalStr)
	}

	// Get MQTT connection parameters from environment variables
	mqttBroker := os.Getenv("MQTT_HOST")
	if mqttBroker == "" {
		log.Fatalf("MQTT_HOST not configured\n")
	}
	mqttPort := os.Getenv("MQTT_PORT")
	if mqttPort == "" {
		mqttPort = "1883"
	}
	mqttUser := os.Getenv("MQTT_USER")
	if mqttUser == "" {
		log.Fatalf("MQTT_USER not configured\n")
	}
	mqttPassword := os.Getenv("MQTT_PASSWORD")

	mqttClientId := os.Getenv("MQTT_CLIENT_ID")
	if mqttClientId == "" {
		mqttClientId = "pwsmqttdispatcher"
	}
	// Set up MQTT client options
	mqttConnUri := fmt.Sprintf("tcp://%s:%s", mqttBroker, mqttPort)
	opts := mqtt.NewClientOptions().AddBroker(mqttConnUri)
	opts.SetClientID(mqttClientId)
	opts.SetUsername(mqttUser)
	opts.SetPassword(mqttPassword)

	// Create MQTT client
	client := mqtt.NewClient(opts)

	// Connect to MQTT broker
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		// panic(token.Error())
		log.Fatalf("Cannot connect to MQTT broker: %s\n", token.Error())
	}
	defer client.Disconnect(250)

	log.Printf("Connected to MQTT broker %s\n", opts.Servers[0])
	return client
}

func main() {

	debugEnabled = false

	if os.Getenv("DEBUG_ENABLED") == "true" {
		debugEnabled = true
	}

	client := newMqttClient()

	mqttTopic := os.Getenv(("MQTT_TOPIC"))
	if mqttTopic == "" {
		mqttTopic = "personal_weather_station"
	}
	log.Printf("Publishing to MQTT topic %s\n", mqttTopic)

	for {
		doc := fetchDocumentFromPws()

		if doc != nil {
			weatherData := parseHtml(doc)
			if weatherData != nil {
				token := client.Publish(mqttTopic, 0, false, weatherData)
				if debugEnabled {
					log.Printf("Published data to MQTT topic\n")
				}
				token.Wait()
				if debugEnabled {
					log.Println(string(weatherData))
				}
			}
		}
		time.Sleep(time.Duration(fetchInterval) * time.Second)
	}

}
