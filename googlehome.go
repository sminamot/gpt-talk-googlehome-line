package main

import (
	"context"

	"github.com/barnybug/go-cast"
	"github.com/evalphobia/google-home-client-go/googlehome"
)

var (
	googleHomeConfig *googlehome.Config
	googleHomeClient *googlehome.Client
)

func googleHomePlay(url string) error {
	return googleHomeClient.Play(url)
}

func isPlaying(ctx context.Context) (bool, error) {
	ip, _ := googleHomeConfig.GetIPv4()
	port := googleHomeConfig.GetPort()
	client := cast.NewClient(ip, port)

	if err := client.Connect(ctx); err != nil {
		return false, err
	}
	defer client.Close()

	return client.IsPlaying(ctx), nil
}
