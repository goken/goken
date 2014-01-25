package main

import (
	chat "jxck/tcpchat"
	"log"
	"os"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("port missing")
	}
	port := os.Args[1]
	log.Println(chat.ListenAndServe(port))
}
