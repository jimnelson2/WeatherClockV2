package darksky

import (
	"fmt"

	"github.com/shawntoffel/darksky"
)

// Tinker is me just tinkering
func Tinker(token string, latitude float64, longitude float64) {
	client := darksky.New(token)
	request := darksky.ForecastRequest{}
	request.Latitude = darksky.Measurement(latitude)
	request.Longitude = darksky.Measurement(longitude)
	forecast, err := client.Forecast(request)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(forecast.Minutely.Data[0].PrecipType)
}
