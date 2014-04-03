package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"sync"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	re := regexp.MustCompile(os.Args[1])
	f, _ := os.Open(os.Args[2])
	var wg sync.WaitGroup

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		wg.Add(1)
		str := scanner.Text()
		go func() {
			defer wg.Done()
			if re.MatchString(str) {
				fmt.Println(str)
			}
		}()
	}
	wg.Wait()
}
