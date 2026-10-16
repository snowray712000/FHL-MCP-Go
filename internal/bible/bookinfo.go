package bible

// BookEntry is one row of get_book_list (field order is part of the contract).
type BookEntry struct {
	ID       int    `json:"id"`
	EngShort string `json:"eng_short"`
	EngFull  string `json:"eng_full"`
	ChiShort string `json:"chi_short"`
	ChiFull  string `json:"chi_full"`
}

// BookListResult is the get_book_list result object.
type BookListResult struct {
	TotalCount int         `json:"total_count"`
	Testament  string      `json:"testament"`
	Books      []BookEntry `json:"books"`
}

// BookList lists the static book table like the Python reference: "OT" (IDs 1-39)
// and "NT" (IDs 40-66) filter case-insensitively; anything else, including an
// unknown value, returns every table entry (Apocrypha and Apostolic Fathers
// included). The testament value is echoed as given, or "all" when empty.
func BookList(testament string) *BookListResult {
	lo, hi := 0, 0
	switch upperASCII(testament) {
	case "OT":
		lo, hi = 1, 39
	case "NT":
		lo, hi = 40, 66
	}
	list := make([]BookEntry, 0, len(books))
	for _, b := range books {
		if hi != 0 && (b.ID < lo || b.ID > hi) {
			continue
		}
		list = append(list, BookEntry{ID: b.ID, EngShort: b.EngShort, EngFull: b.EngFull, ChiShort: b.ChiShort, ChiFull: b.ChiFull})
	}
	label := testament
	if label == "" {
		label = "all"
	}
	return &BookListResult{TotalCount: len(list), Testament: label, Books: list}
}

// NamePair holds one language's short and full names; a nil pointer is JSON null.
type NamePair struct {
	Short *string `json:"short"`
	Full  *string `json:"full"`
}

// BookNames groups the names by language.
type BookNames struct {
	English NamePair `json:"english"`
	Chinese NamePair `json:"chinese"`
}

// BookInfoResult is the get_book_info result object.
type BookInfoResult struct {
	ID        int       `json:"id"`
	Testament string    `json:"testament"`
	Names     BookNames `json:"names"`
}

// BookInfo describes one book. Like the Python reference, a valid numeric code that
// has no table entry (e.g. 113) yields null names, and every book outside 1-39
// (Apocrypha and Apostolic Fathers included) reports testament "NT".
func BookInfo(name string) (*BookInfoResult, error) {
	id, ok := ResolveBookID(name)
	if !ok {
		return nil, &InvalidParameterError{Name: "book", Value: name, Reason: "Book not found. Please check the spelling or use a valid book code."}
	}
	res := &BookInfoResult{ID: id, Testament: "NT"}
	if id >= 1 && id <= 39 {
		res.Testament = "OT"
	}
	if b, ok := byID[id]; ok {
		res.Names = BookNames{
			English: NamePair{Short: &b.EngShort, Full: &b.EngFull},
			Chinese: NamePair{Short: &b.ChiShort, Full: &b.ChiFull},
		}
	}
	return res, nil
}

func upperASCII(s string) string {
	out := []byte(s)
	for i, c := range out {
		if c >= 'a' && c <= 'z' {
			out[i] = c - 'a' + 'A'
		}
	}
	return string(out)
}
