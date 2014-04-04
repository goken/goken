package main

// CR, CRLF, LF のファイルに対応
// bufioでバッファリングしながら読み込み
// ファイルを全部読むまで待たないので大きいファイルでも大丈夫
// 単語の検索はstrings.Contains

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func readFile(filename, word string) {
	var err error
	var file *os.File
	if file, err = os.Open(filename); err != nil {
		log.Fatal(err)
		return
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	for {
		line, err := readLine(reader)
		if err != nil {
			if err != io.EOF {
				log.Fatal(err)
				return
			}
			break
		} else {
			if strings.Contains(line, word) {
				fmt.Println(line)
			}
		}
	}
}

func readLine(r *bufio.Reader) (string, error) {
	buffer := bytes.NewBuffer(make([]byte, 0))
	for {
		r1, err := readRune(r)
		if err != nil {
			if err == io.EOF {
				return buffer.String(), io.EOF
			}
			return "", err
		} else if r1 == '\n' {
			return buffer.String(), nil
		}
		buffer.WriteRune(r1)
	}
	panic("unreachable")
}

func readRune(r *bufio.Reader) (rune, error) {
	r1, _, err := r.ReadRune()
	if r1 == '\r' {
		r1, _, err = r.ReadRune()
		if err == nil {
			if r1 != '\n' {
				r.UnreadRune()
				r1 = '\r'
			}
		}
	}
	return r1, err
}

func main() {
	flag.Parse()
	if flag.NArg() != 2 {
		log.Fatal("Usage: goken-grep word filename")
		return
	}
	word := flag.Arg(0)
	filename := flag.Arg(1)
	readFile(filename, word)
}
