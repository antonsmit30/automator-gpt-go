package api

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	openai "github.com/sashabaranov/go-openai"
)

type Bot struct {
	key      string
	messages []openai.ChatCompletionMessage
}

var g = Bot{
	os.Getenv("OPENAI_API_KEY"),
	[]openai.ChatCompletionMessage{},
}

func (g *Bot) setupSystemMessage() {
	cd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}
	f, err := os.ReadFile(filepath.Join(cd, "files", "masterbot.txt"))
	if err != nil {
		fmt.Println(err)
	}
	fstring := string(f)
	g.messages = []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: fstring,
		},
	}
}

func (g *Bot) addMessage(m openai.ChatCompletionMessage) {
	g.messages = append(g.messages, m)
}

func (g *Bot) getMessages() []openai.ChatCompletionMessage {
	return g.messages
}

func (g *Bot) connect() *openai.Client {
	// connect to openai
	client := openai.NewClient(g.key)
	return client

}

func (g *Bot) send(s string) (string, error) {

	client := g.connect()

	fmt.Printf("Full messages: %v\n", g.getMessages())

	// create a bot message and append to full message
	message := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: s,
	}
	g.addMessage(message)

	res, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    openai.GPT3Dot5Turbo1106,
			Messages: g.getMessages(),
		},
	)

	// append bot message to full messages
	g.addMessage(
		openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: res.Choices[0].Message.Content,
		},
	)

	return res.Choices[0].Message.Content, err

}
