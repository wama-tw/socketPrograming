package main

import (
	"fmt"
	"log"
	"net"
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
	request := make([]byte, 128)                          // set maxium request length to 128B to prevent flood attack
	defer conn.Close()                                    // close connection before exit
	for {
		// Get request
		read_len, err := conn.Read(request)
		request_msg := string(request)
		request = make([]byte, 128) // clear last read content
		if err != nil {
			fmt.Println("Fatal error: ", err.Error())
		}

		// Request process
		// strings.Replace(request_msg, " ", "", -1)
		if read_len == 0 || request_msg == "" || request_msg == "\n" {
			println("break")
			break // connection already closed by client
		}

		// Response
		println("client sent: ", request_msg)
		response := ("Received: " + request_msg)
		conn.Write([]byte(response))
		println("Sent to client: ", []byte(response), " ", response)

		println("-------------------------------------")
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
