package main

type Foo interface {
	DoFoo()
}

type Bar struct {
}

func (bar *Bar) DoFoo() {
}

func do(foo Foo) {
	foo.DoFoo()
}

func main() {
	do(&Bar{})
}
