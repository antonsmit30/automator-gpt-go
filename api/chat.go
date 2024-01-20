package api

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	b64 "encoding/base64"

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

func (g *Bot) transcribe(s *Message) (*Message, error) {

	// basically recieve a b64string and write it to a file
	// then use the openai transcription api to transcribe it
	// then return the transcription.
	decodedString, err := b64.StdEncoding.DecodeString(s.Message)
	if err != nil {
		fmt.Printf("Error decoding string: %v", err)
		return s, err
	}
	// write to file
	cd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		return s, err
	}
	f_err := os.WriteFile(filepath.Join(cd, "audio", "client", "client.webm"), decodedString, 0644)

	if f_err != nil {
		fmt.Println(f_err)
		return s, f_err
	}

	// use api to transcribe
	client := g.connect()
	ctx := context.Background()
	req := openai.AudioRequest{
		Model:    openai.Whisper1,
		FilePath: filepath.Join(cd, "audio", "client", "client.webm"),
	}
	resp, err := client.CreateTranscription(ctx, req)
	if err != nil {
		fmt.Println(err)
		return s, err
	}
	s.Message = resp.Text
	return s, nil

}

func (g *Bot) create_audio(s string) error {
	// basicall use openai to create a audio file from a string
	c := g.connect()
	ctx := context.Background()
	req := openai.CreateSpeechRequest{
		Model:          openai.TTSModel1,
		Voice:          openai.VoiceNova,
		Input:          s,
		ResponseFormat: openai.SpeechResponseFormatMp3,
	}

	resp, err := c.CreateSpeech(ctx, req)
	if err != nil {
		fmt.Printf("Error creating audio: %v", err)
		return err
	}

	// copy to file
	cd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current directory: %v", err)
		return err
	}
	_err := copy(resp, filepath.Join(cd, "audio", "server", "server.mp3"))

	if _err != nil {
		fmt.Printf("Error copying file: %v", _err)
		return _err
	}

	return nil

}
