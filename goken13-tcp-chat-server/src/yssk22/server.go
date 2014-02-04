package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path"
	"strconv"
	"bufio"
	"bytes"
)

func main() {
	var port int = 3000
	var e error

	args := os.Args
	if len(args) == 2 {
		port, e = strconv.Atoi(args[1])
		if e != nil {
			fmt.Printf("Invalid port number\n")
			help()
			os.Exit(1)
		}
	} else {
		if len(args) > 2 {
			fmt.Printf("Too many arguments\n")
			help()
			os.Exit(1)
		}
	}
	NewServer(port).Run()
}

func help() {
	progname := path.Base(os.Args[0])
	fmt.Printf("Usage: %v [port]\n", progname)
}

type Server struct {
	Port int
	receiver chan *Message
	clients  map[string](chan *Message)
}

type Message struct {
	Body []byte
	From string
	// Don't care since all messages are broad cast.
	// To   string
}

var CMD_QUIT = []byte("\\q")

func NewServer(port int) *Server {
	s := new(Server)
	s.Port = port
	s.receiver = make(chan *Message)
	s.clients = make(map[string] chan *Message)
	return s
}

func (s *Server) Run() {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Port))
	if err != nil {
		panic(err.Error())
	}
	// Handle messages from clients message
	go func(){
		for msg := range s.receiver {
			for _, ch := range s.clients {
				// TODO: race condition
				// There would be a change to write on closed channel..
				ch <- msg
			}
		}
	}()


	Debug("Server started at %d", s.Port)
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go func() {
			// TODO: Should use more reliable unique id.
			// otherwise it would break s.clients under race condition.
			remote := conn.RemoteAddr().String()
			ch := make(chan *Message)
			s.clients[remote] = ch

			Debug("[%s] Connected.", remote)
			defer func(){
				close(ch)
			}()
			defer func(){
				delete(s.clients, remote)
			}()
			defer func() {
				Debug("[%s] Closing connection....", remote)
				conn.Close()
			}()


			// a goroutin which receive a message from server and
			// write it into connection.
			go func(){
				for msg := range ch {
					// TODO: error handling
					conn.Write(msg.Body)
				}
			}()

			// loop for receiving messages from clients.
			reader := bufio.NewReader(conn)
			for {
				line, _, err := reader.ReadLine()
				if err != nil  {
					if err == io.EOF {
						break
					}else{
						Debug("[%s] Error: %v", remote, err)
					}
				}

				if bytes.HasPrefix(line, CMD_QUIT) {
					Debug("[%s] Quiting...", remote)
					break
				}
				Debug("[%s] Received: %s", remote, line)
				s.receiver <- &Message{
					Body: append(line, '\n'),
					From: remote,
				}
			}
			// no more packets would come!
		}()
	}
}

func Debug(s string, args ...interface{}) {
	log.Printf(s, args...)
}
