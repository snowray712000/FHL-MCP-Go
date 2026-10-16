package bible

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/snowray712000/FHL-MCP-Go/internal/fhlapi"
)

// ScriptureAPI is the part of the upstream client the Apocrypha and Apostolic Fathers tools need
// (qsub / sesub / qaf / seaf.php).
type ScriptureAPI interface {
	QSub(ctx context.Context, bid, chapter int, verse *string) (*fhlapi.SubVerseResponse, error)
	QAF(ctx context.Context, bid, chapter int, verse *string) (*fhlapi.SubVerseResponse, error)
	SESub(ctx context.Context, q string, limit, offset int) (*fhlapi.SubSearchResponse, error)
	SEAF(ctx context.Context, q string, limit, offset int) (*fhlapi.SubSearchResponse, error)
}

// scriptureBooksJSON is the book table of both collections, in the upstream's own numbering
// (D24: the Python reference numbered them differently from what qsub.php / qaf.php really use).
// The aliases are listed in output order: short Chinese, full Chinese, English abbreviation, English name.
//
//go:embed scripture_books.json
var scriptureBooksJSON []byte

type scriptureBook struct {
	ID      int      `json:"id"`
	NameZH  string   `json:"name_zh"`
	NameEN  string   `json:"name_en"`
	Aliases []string `json:"aliases"`
}

// ScriptureSet is one collection (Apocrypha or Apostolic Fathers) and everything that differs between them.
type ScriptureSet struct {
	JSONKey   string          // key in scripture_books.json
	QueryType string          // prefix of the query_type values
	IDRange   string          `json:"id_range"`
	Min       int             `json:"min"`
	Max       int             `json:"max"`
	Books     []scriptureBook `json:"books"`

	// DefaultCode / DefaultName fill in the version when the upstream leaves it out (the HTTP tools of
	// the Python reference do the same).
	DefaultCode, DefaultName string

	verse  func(api ScriptureAPI) func(context.Context, int, int, *string) (*fhlapi.SubVerseResponse, error)
	search func(api ScriptureAPI) func(context.Context, string, int, int) (*fhlapi.SubSearchResponse, error)

	byAlias map[string]*scriptureBook
	byID    map[int]*scriptureBook
}

var (
	// Apocrypha is the Apocrypha collection (books 101-115, version c1933).
	Apocrypha = &ScriptureSet{
		JSONKey: "apocrypha", QueryType: "apocrypha",
		DefaultCode: "c1933", DefaultName: "1933年聖公會出版",
		verse: func(a ScriptureAPI) func(context.Context, int, int, *string) (*fhlapi.SubVerseResponse, error) {
			return a.QSub
		},
		search: func(a ScriptureAPI) func(context.Context, string, int, int) (*fhlapi.SubSearchResponse, error) {
			return a.SESub
		},
	}
	// ApostolicFathers is the Apostolic Fathers collection (books 201-217, version afhuang).
	ApostolicFathers = &ScriptureSet{
		JSONKey: "apostolic_fathers", QueryType: "apostolic_fathers",
		DefaultCode: "af", DefaultName: "黃錫木主編《使徒教父著作》",
		verse: func(a ScriptureAPI) func(context.Context, int, int, *string) (*fhlapi.SubVerseResponse, error) {
			return a.QAF
		},
		search: func(a ScriptureAPI) func(context.Context, string, int, int) (*fhlapi.SubSearchResponse, error) {
			return a.SEAF
		},
	}
)

func init() {
	var all map[string]json.RawMessage
	if err := json.Unmarshal(scriptureBooksJSON, &all); err != nil {
		panic("scripture_books.json: " + err.Error())
	}
	for _, set := range []*ScriptureSet{Apocrypha, ApostolicFathers} {
		if err := json.Unmarshal(all[set.JSONKey], set); err != nil {
			panic("scripture_books.json: " + err.Error())
		}
		set.byAlias = map[string]*scriptureBook{}
		set.byID = map[int]*scriptureBook{}
		for i := range set.Books {
			b := &set.Books[i]
			set.byID[b.ID] = b
			for _, a := range b.Aliases {
				set.byAlias[aliasKey(a)] = b
			}
		}
	}
}

// aliasKey folds the case of ASCII names only, like ResolveBookID.
func aliasKey(name string) string {
	if isASCII(name) {
		return strings.ToLower(name)
	}
	return name
}

func (s *ScriptureSet) notBook(book string) error {
	return &InvalidParameterError{Name: "book", Value: book, Reason: fmt.Sprintf("Not a book of this collection (%s). Use the matching list_*_books tool to see the valid names and ids.", s.IDRange)}
}

// resolve finds a book by any alias (exact Chinese, case-insensitive English) or by its numeric id.
func (s *ScriptureSet) resolve(book string) (*scriptureBook, error) {
	if book != "" && isDigits(book) {
		if id, err := strconv.Atoi(book); err == nil {
			if b, ok := s.byID[id]; ok {
				return b, nil
			}
		}
		return nil, s.notBook(book)
	}
	if b, ok := s.byAlias[aliasKey(book)]; ok {
		return b, nil
	}
	return nil, s.notBook(book)
}

// ScriptureBookList is the list_apocrypha_books / list_apostolic_fathers_books result.
type ScriptureBookList struct {
	Status    string              `json:"status"`
	QueryType string              `json:"query_type"`
	IDRange   string              `json:"id_range"`
	Count     int                 `json:"book_count"`
	Books     []ScriptureBookInfo `json:"books"`
}

// ScriptureBookInfo is one book of the list.
type ScriptureBookInfo struct {
	ID            int      `json:"id"`
	NameZH        string   `json:"name_zh"`
	NameEN        string   `json:"name_en"`
	Abbreviations []string `json:"abbreviations"`
}

// ListScriptureBooks lists the books of a collection (static data, no upstream request).
func ListScriptureBooks(set *ScriptureSet) *ScriptureBookList {
	out := &ScriptureBookList{Status: "success", QueryType: "list_" + set.QueryType + "_books", IDRange: set.IDRange, Books: make([]ScriptureBookInfo, 0, len(set.Books))}
	for _, b := range set.Books {
		out.Books = append(out.Books, ScriptureBookInfo{ID: b.ID, NameZH: b.NameZH, NameEN: b.NameEN, Abbreviations: b.Aliases})
	}
	out.Count = len(out.Books)
	return out
}

// ScriptureVersion is the text edition of a collection.
type ScriptureVersion struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// ScriptureVerse is one verse of get_apocrypha_verse / get_apostolic_fathers_verse.
type ScriptureVerse struct {
	Book    string          `json:"book"`
	BookID  json.RawMessage `json:"book_id"`
	Chapter int             `json:"chapter"`
	Verse   json.RawMessage `json:"verse"`
	Text    string          `json:"text"`
}

// ScriptureText is the get_apocrypha_verse / get_apostolic_fathers_verse result.
type ScriptureText struct {
	Status     string           `json:"status"`
	QueryType  string           `json:"query_type"`
	Book       string           `json:"book"`
	BookID     json.RawMessage  `json:"book_id"`
	Chapter    int              `json:"chapter"`
	Verse      *string          `json:"verse"`
	Version    ScriptureVersion `json:"version"`
	VerseCount json.RawMessage  `json:"verse_count"`
	Verses     []ScriptureVerse `json:"verses"`
}

func rawOrDefault(raw json.RawMessage, def string) json.RawMessage {
	if len(raw) == 0 {
		return json.RawMessage(def)
	}
	return raw
}

func strOr(s *string, def string) string {
	if s == nil {
		return def
	}
	return *s
}

// verseSpec is what the upstream accepts for sec: 1, 1-5, 1,3,5 or a mix such as 1-2,5,8-10.
// Anything else makes qsub.php / qaf.php answer with a failure that quotes its own SQL, so it is refused here.
var verseSpec = regexp.MustCompile(`^[0-9]+(-[0-9]+)?(,[0-9]+(-[0-9]+)?)*$`)

// GetScriptureText returns verses of the Apocrypha or the Apostolic Fathers. verse is passed to the
// upstream as given; nil returns the whole chapter. The book name in the result is the table's Chinese
// name for the book that was really resolved (the Python reference echoed its own, wrongly numbered, table).
func (s *Service) GetScriptureText(ctx context.Context, set *ScriptureSet, book string, chapter int, verse *string) (*ScriptureText, error) {
	b, err := set.resolve(book)
	if err != nil {
		return nil, err
	}
	if verse != nil && !verseSpec.MatchString(strings.TrimSpace(*verse)) {
		return nil, &InvalidParameterError{Name: "verse", Value: *verse, Reason: "Use forms such as 1, 1-5, 1,3,5 or 1-2,5,8-10."}
	}
	sent := verse
	if verse != nil {
		t := strings.TrimSpace(*verse)
		sent = &t
	}
	resp, err := set.verse(s.API)(ctx, b.ID, chapter, sent)
	if err != nil {
		return nil, err
	}
	bid := json.RawMessage(`"未知"`)
	if len(resp.Record) > 0 && len(resp.Record[0].Bid) > 0 {
		bid = resp.Record[0].Bid
	}
	out := &ScriptureText{
		Status: "success", QueryType: set.QueryType + "_verse",
		Book: b.NameZH, BookID: bid, Chapter: chapter, Verse: verse,
		Version:    ScriptureVersion{Code: strOr(resp.Version, set.DefaultCode), Name: strOr(resp.VName, set.DefaultName)},
		VerseCount: rawOrDefault(resp.RecordCount, "0"),
		Verses:     make([]ScriptureVerse, 0, len(resp.Record)),
	}
	for _, r := range resp.Record {
		out.Verses = append(out.Verses, ScriptureVerse{
			Book: b.NameZH, BookID: rawOrDefault(r.Bid, string(bid)), Chapter: chapter,
			Verse: rawOrDefault(r.Sec, `""`), Text: strOr(r.BibleText, ""),
		})
	}
	return out, nil
}

// ScriptureHit is one hit of search_apocrypha / search_apostolic_fathers.
type ScriptureHit struct {
	Book    string          `json:"book"`
	BookID  json.RawMessage `json:"book_id"`
	Chapter json.RawMessage `json:"chapter"`
	Verse   json.RawMessage `json:"verse"`
	Text    string          `json:"text"`
}

// ScriptureSearch is the search_apocrypha / search_apostolic_fathers result. TotalCount is the
// upstream's record_count, which is the number returned for this page when a limit is given.
type ScriptureSearch struct {
	Status        string          `json:"status"`
	QueryType     string          `json:"query_type"`
	Keyword       string          `json:"keyword"`
	TotalCount    json.RawMessage `json:"total_count"`
	ReturnedCount int             `json:"returned_count"`
	Offset        int             `json:"offset"`
	Results       []ScriptureHit  `json:"results"`
}

// SearchScripture searches a collection by keyword. The book name of a hit comes from the book table by
// the hit's book id, falling back to the name the upstream sent.
func (s *Service) SearchScripture(ctx context.Context, set *ScriptureSet, keyword string, limit, offset int) (*ScriptureSearch, error) {
	if strings.TrimSpace(keyword) == "" {
		return nil, &InvalidParameterError{Name: "query", Reason: "must not be empty"}
	}
	resp, err := set.search(s.API)(ctx, keyword, limit, offset)
	if err != nil {
		return nil, err
	}
	out := &ScriptureSearch{
		Status: "success", QueryType: set.QueryType + "_search", Keyword: keyword,
		TotalCount: rawOrDefault(resp.RecordCount, "0"), Offset: offset,
		Results: make([]ScriptureHit, 0, len(resp.Record)),
	}
	for _, r := range resp.Record {
		name := ""
		if truthy(rawOr(r.Bid, nil)) {
			name = strOr(r.Chineses, "")
			if id, ok := rawInt(r.Bid); ok {
				if b, found := set.byID[id]; found {
					name = b.NameZH
				}
			}
		}
		out.Results = append(out.Results, ScriptureHit{
			Book: name, BookID: rawOrDefault(r.Bid, `""`), Chapter: rawOrDefault(r.Chap, `""`),
			Verse: rawOrDefault(r.Sec, `""`), Text: strOr(r.BibleText, ""),
		})
	}
	out.ReturnedCount = len(out.Results)
	return out, nil
}

// rawInt reads a JSON number or numeric string as an int.
func rawInt(raw json.RawMessage) (int, bool) {
	var n json.Number
	if err := json.Unmarshal(raw, &n); err == nil {
		if v, err := strconv.Atoi(n.String()); err == nil {
			return v, true
		}
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		if v, err := strconv.Atoi(strings.TrimSpace(s)); err == nil {
			return v, true
		}
	}
	return 0, false
}
