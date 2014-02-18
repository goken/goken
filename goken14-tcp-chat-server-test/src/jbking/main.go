package main

/*
これを参考にした習作
https://github.com/akrennmair/telnet-chat/blob/master/chat.go
*/

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

type NetClient struct {
	con net.Conn
	c   chan Message
}

type Client interface {
	Id() string
	Channel() chan<- Message
	Join(Room)
}

type Room interface {
	AddClient(Client)
	DeleteClient(Client)
	Message(Message)
}

type SimpleRoom struct {
	addclient    chan Client
	deleteclient chan Client
	msgchan      chan Message
}

type Message string

const MAX_MSG_BUF int = 64

func (nc *NetClient) Channel() chan<- Message {
	return nc.c
}

func (nc *NetClient) Id() string {
	return fmt.Sprint(nc.con.RemoteAddr())
}

func (room *SimpleRoom) AddClient(client Client) {
	room.addclient <- client
}

func (room *SimpleRoom) DeleteClient(client Client) {
	room.deleteclient <- client
}

func (room *SimpleRoom) Message(msg Message) {
	room.msgchan <- msg
}

func (client *NetClient) Join(room Room) {
	io.WriteString(client.con, "> ")
	go func() {
		defer client.con.Close()
		for s := range client.c {
			if _, err := io.WriteString(client.con, string(s)); err != nil {
				room.DeleteClient(client)
				return
			}
			io.WriteString(client.con, "> ")
		}
	}()

	room.AddClient(client)

	buf := bufio.NewReader(client.con)
	for {
		l, _, err := buf.ReadLine()
		if err != nil {
			break
		}
		room.Message(Message(string(l) + "\r\n"))
	}
}

func (room *SimpleRoom) distribute() {
	clients := make(map[Client]bool)
	for {
		select {
		case client := <-room.addclient:
			fmt.Printf("new client: %v\n", client.Id())
			clients[client] = true
		case client := <-room.deleteclient:
			fmt.Printf("delete client: %v\n", client.Id())
			delete(clients, client)
		case msg := <-room.msgchan:
			for client, _ := range clients {
				select {
				case client.Channel() <- msg:
				default:
				}
			}
		}
	}
}

func NewClient(conn net.Conn) NetClient {
	c := make(chan Message, MAX_MSG_BUF)
	client := NetClient{conn, c}
	return client
}

func main() {
	port := ":8080"
	if len(os.Args) > 1 {
		port = ":" + os.Args[1]
	}
	ln, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal(err)
	}

	room := SimpleRoom{
		make(chan Client),
		make(chan Client),
		make(chan Message),
	}

	go room.distribute()

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}

		client := NewClient(conn)
		go client.Join(&room)
	}
}
