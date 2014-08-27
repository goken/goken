package main

import (
	"io"
	"log"
	"net/http"
	"os"
)

func downloadImgs(imgch <-chan string, numWorker int) {
	for i := 0; i < numWorker; i++ {
		go func() {
			for imgURL := range imgch {
				resp, err := http.Get(imgURL)
				if err != nil {
					log.Println("Error:" + err.Error())
					continue
				}
				file, err := os.Create(string.Split(imgURL, "/")[0])
				io.Copy(file, resp.Body)
				file.Close()
				resp.Body.Close()
			}
		}()
	}
}

func main() {
	imgch := make(chan string)
	downloadImgs(imgch, 5)
}
