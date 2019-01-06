package main

import (
	"github.com/jimnelson2/WeatherClockV2/pkg/darksky"
	"github.com/spf13/viper"
)

func main() {

	var dsc darksky.Job
	var darkskyPollSec int
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

		darkskyPollSec = viper.GetInt("DARKSKY_POLL_SEC")
		if !viper.IsSet("DARKSKY_POLL_SEC") {
			panic("MISSING DARKSKY_POLL_SEC\n")
		}
		if darkskyPollSec < 87 { //TODO make a class constant
			panic("DARKSKY_POLL_SEC less than 87 will exceed api terms for free use of 1000 calls per day\n")
		}
	}

	dsc.Run()
}
