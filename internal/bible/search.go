package bible

import (
	"context"
	"strconv"
	"strings"

	"github.com/snowray712000/FHL-MCP-Go/internal/fhlapi"
)

// SearchParams are the search_bible inputs.
type SearchParams struct {
	Query      string
	SearchType string // keyword | greek_number | hebrew_number
	Scope      string // all | ot | nt
	Version    string
	Limit      int
	Offset     int
	Simplified bool
}

// SearchResult is the search_bible result object (field order is part of the contract).
type SearchResult struct {
	TotalCount int         `json:"total_count"`
	Query      string      `json:"query"`
	SearchType string      `json:"search_type"`
	Scope      string      `json:"scope"`
	Limit      int         `json:"limit"`
	Offset     int         `json:"offset"`
	Results    []VerseItem `json:"results"`
}

var searchTypeOrig = map[string]int{"keyword": 0, "greek_number": 1, "hebrew_number": 2}

// scopeRange maps the scope names to the upstream RANGE value.
var scopeRange = map[string]int{"all": 0, "nt": 1, "ot": 2}

func invalidSearchType(v string) error {
	return &InvalidParameterError{Name: "search_type", Value: v, Reason: "Invalid search type. Use 'keyword', 'greek_number' or 'hebrew_number'."}
}

func emptyQuery() error {
	return &InvalidParameterError{Name: "query", Reason: "must not be empty"}
}

// SearchBible searches the Bible by keyword or by Strong's number (se.php).
func (s *Service) SearchBible(ctx context.Context, p SearchParams) (*SearchResult, error) {
	orig, ok := searchTypeOrig[p.SearchType]
	if !ok {
		return nil, invalidSearchType(p.SearchType)
	}
	rng, ok := scopeRange[p.Scope]
	if !ok {
		return nil, &InvalidParameterError{Name: "scope", Value: p.Scope, Reason: "Invalid scope. Use 'all', 'ot' or 'nt'."}
	}
	// D7 (approved): an empty query is a clear parameter error, not the leaked KeyError of the Python reference.
	if strings.TrimSpace(p.Query) == "" {
		return nil, emptyQuery()
	}
	resp, err := s.searchPaged(ctx, fhlapi.SERequest{
		Query: p.Query, Version: p.Version, Orig: orig, Range: rng,
		Offset: p.Offset, Limit: p.Limit, Simplified: p.Simplified,
	})
	if err != nil {
		return nil, err
	}
	return &SearchResult{
		TotalCount: resp.RecordCount, Query: p.Query, SearchType: p.SearchType, Scope: p.Scope,
		Limit: p.Limit, Offset: p.Offset, Results: hits(resp.Record),
	}, nil
}

// seMaxPage is the most records se.php returns per request: any limit above it is answered with
// status "Fail: record count > 500". seMaxLimit caps one call so a huge limit cannot fan out
// into unbounded upstream requests (10 pages).
const (
	seMaxPage  = 500
	seMaxLimit = 5000
)

// searchPaged queries se.php, splitting a limit above seMaxPage into consecutive offset pages
// (the same approach as the FHLNUI web client) and merging them. A page shorter than requested
// ends the walk. The merged RecordCount is the number of records returned, as for a single page.
func (s *Service) searchPaged(ctx context.Context, r fhlapi.SERequest) (*fhlapi.SEResponse, error) {
	if r.Limit > seMaxLimit {
		return nil, &InvalidParameterError{Name: "limit", Value: strconv.Itoa(r.Limit), Reason: "too large, at most " + strconv.Itoa(seMaxLimit) + " results per call; use a smaller limit (上限 " + strconv.Itoa(seMaxLimit) + " 筆)"}
	}
	if r.Limit <= seMaxPage {
		return s.API.SE(ctx, r)
	}
	var merged *fhlapi.SEResponse
	for remaining, offset := r.Limit, r.Offset; remaining > 0; {
		page := r
		page.Offset, page.Limit = offset, min(remaining, seMaxPage)
		resp, err := s.API.SE(ctx, page)
		if err != nil {
			return nil, err
		}
		if merged == nil {
			merged = resp
		} else {
			merged.Record = append(merged.Record, resp.Record...)
		}
		if len(resp.Record) < page.Limit {
			break
		}
		offset += len(resp.Record)
		remaining -= len(resp.Record)
	}
	merged.RecordCount = len(merged.Record)
	return merged, nil
}

func hits(rs []fhlapi.SERecord) []VerseItem {
	out := make([]VerseItem, 0, len(rs))
	for _, r := range rs {
		out = append(out, VerseItem{Book: r.Chineses, BookEng: r.Engs, Chapter: r.Chap, Verse: r.Sec, Text: r.BibleText})
	}
	return out
}

// AdvancedParams are the search_bible_advanced inputs. A nil or zero range bound means "not given".
type AdvancedParams struct {
	Query      string
	SearchType string
	RangeStart *int
	RangeEnd   *int
	Version    string
	Limit      int
	Offset     int
	Simplified bool
}

// RangeEcho is the "range" object of an advanced result (values echoed as given).
type RangeEcho struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

// AdvancedResult is the search_bible_advanced result object for a book range.
type AdvancedResult struct {
	TotalCount int         `json:"total_count"`
	Query      string      `json:"query"`
	SearchType string      `json:"search_type"`
	Range      RangeEcho   `json:"range"`
	Limit      int         `json:"limit"`
	Offset     int         `json:"offset"`
	Results    []VerseItem `json:"results"`
}

// SearchAdvanced searches inside a book range. Like the Python reference, when either
// bound is missing or 0 it silently falls back to a whole-Bible search and returns the
// plain search_bible shape; the result is *SearchResult in that case, *AdvancedResult otherwise.
// Unlike the Python reference (D19, assumption pending approval) the range path accepts the
// documented search types 'greek_number' and 'hebrew_number' instead of only 'keyword'.
func (s *Service) SearchAdvanced(ctx context.Context, p AdvancedParams) (any, error) {
	if p.RangeStart == nil || p.RangeEnd == nil || *p.RangeStart == 0 || *p.RangeEnd == 0 {
		return s.SearchBible(ctx, SearchParams{
			Query: p.Query, SearchType: p.SearchType, Scope: "all", Version: p.Version,
			Limit: p.Limit, Offset: p.Offset, Simplified: p.Simplified,
		})
	}
	start, end := *p.RangeStart, *p.RangeEnd
	startID, okStart := ResolveBookID(strconv.Itoa(start))
	endID, okEnd := ResolveBookID(strconv.Itoa(end))
	if !okStart || !okEnd {
		return nil, &InvalidParameterError{Name: "range", Value: strconv.Itoa(start) + " - " + strconv.Itoa(end), Reason: "Invalid book range. Use FHL book numbers (1-66 for the canon)."}
	}
	if startID > endID {
		return nil, &InvalidParameterError{Name: "range", Value: strconv.Itoa(start) + " > " + strconv.Itoa(end), Reason: "range_start must not be greater than range_end."}
	}
	orig, ok := searchTypeOrig[p.SearchType]
	if !ok {
		return nil, invalidSearchType(p.SearchType)
	}
	if strings.TrimSpace(p.Query) == "" {
		return nil, emptyQuery()
	}
	resp, err := s.searchPaged(ctx, fhlapi.SERequest{
		Query: p.Query, Version: p.Version, Orig: orig, Range: 3, Offset: p.Offset, Limit: p.Limit,
		RangeBID: startID, RangeEID: endID, Simplified: p.Simplified,
	})
	if err != nil {
		return nil, err
	}
	return &AdvancedResult{
		TotalCount: resp.RecordCount, Query: p.Query, SearchType: p.SearchType,
		Range: RangeEcho{Start: start, End: end}, Limit: p.Limit, Offset: p.Offset, Results: hits(resp.Record),
	}, nil
}
