package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

func main() {
	port := ":" + os.Args[1]
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("server starts at %v\n", port)

	accept := AcceptLoop(listener)
	connections := make([]net.Conn, 0)
	broadcast := make(chan string)
	for {
		select {
		case conn := <-accept:
			go ReadLoop(conn, broadcast)
			connections = append(connections, conn)
		case message := <-broadcast:
			go BroadCast(connections, message)
		default:
		}
	}
}

func AcceptLoop(listener net.Listener) chan net.Conn {
	accept := make(chan net.Conn)
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Fatal(err)
			}
			accept <- conn
		}
	}()
	return accept
}

func ReadLoop(conn net.Conn, broadcast chan string) {
	fmt.Printf("connect %v\n", conn)
	br := bufio.NewReader(conn)
	for {
		line, _, err := br.ReadLine()
		if err != nil {
			if err == io.EOF {
				fmt.Printf("dissconnect %v\n", conn)
			} else {
				log.Println(err)
			}
			return
		}
		message := string(line) + "\n"
		log.Printf("%q\n", message)
		broadcast <- string(message)
	}
}

func BroadCast(connections []net.Conn, message string) {
	for _, conn := range connections {
		go func(conn net.Conn) {
			bw := bufio.NewWriter(conn)
			bw.WriteString(message)
			bw.Flush()
		}(conn)
	}
}
