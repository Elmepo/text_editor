package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"golang.org/x/term"
)

type Logger struct {
	FileName string
}

type TextEditor struct {
	Logger
	fileContents          string
	cursorPosition        [2]int
	commandCursorPosition int
	fileName              string
	fileObject            *os.File
	pageWidth             int
	pageHeight            int
	width                 int
	height                int
	IN_COMMAND_MODE       bool
	command               string
	lineNumBuffer         int
}

func (l *Logger) Log(message string) {
	f, err := os.OpenFile(l.FileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	now := time.Now()
	logMessage := fmt.Sprintf("%s: %s\r\n", now.Format(time.DateTime), message)
	if _, err := f.Write([]byte(logMessage)); err != nil {
		panic(err)
	}
}

func (te *TextEditor) debugPrint() {
	hw := te.pageWidth / 2
	moveCursor(te.pageWidth+(hw/2), 0)
	fmt.Printf("Cursor Position: %d, %d", te.cursorPosition[0], te.cursorPosition[1])
	moveCursor(te.pageWidth+(hw/2), 1)
	fmt.Printf("Command Cursor Position: %d", te.commandCursorPosition)
	moveCursor(te.pageWidth+(hw/2), 2)
	fmt.Printf("Command: %s", te.command)
	moveCursor(te.pageWidth+(hw/2), 3)
	fmt.Printf("Terminal Dimensions: %d, %d", te.width, te.height)
	moveCursor(te.pageWidth+(hw/2), 4)
	fmt.Printf("Page Dimensions: %d, %d", te.pageWidth, te.height-1)
}

func getLineNumWidth(num int) int {
	if num%10 == 0 {
		return 1
	}
	return 1 + getLineNumWidth(num/10)
}

func (t *TextEditor) printContents(xa, xb, ya, yb int) {
	moveCursor(xa, ya)
	lines := 0
	t.Logger.Log(fmt.Sprintf("xa: %d, xb: %d, ya: %d, yb: %d", xa, xb, ya, yb))
	// pagePos := 0
	pageSize := xb - xa
	lineLength := 0
	for ci, c := range t.fileContents {
		t.Logger.Log(fmt.Sprintf("Bytes: %v, %q - CI: %d - lineLength%%pageSize==%d - pageSize: %d - lineLength: %d", c, c, ci, lineLength%pageSize, pageSize, lineLength))
		// if string(c) == "\r" {
		// 	moveCursor(xa, ya+lines)
		// 	// } else if string(c) == "\n" || ci%xb == 0 {
		// 	// } else if c == 10 || ci%xb == 0 {
		// 	// } else if c == 10 || (ci+1)%pageSize == 0 {
		// } else if c == 10 {
		if c == 10 {
			lines += 1
			lineLength = 0
			moveCursor(xa, ya+lines)
			t.Logger.Log(fmt.Sprintf("Moved cursor to (%d,%d)", xa, ya+lines))
			continue
		} else {
			fmt.Print(string(c))
			lineLength += 1
			if lineLength%pageSize == 0 {
				lines += 1
				lineLength = 0
				moveCursor(xa, ya+lines)
			}
			// if ci > 0 {
			// 	if lines > 0 {
			// 		if (ci+1)%(pageSize*lines) == 0 {
			// 			lines += 1
			// 			moveCursor(xa, ya+lines)
			// 		}
			// 	} else {
			// 		if (ci+1)%pageSize == 0 {
			// 			lines += 1
			// 			moveCursor(xa, ya+lines)
			// 		}
			// 	}
			// }
		}
	}
}

func (te *TextEditor) render() {
	// Wipe the current window, then redraw all the file contents?
	// lineNumBuffer := getLineNumWidth(te.height) + 1
	fmt.Print("\033[2J")
	// te.fileContents += "this is a render test"
	for i := range te.height - 1 {
		moveCursor(0, i)
		fmt.Printf("%2d", i)
		moveCursor(te.pageWidth+te.lineNumBuffer, i)
		fmt.Print("\u2590")
	}
	// fmt.Print("\033[H")
	// moveCursor(lineNumBuffer, 0)
	// fmt.Print(te.fileContents)
	// for ci, c := range te.fileContents {
	// 	// if string(c) == "\r" {
	// 	// 	te.Logger.Log("Found carraige return")
	// 	// }
	// 	if ci%te.pageWidth == 0 {
	//
	// 	}
	// 	fmt.Print(string(c))
	// }
	// te.Logger.Log(te.fileContents)
	moveCursor(te.lineNumBuffer, te.height-2)
	for j := range te.width - te.lineNumBuffer {
		if j == te.pageWidth {
			fmt.Print("\u259F")
		} else {
			fmt.Print("\u2584")
		}
	}
	// te.Logger.Log(fmt.Sprintf("Page Width: %d", te.pageWidth))
	// Think I shouldn't need te.lineNumBuffer for xb. Suspect I can refactor to make pagewidth
	// relative to the overall lineNumBuffer???
	te.printContents(te.lineNumBuffer, te.pageWidth+te.lineNumBuffer, 0, te.height-2)
	moveCursor(te.lineNumBuffer, te.height)
	fmt.Print(te.command)
	te.debugPrint()
	if te.IN_COMMAND_MODE {
		moveCursor(te.commandCursorPosition+te.lineNumBuffer, te.height)
	} else {
		moveCursor(te.cursorPosition[0]+te.lineNumBuffer, te.cursorPosition[1])
	}
}

func moveCursor(x int, y int) {
	// ESC [ LINE ; COL H
	fmt.Printf("\033[%d;%dH", y+1, x)
}

// func runCommand(com string, temp string, width int, height int) {
func (te TextEditor) runCommand() {
	// fmt.Print(com)
	commandSlice := strings.Split(te.command, " ")
	switch commandSlice[0] {
	case "save":
		if len(commandSlice) > 1 {
			fileName := commandSlice[1]
			fo, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR, 0644)
			if err != nil {
				panic(err)
			}
			defer func() {
				if err := fo.Close(); err != nil {
					panic(err)
				}
			}()
			te.fileObject = fo
		}
		if _, err := te.fileObject.Write([]byte(te.fileContents)); err != nil {
			panic(err)
		}
		moveCursor(0, te.height)
		for _ = range te.width {
			fmt.Print(" ")
		}
		fmt.Printf("Saved file in %s", te.fileName)
	}
}

func main() {
	logFileName := "log_file.txt"
	err := os.Remove(logFileName)
	if err != nil {
		panic(err)
	}
	l := Logger{
		FileName: logFileName,
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
	fmt.Print("\033[?7l")
	defer fmt.Print("\033[?7h")
	l.Log("Saved previous state and moved to alternate screen")

	w, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		panic(err)
	}
	l.Log(fmt.Sprintf("Terminal size: %d x %d", w, h))
	// fmt.Printf("Terminal size: %d x %d\r\n", width, height)

	l.Log(fmt.Sprintf("CommandLine Args: %v", os.Args))
	te := &TextEditor{
		// pageWidth: w/ 2,
		pageWidth:     w / 6,
		width:         w,
		height:        h,
		Logger:        l,
		lineNumBuffer: getLineNumWidth(h) + 1,
	}
	if len(os.Args) > 1 {
		// te = &TextEditor{
		// 	fileName: os.Args[1],
		// 	width:    w - 20,
		// 	height:   h,
		// 	Logger:   l,
		// }
		te.fileName = os.Args[1]
		// dat, err := os.ReadFile(te.fileName)
		// fo, err := os.Open
		te.fileObject, err = os.OpenFile(te.fileName, os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			panic(err)
		}
		defer te.fileObject.Close()
		fileBytes, err := io.ReadAll(te.fileObject)
		if err != nil {
			panic(err)
		}
		te.fileContents = string(fileBytes)
		// } else {
		// 	te = &TextEditor{
		// 		width:  w - 20,
		// 		height: h,
		// 		Logger: l,
		// 	}
	}
	fmt.Print(te.fileContents)

	// Using a 3 byte buffer because the arrow keys occur on the third index of
	// an ESC command (ESC [ A,B,C,D)
	buffer := make([]byte, 3)
	l.Log(fmt.Sprintf("Cursor Positions %d, %d - Command Cursor %d - IN_COMMAND_MODE %b", te.cursorPosition[0], te.cursorPosition[1], te.commandCursorPosition, te.IN_COMMAND_MODE))
	te.render()
MAIN_LOOP:
	for {
		_, err := os.Stdin.Read(buffer)
		if err != nil {
			break
		}
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
				if te.IN_COMMAND_MODE {
					// fmt.Printf("\033[%d;%dH", cursorPosition[0], commandCursorPosition)
					moveCursor(te.cursorPosition[0], te.cursorPosition[1])
					te.IN_COMMAND_MODE = false
				} else {
					// fmt.Printf("\033[%d;%dH", height, commandCursorPosition)
					moveCursor(te.commandCursorPosition, te.height)
					te.IN_COMMAND_MODE = true
				}
			} else {
				// Arrow key
				if buffer[2] == 65 && te.cursorPosition[1] > 0 {
					// UP
					te.cursorPosition[1] -= 1
				} else if buffer[2] == 66 && te.cursorPosition[1] < te.height {
					// DOWN
					te.cursorPosition[1] += 1
				} else if buffer[2] == 67 && te.cursorPosition[0] < te.width {
					// RIGHT
					te.cursorPosition[0] += 1
				} else if buffer[2] == 68 && te.cursorPosition[0] > 0 {
					// LEFT
					te.cursorPosition[0] -= 1
				}
				// Eventually need to make this relative to content
				moveCursor(te.cursorPosition[0], te.cursorPosition[1])
			}
			// break MAIN_LOOP
		// BACKSPACE?
		case 127:
			// fmt.Print("DEBUG")
			if te.IN_COMMAND_MODE {
				te.command = te.command[:len(te.command)-1]
				// moveCursor(len(te.command), te.height)
				// fmt.Print(" ")
				// moveCursor(len(te.command), te.height)
				te.commandCursorPosition -= 1
			} else {
				// Handle cross line later
				// if te.cursorPosition[0] > 0 {
				// 	te.fileContents = te.fileContents[:len(te.fileContents)]
				// 	// te.cursorPosition[0] -= 1
				// 	// moveCursor(te.cursorPosition[0], te.cursorPosition[1])
				// 	// fmt.Print(" ")
				// 	// moveCursor(te.cursorPosition[0], te.cursorPosition[1])
				// }
				te.fileContents = te.fileContents[:len(te.fileContents)-1]
			}
		// ENTER
		case 13:
			// l.Log(fmt.Sprintf("Enter pressed: Cursor Positions %d, %d - Command Cursor %d - IN_COMMAND_MODE %b", cursorPosition[0], cursorPosition[1], commandCursorPosition, IN_COMMAND_MODE))
			if te.IN_COMMAND_MODE {
				te.runCommand()
				te.command = ""
				te.commandCursorPosition = 0
				moveCursor(te.cursorPosition[0], te.cursorPosition[1])
				te.IN_COMMAND_MODE = false
			} else {
				// fmt.Print("\r\n")
				l.Log(fmt.Sprintf("%d, %d", te.cursorPosition[0], te.cursorPosition[1]))
				te.cursorPosition[0] = 0
				te.cursorPosition[1] += 1
				// moveCursor(cursorPosition[0], cursorPosition[1])
				// fmt.Print("\r\n")
				te.fileContents += "\r\n"
				// fmt.Printf("Cursor Positions: %d, %d", cursorPosition[0], cursorPosition[1])
				// l.Log(fmt.Sprintf("After manipulation: %d, %d", cursorPosition[0], cursorPosition[1]))
			}
		default:
			l.Log(fmt.Sprintf("Default Case. Buffer: %v", buffer))
			// fmt.Printf("%v", buffer)
			// fmt.Printf("%c", key)
			if te.IN_COMMAND_MODE {
				te.commandCursorPosition += 1
				te.command += fmt.Sprintf("%c", key)
			} else {
				te.cursorPosition[0] += 1
				te.fileContents += fmt.Sprintf("%c", key)
			}
		}

		// Quickly reset the buffer
		// buffer = []byte{0,0,0,0}
		for bi := range len(buffer) {
			buffer[bi] = 0
		}
		te.render()
	}
}
