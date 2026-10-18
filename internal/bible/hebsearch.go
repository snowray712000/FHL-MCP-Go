package bible

import (
	"context"
	"regexp"
	"strings"

	"github.com/snowray712000/FHL-MCP-Go/internal/fhlapi"
)

// Hebrew keyword search in bhs. se.php stores bhs in FHL's old ASCII code and answers a Unicode query with
// 0 records, so the search runs on the pre-built Unicode file (the numbers <WH..> are dropped for matching).
//
// The rules follow FHLNUI (docs z260927a): the query is matched per letter cluster (a letter with its points).
//   - final forms equal the normal letter; dagesh and shin / sin dots must match only when the query has them;
//   - vowels must match exactly when the query letter has any, and are free when it has none (Hebrew keyboards
//     mostly type the bare consonants);
//   - cantillation marks and meteg are ignored; spaces and maqaf separate words, and every word of the query
//     must occur in the verse (any order, a word may be a part of a longer word).

// cluster is one Hebrew letter with its points; letter 0 is a boundary (space, maqaf, punctuation, line break).
type cluster struct {
	letter rune
	marks  uint16 // bits 0..11 vowels U+05B0..05BB, 12 dagesh, 13 shin dot, 14 sin dot
}

const (
	markDagesh = 1 << 12
	markShin   = 1 << 13
	markSin    = 1 << 14
	vowelBits  = 1<<12 - 1
)

var hebSofit = map[rune]rune{'ך': 'כ', 'ם': 'מ', 'ן': 'נ', 'ף': 'פ', 'ץ': 'צ'}

var snMarker = regexp.MustCompile(`<W[GH]\d+[a-z]?>`)

func isHebLetter(r rune) bool { return r >= 'א' && r <= 'ת' }

// hebClusters splits text into clusters; ignored marks vanish, every other character is a boundary.
func hebClusters(text string) []cluster {
	var out []cluster
	for _, r := range snMarker.ReplaceAllString(text, "") {
		switch {
		case isHebLetter(r):
			if f, ok := hebSofit[r]; ok {
				r = f
			}
			out = append(out, cluster{letter: r})
		case r >= 0x05B0 && r <= 0x05BB, r == 0x05C7: // vowels (qamats qatan counts as qamats)
			if n := len(out); n > 0 && out[n-1].letter != 0 {
				bit := uint16(1) << (r - 0x05B0)
				if r == 0x05C7 {
					bit = 1 << (0x05B8 - 0x05B0)
				}
				out[n-1].marks |= bit
			}
		case r == 0x05BC, r == 0x05C1, r == 0x05C2:
			if n := len(out); n > 0 && out[n-1].letter != 0 {
				out[n-1].marks |= map[rune]uint16{0x05BC: markDagesh, 0x05C1: markShin, 0x05C2: markSin}[r]
			}
		case r >= 0x0591 && r <= 0x05AF, r == 0x05BD, r == 0x05BF, r == 0x05C4, r == 0x05C5: // cantillation, meteg, rafe
		default:
			if n := len(out); n == 0 || out[n-1].letter != 0 {
				out = append(out, cluster{})
			}
		}
	}
	return out
}

// hebWord matches when the query word q occurs in the cluster stream v.
func hebWord(v, q []cluster) bool {
	for i := 0; i+len(q) <= len(v); i++ {
		ok := true
		for j, c := range q {
			x := v[i+j]
			consonant := uint16(markDagesh | markShin | markSin)
			if x.letter != c.letter || c.marks&consonant&^x.marks != 0 ||
				c.marks&vowelBits != 0 && c.marks&vowelBits != x.marks&vowelBits {
				ok = false
				break
			}
		}
		if ok {
			return true
		}
	}
	return false
}

// hasHebrew reports whether the query has a Hebrew letter (otherwise it may be FHL's ASCII code and goes upstream).
func hasHebrew(q string) bool { return strings.IndexFunc(q, isHebLetter) >= 0 }

// hebQueryWords splits a Hebrew query into words of clusters.
func hebQueryWords(q string) [][]cluster {
	var words [][]cluster
	var cur []cluster
	for _, c := range hebClusters(q) {
		if c.letter == 0 {
			if len(cur) > 0 {
				words, cur = append(words, cur), nil
			}
			continue
		}
		cur = append(cur, c)
	}
	if len(cur) > 0 {
		words = append(words, cur)
	}
	return words
}

// origKeyword answers a keyword se.php request for bhs (Hebrew query), fhlwh or lxx (Greek query) from the pre-built file.
// ok is false for anything else, and the caller asks upstream.
func (s *Service) origKeyword(ctx context.Context, r fhlapi.SERequest) (*fhlapi.SEResponse, bool, error) {
	version := strings.ToLower(r.Version)
	if r.Orig != 0 || !s.hasOrigSN(version) ||
		!(version == "bhs" && hasHebrew(r.Query) || (version == "fhlwh" || version == "lxx") && hasGreek(r.Query)) {
		return nil, false, nil
	}
	ob, err := s.loadOrig(ctx, s.API.(OrigDataAPI), version)
	if err != nil {
		// se.php still searches lxx (exact forms), so a failed download only costs the accent-insensitivity
		return nil, version != "lxx", err
	}
	var match func(i int) bool
	if version == "fhlwh" || version == "lxx" {
		words := strings.Fields(greekFolded(r.Query))
		if len(words) == 0 {
			return nil, false, nil
		}
		ob.greekOnce.Do(func() {
			ob.greek = make([]string, len(ob.rows))
			for i, row := range ob.rows {
				ob.greek[i] = greekFolded(row.text)
			}
		})
		match = func(i int) bool { return greekMatches(ob.greek[i], words) }
	} else {
		words := hebQueryWords(r.Query)
		ob.hebOnce.Do(func() {
			ob.heb = make([][]cluster, len(ob.rows))
			for i, row := range ob.rows {
				ob.heb[i] = hebClusters(row.text)
			}
		})
		match = func(i int) bool {
			for _, w := range words {
				if !hebWord(ob.heb[i], w) {
					return false
				}
			}
			return true
		}
	}
	var recs []fhlapi.SERecord
	for i, row := range ob.rows {
		if !inSERange(r, row.bid) {
			continue
		}
		if match(i) {
			b := byID[row.bid]
			recs = append(recs, fhlapi.SERecord{Chineses: b.ChiShort, Engs: b.EngShort, Chap: row.chap, Sec: row.sec, BibleText: snMarker.ReplaceAllString(row.text, "")})
		}
	}
	return &fhlapi.SEResponse{Status: "success", RecordCount: len(recs), Record: pageRecords(recs, r.Offset, r.Limit)}, true, nil
}
