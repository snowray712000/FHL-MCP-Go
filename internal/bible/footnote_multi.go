package bible

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/snowray712000/FHL-MCP-Go/internal/fhlapi"
)

// FootnoteMultiAPI is the optional upstream part for footnotes of versions other than TCV (rt.php / rt-nui.php).
// The real client implements it; test fakes that only know the TCV call need not.
type FootnoteMultiAPI interface {
	RTMulti(ctx context.Context, q fhlapi.RTQuery) (*fhlapi.RTResponse, error)
}

// FootnoteParams are the get_bible_footnote parameters. FootnoteID < 0 means "not given" (list them all).
type FootnoteParams struct {
	BookID     int
	FootnoteID int
	Version    string // default "tcv"
	Chapter    int    // 0 = not given
	Simplified bool
}

// IsClassicFootnote reports whether the request is the original one (TCV, one footnote), which keeps the old path.
func (p FootnoteParams) IsClassicFootnote() bool {
	v := strings.TrimSpace(p.Version)
	return p.FootnoteID >= 0 && (v == "" || strings.EqualFold(v, "tcv"))
}

// FootnoteItem is one footnote of a list result.
type FootnoteItem struct {
	FootnoteID json.RawMessage `json:"footnote_id"`
	Text       string          `json:"text"`
}

// FootnoteOther is the result for one footnote of a non-TCV version.
type FootnoteOther struct {
	Status     string          `json:"status"`
	QueryType  string          `json:"query_type"`
	Version    string          `json:"version"`
	BookID     int             `json:"book_id"`
	BookName   string          `json:"book_name"`
	Chapter    int             `json:"chapter,omitempty"`
	FootnoteID json.RawMessage `json:"footnote_id"`
	Text       string          `json:"text"`
}

// FootnoteList is the result when footnote_id is omitted: every footnote of a chapter (lcc, csb) or of the book.
type FootnoteList struct {
	Status      string         `json:"status"`
	QueryType   string         `json:"query_type"`
	Version     string         `json:"version"`
	BookID      int            `json:"book_id"`
	BookName    string         `json:"book_name"`
	Scope       string         `json:"scope"`
	Chapter     int            `json:"chapter,omitempty"`
	RecordCount int            `json:"record_count"`
	Footnotes   []FootnoteItem `json:"footnotes"`
}

var versionCodeRe = regexp.MustCompile(`^[A-Za-z0-9]{1,16}$`)

// rtEngsExceptions are books whose upstream footnote abbreviation differs from the verse-query one
// (checked against rt-nui.php on 2026-10-19; every other book is the verse abbreviation, with a space after
// a leading book number: "1Sam" -> "1 Sam").
var rtEngsExceptions = map[int]string{
	11: "1 Kin", 12: "2 Kin", 13: "1 Chr", 14: "2 Chr", 17: "Esth", 28: "Hos", 32: "Jon", 33: "Mic", 34: "Nah",
}

// rtEngs returns the abbreviation rt.php / rt-nui.php use for a canon book (1-66).
func rtEngs(bookID int) (string, bool) {
	if bookID < 1 || bookID > 66 {
		return "", false
	}
	if e, ok := rtEngsExceptions[bookID]; ok {
		return e, true
	}
	e := byID[bookID].EngShort
	if len(e) > 1 && e[0] >= '1' && e[0] <= '3' && e[1] != ' ' {
		e = e[:1] + " " + e[1:]
	}
	return e, true
}

// GetFootnoteMulti answers the extended get_bible_footnote forms: another version, a chapter, or all footnotes.
func (s *Service) GetFootnoteMulti(ctx context.Context, p FootnoteParams) (any, error) {
	version := strings.ToLower(strings.TrimSpace(p.Version))
	if version == "" {
		version = "tcv"
	}
	if !versionCodeRe.MatchString(version) {
		return nil, &InvalidParameterError{Name: "version", Value: p.Version, Reason: "Version code must be 1-16 letters or digits (see list_bible_versions)."}
	}
	engs, ok := rtEngs(p.BookID)
	if !ok {
		return nil, &InvalidParameterError{Name: "book_id", Value: fmt.Sprint(p.BookID), Reason: "Footnotes exist for books 1-66 only."}
	}
	if p.Chapter < 0 {
		return nil, &InvalidParameterError{Name: "chapter", Value: fmt.Sprint(p.Chapter), Reason: "Chapter must be positive."}
	}
	perChapter := version == "lcc" || version == "csb" // their footnote numbers restart in every chapter
	if perChapter && p.Chapter == 0 {
		return nil, &InvalidParameterError{Name: "chapter", Reason: "Version " + version + " numbers its footnotes per chapter, so chapter is required."}
	}
	api, ok := s.API.(FootnoteMultiAPI)
	if !ok {
		return nil, fmt.Errorf("footnotes of other versions are not supported by this upstream client")
	}
	q := fhlapi.RTQuery{Engs: engs, Version: version, ID: p.FootnoteID, Simplified: p.Simplified}
	if perChapter {
		q.Chap = p.Chapter
	}
	if q.ID < 0 {
		q.ID = -1
	}
	resp, err := api.RTMulti(ctx, q)
	if err != nil {
		return nil, err
	}
	name := byID[p.BookID].EngShort
	if p.FootnoteID >= 0 {
		if resp.RecordCount <= 0 || len(resp.Record) == 0 {
			return &FootnoteNotFound{Status: "not_found", Message: fmt.Sprintf("找不到書卷 %d 的註腳 #%d（譯本 %s）", p.BookID, p.FootnoteID, version)}, nil
		}
		rec := resp.Record[0]
		id := rec.ID
		if len(id) == 0 {
			id = json.RawMessage(fmt.Sprint(p.FootnoteID))
		}
		out := &FootnoteOther{Status: "success", QueryType: "bible_footnote", Version: version, BookID: p.BookID, BookName: name,
			FootnoteID: norm(id), Text: strOr(rec.Text, "")}
		if perChapter {
			out.Chapter = p.Chapter
		}
		return out, nil
	}
	list := &FootnoteList{Status: "success", QueryType: "bible_footnotes", Version: version, BookID: p.BookID, BookName: name,
		Scope: resp.Scope, Footnotes: make([]FootnoteItem, 0, len(resp.Record))}
	if list.Scope == "" {
		list.Scope = "book"
		if perChapter {
			list.Scope = "chapter"
		}
	}
	if list.Scope == "chapter" {
		list.Chapter = p.Chapter
	}
	for _, r := range resp.Record {
		list.Footnotes = append(list.Footnotes, FootnoteItem{FootnoteID: norm(r.ID), Text: strOr(r.Text, "")})
	}
	list.RecordCount = len(list.Footnotes)
	return list, nil
}
