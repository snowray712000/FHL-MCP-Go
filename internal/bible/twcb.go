package bible

import (
	"context"
	"fmt"
	"html"
	"regexp"
	"strings"

	"github.com/snowray712000/FHL-MCP-Go/internal/fhlapi"
)

// TWCBAPI is the optional upstream part for the 浸宣 (TWCB) dictionaries. Test fakes that only know the
// CBOL dictionary need not implement it.
type TWCBAPI interface {
	TWCB(ctx context.Context, nt bool, number int, simplified bool) (*fhlapi.TWCBResponse, error)
}

// TWCBEntry is the lookup_strongs result when only the TWCB dictionary is asked for.
type TWCBEntry struct {
	StrongsNumber string `json:"strongs_number"`
	Dictionary    string `json:"dictionary"`
	Testament     string `json:"testament"`
	Definition    string `json:"definition"`
}

// StrongsBoth is the lookup_strongs result for dictionary "both": the CBOL entry plus the TWCB text
// (twcb_definition is absent when TWCB has no entry for the number).
type StrongsBoth struct {
	*StrongsEntry
	TWCBDefinition string `json:"twcb_definition,omitempty"`
}

// LookupStrongsIn is lookup_strongs for a dictionary other than the default: "twcb" or "both" (case-insensitive;
// "cbol" is the default path LookupStrongs).
func (s *Service) LookupStrongsIn(ctx context.Context, number string, testament *string, dictionary string, simplified bool) (any, error) {
	dict := strings.ToLower(strings.TrimSpace(dictionary))
	if dict != "twcb" && dict != "both" && dict != "cbol" {
		return nil, &InvalidParameterError{Name: "dictionary", Value: dictionary, Reason: "must be 'cbol', 'twcb' or 'both'"}
	}
	n, ot, err := ParseStrongsInput(number, testament)
	if err != nil {
		return nil, err
	}
	if dict == "cbol" {
		return s.lookup(ctx, number, n, ot, simplified)
	}
	api, ok := s.API.(TWCBAPI)
	if !ok {
		return nil, fmt.Errorf("the TWCB dictionary is not supported by this upstream client")
	}
	text, found, err := twcbText(ctx, api, n, ot, simplified)
	if err != nil {
		return nil, err
	}
	if dict == "twcb" {
		if !found {
			return nil, &InvalidParameterError{Name: "number", Value: number, Reason: fmt.Sprintf("Strong's #%d (%s) not found in the TWCB dictionary", n, ot)}
		}
		return &TWCBEntry{StrongsNumber: fmt.Sprintf("%05d", n), Dictionary: "TWCB", Testament: ot, Definition: text}, nil
	}
	entry, err := s.lookup(ctx, number, n, ot, simplified)
	if err != nil {
		return nil, err
	}
	return &StrongsBoth{StrongsEntry: entry, TWCBDefinition: text}, nil
}

func twcbText(ctx context.Context, api TWCBAPI, n int, testament string, simplified bool) (string, bool, error) {
	resp, err := api.TWCB(ctx, testament == "NT", n, simplified)
	if err != nil {
		return "", false, err
	}
	if len(resp.Record) == 0 {
		return "", false, nil
	}
	text := cleanTWCB(resp.Record[0].DicText)
	return text, text != "", nil
}

var (
	twcbDivOpen = regexp.MustCompile(`<div[^>]*>`)
	twcbTag     = regexp.MustCompile(`</?[A-Za-z][^>]*>`)
	twcbRef     = regexp.MustCompile(`#([^#|\n]*)\|`)
	twcbBlank   = regexp.MustCompile(`\n{3,}`)
)

// cleanTWCB turns the dictionary markup into plain text: line breaks become "\n", each <div class="idt"> starts
// a new line, the other tags (<span class="exp"> …) are dropped, and the citation wrapper #太3:11| becomes
// the bare citation 太3:11.
func cleanTWCB(raw string) string {
	t := strings.ReplaceAll(raw, "\r\n", "\n")
	t = strings.ReplaceAll(t, "\r", "\n")
	t = twcbDivOpen.ReplaceAllString(t, "\n")
	t = strings.ReplaceAll(t, "</div>", "\n")
	t = twcbTag.ReplaceAllString(t, "")
	t = html.UnescapeString(t)
	t = twcbRef.ReplaceAllString(t, "$1")
	lines := strings.Split(t, "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, " \t\u3000")
	}
	t = twcbBlank.ReplaceAllString(strings.Join(lines, "\n"), "\n\n")
	return strings.TrimSpace(t)
}
