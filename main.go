package main

import (
	"fmt"

	"github.com/antonsmit30/automator-gpt-go/api"
	"github.com/gin-gonic/gin"
)

func main() {

	// msg := []byte("Hello World")
	fmt.Println("Starting go websocket server")

	// Set up Router
	r := gin.Default()

	api.StartManager()

	r.GET("/message", api.SocketHandler)

	// Run the api server
	r.Run("127.0.0.1:5000")
}
