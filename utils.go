package main

import "fmt"

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

func moveCursor(x int, y int) {
	// ESC [ LINE ; COL H
	fmt.Printf("\033[%d;%dH", y+1, x)
}
