package main

import (
	"bufio"
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
	name   string
	conn   net.Conn
	reader *bufio.Reader
	ch     chan string
}

type Room struct {
	sync.Mutex
	clients map[string]*Client
	recvCh  chan *Message
}

func NewRoom() *Room {
	room := &Room{
		clients: make(map[string]*Client),
		recvCh:  make(chan *Message)}

	go func() { // broadcast
		for {
			msg := <-room.recvCh
			text := fmt.Sprintf("%s: %s", msg.sender, msg.message)
			fmt.Println(text)
			room.Lock()
			for _, c := range room.clients {
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
	client := NewClient(conn, room)
	room.Lock()
	room.clients[client.name] = client
	room.Unlock()
	fmt.Println("joined: ", client.name)
	client.Start(room)
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
	reader := bufio.NewReader(c)
	c.Write([]byte("Your name?> "))
	name, err := reader.ReadString('\n')
	if err != nil {
		return nil
	}
	name = name[:len(name)-1]
	return &Client{
		name:   name,
		conn:   c,
		reader: reader,
		ch:     make(chan string, BUFSIZE)}
}

func (client *Client) Start(room *Room) {
	go func() { // sender
		defer func() {
			if r := recover(); r != nil {
				fmt.Println(r)
			}
		}()
		defer client.conn.Close()
		for {
			msg, ok := <-client.ch
			if !ok {
				return
			}
			_, err := client.conn.Write([]byte(msg))
			if err != nil {
				return // todo どうやって receiver 止める？
			}
		}
	}()

	go func() { // receiver
		defer func() {
			if r := recover(); r != nil {
				fmt.Println(r)
			}
		}()
		for {
			read, err := client.reader.ReadString('\n')
			if err != nil {
				room.Apart(client.name)
				break
			}
			fmt.Println("name: ", client.name, " message: ", read)
			room.recvCh <- &Message{sender: client.name, message: read}
		}
	}()
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
		go room.Join(conn)
	}
}
