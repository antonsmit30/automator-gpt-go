package api

import (
	"context"

	openai "github.com/sashabaranov/go-openai"
)

type GPTConnection struct {
	key string
}

func (g *GPTConnection) connect() *openai.Client {
	// connect to openai
	client := openai.NewClient(g.key)
	return client

}

func (g *GPTConnection) send(s string) (string, error) {

	client := g.connect()

	res, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo1106,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "Hello!",
				},
			},
		},
	)

	return res.Choices[0].Message.Content, err

}
