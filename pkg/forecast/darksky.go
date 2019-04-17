// Package forecast wraps and periodically calls
// the darksky.io API
package forecast

import (
	"time"

	"github.com/shawntoffel/darksky"
	log "github.com/sirupsen/logrus"
)

// DarkskyConfig defines the details for calls to the darksky api
type DarkskyConfig struct {
	DarkskyToken    string
	Latitude        float64
	Longitude       float64
	PollIntervalSec int
}

// Run calls the darksky.io service forever
func Run(c chan darksky.ForecastResponse, dsc DarkskyConfig) {
	for {
		log.Debug("Calling darksky.io")
		client := darksky.New(dsc.DarkskyToken)
		request := darksky.ForecastRequest{}
		request.Latitude = darksky.Measurement(dsc.Latitude)
		request.Longitude = darksky.Measurement(dsc.Longitude)
		forecast, err := client.Forecast(request)
		if err != nil {
			log.Error("Failed to get response from darksky.io")
			log.Error(err)
		} else {
			log.Debug("Got response from darksky.io")
			c <- forecast
		}
		time.Sleep(time.Duration(dsc.PollIntervalSec*1000) * time.Millisecond)
	}
}
