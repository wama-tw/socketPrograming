package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

func main() {
	service := "localhost:1200"
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	checkError(err)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	fmt.Println(tcpAddr)
	checkError(err)
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	conn.SetReadDeadline(time.Now().Add(2 * time.Minute)) // set 2 minutes timeout
	defer conn.Close()                                    // close connection before exit

	println("connected: ", conn.RemoteAddr())
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(2)

	var done context.CancelFunc
	ctx, done := context.WithCancel(context.Background())

	message := make(chan string, 10)
	go getRequest(conn, message, done, waitGroup)  // Get request
	go sendResponse(conn, message, ctx, waitGroup) // Response

	waitGroup.Wait()
	println(conn.RemoteAddr(), " :handleClient return")
	return
}

func getRequest(conn net.Conn, message chan string, done context.CancelFunc, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	request := make([]byte, 128) // set maxium request length to 128B to prevent flood attack
	for {
		read_len, err := conn.Read(request)
		request_msg := string(request)
		message <- request_msg
		request = make([]byte, 128) // clear last read content
		if err != nil {
			fmt.Println("Fatal error: ", err.Error())
		}

		// Request process
		if read_len == 0 || request_msg == "" || request_msg == "\n" {
			println("break")
			done()
			return // connection already closed by client
		}
	}
}

func sendResponse(conn net.Conn, message chan string, ctx context.Context, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	for {
		request_msg := <-message
		println("client sent: ", request_msg)
		response := ("Received: " + request_msg)
		conn.Write([]byte(response))
		println("Sent to client: ", []byte(response), " ", response)
		println("-------------------------------------")

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
