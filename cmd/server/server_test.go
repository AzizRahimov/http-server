package main

import (
	"bufio"
	"io/ioutil"
	"net"
	"testing"
	"time"
)

func Test_serverStarting(t *testing.T) {
	go func() {
		err := start("localhost:3333")
		if err != nil {
			t.Fatalf("can't start server: %v", err)
		}
	}()
	time.Sleep(time.Second)
	conn, err := net.Dial("tcp", "localhost:3333")
	if err != nil {
		t.Fatalf("can't connect to server: %v", err)
	}
	defer func() {
		err := conn.Close()
		if err != nil {
			t.Fatalf("can't close conn: %v\n", err)
		}
	}()
	writer := bufio.NewWriter(conn)
	_, err = writer.WriteString("GET /http.jpg HTTP/1.1")
	if err != nil {
		t.Fatalf("can't write: %v\n", err)
	}
	_, err = writer.WriteString("Host: localhost\r\n")
	if err != nil {
		t.Fatalf("can't write: %v\n", err)
	}
	_, err = writer.WriteString("\r\n")
	if err != nil {
		t.Fatalf("can't write: %v\n", err)
	}
	err = writer.Flush()
	if err != nil {
		t.Fatalf("can't flush: %v\n", err)
	}
	_, err = ioutil.ReadAll(conn)
	if err != nil {
		t.Fatalf("can't read response from server: %v", err)
	}

}