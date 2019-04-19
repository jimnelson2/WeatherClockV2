// Package transform converts radar/weather data into
// structures for display
package transform

import (
	"github.com/jimnelson2/WeatherClockV2/pkg/color"
	"github.com/shawntoffel/darksky"
	log "github.com/sirupsen/logrus"
)

// Transform define intensity boundary to color
type Transform struct {
	Intensity float64
	Color     color.WCColor
}

// Transformer holds specific transformations and provides methods to use them
type Transformer struct {
	RainTransform  []Transform
	SleetTransform []Transform
	SnowTransform  []Transform
}

// NewTransformer returns a pointer to a transformer with default transforms
func NewTransformer() *Transformer {
	t := new(Transformer)
	t.RainTransform = []Transform{
		Transform{0.00, color.Black},
		Transform{0.01, color.Green},
		Transform{0.07, color.Yellow},
		Transform{0.20, color.Orange},
		Transform{1.00, color.Red},
		Transform{2.00, color.Purple}}
	t.SleetTransform = []Transform{
		Transform{0.00, color.Black},
		Transform{0.05, color.Pink},
		Transform{0.20, color.Purple}}
	t.SnowTransform = []Transform{
		Transform{0.00, color.Black},
		Transform{0.05, color.LightBlue},
		Transform{0.30, color.DarkBlue},
		Transform{0.60, color.Purple}}

	return t
}

// ForecastToColor maps the forecast to display colors
func (tr *Transformer) ForecastToColor(f darksky.ForecastResponse) []color.WCColor {

	// tbd seem to be 61 items in Minutely, but...weird...need to understand better.
	// I know I'll only have 60 LEDs to light, so...
	colors := make([]color.WCColor, len(f.Minutely.Data))

	// we also might want to consider including probability in here?
	for idx, m := range f.Minutely.Data {
		switch m.PrecipType {
		case "rain":
			{
				colors[idx] = intensityToColor(float64(m.PrecipIntensity), tr.RainTransform)
			}
		case "sleet":
			{
				colors[idx] = intensityToColor(float64(m.PrecipIntensity), tr.SleetTransform)
			}
		case "snow":
			{
				colors[idx] = intensityToColor(float64(m.PrecipIntensity), tr.SnowTransform)
			}
		default:
			{
				colors[idx] = color.Black
			}
		}
	}
	return colors
}

func intensityToColor(intensity float64, t []Transform) color.WCColor {

	var c = color.Black
	for i := 0; i < len(t); i++ {
		if intensity >= t[i].Intensity {
			c = t[i].Color
		} else {
			break
		}
	}
	return c
}

// Dim reduces each color value to product with dimVal
func Dim(c []color.WCColor, dimVal float32) []color.WCColor {

	colors := make([]color.WCColor, 60)

	for i := 0; i < 60; i++ {

		r := uint8(float32(c[i].R) * dimVal)
		g := uint8(float32(c[i].G) * dimVal)
		b := uint8(float32(c[i].B) * dimVal)

		colors[i] = color.WCColor{R: r, G: g, B: b}
	}
	return colors
}

// OverlayColors adds the supplied colors together
func OverlayColors(fc []color.WCColor, ac []color.WCColor) []color.WCColor {

	// not really sure what I want to see. For now...we're just gonna add 'em up
	cs := make([]color.WCColor, 60)
	var r, g, b uint
	for i := 0; i < 60; i++ {

		r = uint(fc[i].R + ac[i].R)
		g = uint(fc[i].G + ac[i].G)
		b = uint(fc[i].B + ac[i].B)

		if r > 255 {
			r = 255
		}
		if g > 255 {
			g = 255
		}
		if b > 255 {
			b = 255
		}

		cs[i].R = uint8(r)
		cs[i].G = uint8(g)
		cs[i].B = uint8(b)
	}

	//return cs  ignore what we're doing here for now
	return fc
}

// AllSameColors returns an array of all the same color
func AllSameColors(c color.WCColor) []color.WCColor {
	cs := make([]color.WCColor, 60)
	for i := 0; i < 60; i++ {
		cs[i] = c
	}
	return cs
}

func testColors() []color.WCColor {
	cs := make([]color.WCColor, 60)
	for i := 0; i < 60; i++ {
		cs[i] = color.Black
	}

	cs[0] = color.Green
	cs[1] = color.Yellow
	cs[2] = color.Orange
	cs[3] = color.Red
	cs[4] = color.Purple
	return cs

}

func traceMapping(f darksky.ForecastResponse, cs []color.WCColor) {

	for i := 0; i < 60; i++ {
		log.Tracef("%s i:%f p:%f- %d %d %d", f.Minutely.Data[i].PrecipType, f.Minutely.Data[i].PrecipIntensity, f.Minutely.Data[i].PrecipProbability,
			cs[i].R, cs[i].G, cs[i].B)
	}
}
