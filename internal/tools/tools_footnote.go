package tools

import (
	"context"

	"github.com/snowray712000/FHL-MCP-Go/internal/bible"
)

// registerFootnote wires the footnote tool.
func (r *Registry) registerFootnote(svc *bible.Service) {
	r.handlers["get_bible_footnote"] = func(ctx context.Context, a args) (string, error) {
		bookID, err := a.requiredInt("book_id")
		if err != nil {
			return "", err
		}
		footnoteID, err := a.optInt("footnote_id", -1)
		if err != nil {
			return "", err
		}
		version, err := a.optString("version", "")
		if err != nil {
			return "", err
		}
		chapter, err := a.optInt("chapter", 0)
		if err != nil {
			return "", err
		}
		simplified, err := a.optBool("use_simplified", false)
		if err != nil {
			return "", err
		}
		p := bible.FootnoteParams{BookID: bookID, FootnoteID: footnoteID, Version: version, Chapter: chapter, Simplified: simplified}
		var res any
		if p.IsClassicFootnote() {
			res, err = svc.GetFootnote(ctx, bookID, footnoteID, simplified)
		} else {
			res, err = svc.GetFootnoteMulti(ctx, p)
		}
		if err != nil {
			return "", err
		}
		return bible.MarshalIndent2(res)
	}
}
