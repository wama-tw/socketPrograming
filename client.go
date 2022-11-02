package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
)

func main() {
	service := "localhost:1200"
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	checkError(err)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	checkError(err)

	// done := make(chan bool)
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(2)

	var done context.CancelFunc
	ctx, done := context.WithCancel(context.Background())
	go write(conn, done, waitGroup) // WRITE
	go read(conn, ctx, waitGroup)   // READ

	waitGroup.Wait()
	os.Exit(0)
}

func write(conn net.Conn, done context.CancelFunc, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	for {
		msg := ""
		fmt.Scanf("%s", &msg)
		if msg == "exit" {
			conn.Write([]byte("client exit"))
			done()
			return
		}
		strings.Replace(msg, " ", "", -1)
		_, err := conn.Write([]byte(msg))
		checkError(err)
	}
}

func read(conn net.Conn, ctx context.Context, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	response := make([]byte, 128)

	for {
		_, err := conn.Read(response) // blocking
		checkError(err)
		// println(response)
		response_msg := string(response)
		// println(response_msg)
		// println(notEmpty(response_msg))
		if notEmpty(response_msg) {
			println("server response: ", response_msg)
			response = make([]byte, 128) // clear last read content
		}

		select {
		case <-ctx.Done():
			return
		default:
			continue
		}
	}
}

func notEmpty(msg string) bool {
	return (msg != "" && msg != " " && msg != "\n")
}

func checkError(err error) {
	if err != nil {
		error_msg := fmt.Sprintf("Fatal error: %s", err.Error())
		log.Fatal(error_msg)
	}
}
