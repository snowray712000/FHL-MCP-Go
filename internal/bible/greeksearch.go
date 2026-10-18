package bible

import (
	"strings"
)

// Greek keyword search in fhlwh (the NT original text) and lxx. se.php answers VERSION=fhlwh with "not found" and
// matches lxx only by exact form (καί does not find καὶ), so the search runs on the pre-built Unicode file. Matching ignores case, accents, breathings, iota subscript
// and the sigma / final sigma difference (FHL spells accents with oxia, keyboards produce tonos, so exact
// matching would miss most typed queries); a keyword may be a part of a word, and every space-separated word
// of the query must occur in the verse.

// greekFolded turns text into lower-case bare letters α..ω; every other character becomes one space (a word
// boundary), runs of them collapse, and the <WG..> numbers are dropped.
func greekFolded(text string) string {
	var b strings.Builder
	space := true
	for _, r := range snMarker.ReplaceAllString(text, "") {
		if f, ok := greekFold[r]; ok {
			if f == 0 {
				continue
			}
			r = f
		}
		if r >= 'α' && r <= 'ω' {
			b.WriteRune(r)
			space = false
		} else if !space {
			b.WriteByte(' ')
			space = true
		}
	}
	return b.String()
}

// hasGreek reports whether the query has a Greek letter (otherwise upstream gets it as before).
func hasGreek(q string) bool {
	return strings.IndexFunc(q, func(r rune) bool { return r >= 0x0370 && r <= 0x03FF || r >= 0x1F00 && r <= 0x1FFF }) >= 0
}

// greekMatches reports whether every word of the folded query occurs in the folded verse.
func greekMatches(verse string, words []string) bool {
	for _, w := range words {
		if !strings.Contains(verse, w) {
			return false
		}
	}
	return true
}
