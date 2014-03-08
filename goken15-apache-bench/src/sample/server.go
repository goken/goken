package main

import (
	"log"
	"net/http"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

func main() {
	log.Fatal(http.ListenAndServe(":3000", http.FileServer(http.Dir("."))))
}
