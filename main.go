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

	// allow cors
	r.Use(api.CORSMiddleware())

	api.StartManager()

	// Handle websocket connection
	r.GET("/server", api.MessageSocketHandler)

	// Run the api server
	r.Run("127.0.0.1:5000")
}
