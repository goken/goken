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

func (c *Client) Read(br *bufio.Reader) error {
	message, err := br.ReadString('\n')
	if err != nil {
		return err
	}
	fmt.Printf("send message from client: %q\n", message)
	c.BroadCastChan <- message
	return nil
}

func (c *Client) ReadLoop() {
	br := bufio.NewReader(c.Conn)
	for {
		err := c.Read(br)
		if err != nil {
			if err == io.EOF {
				c.LeaveChan <- c
				c.Conn.Close()
				break
			}
			if err.Error() == "use of closed network connection" {
				break
			}
			continue
		}
	}
}

func (c *Client) Write(bw *bufio.Writer, message string) (err error) {
	fmt.Printf("send message to client: %q\n", message)
	_, err = bw.WriteString(message)
	if err != nil {
		return err
	}
	err = bw.Flush()
	if err != nil {
		return err
	}
	return err
}

func (c *Client) WriteLoop() {
	bw := bufio.NewWriter(c.Conn)
	for message := range c.WriteChan {
		err := c.Write(bw, message)
		if err != nil {
			continue
		}
	}
}

func (c *Client) String() string {
	return fmt.Sprintf("{id: %d}", c.Id)
}
