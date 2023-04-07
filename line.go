package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/line/line-bot-sdk-go/linebot"
)

var lineBot *linebot.Client

func saveAudio(id string) (string, error) {
	c, err := lineBot.GetMessageContent(id).Do()
	if err != nil {
		return "", err
	}
	defer c.Content.Close()

	name := fmt.Sprintf("%s.m4a", uuid.NewString())

	f, err := os.Create(filepath.Join("input", name))
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := io.Copy(f, c.Content); err != nil {
		return "", err
	}

	return f.Name(), nil
}
