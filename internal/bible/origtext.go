package bible

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

// OrigDataAPI is the optional offline-data part of the upstream client: the decompressed JSON of one of FHL's
// pre-built original-text files (bible.fhl.net/NUI/index/bible_<name>.json.gz). The real client implements it;
// test fakes need not.
type OrigDataAPI interface {
	OrigData(ctx context.Context, name string) ([]byte, error)
}

// origVersionNames are the versions that carry Strong's numbers only in the pre-built files: qb.php answers
// include_strong=1 for them with plain text. The names are those of uiabv.php.
var origVersionNames = map[string]string{
	"lxx":   "七十士譯本",
	"bhs":   "舊約馬索拉原文",
	"fhlwh": "新約原文",
}

type origRow struct {
	bid, chap, sec int
	text           string
}

// origBible is one parsed file: rows in canonical order plus the [start,end) slice of every chapter.
type origBible struct {
	rows     []origRow
	chapters map[[2]int][2]int
}

func parseOrig(b []byte) (*origBible, error) {
	var f struct {
		Col  []string `json:"col"`
		Data [][]any  `json:"data"`
	}
	if err := json.Unmarshal(b, &f); err != nil {
		return nil, err
	}
	if len(f.Col) != 4 || f.Col[0] != "book" || f.Col[3] != "text" {
		return nil, fmt.Errorf("unexpected columns %v", f.Col)
	}
	ob := &origBible{rows: make([]origRow, 0, len(f.Data)), chapters: map[[2]int][2]int{}}
	for i, d := range f.Data {
		if len(d) != 4 {
			return nil, fmt.Errorf("row %d has %d columns", i, len(d))
		}
		bid, ok1 := d[0].(float64)
		chap, ok2 := d[1].(float64)
		sec, ok3 := d[2].(float64)
		text, ok4 := d[3].(string)
		if !ok1 || !ok2 || !ok3 || !ok4 {
			return nil, fmt.Errorf("row %d has the wrong types", i)
		}
		ob.rows = append(ob.rows, origRow{int(bid), int(chap), int(sec), text})
	}
	for i, r := range ob.rows {
		k := [2]int{r.bid, r.chap}
		span, seen := ob.chapters[k]
		if !seen {
			span[0] = i
		}
		span[1] = i + 1
		ob.chapters[k] = span
	}
	return ob, nil
}

type origCache struct {
	mu sync.Mutex
	m  map[string]*origBible
}

// loadOrig returns the parsed file for name, downloading it on first use (the client caches the download).
func (s *Service) loadOrig(ctx context.Context, api OrigDataAPI, name string) (*origBible, error) {
	s.orig.mu.Lock()
	defer s.orig.mu.Unlock()
	if ob, ok := s.orig.m[name]; ok {
		return ob, nil
	}
	b, err := api.OrigData(ctx, name)
	if err != nil {
		return nil, err
	}
	ob, err := parseOrig(b)
	if err != nil {
		return nil, fmt.Errorf("original-text file %s: %w", name, err)
	}
	if s.orig.m == nil {
		s.orig.m = map[string]*origBible{}
	}
	s.orig.m[name] = ob
	return ob, nil
}

// getVerseOrig answers a get_bible_verse request for lxx / bhs / fhlwh with Strong's numbers from the pre-built
// files. ok is false when it cannot (not such a request, client without the files, download failure) and the
// caller then asks the upstream API as before, which has the text but no numbers for these versions.
func (s *Service) getVerseOrig(ctx context.Context, bid int, p VerseParams, version string) (res *VerseResult, ok bool) {
	vname, known := origVersionNames[version]
	api, hasAPI := s.API.(OrigDataAPI)
	if !p.IncludeStrong || !known || !hasAPI {
		return nil, false
	}
	ob, err := s.loadOrig(ctx, api, version)
	if err != nil {
		return nil, false
	}
	span := ob.chapters[[2]int{bid, p.Chapter}]
	var ranges []verseRange
	if p.Verse != "" {
		ranges, _ = parseVerseSpec(p.Verse)
	}
	var keep []int // indexes into ob.rows
	for i := span[0]; i < span[1]; i++ {
		if len(ranges) == 0 {
			keep = append(keep, i)
			continue
		}
		for _, vr := range ranges {
			if vr.has(ob.rows[i].sec) {
				keep = append(keep, i)
				break
			}
		}
	}
	res = &VerseResult{Version: version, VersionName: vname, RecordCount: len(keep), Verses: make([]VerseItem, 0, len(keep))}
	for _, i := range keep {
		r := ob.rows[i]
		b := byID[r.bid]
		res.Verses = append(res.Verses, newVerseItem(b.ChiShort, b.EngShort, r.chap, r.sec, r.text))
	}
	if len(keep) > 0 {
		// like upstream: the verse before the first one (itself at the start of the file) and the one after the last
		prev, next := max(keep[0]-1, 0), keep[len(keep)-1]+1
		nav := &Navigation{Prev: origRef(ob.rows[prev])}
		if next < len(ob.rows) {
			nav.Next = origRef(ob.rows[next])
		}
		res.Navigation = nav
	}
	return res, true
}

func origRef(r origRow) *NavRef {
	b := byID[r.bid]
	return &NavRef{Book: b.ChiShort, BookEng: b.EngShort, Chapter: r.chap, Verse: r.sec}
}
