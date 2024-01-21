package api

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"encoding/base64"
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var manager = ClientManager{
	broadcast:  make(chan []byte),
	register:   make(chan *Client),
	unregister: make(chan *Client),
	clients:    make(map[*Client]bool),
}

func (manager *ClientManager) start() {
	// initialize openai messages
	g.setupSystemMessage()
	// infinite loop
	for {
		// select blocks goroutine until one of its cases can run then runs it
		select {

		// if there is a connection, pass the connection
		case conn := <-manager.register:
			manager.clients[conn] = true
			jsonMessage, _ := json.Marshal(&Message{Message: "/A new socket has connected."})
			manager.send(jsonMessage, conn)

			// If disconnected
		case conn := <-manager.unregister:
			// if connection is true, switch off
			fmt.Print("switching off")
			if _, ok := manager.clients[conn]; ok {
				fmt.Print("switching off gracefully and deferring close")
				defer close(conn.send)
			}

		// broadcast
		case message := <-manager.broadcast:
			fmt.Printf("Broadcast message received: %s\n", message)

			// here we want to print out the message contents as
			// we might want to do different things to it!
			messageString := string(message)
			fmt.Printf("Message string: %v\n", messageString)

			clientMessage := Message{}
			json.Unmarshal([]byte(messageString), &clientMessage)
			fmt.Printf("Client: %v\n", clientMessage.Message)

			// We basically want to convert to audio first, i.e transcribe using openai
			fmt.Printf("Bot id: %v\n", clientMessage.BotId)
			if clientMessage.Type == "audio" {

				resp, err := g.transcribe(&clientMessage)
				if err != nil {
					fmt.Printf("Error transcribing audio: %v", err)
				}
				fmt.Printf("Transcribed audio: %v", resp)
			}

			// send to openai
			openaiMessage, err := g.send(string(clientMessage.Message))
			if err != nil {
				fmt.Printf("Error sending message to openai: %v", err)

			}

			newMessage, _ := json.Marshal(&Message{Message: openaiMessage, EnableAudio: true, BotId: "masterbot", Type: "text"})
			for conn := range manager.clients {
				select {
				case conn.send <- newMessage:
				default:
					close(conn.send)
					delete(manager.clients, conn)
				}
			}

			// if enable audio create an audio file
			if clientMessage.EnableAudio {
				// create audio file
				err := g.create_audio(openaiMessage)
				if err != nil {
					fmt.Printf("Error creating audio file: %v", err)
				}
				//read in audio file and base64 encode
				cd, err := os.Getwd()
				if err != nil {
					fmt.Printf("Error getting working directory in client: %v", err)
				}
				audio, err := os.ReadFile(filepath.Join(cd, "audio", "server", "server.mp3"))
				if err != nil {
					fmt.Printf("Error reading audio file: %v", err)
				}
				b64s := base64.StdEncoding.EncodeToString(audio)
				audioMessage, _ := json.Marshal(&Message{Message: b64s, EnableAudio: true, BotId: "masterbot", Type: "audio"})

				for conn := range manager.clients {
					select {
					case conn.send <- audioMessage:
					default:
						close(conn.send)
						delete(manager.clients, conn)
					}
				}
			}

		}
	}
}

// Define func for manager to send a message
func (manager *ClientManager) send(message []byte, ignore *Client) {
	for conn := range manager.clients {
		// if connection is not ignored, send message
		if conn != ignore {
			conn.send <- message
		}
	}
}

// Read method
func (c *Client) read() {
	// close connection when function returns
	defer func() {
		manager.unregister <- c
		_ = c.socket.Close()
	}()

	for {
		// Read message from the client
		// m := Message{}
		_, message, err := c.socket.ReadMessage()
		if err != nil {
			fmt.Printf("read: %v", err)
			manager.unregister <- c
			_ = c.socket.Close()
			break
		}

		// jsonMessage, _ := json.Marshal(&Message{Message: string(message)}) // bad bad bad code! its already json encoded
		manager.broadcast <- message
	}
}

// Write method
func (c *Client) write() {
	// close connection when function returns
	defer func() {
		_ = c.socket.Close()
	}()

	for {
		select {
		// Read the message from send method
		case message, ok := <-c.send:
			if !ok {
				fmt.Println("Closing socket as issue with send")
				_ = c.socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Write message to the client
			_ = c.socket.WriteMessage(websocket.TextMessage, message)
		}
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// TODO: modify check origin to only allow certain origins
	CheckOrigin: func(r *http.Request) bool {
		fmt.Printf("Origin: %v\n", r.Header.Get("Origin"))
		return true
	},
}

func MessageSocketHandler(c *gin.Context) {

	// Upgrade initial request to a websocket connection
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Printf("socketHandler: %v", err)
		c.JSON(500, gin.H{
			"message": "Error",
		})
	}
	// all connections open a client
	client := &Client{id: uuid.NewString(), socket: conn, send: make(chan []byte)}
	// register client
	manager.register <- client
	//start goroutines
	go client.read()
	go client.write()

}

func StartManager() {
	go manager.start()
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
