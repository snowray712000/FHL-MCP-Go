package bible

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// resolveCanonBook resolves a book for the original-language tools, which only cover the 66 canon books.
func resolveCanonBook(book, tool string) (int, error) {
	id, ok := ResolveBookID(book)
	if !ok {
		return 0, &InvalidParameterError{Name: "book", Value: book, Reason: "Book not found. Please check the spelling or use a valid book code."}
	}
	if id > 66 {
		return 0, &InvalidParameterError{Name: "book", Value: book, Reason: tool + " only supports the 66 canon books (book codes 1-66)."}
	}
	return id, nil
}

// WordItem is one analysed word (field order is part of the contract).
type WordItem struct {
	Position      int `json:"position"`
	Word          any `json:"word"`
	StrongsNumber any `json:"strongs_number"`
	PartOfSpeech  any `json:"part_of_speech"`
	Morphology    any `json:"morphology"`
	Lemma         any `json:"lemma"`
	Gloss         any `json:"gloss"`
	Remark        any `json:"remark,omitempty"`
}

// WordAnalysisResult is the get_word_analysis result object.
type WordAnalysisResult struct {
	Testament    string     `json:"testament"`
	Book         string     `json:"book"` // echoed as the caller wrote it
	Chapter      int        `json:"chapter"`
	Verse        int        `json:"verse"`
	OriginalText any        `json:"original_text"`
	Translation  any        `json:"translation"`
	WordCount    int        `json:"word_count"`
	Words        []WordItem `json:"words"`
	Remark       any        `json:"remark,omitempty"`
}

// truthy mirrors Python truthiness for a decoded JSON value.
func truthy(v any) bool {
	switch x := v.(type) {
	case nil:
		return false
	case string:
		return x != ""
	case bool:
		return x
	case interface{ Float64() (float64, error) }:
		f, err := x.Float64()
		return err != nil || f != 0
	case []any:
		return len(x) > 0
	case map[string]any:
		return len(x) > 0
	}
	return true
}

// WordAnalysis returns the word-by-word original-language analysis of one verse (qp.php).
func (s *Service) WordAnalysis(ctx context.Context, book string, chapter, verse int, simplified bool) (*WordAnalysisResult, error) {
	id, err := resolveCanonBook(book, "get_word_analysis")
	if err != nil {
		return nil, err
	}
	resp, err := s.API.QP(ctx, id, chapter, verse, simplified)
	if err != nil {
		return nil, err
	}
	res := &WordAnalysisResult{
		Testament: "NT", Book: book, Chapter: chapter, Verse: verse,
		OriginalText: "", Translation: "", Words: []WordItem{},
	}
	if resp.N == 1 {
		res.Testament = "OT"
	}
	if len(resp.Record) > 0 {
		if wid, ok := widOf(resp.Record[0].WID); ok && wid == 0 {
			first := resp.Record[0]
			res.OriginalText = rawOr(first.Word, "")
			res.Translation = rawOr(first.Exp, "")
			if r := rawOr(first.Remark, ""); truthy(r) {
				res.Remark = r
			}
		}
	}
	for _, r := range resp.Record {
		wid, ok := widOf(r.WID)
		if !ok || wid <= 0 {
			continue
		}
		item := WordItem{
			Position: wid, Word: rawOr(r.Word, ""), StrongsNumber: rawOr(r.SN, ""),
			PartOfSpeech: rawOr(r.Pro, ""), Morphology: rawOr(r.WForm, ""), Lemma: rawOr(r.Orig, ""), Gloss: rawOr(r.Exp, ""),
		}
		if rm := rawOr(r.Remark, nil); truthy(rm) {
			item.Remark = rm
		}
		res.Words = append(res.Words, item)
	}
	res.WordCount = len(res.Words)
	return res, nil
}

// ParsingParams are the get_parsing_reference inputs.
type ParsingParams struct {
	Book       string
	Chapter    int
	Verse      string // "1", "1-4", "1,3,5", "1-2,5,8-10"
	EndChapter *int
	EndVerse   *int
}

// MaxParsingVerses is the most verses one get_parsing_reference request may cover (D21, approved):
// every verse costs one upstream request, so a larger range is refused up front instead of truncated.
const MaxParsingVerses = 30

func tooManyVerses(verse string) error {
	return &InvalidParameterError{Name: "verse", Value: verse, Reason: fmt.Sprintf("the range covers more than %d verses; get_parsing_reference allows at most %d verses per request, split it into smaller requests", MaxParsingVerses, MaxParsingVerses)}
}

// expandVerseString expands "1-2,5,8-10" into the verse numbers, like the Python reference.
// A list longer than MaxParsingVerses is refused before it is expanded.
func expandVerseString(verse string) ([]int, error) {
	var out []int
	for _, part := range strings.Split(verse, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if i := strings.Index(part, "-"); i >= 0 {
			a, errA := strconv.Atoi(strings.TrimSpace(part[:i]))
			b, errB := strconv.Atoi(strings.TrimSpace(part[i+1:]))
			if errA != nil || errB != nil {
				return nil, &InvalidParameterError{Name: "verse", Value: verse, Reason: fmt.Sprintf("Cannot parse verse range: '%s'", part)}
			}
			if b >= a && (b-a < 0 || b-a >= MaxParsingVerses-len(out)) {
				return nil, tooManyVerses(verse)
			}
			for v := a; v <= b; v++ {
				out = append(out, v)
			}
			continue
		}
		n, err := strconv.Atoi(part)
		if err != nil {
			return nil, &InvalidParameterError{Name: "verse", Value: verse, Reason: fmt.Sprintf("Cannot parse verse number: '%s'", part)}
		}
		out = append(out, n)
	}
	if len(out) > MaxParsingVerses {
		return nil, tooManyVerses(verse)
	}
	return out, nil
}

type chapterVerse struct{ chapter, verse int }

func (s *Service) lastVerse(ctx context.Context, engShort string, chapter int) (int, error) {
	res, err := s.GetChapter(ctx, engShort, chapter, "unv", false)
	if err != nil {
		return 0, err
	}
	return len(res.Verses), nil
}

// seq lists from..to, refusing a span that would not fit in room more verses.
func seq(from, to, room int, verse string) ([]int, error) {
	if to >= from && (to-from < 0 || to-from >= room) {
		return nil, tooManyVerses(verse)
	}
	var out []int
	for v := from; v <= to; v++ {
		out = append(out, v)
	}
	return out, nil
}

// resolveRefs works out the ordered (chapter, verse) list to query.
func (s *Service) resolveRefs(ctx context.Context, engShort string, p ParsingParams) ([]chapterVerse, error) {
	start, err := expandVerseString(p.Verse)
	if err != nil {
		return nil, err
	}
	if len(start) == 0 {
		return nil, &InvalidParameterError{Name: "verse", Value: p.Verse, Reason: "Cannot parse the verse range"}
	}
	in := func(ch int, verses []int) []chapterVerse {
		out := make([]chapterVerse, 0, len(verses))
		for _, v := range verses {
			out = append(out, chapterVerse{ch, v})
		}
		return out
	}
	// span lists from..to of one chapter; room is how many more verses the request may still take.
	span := func(ch, from, to, room int) ([]chapterVerse, error) {
		vs, err := seq(from, to, room, p.Verse)
		if err != nil {
			return nil, err
		}
		return in(ch, vs), nil
	}
	if p.EndChapter == nil {
		if p.EndVerse != nil {
			return span(p.Chapter, start[0], *p.EndVerse, MaxParsingVerses)
		}
		return in(p.Chapter, start), nil
	}
	endChapter := *p.EndChapter
	if endChapter == p.Chapter {
		end := start[len(start)-1]
		if p.EndVerse != nil {
			end = *p.EndVerse
		}
		return span(p.Chapter, start[0], end, MaxParsingVerses)
	}
	if endChapter < p.Chapter {
		return nil, &InvalidParameterError{Name: "end_chapter", Value: strconv.Itoa(endChapter), Reason: "end_chapter must be greater than or equal to chapter"}
	}
	if p.EndVerse == nil {
		return nil, &InvalidParameterError{Name: "end_verse", Reason: "required when end_chapter is in a later chapter"}
	}
	// Every chapter has at least one verse, so a span of more than MaxParsingVerses chapters is
	// already too big and needs no chapter-length lookups.
	if endChapter-p.Chapter >= MaxParsingVerses {
		return nil, tooManyVerses(p.Verse)
	}
	last, err := s.lastVerse(ctx, engShort, p.Chapter)
	if err != nil {
		return nil, err
	}
	refs, err := span(p.Chapter, start[0], last, MaxParsingVerses)
	if err != nil {
		return nil, err
	}
	for c := p.Chapter + 1; c < endChapter; c++ {
		last, err := s.lastVerse(ctx, engShort, c)
		if err != nil {
			return nil, err
		}
		more, err := span(c, 1, last, MaxParsingVerses-len(refs))
		if err != nil {
			return nil, err
		}
		refs = append(refs, more...)
	}
	more, err := span(endChapter, 1, *p.EndVerse, MaxParsingVerses-len(refs))
	if err != nil {
		return nil, err
	}
	return append(refs, more...), nil
}

// ParsingReference builds the Markdown "original text / translation / parsing" reference (qp.php).
func (s *Service) ParsingReference(ctx context.Context, p ParsingParams) (string, error) {
	id, err := resolveCanonBook(p.Book, "get_parsing_reference")
	if err != nil {
		return "", err
	}
	book := byID[id]
	hebrew := id < 40

	refs, err := s.resolveRefs(ctx, book.EngShort, p)
	if err != nil {
		return "", err
	}
	addrs := make([]address, len(refs))
	for i, r := range refs {
		addrs[i] = address{id, r.chapter, r.verse}
	}
	tp := computeAddressTP(addrs)

	blocks := []string{"# 參考資料: 原文", scriptureReferenceLegend}
	for i, r := range refs {
		resp, err := s.API.QP(ctx, id, r.chapter, r.verse, false)
		if err != nil {
			return "", err
		}
		if len(resp.Record) == 0 {
			return "", fmt.Errorf("no original-language data for %s %d:%d", book.EngFull, r.chapter, r.verse)
		}
		summary := resp.Record[0]
		a := addrs[i]

		wordText := pyStr(summary.Word, "")
		if hebrew {
			wordText = normalizeOTWordText(wordText)
		}
		words := bindWords(wordText)

		chineseFull := pyOrEmpty(summary.Chinesef)
		if chineseFull == "" {
			chineseFull = book.ChiFull
		}
		blocks = append(blocks, "## 節: "+chineseFull+" "+strconv.Itoa(r.chapter)+":"+strconv.Itoa(r.verse)+" "+formatBCVW(a, -1, tp))
		blocks = append(blocks, "### 類型: 原文", buildOriginalTextSection(words, a, tp, hebrew))
		blocks = append(blocks, "### 類型: 直譯", formatTranslation(pyStr(summary.Exp, "")))

		var table string
		if hebrew {
			table, err = buildHebrewTable(resp.Record, a, tp)
		} else {
			table, err = buildGreekTable(resp.Record, words, a, tp)
		}
		if err != nil {
			return "", err
		}
		blocks = append(blocks, "### 類型: 原文分析", table)
	}
	return strings.Join(blocks, "\n\n"), nil
}
