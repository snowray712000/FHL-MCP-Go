package bible

import (
	"context"
	"regexp"
	"strconv"
	"strings"
)

// citationRE is the Python reference's pattern `^(.+?)\s+(\d+):(\d+(?:-\d+|,\d+)*)$`
// (approved D10: a space between book and chapter is required).
var citationRE = regexp.MustCompile(`^(.+?)[\s\p{Z}]+([0-9]+):([0-9]+(?:-[0-9]+|,[0-9]+)*)$`)

// ParseCitation splits "約 3:16-17" into book, chapter and verse spec.
func ParseCitation(citation string) (book string, chapter int, verses string, ok bool) {
	m := citationRE.FindStringSubmatch(strings.TrimSpace(citation))
	if m == nil {
		return "", 0, "", false
	}
	ch, err := strconv.Atoi(m[2])
	if err != nil {
		return "", 0, "", false
	}
	return m[1], ch, m[3], true
}

// CitationParams are the query_verse_citation inputs.
type CitationParams struct {
	Citation      string
	Version       string
	IncludeStrong bool
	UseSimplified bool
}

// QueryCitation parses the citation string and queries the verses it names.
func (s *Service) QueryCitation(ctx context.Context, p CitationParams) (*VerseResult, error) {
	book, chapter, verses, ok := ParseCitation(p.Citation)
	if !ok {
		return nil, &InvalidParameterError{Name: "citation", Value: p.Citation, Reason: "Invalid citation format. Expected '<book> <chapter>:<verse>', for example '約 3:16' or 'John 3:16-17' (a space is required after the book name)."}
	}
	return s.GetVerse(ctx, VerseParams{
		Book: book, Chapter: chapter, Verse: verses, Version: p.Version,
		IncludeStrong: p.IncludeStrong, UseSimplified: p.UseSimplified,
	})
}

// GetChapter queries a whole chapter (the same as GetVerse without a verse spec, no Strong's numbers).
func (s *Service) GetChapter(ctx context.Context, book string, chapter int, version string, simplified bool) (*VerseResult, error) {
	return s.GetVerse(ctx, VerseParams{Book: book, Chapter: chapter, Version: version, UseSimplified: simplified})
}
