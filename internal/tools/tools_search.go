package tools

import (
	"context"

	"github.com/snowray712000/FHL-MCP-Go/internal/bible"
)

// optInt reads an optional integer parameter (null == absent, numeric strings accepted).
func (a args) optInt(name string, def int) (int, error) {
	v, ok := a.get(name)
	if !ok {
		return def, nil
	}
	return decodeInt(name, v)
}

// optIntPtr is optInt that reports "not given" as nil.
func (a args) optIntPtr(name string) (*int, error) {
	v, ok := a.get(name)
	if !ok {
		return nil, nil
	}
	n, err := decodeInt(name, v)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

// registerSearch wires the search and Strong's tools.
func (r *Registry) registerSearch(svc *bible.Service) {
	r.handlers["search_bible"] = func(ctx context.Context, a args) (string, error) {
		query, err := a.requiredString("query")
		if err != nil {
			return "", err
		}
		searchType, err := a.optString("search_type", "keyword")
		if err != nil {
			return "", err
		}
		scope, err := a.optString("scope", "all")
		if err != nil {
			return "", err
		}
		version, err := a.optString("version", "unv")
		if err != nil {
			return "", err
		}
		limit, err := a.optInt("limit", 50)
		if err != nil {
			return "", err
		}
		simplified, err := a.optBool("use_simplified", false)
		if err != nil {
			return "", err
		}
		// The contract has no "offset" parameter for search_bible: the first page is always returned.
		res, err := svc.SearchBible(ctx, bible.SearchParams{
			Query: query, SearchType: searchType, Scope: scope, Version: version,
			Limit: limit, Offset: 0, Simplified: simplified,
		})
		if err != nil {
			return "", err
		}
		return bible.MarshalIndent2(res)
	}

	r.handlers["search_bible_advanced"] = func(ctx context.Context, a args) (string, error) {
		query, err := a.requiredString("query")
		if err != nil {
			return "", err
		}
		searchType, err := a.optString("search_type", "keyword")
		if err != nil {
			return "", err
		}
		start, err := a.optIntPtr("range_start")
		if err != nil {
			return "", err
		}
		end, err := a.optIntPtr("range_end")
		if err != nil {
			return "", err
		}
		version, err := a.optString("version", "unv")
		if err != nil {
			return "", err
		}
		limit, err := a.optInt("limit", 50)
		if err != nil {
			return "", err
		}
		offset, err := a.optInt("offset", 0)
		if err != nil {
			return "", err
		}
		simplified, err := a.optBool("use_simplified", false)
		if err != nil {
			return "", err
		}
		res, err := svc.SearchAdvanced(ctx, bible.AdvancedParams{
			Query: query, SearchType: searchType, RangeStart: start, RangeEnd: end, Version: version,
			Limit: limit, Offset: offset, Simplified: simplified,
		})
		if err != nil {
			return "", err
		}
		return bible.MarshalIndent2(res)
	}

	r.handlers["lookup_strongs"] = func(ctx context.Context, a args) (string, error) {
		number, err := a.requiredString("number")
		if err != nil {
			return "", err
		}
		testament, err := a.optStringPtr("testament")
		if err != nil {
			return "", err
		}
		simplified, err := a.optBool("use_simplified", false)
		if err != nil {
			return "", err
		}
		res, err := svc.LookupStrongs(ctx, number, testament, simplified)
		if err != nil {
			return "", err
		}
		return bible.MarshalIndent2(res)
	}

	r.handlers["search_strongs_occurrences"] = func(ctx context.Context, a args) (string, error) {
		number, err := a.requiredString("number")
		if err != nil {
			return "", err
		}
		testament, err := a.optStringPtr("testament")
		if err != nil {
			return "", err
		}
		limit, err := a.optInt("limit", 50)
		if err != nil {
			return "", err
		}
		simplified, err := a.optBool("use_simplified", false)
		if err != nil {
			return "", err
		}
		res, err := svc.SearchStrongsOccurrences(ctx, number, testament, limit, simplified)
		if err != nil {
			return "", err
		}
		return bible.MarshalIndent2(res)
	}
}
