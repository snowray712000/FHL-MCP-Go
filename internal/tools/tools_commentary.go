package tools

import (
	"context"

	"github.com/snowray712000/FHL-MCP-Go/internal/bible"
)

// registerCommentary wires the commentary and topical-study tools.
func (r *Registry) registerCommentary(svc *bible.Service) {
	r.handlers["get_commentary"] = func(ctx context.Context, a args) (string, error) {
		book, err := a.requiredString("book")
		if err != nil {
			return "", err
		}
		chapter, err := a.requiredInt("chapter")
		if err != nil {
			return "", err
		}
		verse, err := a.optIntPtr("verse")
		if err != nil {
			return "", err
		}
		simplified, err := a.optBool("use_simplified", false)
		if err != nil {
			return "", err
		}
		res, err := svc.GetCommentary(ctx, book, chapter, verse, simplified)
		if err != nil {
			return "", err
		}
		return bible.MarshalIndent2(res)
	}

	r.handlers["list_commentaries"] = func(ctx context.Context, a args) (string, error) {
		simplified, err := a.optBool("use_simplified", false)
		if err != nil {
			return "", err
		}
		res, err := svc.ListCommentaries(ctx, simplified)
		if err != nil {
			return "", err
		}
		return bible.MarshalIndent2(res)
	}

	r.handlers["search_commentary"] = func(ctx context.Context, a args) (string, error) {
		keyword, err := a.requiredString("keyword")
		if err != nil {
			return "", err
		}
		simplified, err := a.optBool("use_simplified", false)
		if err != nil {
			return "", err
		}
		res, err := svc.SearchCommentary(ctx, keyword, simplified)
		if err != nil {
			return "", err
		}
		return bible.MarshalIndent2(res)
	}

	r.handlers["get_topic_study"] = func(ctx context.Context, a args) (string, error) {
		keyword, err := a.requiredString("keyword")
		if err != nil {
			return "", err
		}
		source, err := a.optString("source", "all")
		if err != nil {
			return "", err
		}
		countOnly, err := a.optBool("count_only", false)
		if err != nil {
			return "", err
		}
		simplified, err := a.optBool("use_simplified", false)
		if err != nil {
			return "", err
		}
		res, err := svc.GetTopicStudy(ctx, keyword, source, countOnly, simplified)
		if err != nil {
			return "", err
		}
		return bible.MarshalIndent2(res)
	}
}
