package tcpchat

import (
	"fmt"
	"log"
	"net"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

type Server struct {
	clients []*Client
}

func NewServer() *Server {
	server := &Server{
		clients: make([]*Client, 0),
	}
	return server
}

func (s *Server) Listen(port string) {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("server starts at %v\n", port)

	accept := AcceptLoop(listener)
	broadcast := make(chan string)
	for {
		select {
		case client := <-accept:
			s.clients = append(s.clients, client)
			go client.ReadLoop(broadcast)
			go client.WriteLoop()
		case message := <-broadcast:
			go s.BroadCast(message)
		default:
		}
	}
}

func AcceptLoop(listener net.Listener) chan *Client {
	accept := make(chan *Client)
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Fatal(err)
			}
			client := NewClient(conn)
			accept <- client
		}
	}()
	return accept
}

func (s *Server) BroadCast(message string) {
	for _, client := range s.clients {
		select {
		case client.WriteBuf <- message:
		default:
		}
	}
}
