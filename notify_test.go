package main

import (
	"fmt"
	"reflect"
	"testing"

	goteamsnotify "github.com/atc0005/go-teams-notify/v2"
)

func TestSetupMessage(t *testing.T) {
	t.Run("Returns formatted MessageCard", func(t *testing.T) {
		article := Article{
			ID:        "some_id",
			URL:       "some_url",
			Title:     "some_title",
			Excerpt:   "some_excerpt",
			Published: true,
		}

		expected := goteamsnotify.MessageCard{
			Type:    "MessageCard",
			Context: "https://schema.org/extensions",
			Title:   article.Title,
			Text: fmt.Sprintf(
				"<h2>%s</h2><a href=\"%s\">もっと見る</a>",
				article.Excerpt,
				article.URL,
			),
			ThemeColor: "#34D399",
		}

		result := SetupMessage(&article)

		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Formatted MessageCard does not match expected:\n%#v\n", result)
		}
	})
}

type messageSenderMock struct {
	sendMock func(string, goteamsnotify.MessageCard) error
}

func (m messageSenderMock) Send(
	webhook string,
	msg goteamsnotify.MessageCard,
) error {
	return m.sendMock(webhook, msg)
}

func TestSendMessage(t *testing.T) {
	t.Run("Check input params", func(t *testing.T) {
		expectedWebhook := "some_url"
		expectedMsg := goteamsnotify.MessageCard{
			Title:      "some_title",
			Text:       "some_text",
			ThemeColor: "some_color",
		}

		client := messageSenderMock{
			sendMock: func(
				webhook string,
				msg goteamsnotify.MessageCard,
			) error {
				if webhook != expectedWebhook {
					t.Errorf("webhook param does not match expected:\n%#v\n", webhook)
				}

				if msg.Title != expectedMsg.Title ||
					msg.Text != expectedMsg.Text ||
					msg.ThemeColor != expectedMsg.ThemeColor {
					t.Errorf("msg param does not match expected:\n%#v\n", msg)
				}

				return nil
			},
		}

		SendMessage(client, expectedWebhook, expectedMsg)
	})

	t.Run("Returns error if occured", func(t *testing.T) {
		err := fmt.Errorf("some err")
		expectedWebhook := "some_url"
		expectedMsg := goteamsnotify.MessageCard{
			Title:      "some_title",
			Text:       "some_text",
			ThemeColor: "some_color",
		}

		client := messageSenderMock{
			sendMock: func(
				webhook string,
				msg goteamsnotify.MessageCard,
			) error {
				return err
			},
		}

		result := SendMessage(client, expectedWebhook, expectedMsg)

		if result != err {
			t.Errorf("error does not match expected:\n%#v\n", result)
		}
	})
}
