package tools

import (
	"context"

	"github.com/snowray712000/FHL-MCP-Go/internal/bible"
)

// registerWords wires the original-language tools.
func (r *Registry) registerWords(svc *bible.Service) {
	r.handlers["get_word_analysis"] = func(ctx context.Context, a args) (string, error) {
		book, err := a.requiredString("book")
		if err != nil {
			return "", err
		}
		chapter, err := a.requiredInt("chapter")
		if err != nil {
			return "", err
		}
		verse, err := a.requiredInt("verse")
		if err != nil {
			return "", err
		}
		simplified, err := a.optBool("use_simplified", false)
		if err != nil {
			return "", err
		}
		res, err := svc.WordAnalysis(ctx, book, chapter, verse, simplified)
		if err != nil {
			return "", err
		}
		return bible.MarshalIndent2(res)
	}

	r.handlers["get_parsing_reference"] = func(ctx context.Context, a args) (string, error) {
		book, err := a.requiredString("book")
		if err != nil {
			return "", err
		}
		chapter, err := a.requiredInt("chapter")
		if err != nil {
			return "", err
		}
		verse, err := a.requiredString("verse")
		if err != nil {
			return "", err
		}
		endChapter, err := a.optIntPtr("end_chapter")
		if err != nil {
			return "", err
		}
		endVerse, err := a.optIntPtr("end_verse")
		if err != nil {
			return "", err
		}
		// The result is Markdown text, not JSON (D15).
		return svc.ParsingReference(ctx, bible.ParsingParams{
			Book: book, Chapter: chapter, Verse: verse, EndChapter: endChapter, EndVerse: endVerse,
		})
	}
}
