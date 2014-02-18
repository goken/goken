package main

import (
	"testing"
)

type TestClient struct {
	id string
	c  chan Message
}

func (client *TestClient) Id() string {
	return client.id
}

func (client *TestClient) Channel() chan<- Message {
	return client.c
}

func (client *TestClient) Join(Room) {
	panic("Not implemented")
}

func TestRoomType(t *testing.T) {
	room := &SimpleRoom{
		nil,
		nil,
		nil,
	}
	// Type check
	_ = Room(room)
}

func TestDistribute(t *testing.T) {
	room := &SimpleRoom{
		make(chan Client),
		make(chan Client),
		make(chan Message),
	}

	client := &TestClient{
		"test",
		make(chan Message),
	}

	go room.distribute()
	room.AddClient(client)
	room.Message(Message("ping"))
	select {
	case msg := <-client.c:
		if string(msg) != "ping" {
			t.Errorf("Got wrong message: %v", msg)
		}
	}
}
