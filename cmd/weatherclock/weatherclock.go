package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/jimnelson2/WeatherClockV2/pkg/color"
	"github.com/jimnelson2/WeatherClockV2/pkg/display"
	"github.com/jimnelson2/WeatherClockV2/pkg/forecast"
	"github.com/jimnelson2/WeatherClockV2/pkg/io"
	"github.com/jimnelson2/WeatherClockV2/pkg/transform"
	"github.com/shawntoffel/darksky"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {

	// I like big mains. Cannot lie, etc. The idea that main sets up,
	// instantiates, constructs, populates...everything. And then
	// the things just go off and ... do their things.

	// setup runtime variable source
	{
		if os.Getenv("ENVIRONMENT") == "LOCAL" {
			viper.AutomaticEnv()
			viper.SetEnvPrefix("WC")
		} else {
			viper.SetConfigName("wc") // viper is weird, doesn't want a file extension, it'll "figure it out"
			viper.AddConfigPath("/etc/default/")
			viper.AddConfigPath(".")
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

	keypressChannel := make(chan string)
	go io.GetKeys(keypressChannel)

	// loop forever, passing data between channels as it arrives
	go func() {
		var finalColors, lastForecastColors, lastAlertColors []color.WCColor
		var lastForecast darksky.ForecastResponse
		tr := transform.NewTransformer()
		for {
			select {
			case msg0 := <-keypressChannel:
				log.Debugf("keypress: %s", msg0)
				applyTransformDefinition(tr, msg0)
				log.Debugf("new transfrom is %v", tr.RainTransform)
				cs1 := tr.ForecastToColor(lastForecast)
				lastForecastColors = cs1
				finalColors = transform.OverlayColors(lastForecastColors, lastAlertColors)
				finalColors = transform.Dim(finalColors, 0.3)
				m := display.Pixels{Colors: finalColors, PixelCount: 60}
				displayChannel <- m
			case msg1 := <-darkskyChannel:
				lastForecast = msg1
				cs1 := tr.ForecastToColor(lastForecast)
				lastForecastColors = cs1
				finalColors = transform.OverlayColors(lastForecastColors, lastAlertColors)
				finalColors = transform.Dim(finalColors, 0.3)
				m := display.Pixels{Colors: finalColors, PixelCount: 60}
				displayChannel <- m
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
	// until the OS tells us to stop. Thanks go-kit for the code.
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	s := <-c
	log.Infof("Got signal %v", s)

}

func applyTransformDefinition(tr *transform.Transformer, change string) {

	// This feels yucky...not sure how to make it better yet...

	// Here's the idea. We have a default transform that defines the various
	// intensities at which color changes. I want to provide the ability to
	// modify the intensity levels at which a transition occurs. Imagine
	// that we have slider controls on a ruler - one slider to demark a
	// transition between two colors. Let the transitions move around, but
	// don't let sliders pass each other. That's what the below code does.
	// lower-case letters move a slider to a lower value, upper-case moves
	// the slider higher. We go alphabetically. For example:
	// "a" moves the transition between no color and the first color for rain 0.005 lower
	// "A" moves the transition between no color and the first color for rain 0.005 higher, but not higher than the next transition
	// and so on. "b/B" is the second transition, etc. up through "e/E"
	// We follow the same pattern for the other precipitation types
	// The purpose for this is to help tune the transitions in an interactive way...and when we
	// reach a set of transitions that we like...we'll take them and hard-code them as the defaults
	switch change {
	case "a":
		if tr.RainTransform[1].Intensity > tr.RainTransform[0].Intensity {
			tr.RainTransform[1].Intensity = tr.RainTransform[1].Intensity - 0.005
		}
	case "A":
		if tr.RainTransform[1].Intensity < tr.RainTransform[2].Intensity {
			tr.RainTransform[1].Intensity = tr.RainTransform[1].Intensity + 0.005
		}
	case "b":
		if tr.RainTransform[2].Intensity > tr.RainTransform[1].Intensity {
			tr.RainTransform[2].Intensity = tr.RainTransform[2].Intensity - 0.005
		}
	case "B":
		if tr.RainTransform[2].Intensity < tr.RainTransform[3].Intensity {
			tr.RainTransform[2].Intensity = tr.RainTransform[2].Intensity + 0.005
		}
	case "c":
		if tr.RainTransform[3].Intensity > tr.RainTransform[2].Intensity {
			tr.RainTransform[3].Intensity = tr.RainTransform[3].Intensity - 0.005
		}
	case "C":
		if tr.RainTransform[3].Intensity < tr.RainTransform[4].Intensity {
			tr.RainTransform[3].Intensity = tr.RainTransform[3].Intensity + 0.005
		}
	case "d":
		if tr.RainTransform[4].Intensity > tr.RainTransform[3].Intensity {
			tr.RainTransform[4].Intensity = tr.RainTransform[4].Intensity - 0.005
		}
	case "D":
		if tr.RainTransform[4].Intensity < tr.RainTransform[5].Intensity {
			tr.RainTransform[4].Intensity = tr.RainTransform[4].Intensity + 0.005
		}
	case "e":
		if tr.RainTransform[5].Intensity > tr.RainTransform[4].Intensity {
			tr.RainTransform[5].Intensity = tr.RainTransform[5].Intensity - 0.005
		}
	case "E":
		tr.RainTransform[5].Intensity = tr.RainTransform[4].Intensity + 0.005
	default:
	}

}
