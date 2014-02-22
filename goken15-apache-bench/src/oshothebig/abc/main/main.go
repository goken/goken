package main

import (
	"flag"
	"fmt"
	"os"

	"../"
)

func main() {
	var params abc.Params
	flag.IntVar(&params.Concurrency, "c", 1, "concurrency")
	flag.IntVar(&params.Requests, "n", 1, "requests")
	flag.Parse()

	if len(flag.Args()) < 1 {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		return
	}
	params.Url = flag.Arg(0)

	benchmark := abc.NewBenchmark(params)
	benchmark.Run()
}
