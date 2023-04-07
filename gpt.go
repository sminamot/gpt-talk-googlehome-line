package main

import (
	"context"
	"path/filepath"

	"github.com/sashabaranov/go-openai"
)

var openAIClient *openai.Client

func createTranscription(ctx context.Context, path string) (string, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	req := openai.AudioRequest{
		Model:    openai.Whisper1,
		FilePath: path,
	}
	resp, err := openAIClient.CreateTranscription(ctx, req)
	if err != nil {
		return "", err
	}
	return resp.Text, nil
}

func chatCompletion(ctx context.Context, text string) (string, error) {
	req := openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: text,
			},
		},
	}
	resp, err := openAIClient.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Message.Content, nil

}
