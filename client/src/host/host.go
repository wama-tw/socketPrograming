package host

import (
	"client/src/models"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

func StartHost(webSocket *websocket.Conn) {
	defer webSocket.Close()
	service := "localhost:1200"
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	if checkError(err) {
		return
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if checkError(err) {
		return
	}

	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(2)

	var done context.CancelFunc
	ctx, done := context.WithCancel(context.Background())
	go sendRequest(conn, done, waitGroup, webSocket) // WRITE
	go getResponse(conn, ctx, waitGroup, webSocket)  // READ

	waitGroup.Wait()
	println(conn.LocalAddr().String(), "Disconnected")
	return
}

func sendRequest(
	conn net.Conn,
	done context.CancelFunc,
	waitGroup *sync.WaitGroup,
	webSocket *websocket.Conn) {
	defer waitGroup.Done()
	encoder := json.NewEncoder(conn)
	for {
		_, messageBytes, webSocketErr := webSocket.ReadMessage()
		message := string(messageBytes)
		t := time.Now()
		println(conn.LocalAddr().String(), "Received from webSocket: ", message)
		if webSocketErr != nil {
			println(conn.LocalAddr().String(), "webSocketErr: ", webSocketErr.Error())
			encoder.Encode(models.Message{
				Author: conn.LocalAddr().String(),
				Time:   timeFormat(t),
				Exit:   true,
			})
			done()
			return
		}
		println(conn.LocalAddr().String(), "Send to server: ", message)
		err := encoder.Encode(models.Message{
			Author: conn.LocalAddr().String(),
			Text:   message,
			Time:   timeFormat(t),
			Exit:   false,
		})
		if checkError(err) {
			done()
			return
		}
	}
}

func getResponse(
	conn net.Conn,
	ctx context.Context,
	waitGroup *sync.WaitGroup,
	webSocket *websocket.Conn) {
	defer waitGroup.Done()
	var response models.Message
	decoder := json.NewDecoder(conn)

	for {
		err := decoder.Decode(&response) // blocking
		if checkError(err) {
			return
		}

		if notEmpty(response.Text) {
			println(conn.LocalAddr().String(), "Server response: ", response.Text)
			webSocket.WriteJSON(response)
			response = models.Message{} // clear last read content
		}
		if response.Join {
			println(conn.LocalAddr().String(), "Join")
			webSocket.WriteJSON(response)
			response = models.Message{} // clear last read content
		}
		if response.Exit {
			println(conn.LocalAddr().String(), "Left")
			webSocket.WriteJSON(response)
			response = models.Message{} // clear last read content
		}

		select {
		case <-ctx.Done():
			return
		default:
		}
	}
}

func timeFormat(t time.Time) string {
	formatted := ""
	if t.Hour() < 10 {
		formatted += "0"
	}
	formatted += fmt.Sprint(t.Hour())
	formatted += ":"
	if t.Minute() < 10 {
		formatted += "0"
	}
	formatted += fmt.Sprint(t.Minute())

	return formatted
}

func notEmpty(msg string) bool {
	return (msg != "" && msg != " " && msg != "\n")
}

func checkError(err error) bool {
	if err != nil {
		error_msg := fmt.Sprintf("Fatal error: %s", err.Error())
		println(error_msg)

		return true
	}

	return false
}
