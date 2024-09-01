package main

import (
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/xhhuango/json"

	"github.com/PuerkitoBio/goquery"
	"github.com/cdzombak/libwx"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type WeatherData struct {
	ReceiverTime         string  `json:"receiverTime"`
	ReceiverTimestamp    int64   `json:"receiverTimestamp"`
	TemperatureIndoor    float64 `json:"temperatureIndoor"`
	HumidityIndoor       float64 `json:"humidityIndoor"`
	PressureAbsolute     float64 `json:"pressureAbsolute"`
	PressureRelative     float64 `json:"pressureRelative"`
	Temperature          float64 `json:"temperature"`
	Humidity             float64 `json:"humidity"`
	DewPoint             float64 `json:"dewPoint"`
	WindDir              float64 `json:"windDir"`
	WindDirCardinal      string  `json:"windDirCardinal"`
	WindSpeed            float64 `json:"windSpeed"`
	WindGust             float64 `json:"windGust"`
	WindChill            float64 `json:"windChill"`
	SolarRadiation       float64 `json:"solarRadiation"`
	Uv                   float64 `json:"uv"`
	Uvi                  float64 `json:"uvi"`
	PrecipHourlyRate     float64 `json:"precipHourlyRate"`
	PrecipDaily          float64 `json:"precipDaily"`
	PrecipWeekly         float64 `json:"precipWeekly"`
	PrecipMonthly        float64 `json:"precipMonthly"`
	PrecipYearly         float64 `json:"precipYearly"`
	HeatIndex            float64 `json:"heatIndex"`
	IndoorSensorId       string  `json:"indoorSensorId"`
	OutdoorSensorId      string  `json:"outdoorSensorId"`
	IndoorSensorBattery  string  `json:"indoorSensorBattery"`
	OutdoorSensorBattery string  `json:"outdoorSensorBattery"`
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

	htmlData, err := io.ReadAll(resp.Body)
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
		return math.NaN()
	}
	return f
}
func parseHtml(doc *goquery.Document) WeatherData {

	var weatherData WeatherData

	doc.Find("table").Each(func(i int, s *goquery.Selection) {

		rows := s.Find("tr")

		// parse the table rows and extract the data
		rows.Each(func(i int, s *goquery.Selection) {
			inputs := s.Find("input")
			value := inputs.AttrOr("value", "")

			switch i {
			case 8:
				weatherData.ReceiverTime = value
			case 9:
				weatherData.IndoorSensorId = value
				weatherData.IndoorSensorBattery = inputs.Eq(1).AttrOr("value", "")
			case 10:
				weatherData.OutdoorSensorId = value
				weatherData.OutdoorSensorBattery = inputs.Eq(1).AttrOr("value", "")
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

func addCalculatedData(wd WeatherData) WeatherData {
	heatIndex := libwx.HeatIndexC(libwx.TempC(wd.Temperature), libwx.RelHumidity(wd.Humidity))
	wd.HeatIndex = roundFloatTo1Decimal(float64(heatIndex))

	windDirCardinal := windDirToCardinal(int(wd.WindDir))
	wd.WindDirCardinal = windDirCardinal

	windChill := libwx.WindChillC(libwx.TempC(wd.Temperature), libwx.SpeedKmH(wd.WindSpeed).Mph())
	wd.WindChill = roundFloatTo1Decimal(float64(windChill))

	dewPoint := libwx.DewPointC(libwx.TempC(wd.Temperature), libwx.RelHumidity(wd.Humidity))
	wd.DewPoint = roundFloatTo1Decimal(float64(dewPoint))

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
