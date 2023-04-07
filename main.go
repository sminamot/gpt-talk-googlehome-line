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
	"strings"
	"time"
	"unicode/utf8"

	"github.com/evalphobia/google-home-client-go/googlehome"
	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/sashabaranov/go-openai"
	"github.com/tcolgate/mp3"
)

type audios []audio

type audio struct {
	path string
	url  string
}

func init() {
	var err error
	lineBot, err = linebot.New(
		os.Getenv("CHANNEL_SECRET"),
		os.Getenv("CHANNEL_TOKEN"),
	)
	if err != nil {
		log.Fatal(err)
	}

	openAIAPIKey := os.Getenv("OPENAI_API_KEY")
	if openAIAPIKey == "" {
		log.Fatal("OPENAI_API_KEY is empty")
	}
	openAIClient = openai.NewClient(openAIAPIKey)

	if os.Getenv("VOICETEXT_API_KEY") == "" {
		log.Fatal("VOICETEXT_API_KEY is empty")
	}

	googleHomeConfig = &googlehome.Config{
		Hostname: os.Getenv("GOOGLEHOME_IP"),
		Lang:     "ja",
	}
	googleHomeClient, err = googlehome.NewClientWithConfig(*googleHomeConfig)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	http.HandleFunc("/callback", func(w http.ResponseWriter, req *http.Request) {
		events, err := lineBot.ParseRequest(req)
		if err != nil {
			if err == linebot.ErrInvalidSignature {
				w.WriteHeader(400)
			} else {
				w.WriteHeader(500)
			}
			return
		}
		for _, event := range events {
			if event.Type == linebot.EventTypeMessage {
				switch message := event.Message.(type) {
				case *linebot.AudioMessage:
					ctx := context.Background()
					// save input audio on local
					name, err := saveAudio(message.ID)
					if err != nil {
						log.Println(err)
						w.WriteHeader(500)
						return
					}
					defer os.Remove(name)

					// audio to text
					inputText, err := createTranscription(ctx, name)
					if err != nil {
						log.Println(err)
						w.WriteHeader(500)
						return
					}
					log.Println("input:", inputText)

					// chatCompletion
					outputText, err := chatCompletion(ctx, inputText)
					if err != nil {
						log.Println(err)
						w.WriteHeader(500)
						return
					}
					log.Println("output:", outputText)

					ts := splitText(outputText)
					fmt.Printf("len(ts): %v\n", len(ts))

					var audios audios
					for _, t := range ts {
						// text to audio
						outputAudio, err := getVoice(t)
						if err != nil {
							log.Println(err)
							w.WriteHeader(500)
							return
						}
						//defer os.Remove(outputAudio)
						audios = append(audios, audio{
							path: outputAudio,
							url:  getContentsURL(req.Host, outputAudio),
						})
					}
					if err := audios.playAudio(ctx); err != nil {
						log.Println(err)
						w.WriteHeader(500)
						return
					}
				}
			}
		}
	})

	http.Handle("/voice/", http.StripPrefix("/voice/", http.FileServer(http.Dir("./output"))))

	if err := http.ListenAndServe(":"+os.Getenv("PORT"), nil); err != nil {
		log.Fatal(err)
	}
}

func getContentsURL(host, path string) string {
	fileName := filepath.Base(path)
	u := &url.URL{}
	u.Scheme = "https"
	u.Host = host
	return u.JoinPath("voice", fileName).String()
}

func splitText(s string) (ret []string) {
	s = strings.ReplaceAll(s, "\n", "")
	ss := strings.Split(s, "。")

	sp := ""
	for _, v := range ss {
		if l := utf8.RuneCountInString(sp) + utf8.RuneCountInString(v); l < 200 {
			sp = sp + v + "。"
			continue
		}
		ret = append(ret, sp)
		sp = v
	}
	ret = append(ret, sp)

	return
}

func (as audios) playAudio(ctx context.Context) error {
	for _, a := range as {
		if err := func() error {
			defer a.remove()
			d := a.duration()
			if err := googleHomePlay(a.url); err != nil {
				return err
			}
			time.Sleep(d)

			return nil
		}(); err != nil {
			return err
		}
	}

	return nil
}

func (a *audio) remove() {
	log.Println("delete:", a.path)
	os.Remove(a.path)
}

func (a *audio) duration() (duration time.Duration) {

	r, err := os.Open(a.path)
	if err != nil {
		fmt.Println(err)
		return
	}

	d := mp3.NewDecoder(r)
	var f mp3.Frame
	skipped := 0

	for {

		if err := d.Decode(&f, &skipped); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println(err)
			return
		}

		duration = duration + f.Duration()
	}

	return
}
