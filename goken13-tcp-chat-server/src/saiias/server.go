package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

// Client
type ChatClient struct {
	inputstream  chan string
	outputstream chan string
	reader       *bufio.Reader
	writer       *bufio.Writer
}

func (client *ChatClient) Listen() {
	go client.Read()
	go client.Write()
}

func (client *ChatClient) Read() {
	for {
		message, err := client.reader.ReadString('\n')
		if err != nil {
			Log("Buffer Reader error")
		}
		client.inputstream <- message
	}
}

func (client *ChatClient) Write() {
	for temp := range client.outputstream {
		client.writer.WriteString(temp)
		client.writer.Flush()
	}
}

// Server
type ChatServer struct {
	clients      []*ChatClient
	inputstream  chan string
	outputstream chan string
	join         chan net.Conn
}

func (serv *ChatServer) SendMessage(message string) {
	for _, client := range serv.clients {
		client.outputstream <- message
	}
}

func (serv *ChatServer) Listen() {
	go func() {
		counter := MessageID()
		for {
			select {
			case message := <-serv.inputstream:
				Log(fmt.Sprintf("[%d]: %s", counter(),message))
				serv.SendMessage(message)
			case conn := <-serv.join:
				serv.Join(conn)
			}
		}
	}()
}

func (serv *ChatServer) Join(con net.Conn) {
	Log("connected new client\n")

	client := StartClient(con)
	serv.clients = append(serv.clients, client)

	Log(fmt.Sprintf("client count: %d\n",len(serv.clients)))

	go func() {
		for {
			serv.inputstream <- <-client.inputstream
		}
	}()
}

//
func StartClient(conn net.Conn) *ChatClient {

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	client := &ChatClient{
		inputstream:  make(chan string),
		outputstream: make(chan string),
		reader:       reader,
		writer:       writer,
	}

	Log("Start Client\n")
	client.Listen()
	return client
}

func Start() *ChatServer {
	ChatServer := &ChatServer{
		clients:      make([]*ChatClient, 0),
		inputstream:  make(chan string),
		outputstream: make(chan string),
		join:         make(chan net.Conn),
	}
	ChatServer.Listen()
	return ChatServer
}

// Logger
func Log(s string) {
	fmt.Printf(s)
}

func MessageID() func() int{
	cnt := 0
	add := func() int{
		cnt +=1
		return cnt
	}
	return add
}

func main() {

	if len(os.Args) != 2 {
		fmt.Printf("Usage: ./server port\n")
		os.Exit(2)
	}

	port := os.Args[1]
	
	Log("Starting Server...\n")

	serv := Start()

	chat, err := net.Listen("tcp", ":" + port)
	if err != nil {
		Log("Listen Error:")
		os.Exit(1)
	}

	for {
		con, _ := chat.Accept()
		serv.join <- con
	}
}
