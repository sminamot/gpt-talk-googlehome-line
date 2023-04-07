package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

func getVoice(text string) (string, error) {
	v := url.Values{}
	v.Add("text", text)
	v.Add("speaker", "hikari")
	v.Add("speed", "120")
	v.Add("format", "mp3")
	v.Add("volume", "200")

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "https://api.voicetext.jp/v1/tts", nil)
	if err != nil {
		log.Fatal(err)
	}

	req.URL.RawQuery = v.Encode()
	req.SetBasicAuth(os.Getenv("VOICETEXT_API_KEY"), "")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("failed to request VoiceText Web API, status:%d", resp.StatusCode)
	}

	name := fmt.Sprintf("%s.mp3", uuid.NewString())
	f, err := os.Create(filepath.Join("output", name))
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := io.Copy(f, resp.Body); err != nil {
		return "", err
	}
	return f.Name(), nil
}
