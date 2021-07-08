package main

import (
	"fmt"
	"os"

	"github.com/atc0005/go-teams-notify/v2"
)

func SetupMessage(article *Article) goteamsnotify.MessageCard {
	msg := goteamsnotify.NewMessageCard()
	msg.Title = article.Title
	msg.Text = fmt.Sprintf("<h2>%s</h2>", article.Excerpt)
	msg.Text += fmt.Sprintf("<a href=\"%s\">もっと見る</a>", article.URL)
	msg.ThemeColor = "#34D399"

	return msg
}

func NewTeamsClient() goteamsnotify.API {
	return goteamsnotify.NewClient()
}

func GetWebhookUrl() string {
	return os.Getenv("TEAMS_WEBHOOK_URL")
}

type messageSender interface {
	Send(string, goteamsnotify.MessageCard) error
}

func SendMessage(
	client messageSender,
	webhook string,
	msg goteamsnotify.MessageCard,
) error {
	return client.Send(webhook, msg)
}
