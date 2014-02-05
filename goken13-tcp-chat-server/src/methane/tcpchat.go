package main

import (
	"fmt"
	"log"
	"net"
	"sync"
)

const BUFSIZE = 64

type Message struct {
	sender  string
	message string
}

type Client struct {
	Name string
	ch   chan string
}

type Room struct {
	sync.Mutex
	clients map[string]*Client
	recvCh  chan Message
}

func NewRoom() *Room {
	room := &Room{
		clients: make(map[string]*Client),
		recvCh:  make(chan Message)}

	go func() { // broadcast
		for msg := range room.recvCh {
			fmt.Println(msg)
			text := fmt.Sprintf("%s: %s\n", msg.sender, msg.message)
			room.Lock()
			for name, c := range room.clients {
				if name == msg.sender {
					continue
				}
				fmt.Println("Sending to", c)
				select {
				case c.ch <- text:
				default:
				}
			}
			room.Unlock()
		}
	}()

	return room
}

func (room *Room) Join(conn net.Conn) {
	room.Lock()
	defer room.Unlock()
	client := NewClient(conn, room)
	room.clients[client.Name] = client
}

func (room *Room) Apart(name string) {
	room.Lock()
	defer room.Unlock()
	client, ok := room.clients[name]
	if ok {
		delete(room.clients, name)
		close(client.ch)
	}
}

func NewClient(c net.Conn, room *Room) *Client {
	client := &Client{
		Name: c.RemoteAddr().String(),
		ch:   make(chan string, BUFSIZE)}

	go func() {
		defer c.Close()
		for {
			msg, ok := <-client.ch
			if !ok {
				return
			}
			_, err := c.Write([]byte(msg))
			if err != nil {
				return
			}
		}
	}()

	go func() {
		buf := make([]byte, 4096)
		for {
			size, err := c.Read(buf)
			if err != nil {
				room.Apart(client.Name)
				break
			}
			room.recvCh <- Message{client.Name, string(buf[:size])}
		}
	}()

	return client
}

func (client *Client) Send(msg string) bool {
	select {
	case client.ch <- msg:
		return true
	default:
		close(client.ch)
		return false
	}
}

func main() {
	room := NewRoom()

	listener, err := net.Listen("tcp", ":5012")
	if err != nil {
		log.Panic(err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Panic(err)
		}
		room.Join(conn)
	}
}
