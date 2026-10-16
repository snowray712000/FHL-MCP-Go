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
		footnoteID, err := a.requiredInt("footnote_id")
		if err != nil {
			return "", err
		}
		simplified, err := a.optBool("use_simplified", false)
		if err != nil {
			return "", err
		}
		res, err := svc.GetFootnote(ctx, bookID, footnoteID, simplified)
		if err != nil {
			return "", err
		}
		return bible.MarshalIndent2(res)
	}
}
