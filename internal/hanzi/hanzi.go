// Package hanzi converts Chinese text between simplified and traditional characters,
// character by character. The 信望愛 search endpoints pick their simplified or traditional
// table by `gb` and do not convert the keyword, so the keyword must match the mode.
package hanzi

import "strings"

// ToSimplified converts traditional characters to simplified ones.
func ToSimplified(s string) string { return convert(s, t2sChars) }

// ToTraditional converts simplified characters to traditional ones.
func ToTraditional(s string) string { return convert(s, s2tChars) }

// ForMode returns s in the script the upstream expects for the gb flag.
func ForMode(s string, simplified bool) string {
	if simplified {
		return ToSimplified(s)
	}
	return ToTraditional(s)
}

func convert(s string, m map[rune]rune) string {
	if s == "" {
		return s
	}
	changed := false
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if t, ok := m[r]; ok {
			r, changed = t, true
		}
		b.WriteRune(r)
	}
	if !changed {
		return s
	}
	return b.String()
}
