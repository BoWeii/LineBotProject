package main

import (
	"log"
	"net/http"

	"github.com/line/line-bot-sdk-go/linebot"
)

func main() {
	bot, err := linebot.New(
		"b7c8b7a613f8eaf0e2117c67ea33d69c",
		"Vk1uRYNtSh5DZBTZMduCHf3gnMZAgK5cVB9/Q/FGu99Lf9JwKJWyO/GX+QIBjN/XorzCGgucHGVcwsBkrSlzlEiyINmzgIhFGw79QHxhQpw7EMept2st4a2POYtWS4rTy4TrYV9syPo7+6AHyh+2uwdB04t89/1O/w1cDnyilFU=",
	)
	if err != nil {
		log.Fatal(err)
	}
	// Setup HTTP Server for receiving requests from LINE platform
	http.HandleFunc("/callback", func(w http.ResponseWriter, req *http.Request) {
		events, err := bot.ParseRequest(req)
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
				case *linebot.TextMessage:
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(message.Text)).Do(); err != nil {
						log.Print(err)

					}
				}
			}
		}
	})
	// This is just sample code.
	// For actual use, you must support HTTPS by using `ListenAndServeTLS`, a reverse proxy or something else.
	if err := http.ListenAndServe(":3096", nil); err != nil {
		log.Fatal(err)
	}
}
