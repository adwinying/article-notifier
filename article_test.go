package main

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/jomei/notionapi"
)

func newMockArticle() Article {
	return Article{
		ID:        "some_id",
		URL:       "some_url",
		Title:     "some_title",
		Excerpt:   "some_excerpt",
		Published: true,
	}
}

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

		expected := newMockArticle()
		result := formatArticlesResponse(articles)

		if len(result) != 1 {
			t.Error("Number of results does not match expected")
		}

		if !reflect.DeepEqual(result[0], expected) {
			t.Errorf("Formatted article does not match expected:\n%#v\n", result[0])
		}
	})
}

type dbClientMock struct {
	queryMock func(
		context.Context,
		notionapi.DatabaseID,
		*notionapi.DatabaseQueryRequest,
	) (*notionapi.DatabaseQueryResponse, error)
}

func (m dbClientMock) Query(
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

		clientMock := dbClientMock{
			queryMock: func(
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

		FetchArticles(clientMock, expectedDbId)
	})

	t.Run("Returns error from dbClient", func(t *testing.T) {
		dbId := notionapi.DatabaseID("some_db_id")
		expected := fmt.Errorf("")

		clientMock := dbClientMock{
			queryMock: func(
				ctx context.Context,
				dbId notionapi.DatabaseID,
				req *notionapi.DatabaseQueryRequest,
			) (*notionapi.DatabaseQueryResponse, error) {
				return nil, expected
			},
		}

		_, result := FetchArticles(clientMock, dbId)

		if result != expected {
			t.Errorf("Error from dbClient is not returned:\n%#v\n", result)
		}
	})
}

func TestPickRandomArticle(t *testing.T) {
	t.Run("Returns error if array empty", func(t *testing.T) {
		articles := []Article{}

		_, err := PickRandomArticle(articles)

		if err == nil {
			t.Errorf("Expected error from PickRandomArticle")
		}
	})

	t.Run("Returns a single article", func(t *testing.T) {
		articles := []Article{{}}

		result, _ := PickRandomArticle(articles)

		if result != &articles[0] {
			t.Errorf("Returned article does not match expected")
		}
	})
}

type pgClientMock struct {
	updateMock func(
		context.Context,
		notionapi.PageID,
		*notionapi.PageUpdateRequest,
	) (*notionapi.Page, error)
}

func (m pgClientMock) Update(
	ctx context.Context,
	pgId notionapi.PageID,
	req *notionapi.PageUpdateRequest,
) (*notionapi.Page, error) {
	return m.updateMock(ctx, pgId, req)
}

func TestMarkArticleRead(t *testing.T) {
	t.Run("Check input params", func(t *testing.T) {
		pubId := notionapi.ObjectID("some_pub_id")
		expectedArticle := newMockArticle()
		expectedCtx := context.Background()
		expectedReq := notionapi.PageUpdateRequest{
			Properties: notionapi.Properties{
				"Published": notionapi.CheckboxProperty{
					ID:       pubId,
					Checkbox: true,
				},
			},
		}

		clientMock := pgClientMock{
			updateMock: func(
				ctx context.Context,
				pgId notionapi.PageID,
				req *notionapi.PageUpdateRequest,
			) (*notionapi.Page, error) {
				if !reflect.DeepEqual(ctx, expectedCtx) {
					t.Errorf("ctx does not match expected:\n%#v\n", ctx)
				}

				if !reflect.DeepEqual(pgId.String(), expectedArticle.ID.String()) {
					t.Errorf("pgId does not match expected:\n%#v\n", pgId)
				}

				if !reflect.DeepEqual(*req, expectedReq) {
					t.Errorf("req does not match expected:\n%#v\n", req)
				}

				return &notionapi.Page{}, nil
			},
		}

		MarkArticleRead(clientMock, pubId, &expectedArticle)
	})

	t.Run("Returns err if occured", func(t *testing.T) {
		err := fmt.Errorf("some err")
		pubId := notionapi.ObjectID("some_pub_id")
		article := newMockArticle()

		clientMock := pgClientMock{
			updateMock: func(
				ctx context.Context,
				pgId notionapi.PageID,
				req *notionapi.PageUpdateRequest,
			) (*notionapi.Page, error) {
				return nil, err
			},
		}

		MarkArticleRead(clientMock, pubId, &article)
	})
}
