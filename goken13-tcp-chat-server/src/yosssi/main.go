/**
 * メインスレッドの無限ループにて新規TCP接続を受け付ける
 * TCP接続ごとにGoroutineを起動し、各Goroutineにてクライアントからのメッセージ送信を受け付ける
 * 現在接続されているTCP接続の情報はグローバル変数のMapにて管理する（接続元クライアントのIP:ポート番号をMapのキートする）
 * クライアントからの新規接続時はMapにその接続情報を追加し、切断時はMapからその接続情報をMapから除去する
 * クライアントからメッセージが送信された場合は、Mapに格納されている各接続先に対して、そのメッセージを送信する
 **/

package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

var (
	// サービスリスンポート番号
	port string
	// 接続中のTCP接続を保持するMap
	conns = make(map[string]net.Conn)
)

// サービス起動時の初期化処理
// 第1引数で指定されたポート番号をグローバル変数にセットする
func init() {
	if len(os.Args) < 2 {
		log.Fatal("第1引数にポート番号を指定してください。")
	}
	port = os.Args[1]
}

// メインスレッド
// 無限ループで新規TCP接続を受け付ける
func main() {
	// サービスリスン
	l, err := net.Listen("tcp", "localhost:"+port)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("ポート番号%sでリスンしています。\n", port)

	// 新規TCP接続の受付
	for {
		c, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		// TCP接続ごとにGoroutineを起動
		go serve(c)
	}
}

// TCP接続ごとに起動されるGoroutine
// 無限ループでクライアントからの入力を受け付ける
func serve(c net.Conn) {
	// クライアントのIPアドレスを取得する
	addr := c.RemoteAddr().String()

	log.Printf("クライアントから接続されました。[%s]\n", addr)

	// グローバル変数のTCP接続マップにこの接続を追加する
	conns[addr] = c

	// 現在接続中のTCP接続の一覧をコンソールへ出力する
	showConns()

	fmt.Fprintf(c, "Go研チャットサーバへようこそ！[%s]\n", addr)

	// 無限ループでクライアントからの入力を受け付ける
	r := bufio.NewReader(c)
	for {
		b, _, err := r.ReadLine()
		if err != nil {
			if err == io.EOF {
				log.Printf("クライアントからの接続が切断されました。[%s]\n", addr)
			} else {
				log.Printf("クライアントからのメッセージの読み込みでエラーが発生しました。[%s][%s]\n", addr, err.Error())
			}
			// クライアントからの入力受付にてエラーが発生した場合は、そのクライアントとの接続を切断する
			close(c)
			return
		}

		// クライアントから送信されたメッセージを処理する
		handleInput(c, string(b))
	}
}

// クライアントとの接続を切断する
func close(c net.Conn) {
	addr := c.RemoteAddr().String()
	log.Printf("クライアントとの接続を切断します。[%s]\n", addr)
	err := c.Close()
	if err != nil {
		log.Printf("クライアントとの接続の切断に失敗しました。[%s][%s]\n", addr, err.Error())
	} else {
		log.Printf("クライアントとの接続を切断しました。[%s]\n", addr)
		// 現在接続中のTCP接続情報を保持するMapより、今回切断されたTCP接続を除去する
		delete(conns, addr)
	}
	// 現在接続中のTCP接続の一覧をコンソールへ出力する
	showConns()
}

// 現在接続中のTCP接続の一覧をコンソールへ出力する
func showConns() {
	if len(conns) > 0 {
		log.Println("現在接続中のクライアント:")
		i := 0
		for addr, _ := range conns {
			i++
			space := ""
			if i < 10 {
				space = " "
			}
			log.Printf("%s %d %s\n", space, i, addr)
		}
	} else {
		log.Println("現在接続中のクライアントはありません。")
	}
}

// クライアントからの入力内容を処理する
func handleInput(c net.Conn, s string) {
	tokens := strings.Split(s, ":")
	if len(tokens) < 2 {
		fmt.Fprintln(c, "入力内容の形式が不正です。「名前:メッセージ」という形式で入力してください。")
		return
	}
	name := strings.TrimSpace(tokens[0])
	message := strings.TrimSpace(strings.Join(tokens[1:], ":"))
	log.Printf("クライアントからメッセージが送信されました。[%s][%s][%s]", c.RemoteAddr().String(), name, message)
	go broadcast(name, message)
}

// メッセージを全クライアントへ送信する
func broadcast(name string, message string) {
	for _, c := range conns {
		fmt.Fprintln(c, name+": "+message)
	}
}
