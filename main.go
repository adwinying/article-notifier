package main

import (
	"encoding/json"
	"log"

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

	articleClient := NewArticleClient()
	dbId := GetDatabaseId()

	articles, err := FetchArticles(articleClient, dbId)
	if err != nil {
		log.Fatalf("FetchArticles() failed with error: %s\n", err)
	}
	if len(articles) == 0 {
		log.Fatalln("No available articles found.")
	}

	log.Println("Picking random article...")
	article, err := PickRandomArticle(articles)
	if err != nil {
		log.Fatalf("PickRandomArticle() failed with error: %s\n", err)
	}

	// @TODO debug
	json, _ := json.MarshalIndent(article, "", "  ")
	log.Println(string(json))
}
