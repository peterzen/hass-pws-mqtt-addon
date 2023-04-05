package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/PuerkitoBio/goquery"
)

type Variable struct {
	VarName        string `json:"var_name"`
	VarDescription string `json:"var_description"`
	Value          string `json:"value"`
}

func main() {

	// Open the HTML file
	file, err := os.Open("LiveData.html")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// // Read the HTML file
	// resp, err := http.Get("http://example.com/table.html")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer resp.Body.Close()

	// htmlData, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// Parse the HTML file with goquery
	// doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(htmlData)))

	doc, err := goquery.NewDocumentFromReader(file)
	if err != nil {
		log.Fatal(err)
	}

	// Extract the variables from each row
	model := []Variable{
		{VarName: "skip", VarDescription: "", Value: ""},
		{VarName: "skip", VarDescription: "", Value: ""},
		{VarName: "skip", VarDescription: "", Value: ""},
		{VarName: "skip", VarDescription: "", Value: ""},
		{VarName: "skip", VarDescription: "", Value: ""},
		{VarName: "skip", VarDescription: "", Value: ""},
		{VarName: "skip", VarDescription: "", Value: ""},
		{VarName: "skip", VarDescription: "", Value: ""},
		{VarName: "receiver_time", VarDescription: "Receiver Time", Value: ""},
		{VarName: "skip", VarDescription: "", Value: ""},
		{VarName: "skip", VarDescription: "", Value: ""},
		{VarName: "skip", VarDescription: "", Value: ""},
		{VarName: "temp_indoor", VarDescription: "Indoor Temperature", Value: ""},
		{VarName: "humidity_indoor", VarDescription: "Indoor Humidity", Value: ""},
		{VarName: "", VarDescription: "Absolute Pressure", Value: ""},
		{VarName: "", VarDescription: "Relative Pressure", Value: ""},
		{VarName: "", VarDescription: "Outdoor Temperature", Value: ""},
		{VarName: "", VarDescription: "Outdoor Humidity", Value: ""},
		{VarName: "", VarDescription: "Wind Direction", Value: ""},
		{VarName: "", VarDescription: "Wind Speed", Value: ""},
		{VarName: "", VarDescription: "Wind Gust", Value: ""},
		{VarName: "", VarDescription: "Solar Radiation", Value: ""},
		{VarName: "", VarDescription: "UV", Value: ""},
		{VarName: "", VarDescription: "UVI", Value: ""},
		{VarName: "", VarDescription: "Hourly Rain Rate", Value: ""},
		{VarName: "", VarDescription: "Daily Rain", Value: ""},
		{VarName: "", VarDescription: "Weekly Rain", Value: ""},
		{VarName: "", VarDescription: "Monthly Rain", Value: ""},
		{VarName: "", VarDescription: "Yearly Rain", Value: ""},
	}

	var vars []Variable

	doc.Find("table").Each(func(i int, s *goquery.Selection) {

		// parse the table rows and extract the data
		s.Find("tr").Each(func(i int, s *goquery.Selection) {

			if i >= len(model) {
				return
			}

			m := model[i]
			if m.VarName == "skip" {
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
		log.Fatal(err)
	}

	fmt.Println(string(jsonData))
}
