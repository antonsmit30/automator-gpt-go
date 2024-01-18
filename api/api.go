package api

import (
	"fmt"

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

type Message struct {
	Message     string `json:"message"`
	EnableAudio bool   `json:"enable_audio"`
}

func HandleMessage(msg Message) {
	fmt.Println(msg)
}

type ClientManager struct {
	// client map stores
	clients map[*Client]bool

	// Webside message
	broadcast chan []byte

	// long connection client
	register chan *Client

	// cancelled client
	unregister chan *Client
}

type Client struct {
	id     string
	socket *websocket.Conn
	send   chan []byte
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
			fmt.Printf("Sending message to all clients: %s\n", message)

			// here we want to send the message to openai
			// convert message bytes to string
			messageString := string(message)
			clientMessage := Message{}
			json.Unmarshal([]byte(messageString), &clientMessage)

			// send to openai
			openaiMessage, err := g.send(string(clientMessage.Message))
			if err != nil {
				fmt.Printf("Error sending message to openai: %v", err)

			}

			newMessage, _ := json.Marshal(&Message{Message: openaiMessage, EnableAudio: true})
			for conn := range manager.clients {
				select {
				case conn.send <- newMessage:
				default:
					close(conn.send)
					delete(manager.clients, conn)
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
		_, message, err := c.socket.ReadMessage()
		if err != nil {
			fmt.Printf("read: %v", err)
			manager.unregister <- c
			_ = c.socket.Close()
			break
		}
		// if no error put info in broadcast channel
		jsonMessage, _ := json.Marshal(&Message{Message: string(message)})
		manager.broadcast <- jsonMessage
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
	// TODO: add check origin
}

func SocketHandler(c *gin.Context) {

	// Upgrade initial request to a websocket
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
