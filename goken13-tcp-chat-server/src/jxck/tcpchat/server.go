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
	port          string
	clients       []*Client
	broadcastChan chan string
}

func NewServer(port string) *Server {
	return &Server{
		port:          port,
		clients:       make([]*Client, 0, 10),
		broadcastChan: make(chan string),
	}
}

// サーバをスタートし、 BroadCast とクライアント管理用の
// goroutine を開始する
func (s *Server) ListenAndServe() (err error) {
	listener, err := net.Listen("tcp", s.port)
	if err != nil {
		return err
	}
	fmt.Printf("server starts at %v\n", s.port)

	go s.BroadCast()
	dissconnect := s.RemoveClient()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		fmt.Println("new connection")
		client := NewClient(conn, s.broadcastChan)
		s.AddClient(client)
		go client.Handle(dissconnect)
	}
	return nil
}

// 接続してきたクライアントの追加
func (s *Server) AddClient(c *Client) {
	s.clients = append(s.clients, c)
	fmt.Printf("add client (%d client connecting)\n", len(s.clients))
}

// 切断したクライアントの削除
func (s *Server) RemoveClient() (dissconnect chan *Client) {
	dissconnect = make(chan *Client)
	go func() {
		for {
			client := <-dissconnect
			for i, v := range s.clients {
				if v == client {
					err := client.Close()
					if err != nil {
						log.Println(err)
					}
					copy(s.clients[i:], s.clients[i+1:])
					s.clients[len(s.clients)-1] = nil // GC
					s.clients = s.clients[:len(s.clients)-1]
				}
			}
			fmt.Printf("remove client (%d client connecting)\n", len(s.clients))
		}
	}()
	return dissconnect
}

// 全てのクライアントに送信
func (s *Server) BroadCast() {
	for {
		message := <-s.broadcastChan
		go func() {
			for _, client := range s.clients {
				err := client.Send(message)
				if err != nil {
					log.Println(err)
				}
			}
		}()
	}
}

func ListenAndServe(port string) error {
	return NewServer(":" + port).ListenAndServe()
}
