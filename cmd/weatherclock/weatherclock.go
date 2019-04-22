// TODO package comment
package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/jimnelson2/WeatherClockV2/pkg/color"
	"github.com/jimnelson2/WeatherClockV2/pkg/display"
	"github.com/jimnelson2/WeatherClockV2/pkg/forecast"
	//"github.com/jimnelson2/WeatherClockV2/pkg/io"
	"github.com/jimnelson2/WeatherClockV2/pkg/transform"
	"github.com/shawntoffel/darksky"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {

	// determine from where we'll derive runtime variables
	{
		if os.Getenv("ENVIRONMENT") == "LOCAL" {
			viper.AutomaticEnv()
			viper.SetEnvPrefix("WC")
		} else {
			viper.SetConfigName("wc")
			// viper is weird, doesn't want a file extension, it'll "figure it out"
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

	// configure darksky with all it needs to call the api
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

		// Limiting the call interval to 87 or greater ensure we'll not violate
		// the API agreement. If we paid, we'd get more...but isn't the minutely data
		// refreshed on a 5 minute interval regardless?
		if dsc.PollIntervalSec < 87 {
			log.Fatal("DARKSKY_POLL_SEC less than 87 will exceed api terms for free use of 1000 calls per day")
		}
	}

	// setup routines that will run forever
	darkskyChannel := make(chan darksky.ForecastResponse)
	forecastColorChannel := make(chan []color.WCColor)
	pulsedColorChannel := make(chan []color.WCColor)
	displayChannel := make(chan display.Pixels)
	{
		// receive forecast data from darksky
		go forecast.Run(darkskyChannel, dsc)

		// support the ability to pulse the normal display if weather alerts exist.
		// instantiates a base alert color and the current forecast, continuously
		// receives colors that vary between the two...through a black midpoint
		go transform.Pulse(color.Red, forecastColorChannel, pulsedColorChannel)

		// send colors for display
		go display.Run(displayChannel)

		//keypressChannel := make(chan string)
		//go io.GetKeys(keypressChannel)
	}

	// loop forever, receving/sending data across channels
	alerting := false

	//TODO JIM YOU ARE HERE - WE NEED TO DECIDE IF WE ARE ALERTING
	//BASED ON THE INCOMING FORECAST...NOT BASED ON THE HARDCODING
	//WE HAVE RIGHT NOW
	go func() {
		lastForecastColors := color.NewColors(60)
		lastAlertColors := color.NewColors(60)
		displayColors := color.NewColors(60)
		tr := transform.NewTransformer()
		for {
			log.Debug("start main loop")
			select {
			//case msg0 := <-keypressChannel:
			//	log.Debugf("keypress: %s", msg0)
			//	tr.ApplyTransformDefinition(msg0)
			//	log.Debugf("new transfrom is %v", tr.RainTransform)
			//	cs1 := tr.ForecastToColor(lastForecast)
			//	lastForecastColors = cs1
			//	forecastColorChannel <- cs1
			case darkskyForecast := <-darkskyChannel:
				lastForecastColors = tr.ForecastToColor(darkskyForecast)
				//forecastColorChannel <- lastForecastColors
			case alertColors := <-pulsedColorChannel:
				lastAlertColors = alertColors
			default:
			}

			if alerting {
				displayColors = lastAlertColors
			} else {
				displayColors = lastForecastColors
			}

			if len(displayColors) == 60 {
				displayColors = transform.Dim(displayColors, 0.3)
				m := display.Pixels{Colors: displayColors, PixelCount: 60}
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
