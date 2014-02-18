package tcpchat

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
	Conn     io.ReadWriter
	WriteBuf chan string
}

func NewClient(conn io.ReadWriter) *Client {
	client := &Client{
		Conn:     conn,
		WriteBuf: make(chan string, 32),
	}
	return client
}

func (c *Client) Write(message string) error {
	_, err := io.WriteString(c.Conn, message)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) WriteLoop() {
	for message := range c.WriteBuf {
		_, err := io.WriteString(c.Conn, message)
		if err != nil {
			log.Println(err)
		}
	}
}

func (c *Client) ReadLoop(broadcast chan string) {
	fmt.Printf("connect %+v\n", c)
	br := bufio.NewReader(c.Conn)
	for {
		message, err := br.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Printf("dissconnect %+v\n", c.Conn)
			} else {
				log.Println(err)
			}
			return
		}
		fmt.Printf("%q\n", message)
		broadcast <- message
	}
}
