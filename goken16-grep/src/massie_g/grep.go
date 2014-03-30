package main

import (
    //"strings"
    "os"
    "io"
    "fmt"
    "bufio"
)

func scanLine(line string, word string) int {
    return 0
}

// http://tkotobu.cocolog-nifty.com/blog/2011/09/golang-7819.html

func main() {
 // --- trial ----
 //fmt.Printf("args count=%d\n", os.Args.length); //NG
 //fmt.Printf("args count=%d\n", len(os.Args)); // OK
 //fmt.Println(os.Args[1]) // OK

 // --- check args ---
 argCount := len(os.Args)
 if argCount < 3 {
  fmt.Println("usege:   grep word filename") 
  return
 }

 fmt.Println(os.Args[1])
 fmt.Println(os.Args[2])

 // -- read file --
 //var err os.Error
 //var inFile *os.File
 var reader *bufio.Reader
 var flag bool

 inFile,err := os.Open(os.Args[2])
 if err != nil {
  fmt.Printf("Cannot open file: %s\n", os.Args[2])
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

  fmt.Println(string(line))
 }
}

