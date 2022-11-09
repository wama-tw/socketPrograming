package main

import (
	"client/src/controllers"

	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {
	r := gin.Default()
	r.LoadHTMLFiles(
		"src/views/index.html",
	)
	r.GET("/room", controllers.SendHTML)
	r.GET("/room/ws", controllers.ConnectWebSocket)

	return r
}

func main() {
	r := setupRouter()
	// Listen and Server in 0.0.0.0:8000
	r.Run(":8000")
}
