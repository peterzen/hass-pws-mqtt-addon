package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Variable struct {
	VarName        string `json:"var_name"`
	VarDescription string `json:"var_description"`
	Value          string `json:"value"`
}

// Extract the variables from each row
var pwsDataModel = []Variable{
	{VarName: "", VarDescription: "", Value: ""},
	{VarName: "", VarDescription: "", Value: ""},
	{VarName: "", VarDescription: "", Value: ""},
	{VarName: "", VarDescription: "", Value: ""},
	{VarName: "", VarDescription: "", Value: ""},
	{VarName: "", VarDescription: "", Value: ""},
	{VarName: "", VarDescription: "", Value: ""},
	{VarName: "", VarDescription: "", Value: ""},
	{VarName: "receiverTime", VarDescription: "Receiver Time", Value: ""},
	{VarName: "", VarDescription: "", Value: ""},
	{VarName: "", VarDescription: "", Value: ""},
	{VarName: "", VarDescription: "", Value: ""},
	{VarName: "temperatureIndoor", VarDescription: "Indoor Temperature", Value: ""},
	{VarName: "humidityIndoor", VarDescription: "Indoor Humidity", Value: ""},
	{VarName: "pressureAbsolute", VarDescription: "Absolute Pressure", Value: ""},
	{VarName: "pressureRelative", VarDescription: "Relative Pressure", Value: ""},
	{VarName: "temperature", VarDescription: "Outdoor Temperature", Value: ""},
	{VarName: "humidity", VarDescription: "Outdoor Humidity", Value: ""},
	{VarName: "windDir", VarDescription: "Wind Direction", Value: ""},
	{VarName: "windSpeed", VarDescription: "Wind Speed", Value: ""},
	{VarName: "windGust", VarDescription: "Wind Gust", Value: ""},
	{VarName: "solarRadiation", VarDescription: "Solar Radiation", Value: ""},
	{VarName: "uv", VarDescription: "UV", Value: ""},
	{VarName: "uvi", VarDescription: "UVI", Value: ""},
	{VarName: "precipHourlyRate", VarDescription: "Hourly Rain Rate", Value: ""},
	{VarName: "precipDaily", VarDescription: "Daily Rain", Value: ""},
	{VarName: "precipWeekly", VarDescription: "Weekly Rain", Value: ""},
	{VarName: "precipMonthly", VarDescription: "Monthly Rain", Value: ""},
	{VarName: "precipYearly", VarDescription: "Yearly Rain", Value: ""},
}

func fetchDocumentFromPws() *goquery.Document {
	// Read the HTML file
	resp, err := http.Get("http://10.11.50.147/livedata.htm")
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
	return doc
}

func parseHtml(doc *goquery.Document) []byte {

	var vars []Variable

	doc.Find("table").Each(func(i int, s *goquery.Selection) {

		rows := s.Find("tr")

		// parse the table rows and extract the data
		rows.Each(func(i int, s *goquery.Selection) {

			if i >= len(pwsDataModel) {
				return
			}

			m := pwsDataModel[i]
			if m.VarName == "" {
				return
			}

			// ignore the form tag and any non-data rows
			if s.Find("form").Length() > 0 || s.Find("td").Length() < 2 {
				return
			}

			value := s.Find("input").AttrOr("value", "")

			m.Value = value

			vars = append(vars, m)
		})
	})

	// Convert the variables to JSON and print the result
	jsonData, err := json.Marshal(vars)
	if err != nil {
		log.Printf("Unable to marshal JSON: %s\n", err)
		return nil
	}
	return jsonData
}
func main() {

	for {
		doc := fetchDocumentFromPws()

		if doc != nil {
			jsonData := parseHtml(doc)

			if jsonData != nil {
				fmt.Println(string(jsonData))
			}

		}
		time.Sleep(30 * time.Second)
	}

}
