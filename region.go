package main

import "fmt"

func (r Region) drawRegionWindow() {
	flushRegion(r.windowXA, r.windowXB, r.windowYA, r.windowYB)
	for i := range r.windowYB - r.windowYA {
		moveCursor(0, i)
		fmt.Printf("%2d", i)
		moveCursor(r.windowXB, i)
		fmt.Print("\u2590")
	}
	moveCursor(r.windowXA, r.windowYB)
	for _ = range r.windowXB - r.windowXA - 1 {
		fmt.Print("\u2584")
	}
	fmt.Print("\u259f")
	moveCursor(r.xa, r.ya)
}

func (r Region) Redraw() {
	r.drawRegionWindow()
	moveCursor(r.xa, r.ya)
	lines := 0
	// t.Logger.Log(fmt.Sprintf("xa: %d, xb: %d, ya: %d, yb: %d", r.xa, r.xb, r.ya, r.yb))
	pageSize := r.xb - r.xa
	lineLength := 0
	currentLine := ""
	x, y := 0, 0
	for ci, c := range r.Content {
		// r.Logger.Log(fmt.Sprintf("Bytes: %v, %q - CI: %d - lineLength%%pageSize==%d - pageSize: %d - lineLength: %d", c, c, ci, lineLength%pageSize, pageSize, lineLength))
		if c == 10 {
			currentLine = ""
			lines += 1
			y += 1
			x = 0
			lineLength = 0
			moveCursor(r.xa, r.ya+lines)
			// t.Logger.Log(fmt.Sprintf("Moved cursor to (%d,%d)", xa, ya+lines))
			r.fileCoords[ci] = [2]int{x, y}
			r.cursorCoords[[2]int{x, y}] = ci
			continue
		} else {
			currentLine = fmt.Sprintf("%s%s", currentLine, string(c))
			lineLength += 1
			// ' '
			if c == 32 {
				wordSizeLookAhead := getWordLength(r.Content, ci+1)
				// t.Logger.Log(fmt.Sprintf("word lookahead = %s (indexes %d, %d)", t.fileContents[ci:ci+wordSizeLookAhead+1], ci, ci+wordSizeLookAhead))
				if (lineLength + wordSizeLookAhead) >= pageSize {
					// fmt.Print(string(c))
					currentLine = ""
					lines += 1
					y += 1
					x = 0
					lineLength = 0
					moveCursor(r.xa, r.ya+lines)
				} else {
					// if x == r.xb-10 {
					// 	fmt.Print("X")
					// 	continue
					// } else {
					// 	fmt.Print(string(c))
					// }
					fmt.Print(string(c))
				}
			} else {
				fmt.Print(string(c))
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
		r.fileCoords[ci] = [2]int{x, y}
		r.cursorCoords[[2]int{x, y}] = ci
		x += 1
		// t.Logger.Log(fmt.Sprintf("Character Coords for %s: %v", string(c), t.characterCoordMap[ci]))
	}
}

func (r *Region) updateCursorLocation(key byte) {
	currentCursorCoords := r.fileCoords[r.cursorPosition]
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

			if newy < 0 {
				return
			}
			if newx < 0 {
				return
			}
			newPosition := [2]int{newx, newy}
			newCoords, exists := r.cursorCoords[newPosition]
			r.Logger.Log(fmt.Sprintf("newCoords %v newPosition %v exists %v", newCoords, newPosition, exists))
			if !exists {
				xmod += 1
				continue
			} else {
				r.cursorPosition = newCoords
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
			newCoords, exists := r.cursorCoords[newPosition]
			if !exists {
				xmod += 1
				continue
			} else {
				r.cursorPosition = newCoords
				foundPosition = true
			}
		}
	case 67:
		// RIGHT
		if r.cursorPosition == len(r.Content) {
			r.Logger.Log("Hit the end of the file contents")
			return
		}
		newPosition := [2]int{currentCursorCoords[0] + 1, currentCursorCoords[1]}
		newCoords, exists := r.cursorCoords[newPosition]
		if !exists {
			newPosition := [2]int{0, currentCursorCoords[1] + 1}
			newCoords, exists = r.cursorCoords[newPosition]
			if !exists {
				return
			} else {
				r.cursorPosition = newCoords
			}
		} else {
			r.cursorPosition = newCoords
		}
	case 68:
		// LEFT
		if r.cursorPosition == 0 {
			return
		}

		foundPosition := false
		x := currentCursorCoords[0] - 1
		y := currentCursorCoords[1]

		for !foundPosition {
			newPosition := [2]int{x, y}
			newCoords, exists := r.cursorCoords[newPosition]
			r.Logger.Log(fmt.Sprintf("New Coords: %v, exists: %v, newPosition: %v, x: %v, y: %v", newCoords, exists, newPosition, x, y))
			if !exists {
				if x <= 0 {
					if y <= 0 {
						return
					} else {
						x = r.xb
						y -= 1
					}
				}
				x -= 1
			} else {
				r.cursorPosition = newCoords
				foundPosition = true
			}
		}
	default:
		return
	}
}

func (r *Region) update(buffer []byte) {
	shouldRedraw := false
	switch key := buffer[0]; key {
	case 27:
		// ESC
		if buffer[2] >= 65 && buffer[2] <= 68 {
			r.updateCursorLocation(buffer[2])
		}
	case 126:
		r.Content = r.Content[:max(0, r.cursorPosition)] + r.Content[r.cursorPosition+1:]
		shouldRedraw = true
	case 127:
		if (r.cursorPosition - 1) < 0 {
			return
		}
		r.Content = r.Content[:r.cursorPosition-1] + r.Content[r.cursorPosition:]
		r.cursorPosition -= 1
		shouldRedraw = true
	case 13:
		r.Content = r.Content[:r.cursorPosition] + "\r\n" + r.Content[r.cursorPosition:]
		shouldRedraw = true
	default:
		r.Content = r.Content[:r.cursorPosition] + fmt.Sprintf("%c", key) + r.Content[r.cursorPosition:]
		shouldRedraw = true
	}
	if shouldRedraw {
		r.Redraw()
		// r.drawRegionWindow()
	}
	// flushRegion(r.xa, r.xb, r.ya, r.yb)
	// r.Redraw()
	moveCursor(r.fileCoords[r.cursorPosition][0]+r.xa, r.fileCoords[r.cursorPosition][1]+r.ya)
}
