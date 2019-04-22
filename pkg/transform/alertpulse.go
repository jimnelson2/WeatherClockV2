// Package transform converts radar/weather data into
// structures for display
package transform

import (
	"github.com/jimnelson2/WeatherClockV2/pkg/color"
	"math"
	"time"

	log "github.com/sirupsen/logrus"
)

// ci represents an individual point along the 4 phases of a sine wave from 0 to 2pi
type ci struct {
	Cycle      int
	Multiplier float64
}

// Pulse continually returns pulsed colors slices that fade from black to alert to black to forecast to black, forever
func Pulse(alertColor color.WCColor, forecastColors chan []color.WCColor, pulsedColors chan []color.WCColor) {

	// I've got a timing/logic problem somewhere
	// this delay seems to mitigate it :/
	// w/o this...if we're alerting at startup
	// we'll never display anything
	time.Sleep(time.Duration(5000) * time.Millisecond)

	incoming := make([]color.WCColor, 60)
	scaledColors := make([]color.WCColor, 60)
	var interval int
	var cis []ci

	cis = pulseWave()

	for {
		log.Debug("start alert loop")
		select {
		case acs := <-forecastColors:
			log.Debug("alert got new forecast")
			incoming = acs
		default:
		}

		switch cis[interval].Cycle {
		case 0:
			// multiplier in cycle 0 ranges from 0 up to 1
			// fade is from black to full alert color
			for i := 0; i < 60; i++ {
				scaledColors[i] = color.WCColor{
					R: uint8(cis[interval].Multiplier * float64(alertColor.R)),
					G: uint8(cis[interval].Multiplier * float64(alertColor.G)),
					B: uint8(cis[interval].Multiplier * float64(alertColor.B))}
			}
		case 1:
			// multiplier in cycle 1 ranges from 1 down to 0
			// fade is from full alert color to black
			for i := 0; i < 60; i++ {
				scaledColors[i] = color.WCColor{
					R: uint8(cis[interval].Multiplier * float64(alertColor.R)),
					G: uint8(cis[interval].Multiplier * float64(alertColor.G)),
					B: uint8(cis[interval].Multiplier * float64(alertColor.B))}
			}
		case 2:
			// multiplier in cycle 2 ranges from 0 to 1
			// fade is from black to full forecast color
			for i := 0; i < 60; i++ {
				scaledColors[i] = color.WCColor{
					R: uint8(cis[interval].Multiplier * float64(incoming[i].R)),
					G: uint8(cis[interval].Multiplier * float64(incoming[i].G)),
					B: uint8(cis[interval].Multiplier * float64(incoming[i].B))}
			}
		case 3:
			// multiplier in cycle 3 ranges from 1 to 0
			// fade is from full forecast color to black
			for i := 0; i < 60; i++ {
				scaledColors[i] = color.WCColor{
					R: uint8(cis[interval].Multiplier * float64(incoming[i].R)),
					G: uint8(cis[interval].Multiplier * float64(incoming[i].G)),
					B: uint8(cis[interval].Multiplier * float64(incoming[i].B))}
			}

		}

		pulsedColors <- scaledColors
		time.Sleep(time.Duration(100) * time.Millisecond)

		interval++
		if interval == 90 {
			interval = 0
		}

	}
}

// pre-build the map of cyclic intensity we'll use, since
// it's static let's just do the math once and store it -
// we have plenty of memory
func pulseWave() []ci {

	p := make([]ci, 90)

	p[0] = ci{Cycle: 0, Multiplier: 0}

	var d0, d1, x, multiplier float64
	d0 = 0
	d1 = 0

	for i := 1; i < 90; i++ {
		x = x + math.Pi/45
		multiplier = math.Abs(math.Sin(x))
		d1 = d0 + 1/22.5
		d0 = d1
		p[i] = ci{Cycle: int(math.Floor(d1)), Multiplier: multiplier}
	}

	//for i := 1; i < 90; i++ {
	//	log.Infof("%v", p[i])
	//}

	return p
}
