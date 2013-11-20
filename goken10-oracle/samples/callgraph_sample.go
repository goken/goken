package main

import (
	"fmt"
)

func a() {
	b()
}

func b() {
	c()
}

func c() {
	d()
}

func d() {
	fmt.Println("d")
}

func main() {
	a()
}
