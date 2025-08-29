package discord

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

func NewOpenAIClient(apiKey string, maxTokens int, temperature float64) *OpenAIClient {
	client := openai.NewClient(apiKey)
	return &OpenAIClient{
		client:      client,
		maxTokens:   maxTokens,
		temperature: float32(temperature),
	}
}

func (o *OpenAIClient) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	resp, err := o.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			MaxTokens:   o.maxTokens,
			Temperature: o.temperature,
		},
	)

	if err != nil {
		return "", fmt.Errorf("ChatCompletion error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	return resp.Choices[0].Message.Content, nil
}
