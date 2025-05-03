package main

import (
	"fmt"
	"os"
	"golang.org/x/term"
)

func main() {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)
	// ANSI Code to clear the screen, unsure if it will work on all devices
	// Use https://vt100.net/docs/vt100-ug/chapter3.html and https://invisible-island.net/xterm/ctlseqs/ctlseqs.html for reference + lookup
	// Alternate Screen mode + deferring the end alternate screen
	// CSI [ DECSET 1049 + CSI [ DECRST 1049
	fmt.Print("\033[?1049h")
	defer fmt.Print("\033[?1049l")
	// CSI H == clear screen
	// CSI 2J == cursor to top left
	fmt.Print("\033[H\033[2J")

	buffer := make([]byte, 1)
	MAIN_LOOP:
	for {
		_, err := os.Stdin.Read(buffer)
		if err != nil {
			break
		}
		// fmt.Printf("Pressed: %q (%d in bytes)\r\n", buffer[0], buffer[0])
		// if buffer[0] == 27 || buffer[0] == 3 {
		// 	break
		// }
		switch key := buffer[0]; key {
		// CTRL-C
		case 3:
			break MAIN_LOOP
		// ESC
		// case 27:
		// 	break MAIN_LOOP
		// ENTER
		case 13:
			fmt.Print("\r\n")
		default:
			fmt.Printf("%c", key)
		}
	}

}
