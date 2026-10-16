package bible

import (
	"context"
	"errors"
	"strings"

	"github.com/snowray712000/FHL-MCP-Go/internal/fhlapi"
)

// CommentaryAPI is the part of the upstream client the commentary and topic tools need (sc/ssc/st.php).
type CommentaryAPI interface {
	SC(ctx context.Context, bid, chapter int, verse *int, simplified bool) (*fhlapi.SCResponse, error)
	SCList(ctx context.Context, simplified bool) (*fhlapi.SCListResponse, error)
	SSC(ctx context.Context, key string, simplified bool) (*fhlapi.SSCResponse, error)
	ST(ctx context.Context, source int, keyword string, countOnly, simplified bool) (*fhlapi.STResponse, error)
}

// CommentaryEntry is one commentary text for a verse.
type CommentaryEntry struct {
	Name    string `json:"commentary_name"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

// CommentaryResult is the get_commentary result. Book is echoed exactly as given; Verse is null when
// the caller left it out (the upstream has no chapter-level commentary and then returns nothing).
type CommentaryResult struct {
	Book         string            `json:"book"`
	Chapter      int               `json:"chapter"`
	Verse        *int              `json:"verse"`
	Count        int               `json:"commentary_count"`
	Commentaries []CommentaryEntry `json:"commentaries"`
}

func orEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// GetCommentary returns the commentaries on one verse (sc.php). The prev/next navigation the Python
// reference intended is not emitted: it looked for those keys on each record, but the upstream
// puts them at the top level, so the Python output never contains them.
func (s *Service) GetCommentary(ctx context.Context, book string, chapter int, verse *int, simplified bool) (*CommentaryResult, error) {
	bid, ok := ResolveBookID(book)
	if !ok {
		return nil, &InvalidParameterError{Name: "book", Value: book, Reason: "Book not found. Please check the spelling or use a valid book code."}
	}
	resp, err := s.API.SC(ctx, bid, chapter, verse, simplified)
	if err != nil {
		return nil, err
	}
	out := &CommentaryResult{Book: book, Chapter: chapter, Verse: verse, Commentaries: make([]CommentaryEntry, 0, len(resp.Record))}
	for _, r := range resp.Record {
		out.Commentaries = append(out.Commentaries, CommentaryEntry{Name: r.BookName, Title: orEmpty(r.Title), Content: orEmpty(r.ComText)})
	}
	out.Count = len(out.Commentaries)
	return out, nil
}

// CommentaryBook is one commentary book of list_commentaries.
type CommentaryBook struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// CommentaryList is the list_commentaries result.
type CommentaryList struct {
	TotalCount   int              `json:"total_count"`
	Commentaries []CommentaryBook `json:"commentaries"`
}

// ListCommentaries lists the available commentary books (sc.php?validbook=1).
func (s *Service) ListCommentaries(ctx context.Context, simplified bool) (*CommentaryList, error) {
	resp, err := s.API.SCList(ctx, simplified)
	if err != nil {
		return nil, err
	}
	out := &CommentaryList{TotalCount: resp.RecordCount, Commentaries: make([]CommentaryBook, 0, len(resp.Record))}
	for _, r := range resp.Record {
		out.Commentaries = append(out.Commentaries, CommentaryBook{ID: r.ID, Name: r.Name})
	}
	return out, nil
}

// CommentaryHit is one hit of search_commentary. Book keeps the upstream's space padding, like Python.
type CommentaryHit struct {
	CommentaryID   int    `json:"commentary_id"`
	CommentaryName string `json:"commentary_name"`
	Title          string `json:"title"`
	Book           string `json:"book"`
	BookEng        string `json:"book_eng"`
	ChapterStart   int    `json:"chapter_start"`
	VerseStart     int    `json:"verse_start"`
	ChapterEnd     int    `json:"chapter_end"`
	VerseEnd       int    `json:"verse_end"`
}

// CommentarySearch is the search_commentary result.
type CommentarySearch struct {
	Keyword    string          `json:"keyword"`
	TotalCount int             `json:"total_count"`
	Results    []CommentaryHit `json:"results"`
}

// SearchCommentary searches inside the commentaries (ssc.php). D12 (approved): a keyword that
// matches more than 500 records is an explicit error, not an empty result.
func (s *Service) SearchCommentary(ctx context.Context, keyword string, simplified bool) (*CommentarySearch, error) {
	if strings.TrimSpace(keyword) == "" {
		return nil, &InvalidParameterError{Name: "keyword", Reason: "must not be empty"}
	}
	resp, err := s.API.SSC(ctx, keyword, simplified)
	if err != nil {
		if errors.Is(err, fhlapi.ErrTooManyRecords) {
			return nil, &InvalidParameterError{Name: "keyword", Value: keyword, Reason: "too many results (more than 500 records), use a more specific keyword (結果超過 500 筆，請縮小關鍵字)"}
		}
		return nil, err
	}
	out := &CommentarySearch{Keyword: keyword, Results: make([]CommentaryHit, 0, len(resp.Record))}
	for _, r := range resp.Record {
		out.Results = append(out.Results, CommentaryHit{
			CommentaryID: r.Tag, CommentaryName: r.BookName, Title: orEmpty(r.Title),
			Book: r.Chinesef, BookEng: r.Engs,
			ChapterStart: r.BChap, VerseStart: r.BSec, ChapterEnd: r.EChap, VerseEnd: r.ESec,
		})
	}
	out.TotalCount = len(out.Results)
	return out, nil
}

// topicSources maps the source names to the upstream N value.
var topicSources = map[string]int{"all": 4, "torrey_en": 0, "naves_en": 1, "torrey_zh": 2, "naves_zh": 3}

// topicSourceNames labels the upstream record "book" value.
var topicSourceNames = [...]string{"Torrey (English)", "Naves (English)", "Torrey (中文)", "Naves (中文)"}

// TopicEntry is one topic of get_topic_study.
type TopicEntry struct {
	ID      string `json:"id"`
	Source  string `json:"source"`
	Topic   string `json:"topic"`
	Content string `json:"content"`
}

// TopicResult is the get_topic_study result.
type TopicResult struct {
	Keyword    string       `json:"keyword"`
	Source     string       `json:"source"`
	TotalCount int          `json:"total_count"`
	Results    []TopicEntry `json:"results"`
}

// TopicCount is the get_topic_study result when only the count was asked for.
type TopicCount struct {
	Keyword    string `json:"keyword"`
	Source     string `json:"source"`
	TotalCount int    `json:"total_count"`
}

// GetTopicStudy looks a keyword up in the Torrey / Nave's topical Bibles (st.php). The result is a
// *TopicCount when countOnly is set, otherwise a *TopicResult.
func (s *Service) GetTopicStudy(ctx context.Context, keyword, source string, countOnly, simplified bool) (any, error) {
	if strings.TrimSpace(keyword) == "" {
		return nil, &InvalidParameterError{Name: "keyword", Reason: "must not be empty"}
	}
	n, ok := topicSources[source]
	if !ok {
		return nil, &InvalidParameterError{Name: "source", Value: source, Reason: "Must be one of: all, torrey_en, naves_en, torrey_zh, naves_zh"}
	}
	resp, err := s.API.ST(ctx, n, keyword, countOnly, simplified)
	if err != nil {
		return nil, err
	}
	if countOnly {
		total := 0
		if resp.RecordCount != nil {
			total = *resp.RecordCount
		}
		return &TopicCount{Keyword: keyword, Source: source, TotalCount: total}, nil
	}
	out := &TopicResult{Keyword: keyword, Source: source, Results: make([]TopicEntry, 0, len(resp.Record))}
	for _, r := range resp.Record {
		name := ""
		if r.Book >= 0 && r.Book < len(topicSourceNames) {
			name = topicSourceNames[r.Book]
		}
		out.Results = append(out.Results, TopicEntry{ID: r.ID, Source: name, Topic: orEmpty(r.Topic), Content: orEmpty(r.Text)})
	}
	out.TotalCount = len(out.Results)
	if resp.RecordCount != nil {
		out.TotalCount = *resp.RecordCount
	}
	return out, nil
}
