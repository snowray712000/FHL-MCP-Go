package bible

import (
	"strconv"
	"strings"
)

// bookIndex maps every accepted name form to a book ID. Built once at init.
var (
	byID       = map[int]Book{}
	byEngShort = map[string]int{}
	byEngFull  = map[string]int{}
	byChiShort = map[string]int{}
	byChiFull  = map[string]int{}
)

func init() {
	// books is ordered canon-first, so the first entry wins on a duplicate name.
	// The only duplicate is the Chinese short name "多" (Titus 56 / Tobit 101);
	// the Python reference lets the later entry win (Tobit) — Go resolves it to the canon book.
	for _, b := range books {
		byID[b.ID] = b
		setFirst(byEngShort, strings.ToLower(b.EngShort), b.ID)
		setFirst(byEngFull, strings.ToLower(b.EngFull), b.ID)
		setFirst(byChiShort, b.ChiShort, b.ID)
		setFirst(byChiFull, b.ChiFull, b.ID)
	}
}

func setFirst(m map[string]int, k string, id int) {
	if _, ok := m[k]; !ok {
		m[k] = id
	}
}

// validNumericID reports whether a purely numeric book code is accepted
// (same ranges as the Python reference: table IDs, 101-115, 201-217).
func validNumericID(id int) bool {
	if _, ok := byID[id]; ok {
		return true
	}
	return (id >= 101 && id <= 115) || (id >= 201 && id <= 217)
}

// ResolveBookID resolves a book name to its FHL book ID. It accepts a numeric
// book code, English short/full names (case-insensitive), and Chinese
// short/full names, in traditional or simplified form. No whitespace trimming and no alias table: this matches
// the Python reference's BookNameConverter.get_book_id (alias handling is D11, M2).
func ResolveBookID(name string) (int, bool) {
	if id, ok := resolveBookExact(name); ok {
		return id, true
	}
	// Simplified Chinese names (约翰福音, 创, 启) fall back to the traditional table.
	if folded, changed := foldSimplifiedBook(name); changed {
		return resolveBookExact(folded)
	}
	return 0, false
}

func resolveBookExact(name string) (int, bool) {
	if name == "" {
		return 0, false
	}
	if isDigits(name) {
		id, err := strconv.Atoi(name)
		if err != nil || !validNumericID(id) {
			return 0, false
		}
		return id, true
	}
	key := name
	if isASCII(name) {
		key = strings.ToLower(name)
	}
	if id, ok := byEngShort[key]; ok {
		return id, true
	}
	if id, ok := byEngFull[key]; ok {
		return id, true
	}
	if id, ok := byChiShort[name]; ok {
		return id, true
	}
	if id, ok := byChiFull[name]; ok {
		return id, true
	}
	return 0, false
}

func isDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return s != ""
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= 0x80 {
			return false
		}
	}
	return true
}
