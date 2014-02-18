package main

import (
	"bytes"
	"io"
	"net"
	"testing"
	"time"
)

// testConnのモック用
type testAddr string

func (addr testAddr) Network() string {
	return string(addr)
}

func (addr testAddr) String() string {
	return string(addr)
}

// net.Connのテスト用スタブ（モック？）
// net.Connのインターフェイスに準拠（使わなそうなメソッド実装は適当）
type testConn struct {
	readBuf  bytes.Buffer
	writeBuf bytes.Buffer
	readErr  error
	writeErr error
}

// readErrが設定されていたらエラーを返し、そうでなければreadBufから読み込む
func (conn *testConn) Read(b []byte) (n int, err error) {
	if conn.readErr != nil {
		return 0, conn.readErr
	}
	return conn.readBuf.Read(b)
}

// readErrを設定する
func (conn *testConn) SetReadErr(err error) {
	conn.readErr = err
}

// writeErrが設定されていたらエラーを返し、そうでなければwriteBufに書き込む
func (conn *testConn) Write(b []byte) (n int, err error) {
	if conn.writeErr != nil {
		return 0, conn.writeErr
	}
	return conn.writeBuf.Write(b)
}

// writeErrを設定する
func (conn *testConn) SetWriteErr(err error) {
	conn.writeErr = err
}

// 以下、使わなそうなので実装が適当
func (conn *testConn) Close() error {
	return nil
}

func (conn *testConn) LocalAddr() net.Addr {
	return testAddr("127.0.0.1:3000")
}

func (conn *testConn) RemoteAddr() net.Addr {
	return testAddr("127.0.0.1:49000")
}

func (conn *testConn) SetDeadline(t time.Time) error {
	return nil
}

func (conn *testConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (conn *testConn) SetWriteDeadline(t time.Time) error {
	return nil
}

// Server.eventに加入情報を送り、Server.outgoingのサイズが1になっているかをチェック
func TestServerJoin(t *testing.T) {
	incoming := make(chan string)
	event := make(chan membership)
	sut := NewServer(incoming, event)

	addr := "127.0.0.1:40000"

	go sut.Run()
	sut.event <- membership{JOIN, addr, make(chan string)}

	if len(sut.outgoing) != 1 {
		t.Errorf("actual: %q, expected: %q\n", len(sut.outgoing), 1)
	}
	if _, ok := sut.outgoing[addr]; !ok {
		t.Errorf("%s should be registered, but not\n", addr)
	}
}

// Server.eventに離脱情報を送り、Server.outgoingのサイズが0になっているかをチェック
func TestServerLeave(t *testing.T) {
	incoming := make(chan string)
	event := make(chan membership)
	sut := NewServer(incoming, event)

	addr := "127.0.0.1:40000"

	go sut.Run()
	sut.event <- membership{JOIN, addr, make(chan string)}
	go sut.Run()
	sut.event <- membership{LEAVE, addr, nil}

	if len(sut.outgoing) != 0 {
		t.Errorf("actual: %q, expected: %q\n", len(sut.outgoing), 0)
	}
	if _, ok := sut.outgoing[addr]; ok {
		t.Errorf("%s should be unregistered, but not\n", addr)
	}
}

// Server.incomingに文字列を送り、それが外部の出力チャネルに出力しているかをチェック
func TestServerBroadcast(t *testing.T) {
	incoming := make(chan string)
	event := make(chan membership)
	sut := NewServer(incoming, event)

	addr := "127.0.0.1:40000"
	channel := make(chan string)
	expected := "Hello\n"

	go sut.Run()
	sut.event <- membership{JOIN, addr, channel}
	go sut.Run()
	sut.incoming <- expected
	actual := <-channel

	if actual != expected {
		t.Errorf("actual: %q, expected: %q\n", actual, expected)
	}
}

// Client.Read()を実行すると出力用のチャネルに所望の文字列が出力される
func TestClientRead(t *testing.T) {
	conn := new(testConn)
	expected := "Hello\n"
	conn.readBuf.WriteString(expected)

	incoming := make(chan string)
	outgoing := make(chan string)
	leave := make(chan membership)
	sut := NewClient(conn, incoming, outgoing, leave)

	go sut.Read()

	actual := <-outgoing
	if actual != expected {
		t.Errorf("actual: %q, expected: %q\n", actual, expected)
	}
}

// net.ConnでEOFが発生すると、io.EOFを返す
func TestClientReadWhenEOFOccurs(t *testing.T) {
	conn := new(testConn)
	conn.SetReadErr(io.EOF)

	incoming := make(chan string)
	outgoing := make(chan string)
	leave := make(chan membership)
	sut := NewClient(conn, incoming, outgoing, leave)

	err := sut.Read()
	if err != io.EOF {
		t.Errorf("actual: %v, expected: %v\n", err, io.EOF)
	}
}

// net.ConnでEOFが発生すると、離脱情報が送られる
func TestClientReadLoopWhenEOFOccurs(t *testing.T) {
	conn := new(testConn)
	conn.SetReadErr(io.EOF)

	incoming := make(chan string)
	outgoing := make(chan string)
	leave := make(chan membership)
	sut := NewClient(conn, incoming, outgoing, leave)

	go sut.ReadLoop()
	<-leave
}

// Client.Write()を実行するとnet.Connに所望の文字列が書き出される
func TestClientWrite(t *testing.T) {
	conn := new(testConn)
	expected := "Hello\n"

	incoming := make(chan string)
	outgoing := make(chan string)
	leave := make(chan membership)
	sut := NewClient(conn, incoming, outgoing, leave)

	go sut.Write()
	incoming <- expected

	actual := string(conn.writeBuf.Bytes())
	if actual != expected {
		t.Errorf("actual: %q, expected: %q\n", actual, expected)
	}
}

// net.ConnでEOFが発生すると、io.EOFを返す
func TestClientWriteWhenEOFOccurs(t *testing.T) {
	conn := new(testConn)
	conn.SetWriteErr(io.EOF)
	expected := "Hello\n"

	incoming := make(chan string)
	outgoing := make(chan string)
	leave := make(chan membership)
	sut := NewClient(conn, incoming, outgoing, leave)

	var err error
	go func() {
		err = sut.Write()
	}()
	incoming <- expected

	if err != io.EOF {
		t.Errorf("actual: %v, expected: %v\n", err, io.EOF)
	}
}

// net.ConnでEOFが発生すると、離脱情報が送られる
func TestClientWriteLoopWhenEOFOccurs(t *testing.T) {
	conn := new(testConn)
	conn.SetWriteErr(io.EOF)
	expected := "Hello\n"

	incoming := make(chan string)
	outgoing := make(chan string)
	leave := make(chan membership)
	sut := NewClient(conn, incoming, outgoing, leave)

	go sut.WriteLoop()
	incoming <- expected
	<-leave
}
