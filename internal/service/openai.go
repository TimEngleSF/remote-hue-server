package service

import (
	"context"
	"fmt"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

type OpenaiService struct {
	Client            *openai.Client
	SystemRoleMessage *string
}

func SystemRoleMessage(groups Groups, groupNames GroupNames) string {
	message := fmt.Sprintf(`
Given the following action options separated by new lines you are to convert natural language text about Hue light groups into JSON.
'''
status
update
'''

Requests should refer to one of the following groups or all groups:
'''
%v
'''

Here is information on each of the groups that you may use to help create meaningful json responses:
'''
%v
'''

%v

%v

Your response should just be the JSON string not wrapped in any other text.
`, groupNames.String(), fmt.Sprintf("%+v\n", groups), statusExamples(groupNames), updateExamples(groups))
	return message
}

func statusExamples(groupNames GroupNames) string {
	example := fmt.Sprintf(`
Here is an example of a status request and the expected JSON you should respond with:
    request: 
    "What is the status of %v?"
    response:
    {"type": "status", "data": {"room": ["%v"]}}

    request:
    "What is the status of all groups?"
    response:
    {"type": "status", "data": {"room": %v}}
`, strings.ToLower(groupNames[0]), groupNames[0], groupNames.ArrayString())
	return example
}

func updateExamples(groups Groups) string {
	example := fmt.Sprintf(`
Here is an example of an update request to turn groups on or off and the expected JSON you should respond with:
  NOTES: 
  - brightness is optional and should be set to 254 if not provided.
  - if "isOn" is false do not include brightness
  - if asked to turn up or down brightness do so on increments of 25% with 0 being off and 254 being full brightness.
    request:
    "Please turn %v on."
    response:
    {"type": "update", "data": {"group": "%v", "isOn": true, "brightness": 254}}

    request:
    "Please turn %v off."
    response:
    {"type": "update", "data": {"group": "%v", "isOn": false}}
    
    request:
    -Note: act as if current brightness is 254
    "Please turn down brightness of kitchen"
    "{"type": "update", "data": {"group": "kitchen", "isOn": true, "brightness": 191}}"
    
    `, strings.ToLower(groups[0].Name), groups[0].Name, strings.ToLower(groups[0].Name), groups[0].Name)
	return example
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

	completion := CleanGPTResponse(resp.Choices[0].Message.Content)

	return completion, nil
}

// Remove the triple backticks and the "json" keyword if they exist
func CleanGPTResponse(gptResponse string) string {
	cleaned := strings.ReplaceAll(gptResponse, "```json", "")
	cleaned = strings.ReplaceAll(cleaned, "```", "")
	return strings.TrimSpace(cleaned)
}
