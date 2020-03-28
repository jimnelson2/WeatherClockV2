// TODO package comment
package main

import (
	"math"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jimnelson2/WeatherClockV2/pkg/color"
	"github.com/jimnelson2/WeatherClockV2/pkg/display"
	"github.com/jimnelson2/WeatherClockV2/pkg/forecast"
	"github.com/jimnelson2/WeatherClockV2/pkg/transform"
	"github.com/jimnelson2/tsl2591"
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

		file, err := os.OpenFile("/tmp/wc.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			panic(err)
		}
		defer file.Close()
		log.SetOutput(file)

	}

	// setup light sensor
	log.Info("setting up light sensor")
	tsl, err := tsl2591.NewTSL2591(&tsl2591.Opts{
		Gain:   tsl2591.GainMed,
		Timing: tsl2591.Integrationtime600MS,
	})
	if err != nil {
		log.Error(err.Error)
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
	displayChannel := make(chan display.Pixels)
	{
		// receive forecast data from darksky
		go forecast.Run(darkskyChannel, dsc)

		// send colors for display
		go display.Run(displayChannel)

	}

	go func() {
		forecastColors := make([]color.WCColor, 60)
		displayColors := make([]color.WCColor, 60)
		alertColors := make([]color.WCColor, 60)
		tr := transform.NewTransformer()
		var alertOn = false
		var alertToggle = false
		for {
			select {
			case darkskyForecast := <-darkskyChannel:
				forecastColors = tr.ForecastToColor(darkskyForecast)
				alertOn, alertColors = tr.ForecastToAlert(darkskyForecast)
			default:
			}

			// get light sensor
			lux, _ := tsl.Lux()

			// Just display the forecast colors if there's no alerting
			// However, if there is alerting we want to toggle between
			// the forecast colors and alert colors
			if !alertOn {
				log.Info("not alerting - display forecast colors")
				displayColors = forecastColors
			} else {
				if alertOn && alertToggle {
					log.Info("alerting - display alert colors")
					alertToggle = false
					displayColors = alertColors
				} else {
					log.Info("alerting - display forecast colors")
					alertToggle = true
					displayColors = forecastColors
				}
			}

			// dim lights relative to brightness
			if math.IsNaN(lux) {
				log.Info("got NaN from lux sensor, defaulting lux to 5")
				lux = 5.0 // so...icky. defult lux value if we aren't getting one from our sensor
			}

			// overlay current time on top of colors
			displayColors = tr.OverlayColors(displayColors, tr.ClockFace())
			displayChannel <- display.Pixels{Colors: transform.Dim(displayColors, transform.LuxToDim(lux)), PixelCount: 60}

			// sleep 5 seconds
			time.Sleep(time.Duration(5000) * time.Millisecond)
		}
	}()

	// Block until a signal is received. Basically, run forever
	// until the OS tells us to stop. Thanks go-kit for the code.
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	s := <-c
	log.Infof("Got signal %v", s)

}
