package tools

import (
	"context"

	"github.com/snowray712000/FHL-MCP-Go/internal/bible"
)

// registerArticles wires the 信望愛 article tools.
func (r *Registry) registerArticles(svc *bible.Service) {
	r.handlers["search_fhl_articles"] = func(ctx context.Context, a args) (string, error) {
		in := bible.ArticleSearch{}
		for _, f := range []struct {
			name string
			dst  *string
		}{
			{"title", &in.Title}, {"author", &in.Author}, {"content", &in.Content},
			{"abstract", &in.Abstract}, {"column", &in.Column}, {"pub_date", &in.PubDate},
		} {
			v, err := a.optString(f.name, "")
			if err != nil {
				return "", err
			}
			*f.dst = v
		}
		var err error
		if in.Limit, err = a.optInt("limit", bible.DefaultArticleLimit); err != nil {
			return "", err
		}
		if in.IncludeContent, err = a.optBool("include_content", false); err != nil {
			return "", err
		}
		if in.Simplified, err = a.optBool("use_simplified", false); err != nil {
			return "", err
		}
		res, err := svc.SearchArticles(ctx, in)
		if err != nil {
			return "", err
		}
		return bible.MarshalIndent2(res)
	}

	r.handlers["list_fhl_article_columns"] = func(context.Context, args) (string, error) {
		return bible.MarshalIndent2(bible.ListArticleColumns())
	}
}
