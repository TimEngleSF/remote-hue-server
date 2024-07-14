package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/amimof/huego"
	openai "github.com/sashabaranov/go-openai"
)

type OpenaiService struct {
	Client            *openai.Client
	SystemRoleMessage *string
}

func SystemRoleMessage(groups []huego.Group, groupNames GroupNames) string {
	message := fmt.Sprintf(`
Given the following action options seperated by new lines you are to convert natural langauge text about Hue light groups into JSON.
'''
status
'''

Requests should refer to one of the following groups or all groups:
'''
%v
'''

Here is an example of a status request and the expected JSON you should respond with:
    request: 
    "What is the status of %v?"
    response:
    {"type": "status", "data": {"room": ["%v"]}}

    request:
    "What is the status of all groups?"
    response:
    {"type": "status", "data": {"room": %v}}


    Your response should just be the json string not wrapped in any other text.

`, groupNames.String(), strings.ToLower(groupNames[0]), groupNames[0], groupNames.ArrayString())
	return message
}

func (s *OpenaiService) TranformTextBodyToJSON(systemRoleMessage, userMessage string) (string, error) {
	resp, err := s.Client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4o,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemRoleMessage,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: userMessage,
				},
			},
		})
	if err != nil {
		return "", err
	}

	completion := resp.Choices[0].Message.Content

	return completion, nil
}
