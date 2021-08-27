package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"

	"github.com/jomei/notionapi"
)

type Article struct {
	ID        notionapi.ObjectID
	URL       string
	Title     string
	Excerpt   string
	Published bool
}

func formatArticlesResponse(res *notionapi.DatabaseQueryResponse) []Article {
	var articles []Article
	for _, obj := range res.Results {
		titleWrapper := obj.Properties["Name"].(*notionapi.TitleProperty).Title
		var title string
		if len(titleWrapper) > 0 {
			title = titleWrapper[0].PlainText
		}

		excerptWrapper := obj.Properties["Excerpt"].(*notionapi.RichTextProperty).RichText
		var excerpt string
		if len(excerptWrapper) > 0 {
			excerpt = excerptWrapper[0].PlainText
		}

		published := obj.Properties["Published"].(*notionapi.CheckboxProperty).Checkbox

		articles = append(articles, Article{
			ID:        obj.ID,
			URL:       obj.URL,
			Title:     title,
			Excerpt:   excerpt,
			Published: published,
		})
	}

	return articles
}

type dbClient interface {
	Query(
		context.Context,
		notionapi.DatabaseID,
		*notionapi.DatabaseQueryRequest,
	) (*notionapi.DatabaseQueryResponse, error)
}

func NewNotionClient() *notionapi.Client {
	token := notionapi.Token(os.Getenv("NOTION_API_TOKEN"))

	return notionapi.NewClient(token)
}

func GetDatabaseId() notionapi.DatabaseID {
	return notionapi.DatabaseID(os.Getenv("NOTION_DB_ID"))
}

func GetPublishedCheckboxId() notionapi.ObjectID {
	return notionapi.ObjectID(os.Getenv("NOTION_PUBLISHED_CHECKBOX_ID"))
}

func FetchArticles(
	dbClient dbClient,
	dbId notionapi.DatabaseID,
) ([]Article, error) {
	res, err := dbClient.Query(
		context.Background(),
		dbId,
		&notionapi.DatabaseQueryRequest{
			CompoundFilter: &notionapi.CompoundFilter{
				notionapi.FilterOperatorAND: {
					{
						Property: "Published",
						Checkbox: &notionapi.CheckboxFilterCondition{
							DoesNotEqual: true,
						},
					},
					{
						Property: "Ready",
						Checkbox: &notionapi.CheckboxFilterCondition{
							Equals: true,
						},
					},
				},
			},

			Sorts: []notionapi.SortObject{
				{
					Property:  "Updated",
					Timestamp: notionapi.TimestampLastEdited,
					Direction: notionapi.SortOrderDESC,
				},
			},
		},
	)

	if err != nil {
		return nil, err
	}
	if os.Getenv("DEBUG") == "true" {
		json, _ := json.MarshalIndent(res, ">", "  ")
		fmt.Println(string(json))
	}

	return formatArticlesResponse(res), nil
}

func PickRandomArticle(articles []Article) (*Article, error) {
	if len(articles) == 0 {
		return nil, fmt.Errorf("Given array is empty")
	}

	randIndex := rand.Intn(len(articles))

	return &articles[randIndex], nil
}

type pageClient interface {
	Update(
		context.Context,
		notionapi.PageID,
		*notionapi.PageUpdateRequest,
	) (*notionapi.Page, error)
}

func MarkArticleRead(
	pgClient pageClient,
	pubdId notionapi.ObjectID,
	article *Article,
) error {
	_, err := pgClient.Update(
		context.Background(),
		notionapi.PageID(article.ID),
		&notionapi.PageUpdateRequest{
			Properties: notionapi.Properties{
				"Published": notionapi.CheckboxProperty{
					ID:       pubdId,
					Checkbox: true,
				},
			},
		},
	)

	if err != nil {
		return err
	}

	return nil
}
