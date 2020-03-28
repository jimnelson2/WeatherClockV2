// Package transform converts radar/weather data into
// structures for display
package transform

import (
	"strings"
	"time"

	"github.com/jimnelson2/WeatherClockV2/pkg/color"
	"github.com/shawntoffel/darksky"
	log "github.com/sirupsen/logrus"
)

// Transform defines the minimal intensity at which a precipitation intensity takes effect
type Transform struct {
	Intensity float64
	Color     color.WCColor
}

// Transformer holds slices of precipitation-type specific transformations
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

// ClockFace returns a clock face for current time
func (tr *Transformer) ClockFace() []color.WCColor {

	var c = AllSameColors(color.Black)

	t := time.Now().Local()
	h := t.Hour()
	m := t.Minute()

	// Light up the "hour"
	c[((h%12)*5)%60] = color.White

	// Light up the minutes (incl +/- 1)
	c[(m+59)%60] = color.White
	c[m] = color.White
	c[(m+1)%60] = color.White

	return c
}

// ForecastToAlert determines if an alert is active, and if so what color to represent
func (tr *Transformer) ForecastToAlert(f darksky.ForecastResponse) (active bool, c []color.WCColor) {

	c = make([]color.WCColor, 60)
	active = false

	// Detection of active alert is really not good here
	// should almost certainly be accounting for expires time, etc.
	for _, a := range f.Alerts {
		if a.Severity == "warning" {
			if strings.Contains(strings.ToLower(a.Title), "tornado warning") {
				c = AllSameColors(color.Red)
				active = true
				log.Info("Tornado warning active")
				return active, c
			}
			if strings.Contains(strings.ToLower(a.Title), "thunderstorm warning") {
				c = AllSameColors(color.Yellow)
				active = true
				log.Info("Thunderstorm warning active")
			}
		}
	}

	return active, c
}

// ForecastToColor maps the forecast to display colors
func (tr *Transformer) ForecastToColor(f darksky.ForecastResponse) []color.WCColor {

	colors := make([]color.WCColor, 60)

	// I...don't know if I like this
	if len(f.Minutely.Data) != 61 {
		log.Errorf("Asked to transform a forecast with %d minutes. Expected 61", len(f.Minutely.Data))
		return color.NewColors(60)
	}

	// we also might want to consider including probability in here?
	for idx, m := range f.Minutely.Data {
		if idx < 60 {
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
	}
	return colors
}

// ApplyTransformDefinition modifies the transform based on the provided change string
func (tr *Transformer) ApplyTransformDefinition(change string) {

	// This feels yucky...not sure how to make it better yet...
	// The very worst of this is you can't keyboard-close the process :/

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

	// Also...I've only addressed rain here...
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
func (tr *Transformer) OverlayColors(fc []color.WCColor, ac []color.WCColor) []color.WCColor {

	cs := make([]color.WCColor, 60)
	var r, g, b uint
	for i := 0; i < 60; i++ {

		// this would "add" the colors together, dumbly
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

	return cs
}

// AllSameColors returns an array of all the same color
func AllSameColors(c color.WCColor) []color.WCColor {
	cs := make([]color.WCColor, 60)
	for i := 0; i < 60; i++ {
		cs[i] = c
	}
	return cs
}

// LuxToDim converts a lux value to a 0-100 dim percentage. Hard-coded range.
func LuxToDim(lux float64) float32 {

	var maxLux, minLux float64
	maxLux = 75
	minLux = 0.01

	var maxDim, minDim float64
	maxDim = 0.9
	minDim = 0.1

	// we want our lights to be not-to-dark, not-to-bright.
	// we have a sensor reporting ambient light in lux
	// so use that lux value to control how much to
	// dim the lights. we're capping all ranges, might
	// take some experimentation

	if lux > maxLux {
		log.Infof("override lux of %f to %f", lux, maxLux)
		lux = maxLux
	}
	if lux < minLux {
		log.Infof("override lux of %f to %f", lux, minLux)
		lux = minLux
	}

	// linear transform
	// if x falls b/w a and b, create y falling between c and d
	// y := (x-a)/(b-a) * (d-c) + c
	dim := (lux-minLux)/(maxLux-minLux)*(maxDim-minDim) + minDim

	log.Infof("given lux %f, will dim to %f of max", lux, dim)
	return float32(dim)

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

func traceMapping(f darksky.ForecastResponse, cs []color.WCColor) {

	for i := 0; i < 60; i++ {
		log.Tracef("%s i:%f p:%f- %d %d %d", f.Minutely.Data[i].PrecipType, f.Minutely.Data[i].PrecipIntensity, f.Minutely.Data[i].PrecipProbability,
			cs[i].R, cs[i].G, cs[i].B)
	}
}
