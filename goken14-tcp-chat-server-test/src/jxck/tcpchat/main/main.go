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
	server.Serve(":3000")
}
