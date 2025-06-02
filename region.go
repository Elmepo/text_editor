package main

import "fmt"

func (r Region) Redraw(contents string) (map[int][2]int, map[[2]int]int) {
	characterCoordMap := make(map[int][2]int)
	reverseCharacterCoordMap := make(map[[2]int]int)
	moveCursor(r.xa, r.ya)
	lines := 0
	// t.Logger.Log(fmt.Sprintf("xa: %d, xb: %d, ya: %d, yb: %d", r.xa, r.xb, r.ya, r.yb))
	pageSize := r.xb - r.xa
	lineLength := 0
	currentLine := ""
	x, y := 0, 0
	for ci, c := range contents {
		// r.Logger.Log(fmt.Sprintf("Bytes: %v, %q - CI: %d - lineLength%%pageSize==%d - pageSize: %d - lineLength: %d", c, c, ci, lineLength%pageSize, pageSize, lineLength))
		if c == 10 {
			currentLine = ""
			lines += 1
			y += 1
			x = 0
			lineLength = 0
			moveCursor(r.xa, r.ya+lines)
			// t.Logger.Log(fmt.Sprintf("Moved cursor to (%d,%d)", xa, ya+lines))
			characterCoordMap[ci] = [2]int{x, y}
			reverseCharacterCoordMap[[2]int{x, y}] = ci
			continue
		} else {
			fmt.Print(string(c))
			currentLine = fmt.Sprintf("%s%s", currentLine, string(c))
			lineLength += 1
			// ' '
			if c == 32 {
				wordSizeLookAhead := getWordLength(contents, ci+1)
				// t.Logger.Log(fmt.Sprintf("word lookahead = %s (indexes %d, %d)", t.fileContents[ci:ci+wordSizeLookAhead+1], ci, ci+wordSizeLookAhead))
				if (lineLength + wordSizeLookAhead) >= pageSize {
					currentLine = ""
					lines += 1
					y += 1
					x = 0
					lineLength = 0
					moveCursor(r.xa, r.ya+lines)
				}
			} else {
				if lineLength%pageSize == 0 {
					currentLine = ""
					lines += 1
					y += 1
					x = 0
					lineLength = 0
					moveCursor(r.xa, r.ya+lines)
				}
			}
		}
		characterCoordMap[ci] = [2]int{x, y}
		reverseCharacterCoordMap[[2]int{x, y}] = ci
		x += 1
		// t.Logger.Log(fmt.Sprintf("Character Coords for %s: %v", string(c), t.characterCoordMap[ci]))
	}
	return characterCoordMap, reverseCharacterCoordMap
}
