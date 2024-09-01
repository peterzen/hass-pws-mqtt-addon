package main

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func Test_parseHtml(t *testing.T) {
	type args struct {
		doc *goquery.Document
	}

	file, err := os.Open("./testing/LiveData.html")
	if err != nil {
		t.Errorf("can't open test file: %s\n", err)
		return
	}
	defer file.Close()

	htmlData, err := io.ReadAll(file)
	if err != nil {
		t.Errorf("can't read test file: %s\n", err)
		return
	}

	// Parse the HTML file with goquery
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(htmlData)))

	if err != nil {
		t.Errorf("Failed to process HTML: %s\n", err)
		return
	}

	wd := parseHtml(doc)

	if wd.HumidityIndoor != 48 {
		t.Errorf("HumidityIndoor is not 0")
	}
	if wd.TemperatureIndoor != 27.6 {
		t.Errorf("TemperatureIndoor is not 0")
	}
	if wd.IndoorSensorId != "0xad" {
		t.Errorf("IndoorSensorId is not 0xad")
	}
	if wd.OutdoorSensorId != "0x0e" {
		t.Errorf("OutdoorSensorId is not 0x0e")
	}
	if wd.IndoorSensorBattery != "Normal" {
		t.Errorf("IndoorSensorBattery is not Normal")
	}
}
