package main

import (
	"fmt"

	"github.com/jimnelson2/WeatherClockV2/pkg/darksky"
	"github.com/spf13/viper"

	"os"

	"os/signal"

	"syscall"
)

type runtimeConfig struct {
	DarkskyToken string
}

func main() {

	var rc runtimeConfig
	{
		viper.SetEnvPrefix("WC")
		viper.AutomaticEnv()

		// repeat something like this for each config var
		rc.DarkskyToken = viper.GetString("DARKSKY_TOKEN")
		if !viper.IsSet("DARKSKY_TOKEN") {
			fmt.Printf("MISSING DARKSKY_TOKEN\n")
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
	fmt.Printf("%s\n", rc.DarkskyToken)
	darksky.Tinker()
}
