package main

func main() {
	ch := make(chan bool)
	go func() {
		ch <- true
	}()

	<-ch
}
