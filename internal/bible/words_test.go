package bible

import (
	"context"
	"strings"
	"testing"
)

func ptr(n int) *int { return &n }

// D21 (approved): one get_parsing_reference request covers at most MaxParsingVerses verses; a bigger
// range is refused before any upstream request.
func TestParsingReferenceVerseCap(t *testing.T) {
	svc := &Service{} // same-chapter ranges never reach the API
	count := func(p ParsingParams) (int, error) {
		refs, err := svc.resolveRefs(context.Background(), "John", p)
		return len(refs), err
	}
	ok := []struct {
		name string
		p    ParsingParams
		n    int
	}{
		{"range of 30", ParsingParams{Chapter: 1, Verse: "1-30"}, 30},
		{"list of 30", ParsingParams{Chapter: 1, Verse: "1-10,21-30,40-49"}, 30},
		{"end_verse of 30", ParsingParams{Chapter: 1, Verse: "1", EndVerse: ptr(30)}, 30},
		{"same chapter end_chapter", ParsingParams{Chapter: 1, Verse: "5", EndChapter: ptr(1), EndVerse: ptr(34)}, 30},
	}
	for _, c := range ok {
		if n, err := count(c.p); err != nil || n != c.n {
			t.Errorf("%s: n=%d err=%v, want %d", c.name, n, err, c.n)
		}
	}
	tooMany := []struct {
		name string
		p    ParsingParams
	}{
		{"range of 31", ParsingParams{Chapter: 1, Verse: "1-31"}},
		{"list of 31", ParsingParams{Chapter: 1, Verse: "1-10,21-30,40-50"}},
		{"single numbers", ParsingParams{Chapter: 1, Verse: strings.TrimSuffix(strings.Repeat("1,", 31), ",")}},
		{"end_verse of 31", ParsingParams{Chapter: 1, Verse: "1", EndVerse: ptr(31)}},
		{"same chapter end_chapter", ParsingParams{Chapter: 1, Verse: "5", EndChapter: ptr(1), EndVerse: ptr(35)}},
		{"huge range (no allocation)", ParsingParams{Chapter: 1, Verse: "1-9223372036854775807"}},
		{"huge end_verse", ParsingParams{Chapter: 1, Verse: "1", EndVerse: ptr(1 << 40)}},
		{"too many chapters (no lookups)", ParsingParams{Chapter: 1, Verse: "1", EndChapter: ptr(31), EndVerse: ptr(1)}},
	}
	for _, c := range tooMany {
		_, err := count(c.p)
		if err == nil || !strings.Contains(err.Error(), "allows at most 30 verses per request") || !strings.HasPrefix(err.Error(), "Invalid parameter 'verse': ") {
			t.Errorf("%s: err = %v", c.name, err)
		}
	}
}
