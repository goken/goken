package main

func a() {
	b()
}

func b() {
	c()
}

func c() {
}

func main() {
	a()
	c()
}
