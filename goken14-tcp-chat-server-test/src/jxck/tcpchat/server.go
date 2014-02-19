package chat

import (
	"fmt"
	"log"
	"net"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

type Server struct {
	Listener      net.Listener
	Clients       []*Client
	BroadCastChan chan string
	LeaveChan     chan *Client
	NextId        chan int
}

func NewServer() *Server {
	server := &Server{
		Clients:       make([]*Client, 0),
		BroadCastChan: make(chan string),
		LeaveChan:     make(chan *Client),
		NextId:        NextId(),
	}
	go server.BroadCastLoop()
	go server.LeaveLoop()
	return server
}

func (s *Server) Join(client *Client) {
	s.Clients = append(s.Clients, client)
	fmt.Printf("join new client: %v, totla: %d\n", client.Id, len(s.Clients))
}

func (s *Server) Listen(port string) {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal(err)
	}
	s.Listener = listener
	fmt.Printf("server starts at port %s\n", port)
}

func (s *Server) Accept() error {
	conn, err := s.Listener.Accept()
	if err != nil {
		return err
	}
	id := <-s.NextId
	client := NewClient(id, conn, s.BroadCastChan, s.LeaveChan)
	s.Join(client)
	return nil
}

func (s *Server) AcceptLoop() {
	fmt.Printf("start accept loop\n")
	for {
		err := s.Accept()
		if err != nil {
			if err.Error() == "use of closed network connection" {
				break
			}
			continue
		}
	}
}

func (s *Server) ListenAndAcceptLoop(port string) {
	s.Listen(port)
	s.AcceptLoop()
}

func (s *Server) BroadCast(message string) {
	fmt.Printf("broadcast: %q\n", message)
	for _, client := range s.Clients {
		client.WriteChan <- message
	}
}

func (s *Server) BroadCastLoop() {
	fmt.Printf("start broadcast loop\n")
	for message := range s.BroadCastChan {
		s.BroadCast(message)
	}
}

func (s *Server) Leave(client *Client) {
	fmt.Printf("leave: %v\n", client.Id)
	for i, c := range s.Clients {
		if c == client {
			copy(s.Clients[i:], s.Clients[i+1:])
			length := len(s.Clients) - 1
			s.Clients[length] = nil
			s.Clients = s.Clients[:length]
		}
	}
	fmt.Printf("clients %+v\n", s.Clients)
	fmt.Printf("joining client %d\n", len(s.Clients))
}

func (s *Server) LeaveLoop() {
	fmt.Printf("start leave loop\n")
	for client := range s.LeaveChan {
		s.Leave(client)
	}
}

func NextId() chan int {
	nextid := make(chan int)
	id := 0
	go func() {
		for {
			nextid <- id
			id = id + 1
		}
	}()
	return nextid
}

func (s *Server) Close() error {
	defer func() {
		err := s.Listener.Close()
		fmt.Println("close listener")
		if err != nil {
			log.Println(err)
		}
	}()
	for _, c := range s.Clients {
		err := c.Conn.Close()
		if err != nil {
			return err
		}
		fmt.Printf("close connection %v\n", c)
	}
	return nil
}
