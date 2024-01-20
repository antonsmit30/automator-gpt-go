package api

import (
	"io"
	"os"

	"github.com/gorilla/websocket"
)

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

// function that copies src from io.Reader to file using io.Copy. takes in src, dst path
func copy(src io.Reader, dst string) error {
	// create file, if exists overwrite
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	// close file
	defer out.Close()
	// copy src to dst
	_, err = io.Copy(out, src)
	return err
}
