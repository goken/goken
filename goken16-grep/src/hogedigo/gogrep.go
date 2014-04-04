package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

const USAGE = "usage: gogrep <word> <file>"

/*
	goken課題：grep clone
	usage: gogrep <word> <file>

	素直に書いてみた
*/
func main() {

	if len(os.Args) < 3 {
		log.Fatalf(USAGE)
	}

	word := os.Args[1]

	file, err := os.Open(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(file)
	lineNo := 0
	for scanner.Scan() {
		line := scanner.Text()
		lineNo++
		if strings.Contains(line, word) {
			fmt.Printf("%d: %s\n", lineNo, line)
		}
	}
}
