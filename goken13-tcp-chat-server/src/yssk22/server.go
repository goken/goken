package main

import (
	"fmt"
	"net"
	"os"
	"path"
	"strconv"
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
}

func NewServer(port int) *Server {
	s := new(Server)
	s.Port = port
	return s
}

func (s *Server) Run() {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Port))
	if err != nil {
		panic(err.Error())
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			// error handling
			fmt.Println("%v\n")
			continue
		}
		go func() { s.HandleConnection(conn) }()
	}
}

func (s *Server) HandleConnection(c net.Conn) {
	fmt.Printf("Connected: %v\n", c.RemoteAddr().String())
	c.Close()
}
