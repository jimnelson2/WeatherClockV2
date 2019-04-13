package forecast

import (
	"fmt"

	"time"

	"github.com/shawntoffel/darksky"
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
		fmt.Println("start darksky at ", time.Now().Format("2006-01-02 15:04:05.000000"))
		client := darksky.New(dsc.DarkskyToken)
		request := darksky.ForecastRequest{}
		request.Latitude = darksky.Measurement(dsc.Latitude)
		request.Longitude = darksky.Measurement(dsc.Longitude)
		forecast, err := client.Forecast(request)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		c <- forecast
		fmt.Println("finish darksky at ", time.Now().Format("2006-01-02 15:04:05.000000"))
		time.Sleep(time.Duration(dsc.PollIntervalSec*1000) * time.Millisecond)
	}
}
