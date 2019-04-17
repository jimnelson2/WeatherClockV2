package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/jimnelson2/WeatherClockV2/pkg/color"
	"github.com/jimnelson2/WeatherClockV2/pkg/display"
	"github.com/jimnelson2/WeatherClockV2/pkg/forecast"
	"github.com/jimnelson2/WeatherClockV2/pkg/transform"
	"github.com/shawntoffel/darksky"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {

	// I like big mains. Cannot lie, etc. The idea that main sets up,
	// instantiates, constructs, popluates...everything. And then
	// the things just go off and ... do their things.

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

		if dsc.PollIntervalSec < 87 { //TODO make a constant? If we paid, we'd get more...but isn't the minutely data refreshed on a 5 minute interval regardless?
			log.Fatal("DARKSKY_POLL_SEC less than 87 will exceed api terms for free use of 1000 calls per day")
		}
	}

	darkskyChannel := make(chan darksky.ForecastResponse)
	go forecast.Run(darkskyChannel, dsc)

	displayChannel := make(chan display.Pixels)
	go display.Run(displayChannel)

	pulseChannel := make(chan color.WCColor)
	go transform.Pulse(pulseChannel, color.Red)

	// loop forever, passing data between channels as it arrives
	go func() {
		var finalColors, lastForecastColors, lastAlertColors []color.WCColor
		for {
			select {
			case msg1 := <-darkskyChannel:
				cs1 := transform.ForecastToColor(msg1)
				log.Debug(cs1)
				lastForecastColors = cs1
			case msg2 := <-pulseChannel:
				cs2 := transform.AllSameColors(msg2)
				lastAlertColors = cs2
			}
			// display what we have...if we have it. Doesn't smell right, tbh
			if len(lastForecastColors) > 0 && len(lastAlertColors) > 0 {
				finalColors = transform.OverlayColors(lastForecastColors, lastAlertColors)
				finalColors = transform.Dim(finalColors, 0.3)
				m := display.Pixels{Colors: finalColors, PixelCount: 60}
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
