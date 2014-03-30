package main

import (
 "os"
 "io"
 "fmt"
 "bufio"
 "strings"
)

func scanLine(line string, word string) int {
 index := strings.Index(line, word)
 return index
}

// http://tkotobu.cocolog-nifty.com/blog/2011/09/golang-7819.html

func main() {
 // --- check args ---
 argCount := len(os.Args)
 if argCount < 3 {
  fmt.Println("usege:   grep word filename") 
  return
 }

 // -- read file --
 var err error
 var inFile *os.File
 var reader *bufio.Reader
 var flag bool

 word := os.Args[1]
 filename := os.Args[2]
 inFile,err = os.Open(filename)
 if err != nil {
  fmt.Printf("Cannot open file: %s\n", filename)
  return
 }
 reader = bufio.NewReaderSize(inFile, 4096)

 var line []byte
 for {
  line,flag,err = reader.ReadLine()
  if err == io.EOF {
   return
  }
  if err != nil {
   fmt.Println("File read error")
   return
  }
  if flag {
   fmt.Println("buffer error")
   return
  }

  //fmt.Println(string(line))
  if scanLine(string(line), word) >= 0 {
   fmt.Println(string(line))
  }
 }
}

