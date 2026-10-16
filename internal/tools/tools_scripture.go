package tools

import (
	"context"

	"github.com/snowray712000/FHL-MCP-Go/internal/bible"
)

// registerScripture wires the Apocrypha and Apostolic Fathers tools.
func (r *Registry) registerScripture(svc *bible.Service) {
	for _, c := range []struct {
		set                             *bible.ScriptureSet
		verseTool, searchTool, listTool string
	}{
		{bible.Apocrypha, "get_apocrypha_verse", "search_apocrypha", "list_apocrypha_books"},
		{bible.ApostolicFathers, "get_apostolic_fathers_verse", "search_apostolic_fathers", "list_apostolic_fathers_books"},
	} {
		set := c.set
		r.handlers[c.verseTool] = func(ctx context.Context, a args) (string, error) {
			book, err := a.requiredString("book")
			if err != nil {
				return "", err
			}
			chapter, err := a.requiredInt("chapter")
			if err != nil {
				return "", err
			}
			verse, err := a.optStringPtr("verse")
			if err != nil {
				return "", err
			}
			res, err := svc.GetScriptureText(ctx, set, book, chapter, verse)
			if err != nil {
				return "", err
			}
			return bible.MarshalIndent2(res)
		}

		r.handlers[c.searchTool] = func(ctx context.Context, a args) (string, error) {
			query, err := a.requiredString("query")
			if err != nil {
				return "", err
			}
			limit, err := a.optInt("limit", 10)
			if err != nil {
				return "", err
			}
			offset, err := a.optInt("offset", 0)
			if err != nil {
				return "", err
			}
			res, err := svc.SearchScripture(ctx, set, query, limit, offset)
			if err != nil {
				return "", err
			}
			return bible.MarshalIndent2(res)
		}

		r.handlers[c.listTool] = func(context.Context, args) (string, error) {
			return bible.MarshalIndent2(bible.ListScriptureBooks(set))
		}
	}
}
