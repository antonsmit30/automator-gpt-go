package main

import (
	"fmt"
	"time"

	"github.com/antonsmit30/automator-gpt-go/api"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func main() {

	// msg := []byte("Hello World")
	msg := api.Message{Msg: "Hello World"}
	fmt.Println(msg)

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	// Set up Router
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Pong",
		})
	})

	// Connection setup to websocket
	r.GET("/connect", func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			c.JSON(500, gin.H{
				"message": "Error",
			})
			return
		}
		defer conn.Close()
		for {
			conn.WriteMessage(websocket.TextMessage, []byte("client connected."))
			time.Sleep(5 * time.Second)
		}
	})
	r.Run("127.0.0.1:5000")
}
