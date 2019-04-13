package main

import (
	"fmt"
	"os"

	"github.com/jimnelson2/WeatherClockV2/pkg/display"
	"github.com/jimnelson2/WeatherClockV2/pkg/forecast"
	"github.com/shawntoffel/darksky"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {

	//	var log = logrus.New()
	{
		log.SetOutput(os.Stdout)
		log.SetLevel(log.DebugLevel)
	}

	log.Info("Getting runtime variables")
	var dsc forecast.Job
	{
		viper.SetEnvPrefix("WC")
		viper.AutomaticEnv()

		dsc.DarkskyToken = viper.GetString("DARKSKY_TOKEN")
		if !viper.IsSet("DARKSKY_TOKEN") {
			log.Fatal("MISSING DARKSKY_TOKEN")
		}

		dsc.Latitude = viper.GetFloat64("LATITUDE")
		if !viper.IsSet("LATITUDE") {
			log.Fatal("MISSING LATITUDE")
		}

		dsc.Longitude = viper.GetFloat64("LONGITUDE")
		if !viper.IsSet("LONGITUDE") {
			log.Fatal("MISSING LONGITUDE")
		}

		dsc.PollIntervalSec = viper.GetInt("DARKSKY_POLL_SEC")
		if !viper.IsSet("DARKSKY_POLL_SEC") {
			log.Fatal("MISSING DARKSKY_POLL_SEC")
		}

		if dsc.PollIntervalSec < 87 { //TODO make a class constant
			log.Fatal("DARKSKY_POLL_SEC less than 87 will exceed api terms for free use of 1000 calls per day")
		}
	}
	log.Info("Got runtime variables")

	// call darksky on an interval, forever
	darkskyChannel := make(chan darksky.ForecastResponse)
	go dsc.Run(darkskyChannel)

	displayChannel := make(chan display.Minutes)
	go display.Run(displayChannel)

	//pulseChannel := make(chan display.Color)
	//var pulseColor display.Color
	//go display.Pulse(pulseChannel, pulseColor.Red())

	go func() {
		for {
			select {
			case msg1 := <-darkskyChannel:
				cs := colors(msg1)
				cs = testColors()
				log.Debug(cs)
				m := display.Minutes{Colors: cs, PixelCount: 60}
				displayChannel <- m
				//case msg2 := <-pulseChannel:
				//	cs := allSameColors(msg2)
				//	log.Debug(cs)
				//	//m := display.Minutes{Colors: cs, PixelCount: 60}
				//	//displayChannel <- m
			}
		}
	}()

	var input string
	fmt.Scanln(&input)
}

func allSameColors(c display.Color) []display.Color {
	cs := make([]display.Color, 60)
	for i := 0; i < 60; i++ {
		cs[i] = c
	}
	return cs
}

func testColors() []display.Color {
	cs := make([]display.Color, 60)
	var c display.Color
	for i := 0; i < 60; i++ {
		cs[i] = display.Color{R: 127, G: 127, B: 127}
	}

	cs[0] = c.Green()
	cs[1] = c.Yellow()
	cs[2] = c.Orange()
	cs[3] = c.Red()
	cs[4] = c.Purple()
	return cs

}

func pulse(c display.Color) {
	// not entirely sure...
	// I'm imagining the pixel ring will pulse/throb
	// from the provided color to dark.
	// not sure how i want that to work with displayed precip
	// literally can't imagine it...prob need hardware
	// to see options in action
}

// like...way beyond first try...
// could imagine these cutoffs being changeable via
// front-end sliders. yeah we have no front end yet.
// but if we did how cool would that be

// expecting that intensity-to-color mapping will vary by precip type
// intensity is inches of liquid water per hour
// cutoffs in the maps below are arbitrary at this point, I'd prefer to shift them
// to the lower intensities just so things look more interesting.
// we'll need to run thru historical data and tune things
// needs some test written to ensure we don't have misses/gaps
// will refactor this out to a separate file
func rain(intensity float64) display.Color {

	var c display.Color

	switch {
	case intensity < 0.01:
		{
			return c.Black()
		}
	case intensity < 0.1:
		{
			return c.Green()
		}
	case intensity < 0.3:
		{
			return c.Yellow()
		}
	case intensity < 0.5:
		{
			return c.Orange()
		}
	case intensity < 0.7:
		{
			return c.Red()
		}
	case intensity >= 0.7:
		{
			return c.Purple()
		}
	}
	return c.Black()
}

func sleet(intensity float64) display.Color {

	var c display.Color

	switch {
	case intensity < 0.01:
		{
			return c.Black()
		}
	case intensity < 0.2:
		{
			return c.Pink()
		}
	case intensity >= 0.2:
		{
			return c.Purple()
		}
	}
	return c.Black()
}

func snow(intensity float64) display.Color {

	var c display.Color

	switch {
	case intensity < 0.01:
		{
			return c.Black()
		}
	case intensity < 0.3:
		{
			return c.LightBlue()
		}
	case intensity < 0.6:
		{
			return c.DarkBlue()
		}
	case intensity >= 0.6:
		{
			return c.Purple()
		}
	}
	return c.Black()

}

// colors maps the forecast to colors
func colors(f darksky.ForecastResponse) []display.Color {

	// tbd seem to be 61 items in Minutely, but...weird...need to understand better.
	// I know I'll only have 60 LEDs to light, so...
	colors := make([]display.Color, len(f.Minutely.Data))
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
				var c display.Color
				colors[idx] = c.Black()
			}
		}
	}
	return colors
}
