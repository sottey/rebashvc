//go:build windows

package jotframe

import (
	"os"

	"github.com/Depau/switzerland"
	"golang.org/x/term"
)

var (
	sigwinch = make(chan os.Signal)
)

type terminalSize struct {
	rows    uint16
	cols    uint16
	xPixels uint16
	yPixels uint16
}

func GetTerminalSize() (int, int) {
	return terminalWidth, terminalHeight
}

func getTerminalSize() (int, int) {
	width, height, err := term.GetSize(0)
	if err != nil {
		return -1, -1
	} else {
		return width, height
	}
}

func pollSignals() {
	// set signal handlers
	winch := make(chan switzerland.WinchSignal)
	switz := switzerland.GetSwitzerland()
	switz.Notify(winch)
	defer switz.Stop(winch)
	defer close(winch)

	// watch for events
	for {
		select {
		case <-winch:
			terminalWidth, terminalHeight = getTerminalSize()
			lock := getScreenLock()
			lock.Lock()
			clearScreen()
			refresh()
			lock.Unlock()
		}
	}
}
