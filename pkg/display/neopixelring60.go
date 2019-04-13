package display

import (
	"log"

	"time"

	"github.com/kellydunn/go-opc"
)

type Color struct {
	R, G, B uint8
}

type Minutes struct {
	Colors     [60]Color
	PixelCount int
}

func Run(c chan Minutes) {

	var minutes Minutes

	// TODO: is this necessary?
	// TODO: hard-coded
	for i := 0; i < 60; i++ {
		minutes.Colors[i] = Color{R: 0, G: 0, B: 0}
	}

	// receive from channel
	select {
	case minutes = <-c:
	default:
	}

	// Create a client
	oc := opc.NewClient()

	// TODO: hard-coded server address is hard-coded
	err := oc.Connect("tcp", "localhost:7890")

	// TODO: FATAL is too harsh. on-device we need to be more
	// accomodating in case fcserver isn't up yet when we want it
	if err != nil {
		log.Fatal("Could not connect to Fadecandy server", err)
	}

	for {
		m := opc.NewMessage(0)
		m.SetLength(60) // TODO: hard-coded

		// TODO: hard-coded
		for i := 0; i < 60; i++ {
			m.SetPixelColor(i, minutes.Colors[i].R, minutes.Colors[i].G, minutes.Colors[i].B)
		}

		err = oc.Send(m)
		if err != nil {
			log.Println("couldn't send color", err)
		} else {
			log.Println("send color")
		}

		// TODO: hard-coded
		time.Sleep(time.Duration(100) * time.Millisecond)
	}

}
