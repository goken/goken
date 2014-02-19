package main

import (
	"log"
	"../"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

func main() {
	server := chat.NewServer()
	server.ListenAndAcceptLoop(":3000")
}
