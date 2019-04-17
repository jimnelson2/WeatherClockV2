// Package transform converts radar/weather data into
// structures for display
package transform

import (
	"github.com/jimnelson2/WeatherClockV2/pkg/color"
	"math"
	"time"
)

// Pulse feeds the provided channel with a continual cycle
// of color, fading between the provided base color and
// black on a sinusoidal pattern, 100ms intervals
func Pulse(c chan color.WCColor, base color.WCColor) {

	cycle := 4 * math.Pi
	var x, m float64
	var scaledColor color.WCColor
	x = 0
	for {
		// TODO: Understand this better. I think I'd prefer it
		// to be more like "full cycle every N seconds"
		x = math.Mod((x + math.Pi/60), cycle)
		m = (math.Sin(x) + 1) / 2
		// m is cycling between 0 and 1 in a sinusoidal wave
		scaledColor = color.WCColor{R: uint8(float64(base.R) * m),
			G: uint8(float64(base.G) * m),
			B: uint8(float64(base.B) * m)}

		c <- scaledColor
		time.Sleep(time.Duration(100) * time.Millisecond)

	}
}
