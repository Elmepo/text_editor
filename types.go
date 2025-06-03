package main

import "os"

type Region struct {
	Logger         Logger
	windowXA       int
	windowXB       int
	windowYA       int
	windowYB       int
	xa             int
	xb             int
	ya             int
	yb             int
	cursorPosition int
	fileCoords     map[int][2]int
	cursorCoords   map[[2]int]int
	Content        string
	lineNumBuffer  int
}

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
	// Logically I think combining the line + file regions might be better but
	// might also be more effort than it's worth right now
	lineRegion    *Region
	fileRegion    *Region
	commandRegion *Region
}
