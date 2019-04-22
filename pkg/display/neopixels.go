// Package display provides routines to realize color data
// on a physical device
package display

import (
	"time"

	"github.com/jimnelson2/WeatherClockV2/pkg/color"
	"github.com/kellydunn/go-opc"
	log "github.com/sirupsen/logrus"
)

const (
	// PixelCount sets the number of physical pixels in hardware
	PixelCount int = 60
)

// Pixels holds a full-display's worth of pixel data
type Pixels struct {
	Colors     []color.WCColor
	PixelCount int
}

// Run is a forever-loop to display the incoming Pixel data
func Run(c chan Pixels) {

	// Create a client
	oc := opc.NewClient()

	// Could this be a run-time configuration? Yes. Am am I
	// going to bother for my use case? No.
	err := oc.Connect("tcp", "localhost:7890")

	// Significant assumption, maore of an assertion, really. The fadecandy
	// server is required to be available to us immediately.
	if err != nil {
		log.Fatal("Could not connect to Fadecandy server", err)
	}

	m := Pixels{Colors: make([]color.WCColor, 60), PixelCount: 60}

	// We probably don't have pixel data yet. Default to all black
	// do we need to do this?
	for i := 0; i < PixelCount; i++ {
		m.Colors[i] = color.Black
	}

	for {
		log.Debug("Start display loop")
		select {
		case m = <-c:
			msg := opc.NewMessage(0)

			// reminder that each LED has 3 pixels, in r-g-b order
			msg.SetLength(uint16(PixelCount * 3))

			// Add all pixel data to the message
			for i := 0; i < PixelCount; i++ {
				// reminder this is effectively setting three pixel values at once
				msg.SetPixelColor(i, m.Colors[i].R, m.Colors[i].G, m.Colors[i].B)
			}

			err = oc.Send(msg)
			if err != nil {
				log.Error("didn't send color data to fadecandy board", err)
			} else {
				log.Trace("sent color to fadecandy board")
			}

		default:
		}

		time.Sleep(time.Duration(100) * time.Millisecond)
	}
}
