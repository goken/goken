package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"runtime"
	"sync"
)

/*
ABOUT
- 1行につき1つのgoroutineでマッチするかどうかを調べる

env
- mba corei5 1.7GB Mem:4GB

result
- defer使ってwg.Done():  ./test.sh  4.16s user 4.61s system 157% cpu 5.577 total
- 関数の最後でwg.Done(): ./test.sh  3.14s user 3.82s system 139% cpu 4.997 total
- 外部関数化:            ./test.sh  3.06s user 3.71s system 138% cpu 4.898 total

reference
- http://stackoverflow.com/questions/8757389/reading-file-line-by-line-in-go
- http://golang.org/pkg/regexp/

*/

var wg sync.WaitGroup

func Match(pat *regexp.Regexp, str string) {
	wg.Add(1)
	if pat.MatchString(str) {
		fmt.Println(str)
	}
	wg.Done()
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	// compile
	pattern := regexp.MustCompile(os.Args[1])

	// file open
	file, err := os.Open(os.Args[2])
	if err != nil {
		log.Fatal(err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		go Match(pattern, scanner.Text())
		// 		wg.Add(1)
		// 		go func(str string){
		// //			defer wg.Done()
		// 			if pattern.MatchString(str){
		// 				fmt.Println(str)
		// 			}
		// 			wg.Done()
		// 		}(scanner.Text())
	}
	wg.Wait()
}
