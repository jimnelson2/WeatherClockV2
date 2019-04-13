package display

import (
	"time"

	"github.com/kellydunn/go-opc"
	log "github.com/sirupsen/logrus"
)

type Minutes struct {
	Colors     []Color
	PixelCount int
}

func Run(c chan Minutes) {

	// Create a client
	oc := opc.NewClient()

	// TODO: hard-coded server address is hard-coded
	err := oc.Connect("tcp", "localhost:7890")

	// TODO: FATAL is too harsh. on-device we need to be more
	// accomodating in case fcserver isn't up yet when we want it
	if err != nil {
		log.Fatal("Could not connect to Fadecandy server", err)
	}

	m := Minutes{Colors: make([]Color, 60), PixelCount: 60}

	// TODO: is this necessary?
	// TODO: hard-coded
	for i := 0; i < 60; i++ {
		m.Colors[i] = Color{R: 0, G: 0, B: 0}
	}

	for {
		// receive from channel
		select {
		case m = <-c:
			log.Debug("Display got new message to process")
		default:
			msg := opc.NewMessage(0)
			// reminder each LED has 3 pixels, in r-g-b order
			msg.SetLength(180) // TODO: hard-coded

			// TODO: hard-coded
			for i := 0; i < 60; i++ {
				msg.SetPixelColor(i, m.Colors[i].R, m.Colors[i].G, m.Colors[i].B)
			}

			err = oc.Send(msg)
			if err != nil {
				log.Error("couldn't send color data to fadecandy board", err)
			} else {
				log.Debug("sent color to fadecandy board")
			}
		}

		// TODO: hard-coded
		time.Sleep(time.Duration(1000) * time.Millisecond)
	}

}