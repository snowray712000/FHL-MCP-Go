package tools

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/snowray712000/FHL-MCP-Go/internal/bible"
)

// ErrNotImplemented marks a tool that is in the frozen contract but not built yet (M1 ships one real tool).
var ErrNotImplemented = errors.New("this tool is not implemented yet in this build")

// handler runs one tool and returns the string that becomes the MCP text content.
type handler func(ctx context.Context, a args) (string, error)

// Registry maps tool names to handlers.
type Registry struct {
	handlers map[string]handler
}

// NewRegistry wires the implemented tools to the domain service.
func NewRegistry(svc *bible.Service) *Registry {
	r := &Registry{handlers: map[string]handler{}}
	r.handlers["get_bible_verse"] = func(ctx context.Context, a args) (string, error) {
		return getBibleVerse(ctx, svc, a)
	}
	r.handlers["get_bible_chapter"] = func(ctx context.Context, a args) (string, error) {
		return getBibleChapter(ctx, svc, a)
	}
	r.handlers["query_verse_citation"] = func(ctx context.Context, a args) (string, error) {
		return queryVerseCitation(ctx, svc, a)
	}
	r.handlers["get_book_list"] = func(_ context.Context, a args) (string, error) {
		testament, err := a.optString("testament", "all")
		if err != nil {
			return "", err
		}
		return bible.MarshalIndent2(bible.BookList(testament))
	}
	r.handlers["get_book_info"] = func(_ context.Context, a args) (string, error) {
		book, err := a.requiredString("book")
		if err != nil {
			return "", err
		}
		res, err := bible.BookInfo(book)
		if err != nil {
			return "", err
		}
		return bible.MarshalIndent2(res)
	}
	r.handlers["list_bible_versions"] = func(ctx context.Context, a args) (string, error) {
		simplified, err := a.optBool("use_simplified", false)
		if err != nil {
			return "", err
		}
		res, err := svc.ListVersions(ctx, simplified)
		if err != nil {
			return "", err
		}
		return bible.MarshalIndent2(res)
	}
	r.handlers["search_available_versions"] = func(ctx context.Context, a args) (string, error) {
		testament, err := a.optStringPtr("testament")
		if err != nil {
			return "", err
		}
		hasStrongs, err := a.optBoolPtr("has_strongs")
		if err != nil {
			return "", err
		}
		simplified, err := a.optBool("use_simplified", false)
		if err != nil {
			return "", err
		}
		res, err := svc.SearchVersions(ctx, hasStrongs, testament, simplified)
		if err != nil {
			return "", err
		}
		return bible.MarshalIndent2(res)
	}
	r.registerSearch(svc)
	r.registerWords(svc)
	r.registerCommentary(svc)
	r.registerAudio(svc)
	return r
}

// Implemented reports whether name has a real handler.
func (r *Registry) Implemented(name string) bool {
	_, ok := r.handlers[name]
	return ok
}

// Call runs the named tool with raw JSON arguments. A tool that exists in the
// contract but has no handler returns ErrNotImplemented.
func (r *Registry) Call(ctx context.Context, name string, raw json.RawMessage) (string, error) {
	h, ok := r.handlers[name]
	if !ok {
		return "", ErrNotImplemented
	}
	a, err := parseArgs(raw)
	if err != nil {
		return "", err
	}
	return h(ctx, a)
}

func getBibleVerse(ctx context.Context, svc *bible.Service, a args) (string, error) {
	book, err := a.requiredString("book")
	if err != nil {
		return "", err
	}
	chapter, err := a.requiredInt("chapter")
	if err != nil {
		return "", err
	}
	verse, err := a.optString("verse", "")
	if err != nil {
		return "", err
	}
	version, err := a.optString("version", "unv")
	if err != nil {
		return "", err
	}
	strong, err := a.optBool("include_strong", false)
	if err != nil {
		return "", err
	}
	simplified, err := a.optBool("use_simplified", false)
	if err != nil {
		return "", err
	}
	res, err := svc.GetVerse(ctx, bible.VerseParams{
		Book: book, Chapter: chapter, Verse: verse, Version: version,
		IncludeStrong: strong, UseSimplified: simplified,
	})
	if err != nil {
		return "", err
	}
	return bible.MarshalIndent2(res)
}

func getBibleChapter(ctx context.Context, svc *bible.Service, a args) (string, error) {
	book, err := a.requiredString("book")
	if err != nil {
		return "", err
	}
	chapter, err := a.requiredInt("chapter")
	if err != nil {
		return "", err
	}
	version, err := a.optString("version", "unv")
	if err != nil {
		return "", err
	}
	simplified, err := a.optBool("use_simplified", false)
	if err != nil {
		return "", err
	}
	res, err := svc.GetChapter(ctx, book, chapter, version, simplified)
	if err != nil {
		return "", err
	}
	return bible.MarshalIndent2(res)
}

func queryVerseCitation(ctx context.Context, svc *bible.Service, a args) (string, error) {
	citation, err := a.requiredString("citation")
	if err != nil {
		return "", err
	}
	version, err := a.optString("version", "unv")
	if err != nil {
		return "", err
	}
	strong, err := a.optBool("include_strong", false)
	if err != nil {
		return "", err
	}
	simplified, err := a.optBool("use_simplified", false)
	if err != nil {
		return "", err
	}
	res, err := svc.QueryCitation(ctx, bible.CitationParams{Citation: citation, Version: version, IncludeStrong: strong, UseSimplified: simplified})
	if err != nil {
		return "", err
	}
	return bible.MarshalIndent2(res)
}
