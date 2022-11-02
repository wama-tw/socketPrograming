package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

func main() {
	service := "localhost:1200"
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	checkError(err)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	checkError(err)

	response := make([]byte, 128)
	for {
		msg := ""
		fmt.Scanf("%s", &msg)
		if msg == "exit" {
			break
		}
		strings.Replace(msg, " ", "", -1)
		_, err = conn.Write([]byte(msg))
		checkError(err)

		time.Sleep(200 * time.Millisecond)
		_, err = conn.Read(response)
		checkError(err)
		println("server response: ", string(response))
		response = make([]byte, 128) // clear last read content

		println("-------------------------------------")
	}
	os.Exit(0)
}
func checkError(err error) {
	if err != nil {
		error_msg := fmt.Sprintf("Fatal error: %s", err.Error())
		log.Fatal(error_msg)
	}
}
