package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"sync"
)

var wg sync.WaitGroup

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	// compile regexp
	re := regexp.MustCompile(os.Args[1])

	// open file
	file, _ := os.Open(os.Args[2])

	// scan word
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		wg.Add(1)
		go func(str string) {
			defer wg.Done()
			if re.MatchString(str) {
				fmt.Println(str)
			}
		}(scanner.Text())
	}
	wg.Wait()
}
