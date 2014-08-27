package main

import (
	"bufio"
	"io"
	"log"
	"net/http"
	"os"
)

func downloadImgs(numWorker int) chan<- string {
	imgch := make(chan string)
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

	return imgch
}

func parseCSS(imgch chan<- string, numWorker int) chan<- string {
	cssch := make(chan string)
	for i := 0; i < numWorker; i++ {
		go func() {
			for cssURL := range cssch {
				resp, err := http.Get(cssURL)
				if err != nil {
					log.Println("Error:" + err.Error())
					continue
				}
				scanner := bufio.Scanner(resp.Body)
				for scanner.Scan() {
					// パースする
				}
				if err := scanner.Err(); err != nil {
					log.Println("Error:" + err.Error())
				}
				resp.Body.Close()
			}
		}()
	}

	return cssch
}

func parseHTML(imgch, cssch chan<- string, numWorker int) chan<- string {
	htmlch := make(chan string)
	for i := 0; i < numWorker; i++ {
		go func() {
			for htmlURL := range htmlch {
				resp, err := http.Get(htmlURL)
				if err != nil {
					log.Println("Error:" + err.Error())
					continue
				}
				scanner := bufio.Scanner(resp.Body)
				for scanner.Scan() {
					// パースする
				}
				if err := scanner.Err(); err != nil {
					log.Println("Error:" + err.Error())
				}
				resp.Body.Close()
			}
		}()
	}

	return htmlch
}

func main() {
	imgch := downloadImgs(5)
	cssch := parseCSS(imgch, 5)
	htmlch := parseHTML(imgch, cssch, 5)
	htmlch <- "http://golang.org"
}
