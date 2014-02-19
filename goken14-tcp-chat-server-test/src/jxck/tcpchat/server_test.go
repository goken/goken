package chat

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"testing"
)

var port string = ":3000"

func init() {
	log.SetFlags(log.Lshortfile)
}

func TestServer(t *testing.T) {
	server := NewServer()
	server.Listen(port)
	go server.AcceptLoop()
	defer server.Close()

	expected := "test\r\n"

	conn, _ := net.Dial("tcp", port)
	fmt.Fprintf(conn, expected)
	actual, _ := bufio.NewReader(conn).ReadString('\n')

	t.Logf("%q, %q", actual, expected)
	if actual != expected {
		t.Errorf("\ngot  %v\nwant %v", actual, expected)
	}
}

func TestMultiClient(t *testing.T) {
	server := NewServer()
	server.Listen(port)
	go server.AcceptLoop()
	defer server.Close()

	expected := "test\r\n"

	conn1, _ := net.Dial("tcp", port)
	conn2, _ := net.Dial("tcp", port)

	buf1 := bufio.NewReader(conn1)
	buf2 := bufio.NewReader(conn2)

	fmt.Fprintf(conn1, expected)

	actual1, _ := buf1.ReadString('\n')
	actual2, _ := buf2.ReadString('\n')

	t.Logf("%q, %q", actual1, actual2)
	if actual1 != actual2 {
		t.Errorf("\ngot  %v\nwant %v", actual1, actual2)
	}
}

func TestNextId(t *testing.T) {
	nextid := NextId()
	actual := <-nextid
	if actual != 0 {
		t.Errorf("nextid should start with %d but %d", actual, nextid)
	}
}

func TestServerJoin(t *testing.T) {
	server := NewServer()
	actual := len(server.Clients)
	expected := 0

	if actual != expected {
		t.Errorf("initial clients size should %d but %d", actual, expected)
	}

	client := &Client{}
	server.Join(client)

	actual = len(server.Clients)
	expected = 1

	if actual != expected {
		t.Errorf("initial clients size should %d but %d", actual, expected)
	}
}

func TestLeave(t *testing.T) {
	server := NewServer()
	client := &Client{}
	server.Join(client)

	if len(server.Clients) != 1 {
		t.Errorf("clients size should %d but %d", 1, len(server.Clients))
	}

	server.LeaveChan <- client

	if len(server.Clients) != 0 {
		t.Errorf("clients size should %d but %d", 0, len(server.Clients))
	}
}

func TestClose(t *testing.T) {
	server := NewServer()
	server.Listen(port)
	go server.AcceptLoop()

	client1 := &Client{Id: 1, Conn: new(rwcMock)}
	client2 := &Client{Id: 2, Conn: new(rwcMock)}
	client3 := &Client{Id: 3, Conn: new(rwcMock)}
	client4 := &Client{Id: 4, Conn: new(rwcMock)}

	server.Join(client1)
	server.Join(client2)
	server.Join(client3)
	server.Join(client4)

	server.Close()

	for _, c := range []*Client{client1, client2, client3, client4} {
		if c.Conn.(*rwcMock).Closed != true {
			t.Errorf("fail to close client connection %v", c)
		}
	}
}
