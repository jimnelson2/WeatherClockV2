package forecast

import (
	"time"

	"github.com/shawntoffel/darksky"
	log "github.com/sirupsen/logrus"
)

// Job defines the details for a call to the darksky api
type Job struct {
	DarkskyToken    string
	Latitude        float64
	Longitude       float64
	PollIntervalSec int
}

// Run implements the gron.Job interface
func (dsc Job) Run(c chan darksky.ForecastResponse) {
	for {
		log.Debug("calling darksky.io ")
		client := darksky.New(dsc.DarkskyToken)
		request := darksky.ForecastRequest{}
		request.Latitude = darksky.Measurement(dsc.Latitude)
		request.Longitude = darksky.Measurement(dsc.Longitude)
		forecast, err := client.Forecast(request)
		if err != nil {
			log.Error("failed to get response from darksky.io")
			log.Error(err)
		} else {
			log.Debug("got response from darksky.io")
			c <- forecast
		}
		time.Sleep(time.Duration(dsc.PollIntervalSec*1000) * time.Millisecond)
	}
}
