package main

import (
	"sync"
	"time"
)

type AggregatedData struct {
	City             string    `json:"city"`
	WindowStart      time.Time `json:"window_start"`
	WindowEnd        time.Time `json:"window_end"`
	AvgTemp          float64   `json:"avg_temperature"`
	MinTemp          float64   `json:"min_temperature"`
	MaxTemp          float64   `json:"max_temperature"`
	AvgHumidity      float64   `json:"avg_humidity"`
	AvgWindSpeed     float64   `json:"avg_wind_speed"`
	SampleCount      int       `json:"sample_count"`
	MostCommonCond   string    `json:"most_common_condition"`
}

type CityWindow struct {
	City      string
	Samples   []WeatherData
	StartTime time.Time
	EndTime   time.Time
}

type WindowAggregator struct {
	windows    map[string]*CityWindow
	windowSize time.Duration
	outputChan chan AggregatedData
	mu         sync.Mutex
}

func NewWindowAggregator(windowSize time.Duration) *WindowAggregator {
	return &WindowAggregator{
		windows:    make(map[string]*CityWindow),
		windowSize: windowSize,
		outputChan: make(chan AggregatedData, 100),
	}
}

func (wa *WindowAggregator) Add(data WeatherData) {
	wa.mu.Lock()
	defer wa.mu.Unlock()

	window, exists := wa.windows[data.City]
	if !exists || window.EndTime.Before(data.Timestamp) {
		windowStart := data.Timestamp.Truncate(wa.windowSize)
		windowEnd := windowStart.Add(wa.windowSize)
		window = &CityWindow{
			City:      data.City,
			Samples:   make([]WeatherData, 0),
			StartTime: windowStart,
			EndTime:   windowEnd,
		}
		wa.windows[data.City] = window
	}

	window.Samples = append(window.Samples, data)

	if data.Timestamp.After(window.EndTime) || data.Timestamp.Equal(window.EndTime) {
		wa.flushWindow(data.City)
	}
}

func (wa *WindowAggregator) flushWindow(city string) {
	window := wa.windows[city]
	if window == nil || len(window.Samples) == 0 {
		return
	}

	var sumTemp, sumHumidity, sumWind float64
	var minTemp, maxTemp float64
	conditionCount := make(map[string]int)

	for i, sample := range window.Samples {
		sumTemp += sample.Temperature
		sumHumidity += float64(sample.Humidity)
		sumWind += sample.WindSpeed
		conditionCount[sample.Condition]++

		if i == 0 {
			minTemp = sample.Temperature
			maxTemp = sample.Temperature
		} else {
			if sample.Temperature < minTemp {
				minTemp = sample.Temperature
			}
			if sample.Temperature > maxTemp {
				maxTemp = sample.Temperature
			}
		}
	}

	count := len(window.Samples)
	mostCommonCond := ""
	maxCount := 0
	for cond, cnt := range conditionCount {
		if cnt > maxCount {
			maxCount = cnt
			mostCommonCond = cond
		}
	}

	aggregated := AggregatedData{
		City:           city,
		WindowStart:    window.StartTime,
		WindowEnd:      window.EndTime,
		AvgTemp:        sumTemp / float64(count),
		MinTemp:        minTemp,
		MaxTemp:        maxTemp,
		AvgHumidity:    sumHumidity / float64(count),
		AvgWindSpeed:   sumWind / float64(count),
		SampleCount:    count,
		MostCommonCond: mostCommonCond,
	}

	wa.outputChan <- aggregated
	delete(wa.windows, city)
}

func (wa *WindowAggregator) GetOutputChan() <-chan AggregatedData {
	return wa.outputChan
}

func (wa *WindowAggregator) Close() {
	wa.mu.Lock()
	defer wa.mu.Unlock()
	for city := range wa.windows {
		wa.flushWindow(city)
	}
	close(wa.outputChan)
}