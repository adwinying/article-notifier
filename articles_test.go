package main

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/jomei/notionapi"
)

type ArticleDatabaseServiceMock struct {
	notionapi.DatabaseService
}

func (m *ArticleDatabaseServiceMock) Query(
	ctx context.Context,
	id notionapi.DatabaseID,
	query *notionapi.DatabaseQueryRequest,
) (*notionapi.DatabaseQueryResponse, error) {
	return &notionapi.DatabaseQueryResponse{}, nil
}

func TestFormatArticlesResponse(t *testing.T) {
	t.Run("Format 0 articles", func(t *testing.T) {
		articles := &notionapi.DatabaseQueryResponse{
			Results: []notionapi.Page{},
		}

		result := formatArticlesResponse(articles)

		if len(result) != 0 {
			t.Error("Formatted articles not empty")
		}
	})

	t.Run("Format single article", func(t *testing.T) {
		articles := &notionapi.DatabaseQueryResponse{
			Results: []notionapi.Page{
				{
					ID:  "some_id",
					URL: "some_url",
					Properties: notionapi.Properties{
						"Name": &notionapi.PageTitleProperty{
							Title: []notionapi.RichText{
								{PlainText: "some_title"},
							},
						},
						"Excerpt": &notionapi.RichTextProperty{
							RichText: []notionapi.RichText{
								{PlainText: "some_excerpt"},
							},
						},
						"Published": &notionapi.CheckboxProperty{
							Checkbox: true,
						},
					},
				},
			},
		}

		expected := Article{
			ID:        "some_id",
			URL:       "some_url",
			Title:     "some_title",
			Excerpt:   "some_excerpt",
			Published: true,
		}

		result := formatArticlesResponse(articles)

		if len(result) != 1 {
			t.Error("Number of results does not match expected")
		}

		if !reflect.DeepEqual(result[0], expected) {
			t.Errorf("Formatted article does not match expected:\n%#v\n", result[0])
		}
	})
}

type articleClientMock struct {
	queryMock func(
		context.Context,
		notionapi.DatabaseID,
		*notionapi.DatabaseQueryRequest,
	) (*notionapi.DatabaseQueryResponse, error)
}

func (m articleClientMock) Query(
	ctx context.Context,
	dbId notionapi.DatabaseID,
	req *notionapi.DatabaseQueryRequest,
) (*notionapi.DatabaseQueryResponse, error) {
	return m.queryMock(ctx, dbId, req)
}

func TestFetchArticles(t *testing.T) {
	t.Run("Check dbClient input params", func(t *testing.T) {
		expectedCtx := context.Background()
		expectedDbId := notionapi.DatabaseID("some_db_id")
		expectedFilter := notionapi.CompoundFilter{
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
		}
		expectedSort := []notionapi.SortObject{
			{
				Property:  "Updated",
				Timestamp: notionapi.TimestampLastEdited,
				Direction: notionapi.SortOrderDESC,
			},
		}

		articleMock := articleClientMock{
			func(
				ctx context.Context,
				dbId notionapi.DatabaseID,
				req *notionapi.DatabaseQueryRequest,
			) (*notionapi.DatabaseQueryResponse, error) {
				if !reflect.DeepEqual(ctx, expectedCtx) {
					t.Errorf("ctx does not match expected:\n%#v\n", ctx)
				}

				if !reflect.DeepEqual(dbId, expectedDbId) {
					t.Errorf("dbId does not match expected:\n%#v\n", dbId)
				}

				if !reflect.DeepEqual(req.Filter, expectedFilter) {
					t.Errorf("filter does not match expected:\n%#v\n", req.Filter)
				}

				if !reflect.DeepEqual(req.Sorts, expectedSort) {
					t.Errorf("sort does not match expected:\n%#v\n", req.Sorts)
				}

				return &notionapi.DatabaseQueryResponse{}, nil
			},
		}

		FetchArticles(articleMock, expectedDbId)
	})

	t.Run("Returns error from dbClient", func(t *testing.T) {
		dbId := notionapi.DatabaseID("some_db_id")
		expected := fmt.Errorf("")

		articleMock := articleClientMock{
			func(
				ctx context.Context,
				dbId notionapi.DatabaseID,
				req *notionapi.DatabaseQueryRequest,
			) (*notionapi.DatabaseQueryResponse, error) {
				return nil, expected
			},
		}

		_, result := FetchArticles(articleMock, dbId)

		if result != expected {
			t.Errorf("Error from dbClient is not returned:\n%#v\n", result)
		}
	})
}
