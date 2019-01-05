package main

import (
	"fmt"

	"github.com/jimnelson2/WeatherClockV2/pkg/darksky"
	"github.com/spf13/viper"

	"os"

	"os/signal"

	"syscall"
)

// RuntimeConfig will get cleaned up once we settle on context
type RuntimeConfig struct {
	DarkskyToken string
	Latitude     float64
	Longitude    float64
}

func main() {

	var rc RuntimeConfig
	{
		viper.SetEnvPrefix("WC")
		viper.AutomaticEnv()

		// repeat something like this for each config var
		rc.DarkskyToken = viper.GetString("DARKSKY_TOKEN")
		if !viper.IsSet("DARKSKY_TOKEN") {
			panic("MISSING DARKSKY_TOKEN\n")
		}

		rc.Latitude = viper.GetFloat64("LATITUDE")
		if !viper.IsSet("LATITUDE") {
			panic("MISSING LATITUDE\n")
		}

		rc.Longitude = viper.GetFloat64("LONGITUDE")
		if !viper.IsSet("LONGITUDE") {
			panic("MISSING LONGITUDE\n")
		}

	}

	// play nice and try to exit when asked by the system
	errs := make(chan error)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()

	// do stuff
	darksky.Tinker(rc.DarkskyToken, rc.Latitude, rc.Longitude)

}
