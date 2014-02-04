package main

/**
 * 大きい方の実装
 * Server と Client を分けて実装し
 * Client も独自に Loop をもっている
 * Client の追加、削除をちゃんとやってる。
 */

import (
	chat "jxck/tcpchat"
	"log"
	"os"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("port missing")
	}
	port := os.Args[1]
	log.Println(chat.ListenAndServe(port))
}
