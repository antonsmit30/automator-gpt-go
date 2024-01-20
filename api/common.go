package api

import "github.com/gorilla/websocket"

type Message struct {
	Message     string `json:"message"`
	EnableAudio bool   `json:"enable_audio"`
	BotId       string `json:"bot_id"`
	Type        string `json:"type"`
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
