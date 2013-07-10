package main

import (
	"fmt"
	"log"
	"runtime"
	"time"
)

// server will panic after timeout second
func halt(timeout int64) {
	time.AfterFunc(time.Duration(timeout)*time.Second, func() {
		panic("fin")
	})
}

func Get() string {
	time.Sleep(getInterval)
	return "hello"
}

type Pooling struct {
	result chan string
	fin    chan bool
}

var (
	getInterval       time.Duration = 2 * time.Second
	subscribeInterval               = 1 * time.Second
	publishInterval                 = 200 * time.Millisecond
)

func (p *Pooling) Loop() {
	for {
		log.Println("before select", len(p.result), runtime.NumGoroutine())
		interval := time.After(publishInterval)

		select {
		case <-interval:
			if len(p.result) < cap(p.result) {
				go func() {
					p.result <- Get()
					log.Println("   after read", len(p.result), runtime.NumGoroutine())
				}()
			}
		case <-p.fin:
			log.Println("close")
			return
		}
	}
}

func (p *Pooling) Close() {
	close(p.fin)
}

func main() {

	//go halt(10)

	pooling := &Pooling{
		result: make(chan string, 10),
		fin:    make(chan bool, 1),
	}
	go pooling.Loop()

	var i = time.Tick(subscribeInterval)
	var fin = time.After(4 * time.Second)
	for {
		select {
		case <-i:
			log.Println("main loop")

			fmt.Println(<-pooling.result)
			fmt.Println("=====================================")
		case <-fin:
			log.Println("close loop")
			pooling.Close()
			return
		}
	}
}
