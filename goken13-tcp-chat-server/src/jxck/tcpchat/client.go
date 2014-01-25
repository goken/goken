package tcpchat

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

type Client struct {
	conn          net.Conn
	reader        *bufio.Reader
	writer        *bufio.Writer
	broadcastChan chan string
}

func NewClient(conn net.Conn, broadcastChan chan string) *Client {
	return &Client{
		conn:          conn,
		reader:        bufio.NewReader(conn),
		writer:        bufio.NewWriter(conn),
		broadcastChan: broadcastChan,
	}
}

// メッセージのレシーブループを回し、受信メッセージとエラーを通知するための
// channel を返す
func (c *Client) Receive() (recvChan chan string, errChan chan error) {
	recvChan = make(chan string)
	errChan = make(chan error)
	go func() {
		for {
			message, err := c.reader.ReadString('\n')
			if err != nil {
				errChan <- err
				break
			}
			fmt.Printf("reseive message %q (%d)\n", message, len(message))
			recvChan <- message
		}
	}()
	return recvChan, errChan
}

// メッセージを送信する
func (c *Client) Send(message string) (err error) {
	_, err = c.writer.WriteString(message)
	c.writer.Flush()
	return err
}

// クライアントが受信したメッセージのブロードキャストと
// 切断したクライアントの処理
func (client *Client) Handle(dissconnect chan *Client) {
	recvChan, errChan := client.Receive()
	go func() {
		for {
			select {
			case message := <-recvChan:
				client.broadcastChan <- message
			case err := <-errChan:
				if err == io.EOF {
					fmt.Println("dissconected")
					dissconnect <- client
				} else {
					log.Println(err)
				}
			}
		}
	}()
}

func (client *Client) Close() (err error) {
	fmt.Println("close client")
	err = client.conn.Close()
	return err
}
