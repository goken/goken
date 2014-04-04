package main

import (
	"log"
	"os"
	"io"
	"fmt"
	"bytes"
)

const USAGE = "usage: gogrep <word> <file>"

/*
	goken課題：grep clone
	usage: gogrep <word> <file>

	自前でアルゴリズム書いてみた
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

	buf := make([]byte, 4096)
	lineNo := 1
	wordIdx := 0
	hit := false

	lineBuf := bytes.NewBuffer(make([]byte, 4096))
	lineBuf.Reset()

	for {
		size, err := file.Read(buf)
		if err == io.EOF {
			if hit {
				fmt.Printf("%d: %s\n", lineNo, lineBuf)
			}
			return;
		} else if err != nil {
			log.Fatal(err)
		}

		copyFrom := 0
		for i := 0; i < size; i++ {
			if buf[i] == '\n' {
				lineBuf.Write(buf[copyFrom:i])
				if hit {
					fmt.Printf("%d: %s\n", lineNo, lineBuf)
				}

				wordIdx = 0
				lineNo++
				hit = false
				copyFrom = i + 1
				lineBuf.Reset()
				continue
			}

			if hit {
				continue
			}

			if buf[i] == word[wordIdx] {
				wordIdx++
				if wordIdx == len(word) {
					hit = true
				}
			} else {
				wordIdx = 0
			}
		}
		if copyFrom < size {
			lineBuf.Write(buf[copyFrom:size])
		}
	}
}
