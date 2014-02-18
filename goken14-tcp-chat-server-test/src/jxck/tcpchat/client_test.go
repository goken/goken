package chat

import (
	"bytes"
	"testing"
)

type rwcMock struct {
	bytes.Buffer
	Closed bool
}

func (m *rwcMock) Close() error {
	m.Closed = true
	return nil
}

func TestReadLoop(t *testing.T) {
	id := 1
	conn := &rwcMock{}
	broadCastChan := make(chan string)
	client := NewClient(id, conn, broadCastChan, nil)

	expected := "message\r\n"
	client.Conn.Write([]byte(expected))

	actual := <-broadCastChan

	if actual != expected {
		t.Errorf("\ngot  %v\nwant %v", actual, expected)
	}
}

func TestLeaveChan(t *testing.T) {
	id := 1
	conn := &rwcMock{}
	broadCastChan := make(chan string)
	leaveChan := make(chan *Client)
	client := NewClient(id, conn, broadCastChan, leaveChan)

	client.Conn.Close()

	actual := <-client.LeaveChan
	expected := client

	if actual != expected {
		t.Errorf("\ngot  %v\nwant %v", actual, expected)
	}
}

func TestWriteChan(t *testing.T) {
	id := 1
	conn := &rwcMock{}
	client := NewClient(id, conn, nil, nil)

	expected := "message\r\n"
	client.WriteChan <- expected

	actual := string(conn.Bytes())

	if actual != expected {
		t.Errorf("\ngot  %v\nwant %v", actual, expected)
	}
}
