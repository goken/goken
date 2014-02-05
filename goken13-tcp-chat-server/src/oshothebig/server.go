package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("Usage: %s <port>\n", os.Args[0])
	}

	port := os.Args[1]
	server := NewServer(port)
	server.ListenAndServe()
}

type message string

type Server struct {
	port         string
	clients      []*Client
	addClient    chan *Client
	removeClient chan *Client
	broadcast    chan message
}

func NewServer(port string) *Server {
	return &Server{
		port,
		make([]*Client, 0),
		make(chan *Client),
		make(chan *Client),
		make(chan message),
	}
}

func (s *Server) ListenAndServe() error {
	l, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		return err
	}
	defer l.Close()

	go s.Serve()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("Error: %s\n", err)
		}
		c := NewClient(conn, s)

		s.addClient <- c

		go c.Read()
		go c.Write()
	}

}

func (s *Server) Serve() {
	for {
		select {
		case c := <-s.addClient:
			log.Println("Join: " + c.conn.RemoteAddr().String())
			s.clients = append(s.clients, c)
		case c := <-s.removeClient:
			for i := range s.clients {
				if s.clients[i] == c {
					s.clients = append(s.clients[:i], s.clients[i+1:]...)
					log.Println("Leave: " + c.conn.RemoteAddr().String())
					break
				}
			}
		case m := <-s.broadcast:
			for _, c := range s.clients {
				log.Printf("Broadcast (%s): \"%s\"\n", c.conn.RemoteAddr(), m)
				c.channel <- m
			}
		}
	}
}

type Client struct {
	conn    net.Conn
	server  *Server
	channel chan message
	done    chan bool
}

func NewClient(conn net.Conn, server *Server) *Client {
	return &Client{
		conn,
		server,
		make(chan message),
		make(chan bool),
	}
}

func (c *Client) Read() {
	scanner := bufio.NewScanner(c.conn)
LOOP:
	for scanner.Scan() {
		select {
		case <-c.done:
			break LOOP
		default:
			m := scanner.Text()
			log.Printf("Receive (%s): \"%s\"\n", c.conn.RemoteAddr(), m)
			c.server.broadcast <- message(m)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error (read): %s\n", err)
	}
	c.server.removeClient <- c
	c.done <- true
}

func (c *Client) Write() {
LOOP:
	for {
		select {
		case <-c.done:
			break LOOP
		case m := <-c.channel:
			log.Printf("Send (%s): \"%s\"\n", c.conn.RemoteAddr(), m)
			_, err := fmt.Fprintln(c.conn, m)
			if err != nil {
				log.Printf("Error (write): %s\n", err)
				break LOOP
			}
		}
	}

	c.server.removeClient <- c
	c.done <- true
	return
}
