package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
)

type Server struct {
	sync.RWMutex
	host    string
	clients []*client
}

func NewServer(host string) *Server {
	return &Server{
		host:    host,
		clients: []*client{},
	}
}

func (s *Server) ListenAndServe() error {
	ln, err := net.Listen("tcp", s.host)
	if err != nil {
		return err
	}

	log.Println("listen at", s.host)

	for id := 0; ; id++ {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			continue
		}

		c := newClient(s, id, conn)
		s.add(c)
		go c.handle()
	}

	return nil
}

func (s *Server) add(c *client) {
	s.Lock()
	defer s.Unlock()
	s.clients = append(s.clients, c)
	log.Printf("add new client(id=%d)\n", c.id)
}

func (s *Server) remove(c *client) {
	s.Lock()
	defer s.Unlock()
	for i, _ := range s.clients {
		if s.clients[i] == c {
			s.clients = s.clients[:i+copy(s.clients[i:], s.clients[i+1:])]
			log.Printf("remove client(id=%d)\n", c.id)
			break
		}
	}
}

func (s *Server) broadcast(msg []byte) {
	log.Printf("broadcast '%s'\n", string(msg))
	s.RLock()
	for _, c := range s.clients {
		if !c.send(msg) {
			s.RUnlock()
			c.server.remove(c)
			return
		}
	}
	s.RUnlock()
}

type client struct {
	id     int
	server *Server
	conn   net.Conn
	ch     chan []byte
}

func newClient(server *Server, id int, conn net.Conn) *client {
	return &client{
		server: server,
		id:     id,
		conn:   conn,
		ch:     make(chan []byte),
	}
}

func (c *client) handle() {
	defer func() {
		c.server.remove(c)
		c.conn.Close()
	}()

	done := make(chan bool)

	go func() {
		for {
			msg := <-c.ch
			log.Printf("write '%s' to id=%d\n", string(msg), c.id)
			_, err := c.conn.Write(append(msg, '\n'))
			if err != nil {
				done <- true
				return
			}
		}
	}()

	scanner := bufio.NewScanner(c.conn)
	for {
		select {
		case <-done:
			return
		default:
			if !scanner.Scan() || scanner.Err() != nil {
				return
			}

			msg := scanner.Bytes()
			log.Printf("read '%s' from id=%d\n", string(msg), c.id)
			c.server.broadcast(msg)
		}
	}
}

func (c *client) send(msg []byte) bool {
	log.Printf("send to id=%d\n", c.id)
	select {
	case c.ch <- msg:
		return true
	default:
		log.Printf("send fail (id=%d)\n", c.id)
		return false
	}
}

func main() {
	port := "3000"
	if len(os.Args) > 2 {
		port = os.Args[1]
	}

	if err := NewServer(":" + port).ListenAndServe(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}
}
