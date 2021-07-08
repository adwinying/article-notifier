package main

import (
	"context"
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
		titleWrapper := obj.Properties["Name"].(*notionapi.PageTitleProperty).Title
		var title string
		if len(titleWrapper) > 0 {
			title = titleWrapper[0].PlainText
		}

		excerptWrapper := obj.Properties["Excerpt"].(*notionapi.RichTextProperty).RichText
		var excerpt string
		if len(excerptWrapper) > 0 {
			excerpt = excerptWrapper[0].PlainText
		}

		publishedWrapper := obj.Properties["Published"].(*notionapi.CheckboxProperty).Checkbox
		published := publishedWrapper.(bool)

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

type articleClient interface {
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

func NewArticleClient() articleClient {
	return NewNotionClient().Database
}

func GetDatabaseId() notionapi.DatabaseID {
	return notionapi.DatabaseID(os.Getenv("NOTION_DB_ID"))
}

func FetchArticles(
	dbClient articleClient,
	dbId notionapi.DatabaseID,
) ([]Article, error) {
	res, err := dbClient.Query(
		context.Background(),
		dbId,
		&notionapi.DatabaseQueryRequest{
			Filter: notionapi.CompoundFilter{
				notionapi.FilterOperatorAND: []notionapi.Filter{
					notionapi.PropertyFilter{
						Property: "Published",
						Checkbox: map[notionapi.Condition]bool{
							notionapi.ConditionEquals: false,
						},
					},
					notionapi.PropertyFilter{
						Property: "Ready",
						Checkbox: map[notionapi.Condition]bool{
							notionapi.ConditionEquals: true,
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

	return formatArticlesResponse(res), nil
}
