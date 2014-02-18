package main

import (
	"jxck/tcpchat"
	"log"
	"os"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

func main() {
	port := ":3000"
	if len(os.Args) > 2 {
		port = ":" + os.Args[1]
	}

	server := tcpchat.NewServer()
	server.Listen(port)
}
