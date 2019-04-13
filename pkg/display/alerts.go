package display

import (
	"math"
	"time"
)

func Pulse(c chan Color, base Color) {

	// Given a base color representing the "MAX" amount of color,
	// cotinually emit colors that vary between the base color
	// and black following a sin wave pattern so the color
	// appears to wash back and forth between base and
	// black smoothly

	cycle := 4 * math.Pi
	var x, m float64
	var scaledColor Color
	x = 0
	for {
		x = math.Mod((x + math.Pi/60), cycle)
		m = (math.Sin(x) + 1) / 2
		// m is cycling between 0 and 1 in a sinusoidal wave
		scaledColor = Color{R: uint8(float64(base.R) * m),
			G: uint8(float64(base.G) * m),
			B: uint8(float64(base.B) * m)}

		c <- scaledColor
		time.Sleep(time.Duration(100) * time.Millisecond)

	}
}
