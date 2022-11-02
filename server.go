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

	var connections []net.Conn
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		connections = append(connections, conn)
		go handleClient(conn, &connections)
	}
}

func handleClient(conn net.Conn, connections *[]net.Conn) {
	conn.SetReadDeadline(time.Now().Add(2 * time.Minute)) // set 2 minutes timeout
	defer conn.Close()                                    // close connection before exit

	println("connected: ", conn.RemoteAddr())
	handleClientWaitGroup := &sync.WaitGroup{}
	handleClientWaitGroup.Add(2)

	var done context.CancelFunc
	ctx, done := context.WithCancel(context.Background())

	message := make(chan string, 10)
	go getRequest(conn, message, done, handleClientWaitGroup)               // Get request
	go sendResponse(conn, message, ctx, handleClientWaitGroup, connections) // Response

	handleClientWaitGroup.Wait()
	println(conn.RemoteAddr(), " :handleClient return")
	(*connections) = remove((*connections), conn)
	return
}

func getRequest(
	conn net.Conn, message chan string,
	done context.CancelFunc,
	handleClientWaitGroup *sync.WaitGroup,
) {
	defer handleClientWaitGroup.Done()
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

func sendResponse(
	conn net.Conn,
	message chan string,
	ctx context.Context,
	handleClientWaitGroup *sync.WaitGroup,
	connections *[]net.Conn,
) {
	defer handleClientWaitGroup.Done()
	for {
		requestMsg := <-message
		println("client sent: ", requestMsg)
		broadcastMsg := (fmt.Sprint(conn.RemoteAddr()) + " said: " + requestMsg)
		for _, connection := range *connections {
			println("sending message to ", fmt.Sprint(connection.RemoteAddr()))
			connection.Write([]byte(broadcastMsg))
		}
		// response := ("Received: " + requestMsg)
		// conn.Write([]byte(response))
		println("Sent to clients: ", []byte(broadcastMsg), " ", broadcastMsg)
		println("-------------------------------------")

		select {
		case <-ctx.Done():
			return
		default:
			continue
		}
	}
}

func remove(s []net.Conn, remove net.Conn) []net.Conn {
	for index, element := range s {
		if element == remove {
			s[index] = s[len(s)-1]
			break
		}
	}
	return s[:len(s)-1]
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
