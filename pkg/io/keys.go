package io

import (
	term "github.com/nsf/termbox-go"
	log "github.com/sirupsen/logrus"
)

func reset() {
	term.Sync() // cosmestic purpose
}

// GetKeys sends key press strings, one at a time, on the provided channel
func GetKeys(c chan string) {
	// Based on https://www.socketloop.com/tutorials/golang-get-ascii-code-from-a-key-press-cross-platform-example
	err := term.Init()
	if err != nil {
		log.Errorf("term initialization failure: %v", err)
	}
	defer term.Close()

	for {
		switch ev := term.PollEvent(); ev.Type {
		case term.EventKey:
			switch ev.Key {
			case term.KeyEsc:
				break
			default:
				// we only want to read a single character or one key pressed event
				reset()
				c <- string(ev.Ch)
			}
		}
	}
}
