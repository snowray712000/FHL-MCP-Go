package bible

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/snowray712000/FHL-MCP-Go/internal/fhlapi"
)

// API is the slice of the upstream client the Bible service needs (mockable).
type API interface {
	QB(ctx context.Context, r fhlapi.QBRequest) (*fhlapi.QBResponse, error)
	AB(ctx context.Context, simplified bool) (*fhlapi.ABResponse, error)
	SE(ctx context.Context, r fhlapi.SERequest) (*fhlapi.SEResponse, error)
	SD(ctx context.Context, nt bool, number int, simplified bool) (*fhlapi.SDResponse, error)
	QP(ctx context.Context, bid, chapter, verse int, simplified bool) (*fhlapi.QPResponse, error)
}

// Service holds the Bible domain logic. It does not know about MCP.
type Service struct {
	API API
}

// InvalidParameterError is a caller mistake (bad or missing parameter), reported to the model as a tool error.
type InvalidParameterError struct {
	Name   string
	Value  string // empty when the parameter was missing
	Reason string
}

func (e *InvalidParameterError) Error() string {
	if e.Value == "" {
		return fmt.Sprintf("Invalid parameter '%s': %s", e.Name, e.Reason)
	}
	return fmt.Sprintf("Invalid parameter '%s': %s - %s", e.Name, e.Value, e.Reason)
}

// VerseParams are the get_bible_verse inputs after validation.
type VerseParams struct {
	Book          string
	Chapter       int
	Verse         string // empty = whole chapter
	Version       string // default "unv"
	IncludeStrong bool
	UseSimplified bool
}

// VerseItem fixes the JSON field order of one verse (the order is part of the contract).
type VerseItem struct {
	Book    string `json:"book"`
	BookEng string `json:"book_eng"`
	Chapter int    `json:"chapter"`
	Verse   int    `json:"verse"`
	Text    string `json:"text"`
}

// NavRef is one navigation endpoint.
type NavRef struct {
	Book    string `json:"book"`
	BookEng string `json:"book_eng"`
	Chapter int    `json:"chapter"`
	Verse   int    `json:"verse"`
}

// Navigation mirrors the Python output: "next" is only emitted together with "prev".
type Navigation struct {
	Prev *NavRef `json:"prev"`
	Next *NavRef `json:"next,omitempty"`
}

// VerseResult is the get_bible_verse result object.
type VerseResult struct {
	Version     string      `json:"version"`
	VersionName string      `json:"version_name"`
	RecordCount int         `json:"record_count"`
	Verses      []VerseItem `json:"verses"`
	Navigation  *Navigation `json:"navigation,omitempty"`
}

// GetVerse queries verses (or a whole chapter when p.Verse is empty).
func (s *Service) GetVerse(ctx context.Context, p VerseParams) (*VerseResult, error) {
	bid, ok := ResolveBookID(p.Book)
	if !ok {
		return nil, &InvalidParameterError{Name: "book", Value: p.Book, Reason: "Book not found. Please check the spelling or use a valid book code."}
	}
	version := p.Version
	if version == "" {
		version = "unv"
	}
	resp, err := s.API.QB(ctx, fhlapi.QBRequest{
		BID:        bid,
		Chapter:    p.Chapter,
		Verse:      p.Verse,
		Version:    version,
		Strong:     p.IncludeStrong,
		Simplified: p.UseSimplified,
	})
	if err != nil {
		return nil, err
	}

	res := &VerseResult{
		Version:     resp.Version,
		VersionName: resp.VName,
		RecordCount: resp.RecordCount,
		Verses:      make([]VerseItem, 0, len(resp.Record)),
	}
	for _, r := range resp.Record {
		res.Verses = append(res.Verses, VerseItem{Book: r.Chineses, BookEng: r.Engs, Chapter: r.Chap, Verse: r.Sec, Text: r.BibleText})
	}
	if resp.Prev != nil {
		nav := &Navigation{Prev: refToNav(resp.Prev)}
		if resp.Next != nil {
			nav.Next = refToNav(resp.Next)
		}
		res.Navigation = nav
	}
	return res, nil
}

func refToNav(r *fhlapi.Ref) *NavRef {
	return &NavRef{Book: r.Chineses, BookEng: r.Engs, Chapter: r.Chap, Verse: r.Sec}
}

// MarshalIndent2 renders v like Python json.dumps(v, ensure_ascii=False, indent=2):
// two-space indent, no HTML escaping of < > &, no trailing newline.
func MarshalIndent2(v any) (string, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return "", err
	}
	return string(bytes.TrimRight(buf.Bytes(), "\n")), nil
}
