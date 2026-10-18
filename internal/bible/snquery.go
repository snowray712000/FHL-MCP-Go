package bible

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/snowray712000/FHL-MCP-Go/internal/fhlapi"
)

// Combined Strong's number queries (or / and / not / range), modelled on the FHLNUI web client's
// snQuery: se.php can only search one number at a time, so each number is searched once and the
// verse sets are combined locally.
//
//	or   : "or" "," ";" "|" (and full-width ，；｜、)
//	and  : whitespace, "and", "&" (＆)
//	not  : "not", "!" (！) in front of the number to exclude
//	range: G3114-G3120 (every number of the range is OR-ed)
//
// Precedence: not and range bind tightest, then and, then or:
// "G1 G2 or G3 not G4" = (G1 and G2) or (G3 and not G4).

const (
	snMaxTerms = 60  // distinct numbers per query (each one is a search)
	snMaxRange = 200 // numbers one range may expand to
)

type snTerm struct {
	nt bool // Greek (NT) when true, Hebrew (OT) otherwise
	n  int
}

// snItem is one "and" operand: a number or range (its terms are OR-ed), optionally negated.
type snItem struct {
	not   bool
	terms []snTerm
}

// snQuery is a list of "or" groups, each a list of "and" items.
type snQuery [][]snItem

var snTermRe = regexp.MustCompile(`^([GH]?)(\d+)(?:-([GH]?)(\d+))?$`)

var snNormalizer = strings.NewReplacer(
	"，", ",", "；", ";", "｜", "|", "、", ",", "＆", "&", "！", "!",
	"–", "-", "—", "-", "~", "-", "～", "-", "－", "-",
	"　", " ",
)

var snSpacer = strings.NewReplacer(",", " , ", ";", " , ", "|", " , ", "&", " ", "!", " ! ")

func snErr(input, reason string) error {
	return &InvalidParameterError{Name: "number", Value: input, Reason: reason}
}

// parseSNQuery parses a combined expression. defaultNT says what a number without G/H means
// (nil = not allowed). The returned bool is false when the expression is just one plain number
// (the caller then keeps the single-number path).
func parseSNQuery(input string, defaultNT *bool) (snQuery, bool, error) {
	toks := strings.Fields(snSpacer.Replace(snNormalizer.Replace(input)))
	if len(toks) == 0 {
		return nil, false, snErr(input, "Strong's number must not be empty")
	}
	var q snQuery
	var cur []snItem
	negate := false
	flush := func() error {
		if negate {
			return snErr(input, "'not' must be followed by a Strong's number")
		}
		positive := false
		for _, it := range cur {
			positive = positive || !it.not
		}
		if !positive {
			return snErr(input, "each 'or' group needs at least one number that is not negated")
		}
		q = append(q, cur)
		cur = nil
		return nil
	}
	for _, t := range toks {
		switch strings.ToLower(t) {
		case ",", "or":
			if err := flush(); err != nil {
				return nil, false, err
			}
			continue
		case "and":
			continue
		case "not", "!":
			if negate {
				return nil, false, snErr(input, "'not' must be followed by a Strong's number")
			}
			negate = true
			continue
		}
		terms, err := parseSNTerm(t, defaultNT, input)
		if err != nil {
			return nil, false, err
		}
		cur = append(cur, snItem{not: negate, terms: terms})
		negate = false
	}
	if err := flush(); err != nil {
		return nil, false, err
	}
	seen := map[snTerm]bool{}
	for _, g := range q {
		for _, it := range g {
			for _, t := range it.terms {
				seen[t] = true
			}
		}
	}
	if len(seen) > snMaxTerms {
		return nil, false, snErr(input, fmt.Sprintf("too many Strong's numbers (%d), at most %d per query", len(seen), snMaxTerms))
	}
	single := len(q) == 1 && len(q[0]) == 1 && !q[0][0].not && len(q[0][0].terms) == 1
	return q, !single, nil
}

func parseSNTerm(tok string, defaultNT *bool, input string) ([]snTerm, error) {
	m := snTermRe.FindStringSubmatch(strings.ToUpper(tok))
	if m == nil {
		return nil, snErr(input, fmt.Sprintf("'%s' is not a Strong's number (use e.g. G3056, H430, G3114-G3120, or combine with or / and / not)", tok))
	}
	kind := func(p string) (bool, error) {
		switch p {
		case "G":
			return true, nil
		case "H":
			return false, nil
		}
		if defaultNT == nil {
			return false, snErr(input, fmt.Sprintf("'%s' needs a G/H prefix (or the testament parameter)", tok))
		}
		return *defaultNT, nil
	}
	nt, err := kind(m[1])
	if err != nil {
		return nil, err
	}
	from, _ := strconv.Atoi(m[2])
	to := from
	if m[4] != "" {
		to, _ = strconv.Atoi(m[4])
		nt2, err := kind(m[3])
		if err != nil {
			return nil, err
		}
		if nt2 != nt {
			return nil, snErr(input, fmt.Sprintf("range '%s' mixes G and H", tok))
		}
		if to < from {
			return nil, snErr(input, fmt.Sprintf("range '%s' goes backwards", tok))
		}
		if to-from+1 > snMaxRange {
			return nil, snErr(input, fmt.Sprintf("range '%s' is larger than %d numbers", tok, snMaxRange))
		}
	}
	if from <= 0 {
		return nil, snErr(input, "Strong's number must be greater than 0")
	}
	terms := make([]snTerm, 0, to-from+1)
	for n := from; n <= to; n++ {
		terms = append(terms, snTerm{nt: nt, n: n})
	}
	return terms, nil
}

// positiveTerms lists the distinct non-negated numbers in query order.
func (q snQuery) positiveTerms() []snTerm {
	var out []snTerm
	seen := map[snTerm]bool{}
	for _, g := range q {
		for _, it := range g {
			if it.not {
				continue
			}
			for _, t := range it.terms {
				if !seen[t] {
					seen[t] = true
					out = append(out, t)
				}
			}
		}
	}
	return out
}

type verseKey struct{ book, chap, sec int }

// evalSN searches every distinct number once (base carries version / range / simplified; a whole-Bible
// base is narrowed to the testament of each number) and combines the verse sets.
func (s *Service) evalSN(ctx context.Context, q snQuery, base fhlapi.SERequest) ([]fhlapi.SERecord, error) {
	sets := map[snTerm]map[verseKey]bool{}
	recs := map[verseKey]fhlapi.SERecord{}
	fetch := func(t snTerm) (map[verseKey]bool, error) {
		if set, ok := sets[t]; ok {
			return set, nil
		}
		if found, ok, err := s.origSNRecords(ctx, base, t); ok {
			if err != nil {
				return nil, err
			}
			set := make(map[verseKey]bool, len(found))
			for _, rec := range found {
				id, _ := ResolveBookID(rec.Chineses)
				k := verseKey{id, rec.Chap, rec.Sec}
				set[k] = true
				recs[k] = rec
			}
			sets[t] = set
			return set, nil
		}
		r := base
		r.Query, r.Offset, r.Limit = strconv.Itoa(t.n), 0, seMaxLimit
		r.Orig = 2
		if t.nt {
			r.Orig = 1
		}
		if base.Range == 0 {
			r.Range = 2
			if t.nt {
				r.Range = 1
			}
		}
		resp, err := s.searchPaged(ctx, r)
		if err != nil {
			return nil, err
		}
		if len(resp.Record) >= seMaxLimit {
			return nil, snErr(r.Query, fmt.Sprintf("Strong's number %s has %d or more verses, too many to combine; narrow the scope", snName(t), seMaxLimit))
		}
		set := make(map[verseKey]bool, len(resp.Record))
		for _, rec := range resp.Record {
			id, ok := ResolveBookID(rec.Chineses)
			if !ok {
				id, _ = ResolveBookID(rec.Engs)
			}
			k := verseKey{id, rec.Chap, rec.Sec}
			set[k] = true
			recs[k] = rec
		}
		sets[t] = set
		return set, nil
	}
	union := func(it snItem) (map[verseKey]bool, error) {
		out := map[verseKey]bool{}
		for _, t := range it.terms {
			set, err := fetch(t)
			if err != nil {
				return nil, err
			}
			for k := range set {
				out[k] = true
			}
		}
		return out, nil
	}
	result := map[verseKey]bool{}
	for _, g := range q {
		var acc map[verseKey]bool
		for _, it := range g { // positives first, so the negatives subtract from a known set
			if it.not {
				continue
			}
			set, err := union(it)
			if err != nil {
				return nil, err
			}
			if acc == nil {
				acc = set
				continue
			}
			for k := range acc {
				if !set[k] {
					delete(acc, k)
				}
			}
		}
		for _, it := range g {
			if !it.not {
				continue
			}
			set, err := union(it)
			if err != nil {
				return nil, err
			}
			for k := range set {
				delete(acc, k)
			}
		}
		for k := range acc {
			result[k] = true
		}
	}
	keys := make([]verseKey, 0, len(result))
	for k := range result {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		a, b := keys[i], keys[j]
		if a.book != b.book {
			return a.book < b.book
		}
		if a.chap != b.chap {
			return a.chap < b.chap
		}
		return a.sec < b.sec
	})
	out := make([]fhlapi.SERecord, len(keys))
	for i, k := range keys {
		out[i] = recs[k]
	}
	return out, nil
}

func snName(t snTerm) string {
	if t.nt {
		return "G" + strconv.Itoa(t.n)
	}
	return "H" + strconv.Itoa(t.n)
}

// page applies offset / limit to the combined records.
func pageRecords(rs []fhlapi.SERecord, offset, limit int) []fhlapi.SERecord {
	if offset < 0 {
		offset = 0
	}
	if offset >= len(rs) {
		return nil
	}
	rs = rs[offset:]
	if limit >= 0 && limit < len(rs) {
		rs = rs[:limit]
	}
	return rs
}

// searchSNCombined returns (records of the requested page, total matches, true) when query is a
// combined expression; ok is false when it is not (the caller keeps its normal path).
// searchType is "greek_number" / "hebrew_number" (what a bare number means) or "" (G/H required).
func (s *Service) searchSNCombined(ctx context.Context, query, searchType string, base fhlapi.SERequest, offset, limit int) (recs []fhlapi.SERecord, total int, ok bool, err error) {
	var def *bool
	switch searchType {
	case "greek_number":
		t := true
		def = &t
	case "hebrew_number":
		f := false
		def = &f
	}
	q, combo, perr := parseSNQuery(query, def)
	if perr != nil {
		return nil, 0, true, perr
	}
	if !combo && !s.hasOrigSN(base.Version) {
		return nil, 0, false, nil // one plain number: the caller keeps its normal path
	}
	all, err := s.evalSN(ctx, q, base)
	if err != nil {
		return nil, 0, true, err
	}
	return pageRecords(all, offset, limit), len(all), true, nil
}
