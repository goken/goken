package main

type Foo interface {
	Do()
}

type Bar struct {
}

func (bar *Bar) Do() {
}

type Hoge struct {
}

func (hoge *Hoge) Do() {
}

func main() {
}
