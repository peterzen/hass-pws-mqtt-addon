package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
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
	ReceiverTimestamp int64   `json:"receiverTimestamp"`
	TemperatureIndoor float64 `json:"temperatureIndoor"`
	HumidityIndoor    float64 `json:"humidityIndoor"`
	PressureAbsolute  float64 `json:"pressureAbsolute"`
	PressureRelative  float64 `json:"pressureRelative"`
	Temperature       float64 `json:"temperature"`
	Humidity          float64 `json:"humidity"`
	DewPoint          float64 `json:"dewPoint"`
	WindDir           float64 `json:"windDir"`
	WindDirCardinal   string  `json:"windDirCardinal"`
	WindSpeed         float64 `json:"windSpeed"`
	WindGust          float64 `json:"windGust"`
	WindChill         float64 `json:"windChill"`
	SolarRadiation    float64 `json:"solarRadiation"`
	Uv                float64 `json:"uv"`
	Uvi               float64 `json:"uvi"`
	PrecipHourlyRate  float64 `json:"precipHourlyRate"`
	PrecipDaily       float64 `json:"precipDaily"`
	PrecipWeekly      float64 `json:"precipWeekly"`
	PrecipMonthly     float64 `json:"precipMonthly"`
	PrecipYearly      float64 `json:"precipYearly"`
	HeatIndex         float64 `json:"heatIndex"`
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
func parseHtml(doc *goquery.Document) WeatherData {

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

	return weatherData
}

func weatherDataAsJson(wd WeatherData) []byte {
	// Convert the variables to JSON and print the result
	jsonData, err := json.Marshal(wd)
	if err != nil {
		log.Printf("Unable to marshal JSON: %s\n", err)
		return nil
	}
	return jsonData
}

func calculateHeatIndex(temperature, humidity float64) float64 {
	// Convert temperature to Fahrenheit
	temperature = (temperature * 1.8) + 32

	// https://www.wpc.ncep.noaa.gov/html/heatindex_equation.shtml
	if temperature >= 80 {
		// Calculate the heat index in Fahrenheit
		heatIndex := -42.379 + 2.04901523*temperature + 10.14333127*humidity - 0.22475541*temperature*humidity - 6.83783e-3*math.Pow(temperature, 2) - 5.481717e-2*math.Pow(humidity, 2) + 1.22874e-3*math.Pow(temperature, 2)*humidity + 8.5282e-4*temperature*math.Pow(humidity, 2) - 1.99e-6*math.Pow(temperature, 2)*math.Pow(humidity, 2)

		// Convert heat index back to Celsius
		heatIndex = (heatIndex - 32) * (5.0 / 9.0)
		return heatIndex
	}
	return 0
}

func windDirToCardinal(windDirDeg int) string {
	dir := []string{"N ⬇️", "NNE ⬇️", "NE ↙️", "ENE ⬅️", "E ⬅️", "ESE ⬅️", "SE ↖️", "SSE ⬆️", "S ⬆️", "SSW ⬆️", "SW ↗️", "WSW ➡️", "W ➡️", "WNW ➡️", "NW ↘️", "NNW ⬇️"}
	wind := windDirDeg % 360
	winddiroffset := (float64(wind) + (360.0 / 32.0)) / 360.0
	winddiridx := int(math.Floor(winddiroffset / (1.0 / 16.0)))

	if winddiridx >= len(dir) {
		if debugEnabled {
			log.Printf("windDirToCardinal calculated invalid index %d (deg: %d)\n", winddiridx, windDirDeg)
		}
		return ""
	}
	winddir := dir[winddiridx]
	return winddir
}

func dateToUnixTimestamp(dateStr string) (int64, error) {
	layout := "15:04 1/2/2006" // day and month are swapped compared to the US format
	location, err := time.LoadLocation("CET")
	if err != nil {
		return 0, err
	}
	t, err := time.ParseInLocation(layout, dateStr, location)
	if err != nil {
		if debugEnabled {
			log.Printf("Cannot convert date from '%s': %s\n", dateStr, err)
		}
		return 0, err
	}
	return t.Unix(), nil
}
func calculateWindChill(windSpeed float64, temp float64) float64 {
	// Calculate the wind chill temperature in Celsius using the National Weather Service's formula
	// where T is the air temperature in Celsius and V is the wind speed in km/h

	// A Wind Chill value cannot be calculated for wind speeds less than 4.8 kilometers/hour
	if windSpeed < 4.8 {
		return temp
	}

	V := windSpeed / 1.609344 // convert wind speed to miles per hour
	T := temp*1.8 + 32        // convert temperature to Fahrenheit
	WCI := 35.74 + 0.6215*T - 35.75*math.Pow(V, 0.16) + 0.4275*T*math.Pow(V, 0.16)
	windChill := (WCI - 32) * 5 / 9 // convert wind chill to Celsius
	return windChill
}

func calculateDewPoint(tempCelsius, humidity float64) float64 {
	a := 17.27
	b := 237.7
	alpha := ((a * tempCelsius) / (b + tempCelsius)) + math.Log(humidity/100.0)
	dewPointCelsius := (b * alpha) / (a - alpha)
	return dewPointCelsius
}

func addCalculatedData(wd WeatherData) WeatherData {
	heatIndex := calculateHeatIndex(wd.Temperature, wd.Humidity)
	wd.HeatIndex = roundFloatTo1Decimal(heatIndex)

	windDirCardinal := windDirToCardinal(int(wd.WindDir))
	wd.WindDirCardinal = windDirCardinal

	windChill := calculateWindChill(wd.WindSpeed, wd.Temperature)
	wd.WindChill = roundFloatTo1Decimal(windChill)

	dewPoint := calculateDewPoint(wd.Temperature, wd.Humidity)
	wd.DewPoint = roundFloatTo1Decimal(dewPoint)

	recTs, err := dateToUnixTimestamp(wd.ReceiverTime)
	if err == nil {
		wd.ReceiverTimestamp = recTs
	}
	return wd
}

func roundFloatTo1Decimal(f float64) float64 {
	return math.Round(f*10) / 10
}

func main() {

	debugEnabled = false

	if os.Getenv("DEBUG_ENABLED") == "true" {
		debugEnabled = true
	}

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

	mqttTopic := os.Getenv(("MQTT_TOPIC"))
	if mqttTopic == "" {
		mqttTopic = "personal_weather_station"
	}
	log.Printf("Publishing to MQTT topic %s\n", mqttTopic)

	for {
		doc := fetchDocumentFromPws()

		if doc != nil {
			weatherData := parseHtml(doc)
			weatherData = addCalculatedData(weatherData)
			weatherDataJson := weatherDataAsJson(weatherData)
			if weatherDataJson != nil {
				token := client.Publish(mqttTopic, 0, false, weatherDataJson)
				if debugEnabled {
					log.Printf("Published data to #%s\n", mqttTopic)
				}
				token.Wait()
				if debugEnabled {
					log.Println(string(weatherDataJson))
				}
			}
		}
		time.Sleep(time.Duration(fetchInterval) * time.Second)
	}

}
