package bible

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Citation grammar (relaxed from the Python reference's `^(.+?)\s+(\d+):(\d+(?:-\d+|,\d+)*)$`,
// modelled on the cases FHLNUI's reference parser handles). A citation is one or more
// ";"-separated segments; a segment is an optional book followed by comma-separated items:
//
//	<book>[ ]<chapter>:<verses>   verses = n | n-m | n-end | n-e | comma list of those
//	<book>[ ]<chapter>            whole chapter
//	c:v-c2:w                      cross-chapter range (also c:v-c2:end)
//	c:v,c2:w                      a "c:" item starts a new chapter inside a segment
//	<book>                        whole chapter 1, only for one-chapter books (Obad, Philem, 2-3John, Jude)
//	<one-chapter book>[ ]<verses> "猶 5" / "Jude 5-7" mean verses of chapter 1
//
// A segment without a book reuses the previous segment's book ("約3:16;18:1" = 約3:16;約18:1); a bare
// number without a chapter in scope is a whole chapter (FHLNUI's rule). Full-width digits and
// punctuation, "#...|" wrappers and a trailing period are accepted.
var (
	verseItem    = `(?:[0-9]+(?:\s*-\s*(?:[0-9]+|[eE][nN][dD]|[eE]))?)`
	citeCrossRE  = regexp.MustCompile(`^([0-9]+):([0-9]+)-([0-9]+):([0-9]+|end|e)$`)
	citeColonRE  = regexp.MustCompile(`^([0-9]+):(` + verseItem + `)$`)
	citePlainRE  = regexp.MustCompile(`^` + verseItem + `$`)
	citeDigitDot = regexp.MustCompile(`([0-9])．([0-9])`)
	citeSpaces   = regexp.MustCompile(`\s+`)
)

// maxCitationQueries caps the upstream requests one citation may expand to (cross-chapter ranges
// and long ";" lists multiply requests).
const maxCitationQueries = 40

// citeNormalizer maps full-width digits/punctuation to their ASCII forms.
var citeNormalizer = strings.NewReplacer(
	"０", "0", "１", "1", "２", "2", "３", "3", "４", "4", "５", "5", "６", "6", "７", "7", "８", "8", "９", "9",
	"：", ":", "，", ",", "、", ",", "；", ";", "－", "-", "–", "-", "—", "-", "~", "-", "～", "-", "〜", "-",
	" ", " ", "　", " ",
)

func normalizeCitation(s string) string {
	s = citeNormalizer.Replace(s)
	s = citeDigitDot.ReplaceAllString(s, "$1:$2")
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "#")
	s = strings.TrimSuffix(s, "|")
	s = strings.TrimSpace(strings.TrimRight(s, ".。;"))
	return s
}

// isOneChapterBook reports the books with a single chapter (Obadiah, Philemon, 2 John, 3 John, Jude).
func isOneChapterBook(id int) bool {
	switch id {
	case 31, 57, 63, 64, 65:
		return true
	}
	return false
}

// resolveBookLoose resolves a book name, tolerating stray spaces ("1John", "1 John", "Song of Solomon"),
// case and a trailing period. It returns the spelling that resolved.
func resolveBookLoose(name string) (spelling string, id int, ok bool) {
	name = strings.TrimSpace(name)
	if name == "" || isDigits(name) {
		return "", 0, false // a bare number is a chapter here, never a numeric book code
	}
	cands := []string{name, strings.TrimRight(name, "."), citeSpaces.ReplaceAllString(name, " "), citeSpaces.ReplaceAllString(name, "")}
	for _, c := range cands {
		if c == "" {
			continue
		}
		if id, ok := ResolveBookID(c); ok {
			return c, id, true
		}
	}
	return "", 0, false
}

// CitationQuery is one upstream request a citation expands to. Spec is empty for a whole chapter.
type CitationQuery struct {
	Book    string
	Chapter int
	Spec    string
}

func compactSpec(s string) string {
	return strings.ToLower(citeSpaces.ReplaceAllString(s, ""))
}

// splitBook peels the book off a segment: the shortest prefix (ending right before a digit) that
// resolves as a book name. Without a resolvable book it reports the text before the first digit
// as the (unknown) book so the caller can surface "Book not found"; a segment starting with a
// digit has no book (rest = whole segment).
func splitBook(seg string) (spelling string, id int, rest string, hasBook bool) {
	for i := 1; i < len(seg); i++ {
		if seg[i] < '0' || seg[i] > '9' {
			continue
		}
		if sp, bid, ok := resolveBookLoose(seg[:i]); ok {
			return sp, bid, strings.TrimSpace(seg[i:]), true
		}
	}
	if sp, bid, ok := resolveBookLoose(seg); ok { // bare book, no digits at all
		return sp, bid, "", true
	}
	if seg == "" || (seg[0] >= '0' && seg[0] <= '9') {
		return "", 0, seg, false
	}
	if i := strings.IndexAny(seg, "0123456789"); i > 0 {
		return strings.TrimSpace(seg[:i]), 0, strings.TrimSpace(seg[i:]), true
	}
	return strings.TrimSpace(seg), 0, "", true
}

// ParseCitationQueries expands a citation into the upstream queries it names.
func ParseCitationQueries(citation string) ([]CitationQuery, bool) {
	s := normalizeCitation(citation)
	if s == "" {
		return nil, false
	}
	var out []CitationQuery
	book, bookID := "", 0
	for _, seg := range strings.Split(s, ";") {
		seg = strings.TrimSpace(seg)
		if seg == "" {
			continue
		}
		if sp, id, rest, has := splitBook(seg); has {
			book, bookID, seg = sp, id, rest
		} else if book == "" {
			return nil, false // the first segment must name a book
		}
		qs, ok := parseSegment(book, bookID, seg)
		if !ok {
			return nil, false
		}
		out = append(out, qs...)
		if len(out) > maxCitationQueries {
			return nil, false
		}
	}
	if len(out) == 0 {
		return nil, false
	}
	return mergeAdjacent(out), true
}

// parseSegment parses the comma-separated items after a book.
func parseSegment(book string, bookID int, rest string) ([]CitationQuery, bool) {
	oneChap := isOneChapterBook(bookID)
	if rest == "" {
		if !oneChap {
			return nil, false
		}
		return []CitationQuery{{Book: book, Chapter: 1}}, true // bare one-chapter book = whole chapter
	}
	var out []CitationQuery
	cur := 0
	if oneChap && !strings.Contains(rest, ":") {
		cur = 1 // "猶 5-7" = verses of chapter 1
	}
	add := func(ch int, spec string) { out = append(out, CitationQuery{Book: book, Chapter: ch, Spec: spec}) }
	for _, it := range strings.Split(rest, ",") {
		it = compactSpec(it)
		switch {
		case citeCrossRE.MatchString(it):
			m := citeCrossRE.FindStringSubmatch(it)
			c1, _ := strconv.Atoi(m[1])
			v1, _ := strconv.Atoi(m[2])
			c2, _ := strconv.Atoi(m[3])
			if c2 < c1 || c1 < 1 || c2-c1 > maxCitationQueries {
				return nil, false
			}
			if c1 == c2 {
				add(c1, fmt.Sprintf("%d-%s", v1, m[4]))
			} else {
				add(c1, fmt.Sprintf("%d-end", v1))
				for c := c1 + 1; c < c2; c++ {
					add(c, "")
				}
				add(c2, "1-"+m[4])
			}
			cur = c2
		case citeColonRE.MatchString(it):
			m := citeColonRE.FindStringSubmatch(it)
			cur, _ = strconv.Atoi(m[1])
			add(cur, m[2])
		case citePlainRE.MatchString(it):
			if cur != 0 {
				add(cur, it)
			} else if n, err := strconv.Atoi(it); err == nil {
				cur = n // a bare number with no chapter in scope is a whole chapter
				add(n, "")
			} else {
				return nil, false // chapter ranges ("3-4") are not supported
			}
		default:
			return nil, false
		}
	}
	return out, true
}

// mergeAdjacent joins consecutive verse specs of the same book and chapter ("16" + "18" -> "16,18")
// so a plain citation stays one upstream request.
func mergeAdjacent(qs []CitationQuery) []CitationQuery {
	out := make([]CitationQuery, 0, len(qs))
	for _, q := range qs {
		if n := len(out); n > 0 {
			p := &out[n-1]
			if p.Book == q.Book && p.Chapter == q.Chapter && p.Spec != "" && q.Spec != "" {
				p.Spec += "," + q.Spec
				continue
			}
		}
		out = append(out, q)
	}
	return out
}

// ParseCitation splits a citation that names exactly one chapter request, such as "約 3:16-17",
// into book, chapter and verse spec. verses is empty for a whole chapter and may contain "end".
func ParseCitation(citation string) (book string, chapter int, verses string, ok bool) {
	qs, ok := ParseCitationQueries(citation)
	if !ok || len(qs) != 1 {
		return "", 0, "", false
	}
	return qs[0].Book, qs[0].Chapter, qs[0].Spec, true
}

// verseRange is one inclusive range of a verse spec; hi == 0 means "to the end of the chapter".
type verseRange struct{ lo, hi int }

// parseVerseSpec parses "16", "16-18", "16-end", "1,3-e". A reversed range is swapped.
func parseVerseSpec(spec string) (rs []verseRange, hasEnd bool) {
	for _, part := range strings.Split(spec, ",") {
		lo, hi, isRange := strings.Cut(part, "-")
		l, _ := strconv.Atoi(lo)
		r := verseRange{lo: l, hi: l}
		if isRange {
			if hi == "end" || hi == "e" {
				r.hi, hasEnd = 0, true
			} else {
				h, _ := strconv.Atoi(hi)
				r.hi = h
				if r.hi < r.lo {
					r.lo, r.hi = r.hi, r.lo
				}
			}
		}
		rs = append(rs, r)
	}
	return rs, hasEnd
}

func (r verseRange) has(v int) bool { return v >= r.lo && (r.hi == 0 || v <= r.hi) }

// CitationParams are the query_verse_citation inputs.
type CitationParams struct {
	Citation      string
	Version       string
	IncludeStrong bool
	UseSimplified bool
}

// QueryCitation parses the citation string and queries the verses it names. A citation that
// expands to several requests (";" lists, cross-chapter ranges) returns one result with the verses
// concatenated in citation order and no navigation.
func (s *Service) QueryCitation(ctx context.Context, p CitationParams) (*VerseResult, error) {
	qs, ok := ParseCitationQueries(p.Citation)
	if !ok {
		return nil, &InvalidParameterError{Name: "citation", Value: p.Citation, Reason: fmt.Sprintf("Invalid citation format. Expected '<book> <chapter>:<verse>', for example '約 3:16', '約3:16-17', 'John 3:16,18', 'John 3:16-end', a whole chapter such as '約 3', several parts such as '約 3:16; 羅 8:28' or a cross-chapter range such as '約 3:16-4:2' (a one-chapter book may omit the chapter: '猶 5'). At most %d chapter requests per citation.", maxCitationQueries)}
	}
	var res *VerseResult
	for _, q := range qs {
		r, err := s.queryOne(ctx, q, p)
		if err != nil {
			return nil, err
		}
		if res == nil {
			res = r
			continue
		}
		res.Verses = append(res.Verses, r.Verses...)
		res.RecordCount += r.RecordCount
		res.Navigation = nil
	}
	return res, nil
}

func (s *Service) queryOne(ctx context.Context, q CitationQuery, p CitationParams) (*VerseResult, error) {
	vp := VerseParams{Book: q.Book, Chapter: q.Chapter, Verse: q.Spec, Version: p.Version, IncludeStrong: p.IncludeStrong, UseSimplified: p.UseSimplified}
	ranges, hasEnd := parseVerseSpec(q.Spec)
	if q.Spec == "" || !hasEnd {
		return s.GetVerse(ctx, vp)
	}
	// "end" needs the chapter's last verse number: fetch the whole chapter and keep the named verses.
	vp.Verse = ""
	res, err := s.GetVerse(ctx, vp)
	if err != nil {
		return nil, err
	}
	kept := make([]VerseItem, 0, len(res.Verses))
	for _, v := range res.Verses {
		for _, r := range ranges {
			if r.has(v.Verse) {
				kept = append(kept, v)
				break
			}
		}
	}
	res.Verses, res.RecordCount, res.Navigation = kept, len(kept), nil
	return res, nil
}

// GetChapter queries a whole chapter (the same as GetVerse without a verse spec, no Strong's numbers).
func (s *Service) GetChapter(ctx context.Context, book string, chapter int, version string, simplified bool) (*VerseResult, error) {
	return s.GetVerse(ctx, VerseParams{Book: book, Chapter: chapter, Version: version, UseSimplified: simplified})
}
