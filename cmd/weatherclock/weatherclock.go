package main

import (
	"fmt"

	//"github.com/carlescere/scheduler"
	"github.com/jimnelson2/WeatherClockV2/pkg/forecast"
	"github.com/shawntoffel/darksky"
	"github.com/spf13/viper"
)

func main() {

	var dsc forecast.Job
	//var darkskyPollSec int
	{
		viper.SetEnvPrefix("WC")
		viper.AutomaticEnv()

		dsc.DarkskyToken = viper.GetString("DARKSKY_TOKEN")
		if !viper.IsSet("DARKSKY_TOKEN") {
			panic("MISSING DARKSKY_TOKEN\n")
		}

		dsc.Latitude = viper.GetFloat64("LATITUDE")
		if !viper.IsSet("LATITUDE") {
			panic("MISSING LATITUDE\n")
		}

		dsc.Longitude = viper.GetFloat64("LONGITUDE")
		if !viper.IsSet("LONGITUDE") {
			panic("MISSING LONGITUDE\n")
		}

		//darkskyPollSec := viper.GetInt("DARKSKY_POLL_SEC")
		//if !viper.IsSet("DARKSKY_POLL_SEC") {
		//	panic("MISSING DARKSKY_POLL_SEC\n")
		//}
		//if darkskyPollSec < 87 { //TODO make a class constant
		//	panic("DARKSKY_POLL_SEC less than 87 will exceed api terms for free use of 1000 calls per day\n")
		//}
	}

	// call darksky on an interval, forever
	//scheduler.Every(darkskyPollSec).Seconds().Run(dsc.Run)
	darkskyChannel := make(chan darksky.ForecastResponse)
	go dsc.Run(darkskyChannel)

	f := <-darkskyChannel
	cs := colors(f)
	for idx, c := range cs {
		fmt.Printf("%d, %+v\n", idx, c)
	}

}

// Color holds a pixel color
type Color struct {
	R, G, B uint8
}

// expecting that intensity-to-color mapping will vary by precip type
// intensity is inches of liquid water per hour
// cutoffs in the maps below are arbitrary at this point, I'd prefer to shift them
// to the lower intensities just so things look more interesting.
// we'll need to run thru historical data and tune things
// needs some test written to ensure we don't have misses/gaps
// will refactor this out to a separate file
func rain(intensity float64) Color {
	switch {
	case intensity < 0.1:
		{
			// green
			return Color{R: 0, G: 255, B: 0}
		}
	case intensity < 0.3:
		{
			// yellow
			return Color{R: 255, G: 255, B: 0}
		}
	case intensity < 0.5:
		{
			// organge
			return Color{R: 255, G: 128, B: 0}
		}
	case intensity < 0.7:
		{
			// red
			return Color{R: 255, G: 0, B: 0}
		}
	case intensity >= 0.7:
		{
			// purple
			return Color{R: 127, G: 0, B: 255}
		}
	}
	// black
	return Color{R: 0, G: 0, B: 0}
}

func sleet(intensity float64) Color {
	switch {
	case intensity < 0.2:
		{
			// pink
			return Color{R: 255, G: 0, B: 255}
		}
	case intensity >= 0.2:
		{
			// purple
			return Color{R: 127, G: 0, B: 255}
		}
	}
	// black
	return Color{R: 0, G: 0, B: 0}
}

func snow(intensity float64) Color {
	switch {
	case intensity < 0.3:
		{
			// light blue
			return Color{R: 0, G: 255, B: 255}
		}
	case intensity < 0.6:
		{
			// dark blue
			return Color{R: 0, G: 0, B: 255}
		}
	case intensity >= 0.6:
		{
			// purple
			return Color{R: 127, G: 0, B: 255}
		}
	}
	// black
	return Color{R: 0, G: 0, B: 0}

}

// colors maps the forecast to colors
func colors(f darksky.ForecastResponse) []Color {

	// tbd seem to be 61 items in Minutely, but...weird...need to understand better.
	// I know I'll only have 60 LEDs to light, so...
	colors := make([]Color, len(f.Minutely.Data))
	// happy path first, but be aware...we cannot trust any data element(s) will be present at
	// any point in time or the forecast
	// we also might want to consider including probability in here?
	fmt.Printf("data points: %d\n", len(f.Minutely.Data))
	for idx, m := range f.Minutely.Data {
		switch m.PrecipType {
		case "rain":
			{
				colors[idx] = rain(float64(m.PrecipIntensity))
			}
		case "sleet":
			{
				colors[idx] = sleet(float64(m.PrecipIntensity))
			}
		case "snow":
			{
				colors[idx] = snow(float64(m.PrecipIntensity))
			}
		default:
			{
				// default color is black
				colors[idx] = Color{R: 0, G: 0, B: 0}
			}
		}
	}
	return colors
}
