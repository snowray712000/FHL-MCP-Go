package bible

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/snowray712000/FHL-MCP-Go/internal/fhlapi"
)

// ParseStrongsInput parses a Strong's number the way the Python reference does:
// "G3056" -> (3056, NT), "H430" -> (430, OT); plain digits need the testament parameter
// ("OT"/"NT", case-insensitive); a G/H prefix wins over the testament parameter;
// leading zeros are dropped.
func ParseStrongsInput(input string, testament *string) (int, string, error) {
	s := strings.ToUpper(strings.TrimSpace(input))
	if s == "" {
		return 0, "", &InvalidParameterError{Name: "number", Reason: "Strong's number must not be empty"}
	}
	var detected, digits string
	switch {
	case strings.HasPrefix(s, "G"):
		detected, digits = "NT", s[1:]
	case strings.HasPrefix(s, "H"):
		detected, digits = "OT", s[1:]
	default:
		if testament == nil {
			return 0, "", &InvalidParameterError{Name: "testament", Reason: fmt.Sprintf("required when number '%s' has no G/H prefix (use 'OT' or 'NT')", input)}
		}
		detected, digits = strings.ToUpper(*testament), s
	}
	if detected != "OT" && detected != "NT" {
		return 0, "", &InvalidParameterError{Name: "testament", Value: detected, Reason: "must be 'OT' or 'NT'"}
	}
	digits = strings.TrimLeft(digits, "0")
	if digits == "" {
		digits = "0"
	}
	n, err := strconv.Atoi(digits)
	if err != nil {
		return 0, "", &InvalidParameterError{Name: "number", Value: input, Reason: "Invalid Strong's number format"}
	}
	if n <= 0 {
		return 0, "", &InvalidParameterError{Name: "number", Value: input, Reason: "Strong's number must be greater than 0"}
	}
	return n, detected, nil
}

// RelatedWord is one related word of a dictionary entry (NT only).
type RelatedWord struct {
	Word        any `json:"word"`
	Number      any `json:"number"`
	Occurrences int `json:"occurrences"`
	Gloss       any `json:"gloss"`
}

// StrongsEntry is the lookup_strongs result object. Optional upstream fields are
// passed through: an absent field is "" and an explicit null stays null, as in Python.
type StrongsEntry struct {
	StrongsNumber     string        `json:"strongs_number"`
	OriginalWord      any           `json:"original_word"`
	ChineseDefinition any           `json:"chinese_definition"`
	EnglishDefinition any           `json:"english_definition"`
	Testament         string        `json:"testament"`
	RelatedWords      []RelatedWord `json:"related_words,omitempty"`
}

// LookupStrongs queries the Strong's dictionary (sd.php).
func (s *Service) LookupStrongs(ctx context.Context, number string, testament *string, simplified bool) (*StrongsEntry, error) {
	n, ot, err := ParseStrongsInput(number, testament)
	if err != nil {
		return nil, err
	}
	return s.lookup(ctx, number, n, ot, simplified)
}

func (s *Service) lookup(ctx context.Context, input string, n int, testament string, simplified bool) (*StrongsEntry, error) {
	resp, err := s.API.SD(ctx, testament == "NT", n, simplified)
	if err != nil {
		return nil, err
	}
	if len(resp.Record) == 0 {
		return nil, &InvalidParameterError{Name: "number", Value: input, Reason: fmt.Sprintf("Strong's #%d (%s) not found", n, testament)}
	}
	rec := resp.Record[0]
	entry := &StrongsEntry{
		StrongsNumber:     rec.SN,
		OriginalWord:      rawOr(rec.Orig, ""),
		ChineseDefinition: rawOr(rec.DicText, ""),
		EnglishDefinition: rawOr(rec.EDicText, ""),
		Testament:         testament,
	}
	for _, w := range rec.Same {
		entry.RelatedWords = append(entry.RelatedWords, RelatedWord{
			Word: rawOr(w.Word, nil), Number: rawOr(w.CSN, nil), Occurrences: countOf(w.CCnt), Gloss: rawOr(w.CExp, ""),
		})
	}
	return entry, nil
}

// rawOr decodes an optional JSON value: absent -> def, null -> nil, otherwise the value
// (numbers keep their upstream text via json.Number).
func rawOr(raw json.RawMessage, def any) any {
	if len(raw) == 0 {
		return def
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return def
	}
	return v
}

// countOf mirrors `int(x) if x else 0` for the ccnt field (number or numeric string).
func countOf(raw json.RawMessage) int {
	switch v := rawOr(raw, nil).(type) {
	case json.Number:
		n, _ := strconv.Atoi(v.String())
		return n
	case string:
		n, _ := strconv.Atoi(strings.TrimSpace(v))
		return n
	}
	return 0
}

// Occurrences is the "occurrences" object of search_strongs_occurrences.
type Occurrences struct {
	TotalCount int         `json:"total_count"`
	Showing    int         `json:"showing"`
	Results    []VerseItem `json:"results"`
}

// StrongsOccurrencesResult is the search_strongs_occurrences result object.
type StrongsOccurrencesResult struct {
	StrongsInfo *StrongsEntry `json:"strongs_info,omitempty"`
	// StrongsInfos replaces StrongsInfo for a combined query (or / and / not / range): the
	// dictionary entries of its non-negated numbers (at most snInfoMax, unknown numbers skipped).
	StrongsInfos []*StrongsEntry `json:"strongs_infos,omitempty"`
	Occurrences  Occurrences     `json:"occurrences"`
}

// SearchStrongsOccurrences looks the number up, then searches the matching testament for it.
func (s *Service) SearchStrongsOccurrences(ctx context.Context, number string, testament *string, limit int, simplified bool) (*StrongsOccurrencesResult, error) {
	if res, ok, err := s.searchStrongsCombined(ctx, number, testament, limit, simplified); ok || err != nil {
		return res, err
	}
	n, ot, err := ParseStrongsInput(number, testament)
	if err != nil {
		return nil, err
	}
	info, err := s.lookup(ctx, number, n, ot, simplified)
	if err != nil {
		return nil, err
	}
	p := SearchParams{Query: strconv.Itoa(n), SearchType: "greek_number", Scope: "nt", Version: "unv", Limit: limit, Simplified: simplified}
	if ot == "OT" {
		p.SearchType, p.Scope = "hebrew_number", "ot"
	}
	found, err := s.SearchBible(ctx, p)
	if err != nil {
		return nil, err
	}
	return &StrongsOccurrencesResult{
		StrongsInfo: info,
		Occurrences: Occurrences{TotalCount: found.TotalCount, Showing: len(found.Results), Results: found.Results},
	}, nil
}

// snInfoMax caps the dictionary lookups made for a combined query.
const snInfoMax = 10

// searchStrongsCombined handles number expressions such as "G34 or G35", "G3114-G3120" or
// "G25 not G26". ok is false for a single plain number (normal path).
func (s *Service) searchStrongsCombined(ctx context.Context, number string, testament *string, limit int, simplified bool) (*StrongsOccurrencesResult, bool, error) {
	var def *bool
	if testament != nil {
		nt := strings.ToUpper(*testament) == "NT"
		if !nt && strings.ToUpper(*testament) != "OT" {
			return nil, false, nil // let the normal path report the bad testament
		}
		def = &nt
	}
	q, combo, err := parseSNQuery(number, def)
	if err != nil {
		return nil, true, err
	}
	if !combo {
		return nil, false, nil
	}
	searchType := ""
	if def != nil {
		searchType = map[bool]string{true: "greek_number", false: "hebrew_number"}[*def]
	}
	recs, total, _, err := s.searchSNCombined(ctx, number, searchType, fhlapi.SERequest{Version: "unv", Simplified: simplified}, 0, limit)
	if err != nil {
		return nil, true, err
	}
	res := &StrongsOccurrencesResult{Occurrences: Occurrences{TotalCount: total, Showing: len(recs), Results: hits(recs)}}
	for i, t := range q.positiveTerms() {
		if i >= snInfoMax {
			break
		}
		if info, err := s.lookup(ctx, snName(t), t.n, map[bool]string{true: "NT", false: "OT"}[t.nt], simplified); err == nil {
			res.StrongsInfos = append(res.StrongsInfos, info)
		}
	}
	return res, true, nil
}
