package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/jimnelson2/WeatherClockV2/pkg/color"
	"github.com/jimnelson2/WeatherClockV2/pkg/display"
	"github.com/jimnelson2/WeatherClockV2/pkg/forecast"
	"github.com/shawntoffel/darksky"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {

	// setup runtime variable source
	{
		if os.Getenv("ENVIRONMENT") == "LOCAL" {
			viper.AutomaticEnv()
			viper.SetEnvPrefix("WC")
		} else {
			viper.SetConfigName("wc") // viper is weird, doesn't want a file extension, it'll "figure it out"
			viper.AddConfigPath("/etc/default/")
			err := viper.ReadInConfig()
			if err != nil {
				log.Fatalf("Fatal error config file: %s \n", err)
			}
		}
	}

	// setup logging
	{
		log.SetOutput(os.Stdout)
		logLevel := viper.GetString("LOG_LEVEL")
		switch logLevel {
		case "TRACE":
			log.SetLevel(log.TraceLevel)
		case "DEBUG":
			log.SetLevel(log.DebugLevel)
		case "INFO":
			log.SetLevel(log.InfoLevel)
		case "ERROR":
			log.SetLevel(log.ErrorLevel)
		case "FATAL":
			log.SetLevel(log.FatalLevel)
		default:
			log.SetLevel(log.DebugLevel)
		}
	}

	// gather darksky configuration
	var dsc forecast.DarkskyConfig
	{
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

	darkskyChannel := make(chan darksky.ForecastResponse)
	go forecast.Run(darkskyChannel, dsc)

	displayChannel := make(chan display.Minutes)
	go display.Run(displayChannel)

	pulseChannel := make(chan color.WCColor)
	go display.Pulse(pulseChannel, color.Red)

	// loop forever, passing data between channels as it arrives
	go func() {
		var finalColors, lastForecastColors, lastAlertColors []color.WCColor
		for {
			select {
			case msg1 := <-darkskyChannel:
				cs1 := colors(msg1)
				//cs = testColors()
				log.Debug(cs1)
				traceMapping(msg1, cs1)
				lastForecastColors = cs1
				//m := display.Minutes{Colors: cs, PixelCount: 60}
				//displayChannel <- m
			case msg2 := <-pulseChannel:
				cs2 := allSameColors(msg2)
				lastAlertColors = cs2
				//m := display.Minutes{Colors: cs, PixelCount: 60}
				//displayChannel <- m
			}
			if len(lastForecastColors) > 0 && len(lastAlertColors) > 0 {
				finalColors = overlayColors(lastForecastColors, lastAlertColors)
				finalColors = dim(finalColors, 0.3)
				m := display.Minutes{Colors: finalColors, PixelCount: 60}
				displayChannel <- m
			}
		}
	}()

	// Block until a signal is received. Basically, run forever
	// until the OS tells us to step
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	s := <-c
	log.Infof("Got signal %v", s)

}

func overlayColors(fc []color.WCColor, ac []color.WCColor) []color.WCColor {

	// not really sure what I want to see. For now...we're just gonna add 'em up
	cs := make([]color.WCColor, 60)
	var r, g, b uint
	for i := 0; i < 60; i++ {

		r = uint(fc[i].R + ac[i].R)
		g = uint(fc[i].G + ac[i].G)
		b = uint(fc[i].B + ac[i].B)

		if r > 255 {
			r = 255
		}
		if g > 255 {
			g = 255
		}
		if b > 255 {
			b = 255
		}

		cs[i].R = uint8(r)
		cs[i].G = uint8(g)
		cs[i].B = uint8(b)
	}

	//return cs  ignore what we're doing here for now
	return fc
}

func allSameColors(c color.WCColor) []color.WCColor {
	cs := make([]color.WCColor, 60)
	for i := 0; i < 60; i++ {
		cs[i] = c
	}
	return cs
}

func testColors() []color.WCColor {
	cs := make([]color.WCColor, 60)
	for i := 0; i < 60; i++ {
		cs[i] = color.Black
	}

	cs[0] = color.Green
	cs[1] = color.Yellow
	cs[2] = color.Orange
	cs[3] = color.Red
	cs[4] = color.Purple
	return cs

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
func rain(intensity float64) color.WCColor {

	// These ranges here could be defined more succinctly,
	// but I find specifying both ends of a given range
	// explicity to be more readable
	switch {
	case intensity < 0.05:
		{
			return color.Black
		}
	case intensity >= 0.05 && intensity < 0.50:
		{
			return color.Green
		}
	case intensity >= 0.50 && intensity < 1.00:
		{
			return color.Yellow
		}
	case intensity >= 1.00 && intensity < 2.00:
		{
			return color.Orange
		}
	case intensity >= 2.00 && intensity < 4.50:
		{
			return color.Red
		}
	case intensity > 4.50:
		{
			return color.Purple
		}
	}
	return color.Black
}

func sleet(intensity float64) color.WCColor {

	// These ranges here could be defined more succinctly,
	// but I find specifying both ends of a given range
	// explicity to be more readable
	switch {
	case intensity < 0.05:
		{
			return color.Black
		}
	case intensity >= 0.05 && intensity < 0.20:
		{
			return color.Pink
		}
	case intensity >= 0.20:
		{
			return color.Purple
		}
	}
	return color.Black
}

func snow(intensity float64) color.WCColor {

	// These ranges here could be defined more succinctly,
	// but I find specifying both ends of a given range
	// explicity to be more readable
	switch {
	case intensity < 0.05:
		{
			return color.Black
		}
	case intensity >= 0.05 && intensity < 0.30:
		{
			return color.LightBlue
		}
	case intensity >= 0.30 && intensity < 0.60:
		{
			return color.DarkBlue
		}
	case intensity >= 0.60:
		{
			return color.Purple
		}
	}
	return color.Black

}

func traceMapping(f darksky.ForecastResponse, cs []color.WCColor) {

	for i := 0; i < 60; i++ {
		log.Tracef("%s i:%f p:%f- %d %d %d", f.Minutely.Data[i].PrecipType, f.Minutely.Data[i].PrecipIntensity, f.Minutely.Data[i].PrecipProbability,
			cs[i].R, cs[i].G, cs[i].B)
	}
}

func dim(c []color.WCColor, dimVal float32) []color.WCColor {

	colors := make([]color.WCColor, 60)

	for i := 0; i < 60; i++ {

		r := uint8(float32(c[i].R) * dimVal)
		g := uint8(float32(c[i].G) * dimVal)
		b := uint8(float32(c[i].B) * dimVal)

		colors[i] = color.WCColor{R: r, G: g, B: b}
	}
	return colors
}

// colors maps the forecast to colors
func colors(f darksky.ForecastResponse) []color.WCColor {

	// tbd seem to be 61 items in Minutely, but...weird...need to understand better.
	// I know I'll only have 60 LEDs to light, so...
	colors := make([]color.WCColor, len(f.Minutely.Data))
	// happy path first, but be aware...we cannot trust any data element(s) will be present at
	// any point in time or the forecast
	// we also might want to consider including probability in here?
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
				colors[idx] = color.Black
			}
		}
	}
	return colors
}
