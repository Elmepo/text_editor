package main

import (
	"math/rand/v2"
	"fmt"
	"os"
	"golang.org/x/term"
	"time"
)

type Logger struct {
	FileName string
}

func (l *Logger) Log(message string) {
	f, err := os.OpenFile(l.FileName, os.O_APPEND | os.O_CREATE | os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	now := time.Now()
	logMessage := fmt.Sprintf("%s: %s\r\n", now.Format(time.DateTime), message)
	if _, err := f.Write([]byte(logMessage)); err != nil {
		panic(err)
	}
}

func moveCursor(x int, y int) {
	// ESC [ LINE ; COL H
	fmt.Printf("\033[%d;%dH", y, x)
}

func runCommand(com string, temp string, width int, height int) {
	// fmt.Print(com)
	switch com {
	case "save":
		temp_file_name := fmt.Sprintf("test_file_%d.txt", rand.IntN(100))
		fo, err := os.Create(temp_file_name)
		if err != nil {
			panic(err)
		}
		defer func() {
			if err := fo.Close(); err != nil {
				panic(err)
			}
		}()
		if _, err := fo.Write([]byte(temp)); err != nil {
			panic(err)
		}
		moveCursor(0, height)
		for _ = range(width) {
			fmt.Print(" ")
		}
		fmt.Printf("Saved file in %s", temp_file_name)
	}
}

func main() {
	l := Logger{
		FileName: "log_file.txt",
	}
	l.Log("Started Program")
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
	l.Log("Saved previous state and moved to alternate screen")

	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		panic(err)
	}
	l.Log(fmt.Sprintf("Terminal size: %d x %d", width, height))
	// fmt.Printf("Terminal size: %d x %d\r\n", width, height)

	cursorPosition := []int{0,0}
	commandCursorPosition := 0
	IN_COMMAND_MODE := false
	command := ""
	fileContents := ""
	// Using a 3 byte buffer because the arrow keys occur on the third index of
	// an ESC command (ESC [ A,B,C,D)
	buffer := make([]byte, 3)
	l.Log(fmt.Sprintf("Cursor Positions %d, %d - Command Cursor %d - IN_COMMAND_MODE %b", cursorPosition[0], cursorPosition[1], commandCursorPosition, IN_COMMAND_MODE))
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
		case 27:
			// fmt.Printf("\n\rBuffer: %v", buffer)
			if buffer[1] == 0 && buffer[2] == 0 {
				// Just a pure ESC
				// break MAIN_LOOP
				// ESC [ LINE ; COL H
				if IN_COMMAND_MODE {
					// fmt.Printf("\033[%d;%dH", cursorPosition[0], commandCursorPosition)
					moveCursor(cursorPosition[0], cursorPosition[1])
					IN_COMMAND_MODE = false
				} else {
					// fmt.Printf("\033[%d;%dH", height, commandCursorPosition)
					moveCursor(commandCursorPosition, height)
					IN_COMMAND_MODE = true
				}
			} else {
				// Arrow key
				if buffer[2] == 65 && cursorPosition[1] > 0 {
					// UP
					cursorPosition[1] -= 1
				} else if buffer[2] == 66 && cursorPosition[1] < height {
					// DOWN
					cursorPosition[1] += 1
				} else if buffer[2] == 67 && cursorPosition[0] < width {
					// RIGHT
					cursorPosition[0] += 1
				} else if buffer[2] == 68 && cursorPosition[0] > 0 {
					// LEFT
					cursorPosition[0] -= 1
				}
			}
			// break MAIN_LOOP
		// BACKSPACE?
		case 127:
			// fmt.Print("DEBUG")
			if IN_COMMAND_MODE {
				command = command[:len(command)]
				moveCursor(len(command), height)
				fmt.Print(" ")
				moveCursor(len(command), height)
				commandCursorPosition -= 1
			} else {
				// Handle cross line later
				if cursorPosition[0] > 0 {
					fileContents = fileContents[:len(fileContents)]
					cursorPosition[0] -= 1
					moveCursor(cursorPosition[0], cursorPosition[1])
					fmt.Print(" ")
					moveCursor(cursorPosition[0], cursorPosition[1])
				}
			}
		// ENTER
		case 13:
			// l.Log(fmt.Sprintf("Enter pressed: Cursor Positions %d, %d - Command Cursor %d - IN_COMMAND_MODE %b", cursorPosition[0], cursorPosition[1], commandCursorPosition, IN_COMMAND_MODE))
			l.Log("ENTER")
			if IN_COMMAND_MODE {
				runCommand(command, fileContents, width, height)
				command = ""
				commandCursorPosition = 0
				moveCursor(cursorPosition[0], cursorPosition[1])
				IN_COMMAND_MODE = false
			} else {
				// fmt.Print("\r\n")
				l.Log(fmt.Sprintf("%d, %d", cursorPosition[0], cursorPosition[1]))
				cursorPosition[0] = 0
				cursorPosition[1] += 1
				// moveCursor(cursorPosition[0], cursorPosition[1])
				fmt.Print("\r\n")
				fileContents += "\r\n"
				// fmt.Printf("Cursor Positions: %d, %d", cursorPosition[0], cursorPosition[1])
				// l.Log(fmt.Sprintf("After manipulation: %d, %d", cursorPosition[0], cursorPosition[1]))
			}
		default:
			l.Log(fmt.Sprintf("Default Case. Buffer: %v", buffer))
			// fmt.Printf("%v", buffer)
			fmt.Printf("%c", key)
			if IN_COMMAND_MODE {
				commandCursorPosition += 1
				command += fmt.Sprintf("%c", key)
			} else {
				cursorPosition[0] += 1
				fileContents += fmt.Sprintf("%c", key)
			}
		}
		// Quickly reset the buffer
		// buffer = []byte{0,0,0,0}
		for bi := range(len(buffer)) {
			buffer[bi] = 0
		}
	}

}
