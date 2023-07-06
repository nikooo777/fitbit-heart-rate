package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

type Value struct {
	Date  string  `json:"date"`
	Value float64 `json:"value"`
	Error float64 `json:"error"`
}

type HeartRate struct {
	DateTime string `json:"dateTime"`
	Value    Value  `json:"value"`
}

func parseJSONFile(filepath string) ([]HeartRate, error) {
	fileBytes, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var heartRates []HeartRate
	err = json.Unmarshal(fileBytes, &heartRates)
	if err != nil {
		return nil, err
	}

	return heartRates, nil
}

func main() {
	// Load all .json files in the data directory
	matches, _ := filepath.Glob("data/*.json")

	for _, match := range matches {
		heartRates, err := parseJSONFile(match)
		if err != nil {
			fmt.Printf("Failed to parse file %s: %s\n", match, err)
			continue
		}

		p := plot.New()
		p.Title.Text = "Resting Heart Rate Over Time"
		p.Y.Label.Text = "BPM"

		// Set the minimum Y-axis value
		p.Y.Min = 30
		p.Y.Max = 100

		//filter out bad data:
		filteredHr := make([]HeartRate, 0, len(heartRates))
		sum := 0.0
		for _, hr := range heartRates {
			if hr.Value.Value < 20 {
				continue
			}
			filteredHr = append(filteredHr, hr)
			sum += hr.Value.Value
		}
		average := sum / float64(len(filteredHr))
		fmt.Printf("Average heart rate for the year: %f\n", average)

		// Make a plotter.XYs value and fill it with the data from the file
		pts := make(plotter.XYs, len(filteredHr))
		for i, hr := range filteredHr {
			// Assuming the dateTime in the json is in the format "MM/DD/YY HH:MM:SS"
			t, _ := time.Parse("01/02/06 15:04:05", hr.DateTime)
			pts[i].X = float64(t.Unix())
			pts[i].Y = hr.Value.Value
		}

		// Make a line plotter and set its style
		lpLine, _ := plotter.NewLine(pts)

		// Add the line plotter to the plot
		p.Add(lpLine)

		// Customizing the X-axis.
		p.X.Tick.Marker = plot.TimeTicks{Format: "2006-01-02"}

		// Get the year from filename and create a PNG file for each year
		filename := filepath.Base(match)
		year := strings.Split(filename, "-")[1]
		err = p.Save(16*vg.Inch, 8*vg.Inch, "heart_rate_"+year+".png") // Bigger resolution
		if err != nil {
			logrus.Fatalf("Failed to save plot: %s", err)
		}
	}
}
