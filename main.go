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
	fileContents string
	// Position of the cursor within the fileContents
	cursorPosition           int
	commandCursorPosition    int
	fileName                 string
	fileObject               *os.File
	pageWidth                int
	pageHeight               int
	width                    int
	height                   int
	IN_COMMAND_MODE          bool
	command                  string
	lineNumBuffer            int
	currentLine              string
	x                        int
	y                        int
	characterCoordMap        map[int][2]int
	reverseCharacterCoordMap map[[2]int]int
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
	fmt.Printf("Cursor Position: %d", te.cursorPosition)
	moveCursor(te.pageWidth+(hw/2), 1)
	fmt.Printf("Command Cursor Position: %d", te.commandCursorPosition)
	moveCursor(te.pageWidth+(hw/2), 2)
	fmt.Printf("Command: %s", te.command)
	moveCursor(te.pageWidth+(hw/2), 3)
	fmt.Printf("Terminal Dimensions: %d, %d", te.width, te.height)
	moveCursor(te.pageWidth+(hw/2), 4)
	fmt.Printf("Page Dimensions: %d, %d", te.pageWidth, te.height-1)
	moveCursor(te.pageWidth+(hw/2), 5)
	fmt.Printf("Current Line: %s", te.currentLine)
}

func getLineNumWidth(num int) int {
	if num%10 == 0 {
		return 1
	}
	return 1 + getLineNumWidth(num/10)
}

func getWordLength(content string, startIdx int) int {
	wordLength := 0
	for _, c := range content[startIdx:] {
		if c == 32 || c == 10 {
			return wordLength
		}
		wordLength += 1
	}
	return wordLength
}

func (t *TextEditor) printContents(xa, xb, ya, yb int) {
	moveCursor(xa, ya)
	lines := 0
	t.Logger.Log(fmt.Sprintf("xa: %d, xb: %d, ya: %d, yb: %d", xa, xb, ya, yb))
	pageSize := xb - xa
	lineLength := 0
	currentLine := ""
	x, y := 0, 0
	for ci, c := range t.fileContents {
		// t.Logger.Log(fmt.Sprintf("Bytes: %v, %q - CI: %d - lineLength%%pageSize==%d - pageSize: %d - lineLength: %d", c, c, ci, lineLength%pageSize, pageSize, lineLength))
		if c == 10 {
			currentLine = ""
			lines += 1
			y += 1
			x = 0
			lineLength = 0
			moveCursor(xa, ya+lines)
			// t.Logger.Log(fmt.Sprintf("Moved cursor to (%d,%d)", xa, ya+lines))
			t.characterCoordMap[ci] = [2]int{x, y}
			t.reverseCharacterCoordMap[[2]int{x, y}] = ci
			continue
		} else {
			fmt.Print(string(c))
			currentLine = fmt.Sprintf("%s%s", currentLine, string(c))
			lineLength += 1
			// ' '
			if c == 32 {
				wordSizeLookAhead := getWordLength(t.fileContents, ci+1)
				// t.Logger.Log(fmt.Sprintf("word lookahead = %s (indexes %d, %d)", t.fileContents[ci:ci+wordSizeLookAhead+1], ci, ci+wordSizeLookAhead))
				if (lineLength + wordSizeLookAhead) >= pageSize {
					currentLine = ""
					lines += 1
					y += 1
					x = 0
					lineLength = 0
					moveCursor(xa, ya+lines)
				}
			} else {
				if lineLength%pageSize == 0 {
					currentLine = ""
					lines += 1
					y += 1
					x = 0
					lineLength = 0
					moveCursor(xa, ya+lines)
				}
			}
		}
		t.characterCoordMap[ci] = [2]int{x, y}
		t.reverseCharacterCoordMap[[2]int{x, y}] = ci
		x += 1
		// t.Logger.Log(fmt.Sprintf("Character Coords for %s: %v", string(c), t.characterCoordMap[ci]))
	}
}

func (te *TextEditor) render() {
	// Wipe the current window, then redraw all the file contents
	fmt.Print("\033[2J")
	for i := range te.height - 1 {
		moveCursor(0, i)
		fmt.Printf("%2d", i)
		moveCursor(te.pageWidth+te.lineNumBuffer, i)
		fmt.Print("\u2590")
	}
	moveCursor(te.lineNumBuffer, te.height-2)
	for j := range te.width - te.lineNumBuffer {
		if j == te.pageWidth {
			fmt.Print("\u259F")
		} else {
			fmt.Print("\u2584")
		}
	}
	// Think I shouldn't need te.lineNumBuffer for xb. Suspect I can refactor to make pagewidth
	// relative to the overall lineNumBuffer???
	te.printContents(te.lineNumBuffer, te.pageWidth+te.lineNumBuffer, 0, te.height-2)
	moveCursor(te.lineNumBuffer, te.height)
	fmt.Print(te.command)
	te.debugPrint()
	if te.IN_COMMAND_MODE {
		moveCursor(te.commandCursorPosition+te.lineNumBuffer, te.height)
	} else {
		moveCursor(te.characterCoordMap[te.cursorPosition][0]+te.lineNumBuffer, te.characterCoordMap[te.cursorPosition][1])
	}
}

func moveCursor(x int, y int) {
	// ESC [ LINE ; COL H
	fmt.Printf("\033[%d;%dH", y+1, x)
}

func (te TextEditor) runCommand() {
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

func (te *TextEditor) updateFileCursor(key byte) {
	currentCursorCoords := te.characterCoordMap[te.cursorPosition]
	switch key {
	case 65:
		// UP
		foundPosition := false
		xmod := 0
		if currentCursorCoords[1] == 0 {
			return
		}
		for !foundPosition {
			newx := currentCursorCoords[0] - xmod
			newy := currentCursorCoords[1] - 1
			te.Logger.Log(fmt.Sprintf("Not found position, xmod %v newx %v newy %v", xmod, newx, newy))

			if newy < 0 {
				te.Logger.Log("Hit newy")
				return
			}
			if newx < 0 {
				te.Logger.Log("Hit newx")
				return
			}
			newPosition := [2]int{newx, newy}
			newCoords, exists := te.reverseCharacterCoordMap[newPosition]
			te.Logger.Log(fmt.Sprintf("newCoords %v newPosition %v exists %v", newCoords, newPosition, exists))
			if !exists {
				xmod += 1
				continue
			} else {
				te.cursorPosition = newCoords
				foundPosition = true
			}
		}
	case 66:
		// DOWN
		foundPosition := false
		xmod := 0
		for !foundPosition {
			newx := currentCursorCoords[0] - xmod
			newy := currentCursorCoords[1] + 1
			if newx < 0 {
				return
			}
			newPosition := [2]int{newx, newy}
			te.Logger.Log(fmt.Sprintf("Current Coords: %v, newCoords: %v, newCursorCoords: %v", currentCursorCoords, newPosition, te.reverseCharacterCoordMap[newPosition]))
			newCoords, exists := te.reverseCharacterCoordMap[newPosition]
			if !exists {
				xmod += 1
				continue
			} else {
				te.cursorPosition = newCoords
				foundPosition = true
			}
		}
	case 67:
		// RIGHT
		if te.cursorPosition == len(te.fileContents) {
			te.Logger.Log("Hit the end of the file contents")
			return
		}
		newPosition := [2]int{currentCursorCoords[0] + 1, currentCursorCoords[1]}
		newCoords, exists := te.reverseCharacterCoordMap[newPosition]
		if !exists {
			newPosition := [2]int{0, currentCursorCoords[1] + 1}
			newCoords, exists = te.reverseCharacterCoordMap[newPosition]
			if !exists {
				return
			} else {
				te.cursorPosition = newCoords
			}
		} else {
			te.cursorPosition = newCoords
		}
	case 68:
		// LEFT
		if te.cursorPosition == 0 {
			return
		}

		foundPosition := false
		xmod := 1
		ymod := 0
		for !foundPosition {
			newPosition := [2]int{currentCursorCoords[0] - xmod, currentCursorCoords[1] - ymod}
			newCoords, exists := te.reverseCharacterCoordMap[newPosition]
			if !exists {
				xmod = te.pageWidth
				ymod -= 1
				continue
			} else {
				te.cursorPosition = newCoords
				foundPosition = true
			}
		}
	default:
		return
	}
	return
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

	l.Log(fmt.Sprintf("CommandLine Args: %v", os.Args))
	te := &TextEditor{
		pageWidth:                w / 6,
		width:                    w,
		height:                   h,
		Logger:                   l,
		lineNumBuffer:            getLineNumWidth(h) + 1,
		characterCoordMap:        make(map[int][2]int),
		reverseCharacterCoordMap: make(map[[2]int]int),
	}
	if len(os.Args) > 1 {
		te.fileName = os.Args[1]
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
	}
	fmt.Print(te.fileContents)

	// Using a 3 byte buffer because the arrow keys occur on the third index of
	// an ESC command (ESC [ A,B,C,D)
	buffer := make([]byte, 3)
	l.Log(fmt.Sprintf("Cursor Position %d - Command Cursor %d - IN_COMMAND_MODE %b", te.cursorPosition, te.commandCursorPosition, te.IN_COMMAND_MODE))
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
			if buffer[1] == 0 && buffer[2] == 0 {
				// Just a pure ESC
				if te.IN_COMMAND_MODE {
					te.IN_COMMAND_MODE = false
				} else {
					te.IN_COMMAND_MODE = true
				}
			} else {
				// Arrow keys (for now)
				if buffer[2] >= 65 && buffer[2] <= 68 {
					if te.IN_COMMAND_MODE {
						if buffer[2] == 67 {
							// RIGHT
							te.commandCursorPosition = min(len(te.command), te.commandCursorPosition+1)
						} else if buffer[2] == 68 {
							// LEFT
							te.commandCursorPosition = max(0, te.commandCursorPosition-1)
						}
					} else {
						te.updateFileCursor(buffer[2])
					}
				}
			}
		case 126:
			// DELETE
			if te.IN_COMMAND_MODE {
				// te.command = te.command[:len(te.command)]
				te.command = te.command[:max(0, te.commandCursorPosition)] + te.command[te.commandCursorPosition+1:]
			} else {
				te.fileContents = te.fileContents[:max(0, te.cursorPosition)] + te.fileContents[te.cursorPosition+1:]
			}
		case 127:
			// BACKSPACE
			if te.IN_COMMAND_MODE {
				te.command = te.command[:te.commandCursorPosition-1] + te.command[te.commandCursorPosition:]
				te.commandCursorPosition -= 1
				// te.command = te.command[:len(te.command)-1]
			} else {
				if te.cursorPosition-1 < 0 {
					continue
				}
				te.fileContents = te.fileContents[:te.cursorPosition-1] + te.fileContents[te.cursorPosition:]
				te.cursorPosition -= 1
			}
		// ENTER
		case 13:
			if te.IN_COMMAND_MODE {
				te.runCommand()
				te.command = ""
				te.commandCursorPosition = 0
				te.IN_COMMAND_MODE = false
			} else {
				te.fileContents = te.fileContents[:te.cursorPosition] + "\r\n" + te.fileContents[te.cursorPosition:]
			}
		default:
			l.Log(fmt.Sprintf("Default Case. Buffer: %v", buffer))
			if te.IN_COMMAND_MODE {
				te.commandCursorPosition += 1
				te.command += fmt.Sprintf("%c", key)
			} else {
				te.fileContents = te.fileContents[:te.cursorPosition] + fmt.Sprintf("%c", key) + te.fileContents[te.cursorPosition:]
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
