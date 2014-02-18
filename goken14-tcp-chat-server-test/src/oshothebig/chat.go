/*
	本来は型毎にファイルを分けた方がよいと思うが、ここでは1ファイルで書いた
	goroutineの起動を行うところを一箇所にまとめるようにした
*/

package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <port>\n", os.Args[0])
	}
	port := ":" + os.Args[1]

	server := NewServer(make(chan string), make(chan membership))
	server.Listen(port)
}

// イベントのタイプを型定義
type eventType int

// イベントは加入と離脱がある
const (
	JOIN eventType = iota
	LEAVE
)

// 加入および離脱が発生した時にはこの構造体をメッセージとしてチャネルに送る
// channelにはブロードキャストを受け取るチャネルを入れる
type membership struct {
	event   eventType
	addr    string
	channel chan<- string
}

//
type Server struct {
	// クライアントから文字列を受け取るチャネル
	// 本当は受信専用にしたい
	incoming chan string
	// クライアント毎のブロードキャスト用の出力チャネル
	outgoing map[string]chan<- string
	// クライアントの加入、離脱のイベントを受け取るチャネル
	// 本当は受信専用にしたい
	event chan membership
}

func NewServer(incoming chan string, event chan membership) *Server {
	return &Server{
		incoming,
		make(map[string]chan<- string),
		event,
	}
}

// ポート番号を指定して待ち受け
func (server *Server) Listen(port string) {
	l, err := net.Listen("tcp", port)
	defer l.Close()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Server started: port=%s\n", port)

	go server.RunLoop()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println(err)
		}

		toClient := make(chan string)
		//	クライアントを追加
		server.event <- membership{JOIN, conn.RemoteAddr().String(), toClient}
		client := NewClient(conn, toClient, server.incoming, server.event)
		go client.ReadLoop()
		go client.WriteLoop()
	}
}

// サーバのメインの処理
// incomingを受信すると、ブロードキャスト
// eventを受信すると、クライアントの更新
func (server *Server) Run() {
	select {
	case m := <-server.incoming:
		server.Broadcast(m)
	case e := <-server.event:
		switch e.event {
		case JOIN:
			server.Join(e.addr, e.channel)
		case LEAVE:
			server.Leave(e.addr)
		}
	}
}

func (server *Server) RunLoop() {
	for {
		server.Run()
	}
}

// クライアントに対して指定された文字列を送信
func (server *Server) Broadcast(message string) {
	for _, outgoing := range server.outgoing {
		select {
		case outgoing <- message:

			// クライアントがつまっている時にselect文でブロックしないようにするためには必要
			// 但し、TestServerLeave()がデッドロックしてしまうのでコメントアウトしている
			// コメントアウトせずにテストを書くにはどうしたらよいのか分かっていない
			// default:
		}
	}
}

// クライアントの追加
func (server *Server) Join(addr string, toClient chan<- string) {
	server.outgoing[addr] = toClient
}

// クライアントの削除
func (server *Server) Leave(addr string) {
	delete(server.outgoing, addr)
}

type Client struct {
	// net.Connを保持していないため
	addr   string
	reader *bufio.Reader
	writer *bufio.Writer
	// サーバからのチャネル
	incoming <-chan string
	// サーバ宛てのチャネル
	outgoing chan<- string
	// 離脱情報を送るチャネル
	leave chan<- membership
}

func NewClient(conn net.Conn, incoming <-chan string, outgoing chan<- string, leave chan<- membership) *Client {
	client := &Client{
		addr:     conn.RemoteAddr().String(),
		reader:   bufio.NewReader(conn),
		writer:   bufio.NewWriter(conn),
		incoming: incoming,
		outgoing: outgoing,
		leave:    leave,
	}

	return client
}

// 改行毎に読み取りを行い、outgoingに送信
func (client *Client) Read() (err error) {
	message, err := client.reader.ReadString('\n')
	if err != nil {
		return err
	}

	log.Printf("Read: %q", message)
	client.outgoing <- message
	return
}

// エラーが発生したら、離脱
func (client *Client) ReadLoop() {
	for {
		err := client.Read()
		if err != nil {
			if err == io.EOF {
				log.Println("Disconnected")
			} else {
				log.Println(err)
			}
			client.Leave()
			return
		}
	}
}

// incomingから受信した文字列を、外部に出力
func (client *Client) Write() (err error) {
	message := <-client.incoming
	_, err = client.writer.WriteString(message)
	if err != nil {
		return
	}
	err = client.writer.Flush()
	log.Printf("Write: %q\n", message)
	return
}

// エラーが発生したら、離脱
func (client *Client) WriteLoop() {
	for {
		err := client.Write()
		if err != nil {
			if err == io.EOF {
				log.Println("Disconnected")
			} else {
				log.Println(err)
			}
			client.Leave()
			return
		}
	}
}

func (client *Client) Leave() {
	client.leave <- membership{LEAVE, client.addr, nil}
}
