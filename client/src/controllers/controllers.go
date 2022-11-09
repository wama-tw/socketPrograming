package controllers

import (
	"client/src/host"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func SendHTML(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
}

func ConnectWebSocket(c *gin.Context) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  128,
		WriteBufferSize: 128,
	}
	webSocket, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	checkError(err)
	host.StartHost(webSocket)
}

func checkError(err error) {
	if err != nil {
		error_msg := fmt.Sprintf("Fatal error: %s", err.Error())
		println(error_msg)
		// log.Fatal(error_msg)
	}
}
