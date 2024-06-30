package service

import (
	"context"

	openai "github.com/sashabaranov/go-openai"
)

type OpenaiService struct {
	Client            *openai.Client
	SystemRoleMessage *string
}

func (s *OpenaiService) TranformTextBodyToJSON(text string) (string, error) {
	resp, err := s.Client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4o,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are a helpful assistant.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: text,
				},
			},
		})
	if err != nil {
		return "", err
	}

	completion := resp.Choices[0].Message.Content

	return completion, nil
}
