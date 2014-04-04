package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	flag.Parse()
	if flag.NArg() != 2 {
		log.Fatal("Usage: goken-grep pattern file")
	}
	pattern := flag.Arg(0)
	filename := flag.Arg(1)
	searchAndPrint(filename, pattern)
}

func searchAndPrint(filename, pattern string) {
	var file *os.File
	var err error
	if file, err = os.Open(filename); err != nil {
		log.Fatal("Error in opening file: %s, error: %v", filename, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, pattern) {
			fmt.Println(line)
		}
	}
}
