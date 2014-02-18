package chat

import (
	"bufio"
	"fmt"
	"io"
	"log"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

type Client struct {
	Id            int
	Conn          io.ReadWriteCloser
	BroadCastChan chan string
	WriteChan     chan string
	LeaveChan     chan *Client
}

func NewClient(id int, conn io.ReadWriteCloser, broadCastChan chan string, leaveChan chan *Client) *Client {
	client := &Client{
		Id:            id,
		Conn:          conn,
		WriteChan:     make(chan string),
		BroadCastChan: broadCastChan,
		LeaveChan:     leaveChan,
	}
	go client.ReadLoop()
	go client.WriteLoop()
	return client
}

func (c *Client) ReadLoop() {
	br := bufio.NewReader(c.Conn)
	for {
		message, err := br.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				c.LeaveChan <- c
				c.Conn.Close()
				break
			}
			log.Println(err)
			continue
		}
		fmt.Printf("send message from client: %q\n", message)
		c.BroadCastChan <- message
	}
}

func (c *Client) WriteLoop() {
	bw := bufio.NewWriter(c.Conn)
	for message := range c.WriteChan {
		fmt.Printf("send message to client: %q\n", message)
		bw.WriteString(message)
		bw.Flush()
	}
}

func (c *Client) String() string {
	return fmt.Sprintf("{id: %d}", c.Id)
}
