package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/atc0005/go-teams-notify/v2"
	"github.com/joho/godotenv"
)

func main() {
	// load .env
	log.Println("Loading .env...")
	err := godotenv.Load()
	if err != nil {
		log.Fatalln("Error loading .env file")
	}

	log.Println("Fetching articles...")
	notionClient := NewNotionClient()
	dbClient := notionClient.Database
	dbId := GetDatabaseId()

	articles, err := FetchArticles(dbClient, dbId)
	if err != nil {
		log.Fatalf("FetchArticles() failed with error: %s\n", err)
	}

	var msg goteamsnotify.MessageCard
	var article *Article

	if len(articles) > 0 {
		log.Println("Picking random article...")
		article, err = PickRandomArticle(articles)
		if err != nil {
			log.Fatalf("PickRandomArticle() failed with error: %s\n", err)
		}
		if os.Getenv("DEBUG") == "true" {
			json, _ := json.MarshalIndent(article, ">", "  ")
			log.Println(string(json))
		}

		log.Println("Composing webhook message...")
		msg = SetupMessage(article)

	} else {
		log.Println("No available articles found. Generating apology...")
		msg = GetUnavailableMessage()
	}

	log.Println("Triggering webhook...")
	teamsClient := NewTeamsClient()
	url := GetWebhookUrl()
	err = SendMessage(teamsClient, url, msg)
	if err != nil {
		log.Fatalf("SendMessage() failed with error: %s\n", err)
	}

	if len(articles) == 0 {
		log.Fatalln("Apology sent.")
	}

	log.Println("Marking article as read...")
	pgClient := notionClient.Page
	pubdCheckboxId := GetPublishedCheckboxId()
	err = MarkArticleRead(pgClient, pubdCheckboxId, article)
	if err != nil {
		log.Fatalf("MarkArticleRead() failed with error: %s\n", err)
	}

	log.Println("Success!")
}
