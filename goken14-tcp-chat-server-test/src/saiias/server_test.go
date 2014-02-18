package main

import(
	"testing"
	"net"
	"os/exec"
	"bytes"
)

type connListener struct{
	conn net.Conn
}

func TestStartServ(t *testing.T){
	_,err := net.Listen("tcp",":8080")
	if err != nil {
		t.Error("Fail to Listen")
	}
}

func TestStart(t *testing.T){
	serv := Start()
	if serv == nil{
		t.Error("Fail to Start Server")
	}
}

func CreateEnv() net.Conn{
	chat, _ := net.Listen("tcp", ":8080")
	for{
		con, _ := chat.Accept()
		return con
	}
}

func StartServ(t *testing.T){
	cmd := exec.Command("go run server.go 8080")
	var out bytes.Buffer
	cmd.Stdout = &out	
	err := cmd.Run()
	if err != nil {
		t.Error("Fail %q",out.String())
	}
}

func TestConn(t *testing.T){
	go StartServ(t)
};

