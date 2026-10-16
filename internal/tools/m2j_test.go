package tools

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/snowray712000/FHL-MCP-Go/internal/fhlapi"
)

// articlesAPI answers json.php from recorded upstream fixtures (or a canned response) and records the queries.
type articlesAPI struct {
	t      *testing.T
	got    []fhlapi.ArticleQuery
	canned *fhlapi.ArticlesResponse
}

func (f *articlesAPI) Articles(_ context.Context, q fhlapi.ArticleQuery) (*fhlapi.ArticlesResponse, error) {
	f.got = append(f.got, q)
	if f.canned != nil {
		return f.canned, nil
	}
	name := "articles_none.json"
	switch {
	case q.Title == "愛" && q.Simplified:
		name = "articles_title_ai_gb1.json"
	case q.Title == "愛":
		name = "articles_title_ai.json"
	case q.Title != "zzqqxx":
		f.t.Fatalf("unexpected article query %+v", q)
	}
	var out fhlapi.ArticlesResponse
	return &out, json.Unmarshal(fixture(f.t, "upstream", name), &out)
}

func (f *articlesAPI) QB(context.Context, fhlapi.QBRequest) (*fhlapi.QBResponse, error) {
	return nil, errors.New("QB not expected in this test")
}

func (f *articlesAPI) AB(context.Context, bool) (*fhlapi.ABResponse, error) {
	return nil, errors.New("AB not expected in this test")
}

func (f *articlesAPI) SE(context.Context, fhlapi.SERequest) (*fhlapi.SEResponse, error) {
	return nil, errors.New("SE not expected in this test")
}

func (f *articlesAPI) SD(context.Context, bool, int, bool) (*fhlapi.SDResponse, error) {
	return nil, errors.New("SD not expected in this test")
}

func (f *articlesAPI) QP(context.Context, int, int, int, bool) (*fhlapi.QPResponse, error) {
	return nil, errors.New("QP not expected in this test")
}

func (f *articlesAPI) SC(context.Context, int, int, *int, bool) (*fhlapi.SCResponse, error) {
	return nil, errors.New("SC not expected in this test")
}

func (f *articlesAPI) SCList(context.Context, bool) (*fhlapi.SCListResponse, error) {
	return nil, errors.New("SCList not expected in this test")
}

func (f *articlesAPI) SSC(context.Context, string, bool) (*fhlapi.SSCResponse, error) {
	return nil, errors.New("SSC not expected in this test")
}

func (f *articlesAPI) ST(context.Context, int, string, bool, bool) (*fhlapi.STResponse, error) {
	return nil, errors.New("ST not expected in this test")
}

func (f *articlesAPI) AU(context.Context, int, int, int, bool) (*fhlapi.AUResponse, error) {
	return nil, errors.New("AU not expected in this test")
}

func (f *articlesAPI) QSub(context.Context, int, int, *string) (*fhlapi.SubVerseResponse, error) {
	return nil, errors.New("QSub not expected in this test")
}

func (f *articlesAPI) QAF(context.Context, int, int, *string) (*fhlapi.SubVerseResponse, error) {
	return nil, errors.New("QAF not expected in this test")
}

func (f *articlesAPI) SESub(context.Context, string, int, int) (*fhlapi.SubSearchResponse, error) {
	return nil, errors.New("SESub not expected in this test")
}

func (f *articlesAPI) SEAF(context.Context, string, int, int) (*fhlapi.SubSearchResponse, error) {
	return nil, errors.New("SEAF not expected in this test")
}

func (f *articlesAPI) RT(context.Context, int, int, bool) (*fhlapi.RTResponse, error) {
	return nil, errors.New("RT not expected in this test")
}

// Goldens come from running the Python reference's own search_articles / http_search_articles code
// (tests/oracle/articles_oracle.py) on the recorded fixtures, with the one approved fix of D25 (text read
// from "txt"). list_fhl_article_columns is the live M0 sample.
func TestM2jGoldens(t *testing.T) {
	cases := []struct{ tool, args, golden string }{
		{"search_fhl_articles", `{"title":"愛","limit":2}`, "search_fhl_articles_preview.json"},
		{"search_fhl_articles", `{"title":"愛","limit":"2","include_content":false,"abstract":null}`, "search_fhl_articles_preview.json"},
		{"search_fhl_articles", `{"title":"愛","limit":2,"include_content":true}`, "search_fhl_articles_full.json"},
		{"search_fhl_articles", `{"title":"愛","limit":1,"use_simplified":true}`, "search_fhl_articles_simplified.json"},
		{"search_fhl_articles", `{"title":"愛","limit":10}`, "search_fhl_articles_all3.json"},
		{"search_fhl_articles", `{"title":"愛"}`, "search_fhl_articles_all3.json"},
		{"search_fhl_articles", `{"title":"zzqqxx"}`, "search_fhl_articles_none.json"},
		{"list_fhl_article_columns", `{}`, "list_fhl_article_columns.json"},
	}
	for _, c := range cases {
		t.Run(c.golden+" "+c.args, func(t *testing.T) {
			got, err := callOn(t, &articlesAPI{t: t}, c.tool, c.args)
			if err != nil {
				t.Fatal(err)
			}
			if want := string(fixture(t, "golden", c.golden)); got != want {
				t.Errorf("differs from the golden:\n got %.500s\nwant %.500s", got, want)
			}
		})
	}
}

func TestArticleQuery(t *testing.T) {
	api := &articlesAPI{t: t, canned: &fhlapi.ArticlesResponse{Status: 0, Result: "no data"}}
	// An empty condition is absent, like Python's truthiness test; use_simplified reaches the site (D4).
	args := `{"title":"愛","author":"","column":"women3","pub_date":"2026.10.04","content":"x","abstract":"y","use_simplified":true}`
	if _, err := callOn(t, api, "search_fhl_articles", args); err != nil {
		t.Fatal(err)
	}
	want := fhlapi.ArticleQuery{Title: "愛", Column: "women3", PubDate: "2026.10.04", Content: "x", Abstract: "y", Simplified: true}
	if len(api.got) != 1 || api.got[0] != want {
		t.Errorf("queries = %+v, want %+v", api.got, want)
	}
}

func TestArticleValidation(t *testing.T) {
	none := "Invalid parameter 'search_params': Must provide at least one search parameter (title, author, content, abstract, column, or pub_date)"
	cases := []struct{ name, args, want string }{
		{"no condition", `{}`, none},
		{"only empty strings", `{"title":"","author":""}`, none},
		{"limit 0", `{"title":"愛","limit":0}`, "Invalid parameter 'limit': 0 - must be between 1 and 200"},
		{"limit 201", `{"title":"愛","limit":201}`, "Invalid parameter 'limit': 201 - must be between 1 and 200"},
		{"limit text", `{"title":"愛","limit":"x"}`, "Invalid parameter 'limit': \"x\" - must be an integer"},
		{"bad title type", `{"title":5}`, "Invalid parameter 'title': 5 - must be a string"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			api := &articlesAPI{t: t}
			_, err := callOn(t, api, "search_fhl_articles", c.args)
			if err == nil || err.Error() != c.want {
				t.Fatalf("err = %v\nwant %q", err, c.want)
			}
			if len(api.got) != 0 {
				t.Errorf("must not reach upstream: %+v", api.got)
			}
		})
	}
}

// The site's failure statuses are ordinary results in the Python reference and stay so.
func TestArticleSiteAnswers(t *testing.T) {
	long := strings.Repeat("字", 205)
	rec := func(txt string) fhlapi.ArticleRecord {
		return fhlapi.ArticleRecord{ID: json.RawMessage(`"1"`), Title: json.RawMessage(`"t"`), Txt: json.RawMessage(txt)}
	}
	one := func(txt string) fhlapi.ArticlesResponse {
		return fhlapi.ArticlesResponse{Status: 1, RecordCount: 1, Record: []fhlapi.ArticleRecord{rec(txt)}}
	}
	cases := []struct {
		name   string
		resp   fhlapi.ArticlesResponse
		args   string
		checks []string
		absent []string
	}{
		{"data too much", fhlapi.ArticlesResponse{Status: 0, Result: "Data too much"}, `{"title":"x"}`,
			[]string{`"status": "error"`, `"error": "資料量過大，請提供至少一個搜尋條件"`, `"hint"`}, nil},
		{"no data", fhlapi.ArticlesResponse{Status: 0, Result: "no data"}, `{"title":"x"}`,
			[]string{`"status": "no_results"`, `"message": "未找到符合條件的文章"`}, []string{"query_type"}},
		{"other failure", fhlapi.ArticlesResponse{Status: 0, Result: "boom"}, `{"title":"x"}`,
			[]string{`"status": "error"`, `"error": "boom"`}, []string{"hint"}},
		{"failure without text", fhlapi.ArticlesResponse{Status: 0}, `{"title":"x"}`,
			[]string{`"error": "Unknown error"`}, nil},
		{"success but empty", fhlapi.ArticlesResponse{Status: 1, RecordCount: 0}, `{"title":"x"}`,
			[]string{`"status": "no_results"`, `"query_type": "article_search"`}, nil},
		{"preview is cut at 200 characters and tags are stripped", one(`"<p>` + long + `</p>"`), `{"title":"x"}`,
			[]string{`"content_preview": "` + strings.Repeat("字", 200) + `..."`}, []string{`"content":`, "<p>"}},
		{"short text has no ellipsis", one(`"<b>abc</b>"`), `{"title":"x"}`,
			[]string{`"content_preview": "abc"`}, nil},
		{"no text, no preview", one(`""`), `{"title":"x"}`,
			nil, []string{"content_preview", `"content"`}},
		{"full content keeps the HTML", one(`"<p>hi</p>"`), `{"title":"x","include_content":true}`,
			[]string{`"content": "<p>hi</p>"`}, []string{"content_preview"}},
		{"limit cuts the list and the count", fhlapi.ArticlesResponse{Status: 1, RecordCount: 3, Record: []fhlapi.ArticleRecord{rec(`""`), rec(`""`), rec(`""`)}}, `{"title":"x","limit":2}`,
			[]string{`"total_count": 2`, `"returned_count": 2`}, nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			resp := c.resp
			out, err := callOn(t, &articlesAPI{t: t, canned: &resp}, "search_fhl_articles", c.args)
			if err != nil {
				t.Fatal(err)
			}
			for _, want := range c.checks {
				if !strings.Contains(out, want) {
					t.Errorf("output lacks %s:\n%s", want, out)
				}
			}
			for _, bad := range c.absent {
				if strings.Contains(out, bad) {
					t.Errorf("output must not contain %s:\n%s", bad, out)
				}
			}
		})
	}
}
