package darksky

import (
	"fmt"
	"github.com/shawntoffel/darksky"
)

// Tinker is me just tinkering
func Tinker() {
	client := darksky.New("api key")
	request := darksky.ForecastRequest{}
	request.Latitude = 40.7128
	request.Longitude = -74.0059
	request.Options = darksky.ForecastRequestOptions{Exclude: "hourly,minutely"}
	forecast, err := client.Forecast(request)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(forecast.Currently.Temperature)
}
