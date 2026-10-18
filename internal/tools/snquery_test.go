package tools

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/snowray712000/FHL-MCP-Go/internal/fhlapi"
)

// snAPI serves fixed verse sets per Strong's number: 34 -> 1:1-3, 35 -> 1:2-4, 36 -> 1:3.
type snAPI struct {
	searchAPI
	reqs []fhlapi.SERequest
}

func (f *snAPI) SE(_ context.Context, r fhlapi.SERequest) (*fhlapi.SEResponse, error) {
	f.reqs = append(f.reqs, r)
	sets := map[string][]int{"34": {1, 2, 3}, "35": {2, 3, 4}, "36": {3}}
	out := &fhlapi.SEResponse{Status: "success"}
	for _, v := range sets[r.Query] {
		out.Record = append(out.Record, fhlapi.SERecord{Chineses: "約", Engs: "Joh", Chap: 1, Sec: v})
	}
	out.RecordCount = len(out.Record)
	return out, nil
}

// SD answers "not found", so the dictionary lookups of a combined query are skipped.
func (f *snAPI) SD(context.Context, bool, int, bool) (*fhlapi.SDResponse, error) {
	return &fhlapi.SDResponse{Status: "success"}, nil
}

func snVerses(t *testing.T, got string) (total int, vs []int) {
	t.Helper()
	var res struct {
		Occ struct {
			Total   int `json:"total_count"`
			Results []struct {
				Verse int `json:"verse"`
			} `json:"results"`
		} `json:"occurrences"`
		Total   int `json:"total_count"`
		Results []struct {
			Verse int `json:"verse"`
		} `json:"results"`
	}
	if err := json.Unmarshal([]byte(got), &res); err != nil {
		t.Fatal(err)
	}
	total = res.Total + res.Occ.Total
	for _, r := range res.Results {
		vs = append(vs, r.Verse)
	}
	for _, r := range res.Occ.Results {
		vs = append(vs, r.Verse)
	}
	return
}

func join(vs []int) string {
	s := make([]string, len(vs))
	for i, v := range vs {
		s[i] = strconv.Itoa(v)
	}
	return strings.Join(s, ",")
}

func TestStrongsOccurrencesCombined(t *testing.T) {
	cases := []struct{ number, want string }{
		{"G34 or G35", "1,2,3,4"},
		{"G34, G35", "1,2,3,4"},
		{"G34，G35", "1,2,3,4"},
		{"G34 G35", "2,3"},
		{"G34 and G35", "2,3"},
		{"G34&G35", "2,3"},
		{"G34 not G35", "1"},
		{"G34 !G35", "1"},
		{"G34-G36", "1,2,3,4"},
		{"G34 G35 or G36 not G34", "2,3"},
		{"g34 OR g36", "1,2,3"},
	}
	for _, c := range cases {
		api := &snAPI{}
		got, err := callOn(t, api, "search_strongs_occurrences", `{"number":"`+c.number+`"}`)
		if err != nil {
			t.Errorf("%q: %v", c.number, err)
			continue
		}
		total, vs := snVerses(t, got)
		if join(vs) != c.want || total != len(vs) {
			t.Errorf("%q: verses %s total %d, want %s", c.number, join(vs), total, c.want)
		}
		if strings.Contains(got, `"strongs_info":`) {
			t.Errorf("%q: combined result must not carry strongs_info", c.number)
		}
	}
}

func TestStrongsOccurrencesCombinedLimitAndRequests(t *testing.T) {
	api := &snAPI{}
	got, err := callOn(t, api, "search_strongs_occurrences", `{"number":"G34 or G35 or G34","limit":2}`)
	if err != nil {
		t.Fatal(err)
	}
	total, vs := snVerses(t, got)
	if total != 4 || join(vs) != "1,2" {
		t.Errorf("total %d verses %s, want total 4 verses 1,2", total, join(vs))
	}
	if len(api.reqs) != 2 { // G34 searched once although it appears twice
		t.Errorf("requests: %+v", api.reqs)
	}
	for _, r := range api.reqs {
		if r.Orig != 1 || r.Range != 1 {
			t.Errorf("Greek number must search the NT: %+v", r)
		}
	}
}

func TestSearchBibleCombinedNumbers(t *testing.T) {
	api := &snAPI{}
	got, err := callOn(t, api, "search_bible", `{"query":"34 or 36","search_type":"greek_number"}`)
	if err != nil {
		t.Fatal(err)
	}
	if _, vs := snVerses(t, got); join(vs) != "1,2,3" {
		t.Errorf("verses %s", join(vs))
	}
	got, err = callOn(t, api, "search_bible_advanced", `{"query":"G34 G35","search_type":"greek_number","range_start":43,"range_end":43}`)
	if err != nil {
		t.Fatal(err)
	}
	if _, vs := snVerses(t, got); join(vs) != "2,3" {
		t.Errorf("advanced verses %s", join(vs))
	}
	if last := api.reqs[len(api.reqs)-1]; last.Range != 3 || last.RangeBID != 43 {
		t.Errorf("advanced must keep its book range: %+v", last)
	}
}

func TestStrongsCombinedErrors(t *testing.T) {
	for _, n := range []string{"not G34", "G34 or", "G34 not", "G34-G33", "G34-H3", "G34 foo", "34 or 35"} {
		api := &snAPI{}
		if _, err := callOn(t, api, "search_strongs_occurrences", `{"number":"`+n+`"}`); err == nil {
			t.Errorf("%q should be rejected", n)
		}
		if len(api.reqs) != 0 {
			t.Errorf("%q must fail before any upstream request", n)
		}
	}
	api := &snAPI{}
	if _, err := callOn(t, api, "search_strongs_occurrences", `{"number":"34 or 35","testament":"NT"}`); err != nil {
		t.Errorf("bare numbers with testament: %v", err)
	}
}
