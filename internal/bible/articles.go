package bible

import (
	"context"
	"encoding/json"
	"regexp"
	"strconv"
	"strings"

	"github.com/snowray712000/FHL-MCP-Go/internal/fhlapi"
)

// ArticleAPI is the part of the upstream client the article tools need (www.fhl.net/api/json.php).
type ArticleAPI interface {
	Articles(ctx context.Context, q fhlapi.ArticleQuery) (*fhlapi.ArticlesResponse, error)
}

// Limits of search_fhl_articles' limit parameter (the documented range; Python left it unchecked).
const (
	MinArticleLimit     = 1
	MaxArticleLimit     = 200
	DefaultArticleLimit = 50
)

// ArticleSearch is the input of SearchArticles. An empty condition counts as absent (Python truthiness).
type ArticleSearch struct {
	Title, Author, Content, Abstract, Column, PubDate string
	Limit                                             int
	IncludeContent, Simplified                        bool
}

// Article is one article of the result. Content (full HTML) and ContentPreview are mutually exclusive.
type Article struct {
	ID             json.RawMessage `json:"id"`
	Title          json.RawMessage `json:"title"`
	Author         json.RawMessage `json:"author"`
	PubTime        json.RawMessage `json:"pubtime"`
	Column         json.RawMessage `json:"column"`
	Abstract       json.RawMessage `json:"abstract"`
	Content        json.RawMessage `json:"content,omitempty"`
	ContentPreview *string         `json:"content_preview,omitempty"`
}

// ArticleResults is the successful search_fhl_articles result.
type ArticleResults struct {
	Status             string    `json:"status"`
	QueryType          string    `json:"query_type"`
	TotalCount         int       `json:"total_count"`
	ReturnedCount      int       `json:"returned_count"`
	IncludeFullContent bool      `json:"include_full_content"`
	Articles           []Article `json:"articles"`
}

// ArticleNoResults is the answer when nothing matched (a normal result, not an error).
type ArticleNoResults struct {
	Status    string `json:"status"`
	QueryType string `json:"query_type,omitempty"`
	Message   string `json:"message"`
}

// ArticleFailure is the answer when the site refused the query (also a normal result, like Python).
type ArticleFailure struct {
	Status string `json:"status"`
	Error  string `json:"error"`
	Hint   string `json:"hint,omitempty"`
}

const noArticlesMessage = "未找到符合條件的文章"

var htmlTag = regexp.MustCompile(`<[^>]+>`)

func rawOrEmptyString(raw json.RawMessage) json.RawMessage { return rawOrDefault(raw, `""`) }

// preview is the first 200 characters of the article text without HTML tags; "" when there is no text.
func preview(txt json.RawMessage) (string, bool) {
	if !truthy(rawOr(txt, nil)) {
		return "", false
	}
	var s string
	if err := json.Unmarshal(txt, &s); err != nil {
		s = string(txt)
	}
	clean := []rune(htmlTag.ReplaceAllString(s, ""))
	if len(clean) > 200 {
		return string(clean[:200]) + "...", true
	}
	return string(clean), true
}

// SearchArticles searches the 信望愛 articles. The result is an *ArticleResults, or an *ArticleNoResults /
// *ArticleFailure for the answers Python also returned as ordinary results.
// Unlike Python, the article text is read from the upstream's "txt" field (D25): Python read "content",
// which does not exist, so its preview never appeared and include_content returned "".
func (s *Service) SearchArticles(ctx context.Context, in ArticleSearch) (any, error) {
	if in.Title == "" && in.Author == "" && in.Content == "" && in.Abstract == "" && in.Column == "" && in.PubDate == "" {
		return nil, &InvalidParameterError{Name: "search_params", Reason: "Must provide at least one search parameter (title, author, content, abstract, column, or pub_date)"}
	}
	if in.Limit < MinArticleLimit || in.Limit > MaxArticleLimit {
		return nil, &InvalidParameterError{Name: "limit", Value: strconv.Itoa(in.Limit), Reason: "must be between 1 and 200"}
	}
	resp, err := s.API.Articles(ctx, fhlapi.ArticleQuery{
		Title: in.Title, Author: in.Author, Content: in.Content, Abstract: in.Abstract,
		Column: in.Column, PubDate: in.PubDate, Simplified: in.Simplified,
	})
	if err != nil {
		return nil, err
	}
	// The limit is applied here, after the upstream answered (the site has no such parameter).
	records, count := resp.Record, resp.RecordCount
	if resp.Status == 1 && len(records) > in.Limit {
		records, count = records[:in.Limit], in.Limit
	}
	switch {
	case resp.Status == 1 && count > 0:
		if len(records) == 0 {
			return &ArticleNoResults{Status: "no_results", QueryType: "article_search", Message: noArticlesMessage}, nil
		}
		out := &ArticleResults{
			Status: "success", QueryType: "article_search", TotalCount: count,
			ReturnedCount: len(records), IncludeFullContent: in.IncludeContent,
			Articles: make([]Article, 0, len(records)),
		}
		for _, r := range records {
			a := Article{
				ID: rawOrEmptyString(r.ID), Title: rawOrEmptyString(r.Title), Author: rawOrEmptyString(r.Author),
				PubTime: rawOrEmptyString(r.PubTime), Column: rawOrEmptyString(r.Column), Abstract: rawOrEmptyString(r.Abst),
			}
			if in.IncludeContent {
				a.Content = rawOrEmptyString(r.Txt)
			} else if p, ok := preview(r.Txt); ok {
				a.ContentPreview = &p
			}
			out.Articles = append(out.Articles, a)
		}
		return out, nil
	case resp.Status == 0:
		msg := resp.Result
		if msg == "" {
			msg = "Unknown error"
		}
		lower := strings.ToLower(msg)
		switch {
		case strings.Contains(lower, "data too much"):
			return &ArticleFailure{Status: "error", Error: "資料量過大，請提供至少一個搜尋條件", Hint: "請使用 title, author, content, abstract, column 或 pub_date 參數"}, nil
		case strings.Contains(lower, "no data"):
			return &ArticleNoResults{Status: "no_results", Message: noArticlesMessage}, nil
		default:
			return &ArticleFailure{Status: "error", Error: msg}, nil
		}
	default:
		return &ArticleNoResults{Status: "no_results", QueryType: "article_search", Message: noArticlesMessage}, nil
	}
}

// ArticleColumn is one article column of list_fhl_article_columns.
type ArticleColumn struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ArticleColumnList is the list_fhl_article_columns result.
type ArticleColumnList struct {
	Status      string          `json:"status"`
	QueryType   string          `json:"query_type"`
	ColumnCount int             `json:"column_count"`
	Columns     []ArticleColumn `json:"columns"`
}

// articleColumns is the hand-kept list of the Python reference (12 columns). Most of these codes return no
// articles from the site (see tool-matrix D25).
var articleColumns = []ArticleColumn{
	{"women3", "麻辣姊妹", "女性信仰生活分享"},
	{"sunday", "主日學", "主日學教材與資源"},
	{"youth", "青少年", "青少年信仰與生活"},
	{"family", "家庭", "家庭生活與信仰"},
	{"theology", "神學", "神學探討與研究"},
	{"bible_study", "查經", "聖經研究與分享"},
	{"devotion", "靈修", "靈修心得與見證"},
	{"mission", "宣教", "宣教事工與分享"},
	{"church", "教會", "教會生活與事奉"},
	{"culture", "文化", "信仰與文化對話"},
	{"history", "歷史", "教會歷史與傳統"},
	{"counseling", "輔導", "心理輔導與關懷"},
}

// ListArticleColumns lists the known article columns (static data, no upstream request).
func ListArticleColumns() *ArticleColumnList {
	return &ArticleColumnList{Status: "success", QueryType: "list_article_columns", ColumnCount: len(articleColumns), Columns: articleColumns}
}
