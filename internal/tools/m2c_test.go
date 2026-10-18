package tools

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/snowray712000/FHL-MCP-Go/internal/bible"
	"github.com/snowray712000/FHL-MCP-Go/internal/fhlapi"
)

// searchAPI answers se.php / sd.php requests from recorded upstream fixtures and records them.
type searchAPI struct {
	t     *testing.T
	se    []fhlapi.SERequest
	sd    []int
	seErr error
}

func (f *searchAPI) QB(context.Context, fhlapi.QBRequest) (*fhlapi.QBResponse, error) {
	return nil, errors.New("QB not expected in this test")
}

func (f *searchAPI) AB(context.Context, bool) (*fhlapi.ABResponse, error) {
	return nil, errors.New("AB not expected in this test")
}

func (f *searchAPI) SE(_ context.Context, r fhlapi.SERequest) (*fhlapi.SEResponse, error) {
	f.se = append(f.se, r)
	if f.seErr != nil {
		return nil, f.seErr
	}
	name := "se_ai_limit3.json"
	switch {
	case r.Range == 3 && r.RangeBID == 43 && r.RangeEID == 43:
		name = "se_ai_john_range_limit3.json"
	case r.Orig == 1 && r.Query == "25":
		name = "se_greek25_nt_limit3.json"
	}
	var out fhlapi.SEResponse
	return &out, json.Unmarshal(fixture(f.t, "upstream", name), &out)
}

func (f *searchAPI) SD(_ context.Context, nt bool, number int, _ bool) (*fhlapi.SDResponse, error) {
	f.sd = append(f.sd, number)
	name := "sd_nt_25.json"
	if number == 99999 {
		name = "sd_nt_99999.json"
	}
	var out fhlapi.SDResponse
	return &out, json.Unmarshal(fixture(f.t, "upstream", name), &out)
}

func callOn(t *testing.T, api bible.API, name, args string) (string, error) {
	t.Helper()
	return NewRegistry(&bible.Service{API: api}).Call(context.Background(), name, json.RawMessage(args))
}

func TestM2cGoldens(t *testing.T) {
	cases := []struct{ tool, args, golden string }{
		{"search_bible", `{"query":"愛","limit":3}`, "search_bible_keyword_ai_limit3.json"},
		{"search_bible_advanced", `{"query":"愛","range_start":43,"range_end":43,"limit":3}`, "search_bible_advanced_john_ai_limit3.json"},
		{"lookup_strongs", `{"number":"25","testament":"NT"}`, "lookup_strongs_nt25.json"},
		{"lookup_strongs", `{"number":"99999","testament":"NT"}`, "lookup_strongs_nt99999_empty.json"},
		{"search_strongs_occurrences", `{"number":"25","testament":"NT","limit":3}`, "search_strongs_occurrences_nt25_limit3.json"},
	}
	for _, c := range cases {
		t.Run(c.golden, func(t *testing.T) {
			got, err := callOn(t, &searchAPI{t: t}, c.tool, c.args)
			if err != nil {
				t.Fatal(err)
			}
			if want := string(fixture(t, "golden", c.golden)); got != want {
				t.Errorf("differs from Python golden:\n got %s\nwant %s", got, want)
			}
		})
	}
}

func TestSearchBibleRequests(t *testing.T) {
	api := &searchAPI{t: t}
	if _, err := callOn(t, api, "search_bible", `{"query":"愛"}`); err != nil {
		t.Fatal(err)
	}
	want := fhlapi.SERequest{Query: "愛", Version: "unv", Orig: 0, Range: 0, Offset: 0, Limit: 50}
	if len(api.se) != 1 || api.se[0] != want {
		t.Errorf("default request = %+v, want %+v", api.se, want)
	}

	api = &searchAPI{t: t}
	if _, err := callOn(t, api, "search_bible", `{"query":"430","search_type":"hebrew_number","scope":"ot","version":"kjv","limit":"7","use_simplified":true,"offset":99}`); err != nil {
		t.Fatal(err)
	}
	want = fhlapi.SERequest{Query: "430", Version: "kjv", Orig: 2, Range: 2, Limit: 7, Simplified: true}
	if len(api.se) != 1 || api.se[0] != want {
		t.Errorf("request = %+v, want %+v (offset is not a search_bible parameter)", api.se, want)
	}
}

func TestSearchBibleValidation(t *testing.T) {
	cases := []struct{ name, args, want string }{
		{"empty query (D7)", `{"query":""}`, "Invalid parameter 'query': must not be empty"},
		{"blank query", `{"query":"   "}`, "Invalid parameter 'query': must not be empty"},
		{"missing query", `{}`, "Invalid parameter 'query': required parameter is missing"},
		{"bad type", `{"query":"a","search_type":"greek"}`, "Invalid parameter 'search_type': greek - Invalid search type. Use 'keyword', 'greek_number' or 'hebrew_number'."},
		{"bad scope", `{"query":"a","scope":"range"}`, "Invalid parameter 'scope': range - Invalid scope. Use 'all', 'ot' or 'nt'."},
		{"bad limit", `{"query":"a","limit":"x"}`, "Invalid parameter 'limit': \"x\" - must be an integer"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			api := &searchAPI{t: t}
			_, err := callOn(t, api, "search_bible", c.args)
			if err == nil || err.Error() != c.want {
				t.Fatalf("err = %v, want %q", err, c.want)
			}
			if len(api.se) != 0 {
				t.Errorf("invalid input must not reach upstream: %+v", api.se)
			}
		})
	}
}

func TestSearchUpstreamFailureIsReported(t *testing.T) {
	api := &searchAPI{t: t, seErr: &fhlapi.Error{Kind: fhlapi.ErrUpstream, Message: `status "Fail: Empty Query"`}}
	_, err := callOn(t, api, "search_bible", `{"query":"x"}`)
	var up *fhlapi.Error
	if !errors.As(err, &up) || !errors.Is(err, fhlapi.ErrUpstream) {
		t.Errorf("err = %v, want the upstream error passed through (D12)", err)
	}
}

func TestSearchAdvanced(t *testing.T) {
	t.Run("one bound only falls back to whole-Bible search (Python)", func(t *testing.T) {
		api := &searchAPI{t: t}
		out, err := callOn(t, api, "search_bible_advanced", `{"query":"愛","range_start":43,"limit":3,"offset":2}`)
		if err != nil {
			t.Fatal(err)
		}
		if len(api.se) != 1 || api.se[0].Range != 0 || api.se[0].Offset != 2 {
			t.Errorf("request = %+v", api.se)
		}
		var v map[string]any
		_ = json.Unmarshal([]byte(out), &v)
		if v["scope"] != "all" || v["range"] != nil {
			t.Errorf("fallback must use the search_bible shape: %v", v)
		}
	})
	t.Run("zero bound is treated as not given", func(t *testing.T) {
		api := &searchAPI{t: t}
		if _, err := callOn(t, api, "search_bible_advanced", `{"query":"愛","range_start":0,"range_end":66}`); err != nil {
			t.Fatal(err)
		}
		if len(api.se) != 1 || api.se[0].Range != 0 {
			t.Errorf("request = %+v", api.se)
		}
	})
	t.Run("range request", func(t *testing.T) {
		api := &searchAPI{t: t}
		if _, err := callOn(t, api, "search_bible_advanced", `{"query":"愛","range_start":"40","range_end":43,"limit":3,"offset":5}`); err != nil {
			t.Fatal(err)
		}
		want := fhlapi.SERequest{Query: "愛", Version: "unv", Range: 3, Offset: 5, Limit: 3, RangeBID: 40, RangeEID: 43}
		if len(api.se) != 1 || api.se[0] != want {
			t.Errorf("request = %+v, want %+v", api.se, want)
		}
	})
	t.Run("documented types work in range mode (D19)", func(t *testing.T) {
		api := &searchAPI{t: t}
		if _, err := callOn(t, api, "search_bible_advanced", `{"query":"25","search_type":"greek_number","range_start":40,"range_end":66}`); err != nil {
			t.Fatal(err)
		}
		if len(api.se) != 1 || api.se[0].Orig != 1 || api.se[0].Range != 3 {
			t.Errorf("request = %+v", api.se)
		}
		if _, err := callOn(t, api, "search_bible_advanced", `{"query":"430","search_type":"hebrew_number","range_start":1,"range_end":39}`); err != nil {
			t.Fatal(err)
		}
		if len(api.se) != 2 || api.se[1].Orig != 2 || api.se[1].Range != 3 || api.se[1].RangeBID != 1 || api.se[1].RangeEID != 39 {
			t.Errorf("request = %+v", api.se)
		}
	})
	errCases := []struct{ name, args, want string }{
		{"reversed", `{"query":"a","range_start":43,"range_end":40}`, "Invalid parameter 'range': 43 > 40 - range_start must not be greater than range_end."},
		{"unknown book number", `{"query":"a","range_start":1,"range_end":999}`, "Invalid parameter 'range': 1 - 999 - Invalid book range. Use FHL book numbers (1-66 for the canon)."},
		{"bad type in range", `{"query":"a","search_type":"greek","range_start":1,"range_end":2}`, "Invalid parameter 'search_type': greek - Invalid search type. Use 'keyword', 'greek_number' or 'hebrew_number'."},
		{"empty query", `{"query":"","range_start":1,"range_end":2}`, "Invalid parameter 'query': must not be empty"},
	}
	for _, c := range errCases {
		t.Run(c.name, func(t *testing.T) {
			api := &searchAPI{t: t}
			_, err := callOn(t, api, "search_bible_advanced", c.args)
			if err == nil || err.Error() != c.want {
				t.Fatalf("err = %v, want %q", err, c.want)
			}
			if len(api.se) != 0 {
				t.Errorf("must not reach upstream: %+v", api.se)
			}
		})
	}
}

func TestParseStrongsInput(t *testing.T) {
	ptr := func(s string) *string { return &s }
	cases := []struct {
		in        string
		testament *string
		n         int
		ot        string
		err       string
	}{
		{"G3056", nil, 3056, "NT", ""},
		{"h430", nil, 430, "OT", ""},
		{" g03056 ", nil, 3056, "NT", ""},
		{"25", ptr("nt"), 25, "NT", ""},
		{"0025", ptr("NT"), 25, "NT", ""},
		{"G25", ptr("OT"), 25, "NT", ""}, // the prefix wins over the testament parameter
		{"", nil, 0, "", "Invalid parameter 'number': Strong's number must not be empty"},
		{"25", nil, 0, "", "Invalid parameter 'testament': required when number '25' has no G/H prefix (use 'OT' or 'NT')"},
		{"25", ptr("both"), 0, "", "Invalid parameter 'testament': BOTH - must be 'OT' or 'NT'"},
		{"25", ptr(""), 0, "", "Invalid parameter 'testament': must be 'OT' or 'NT'"},
		{"Gabc", nil, 0, "", "Invalid parameter 'number': Gabc - Invalid Strong's number format"},
		{"G0", nil, 0, "", "Invalid parameter 'number': G0 - Strong's number must be greater than 0"},
		{"G", nil, 0, "", "Invalid parameter 'number': G - Strong's number must be greater than 0"},
	}
	for _, c := range cases {
		n, ot, err := bible.ParseStrongsInput(c.in, c.testament)
		if c.err != "" {
			if err == nil || err.Error() != c.err {
				t.Errorf("%q: err = %v, want %q", c.in, err, c.err)
			}
			continue
		}
		if err != nil || n != c.n || ot != c.ot {
			t.Errorf("%q: got %d %s %v, want %d %s", c.in, n, ot, err, c.n, c.ot)
		}
	}
}

func TestStrongsRequests(t *testing.T) {
	api := &searchAPI{t: t}
	if _, err := callOn(t, api, "search_strongs_occurrences", `{"number":"H430"}`); err != nil {
		t.Fatal(err)
	}
	want := fhlapi.SERequest{Query: "430", Version: "unv", Orig: 2, Range: 2, Limit: 50}
	if len(api.se) != 1 || api.se[0] != want || len(api.sd) != 1 || api.sd[0] != 430 {
		t.Errorf("se=%+v sd=%v, want se=%+v sd=[430]", api.se, api.sd, want)
	}

	api = &searchAPI{t: t}
	if _, err := callOn(t, api, "lookup_strongs", `{"number":"25"}`); err == nil {
		t.Error("a plain number without testament must be rejected")
	}
	if len(api.sd) != 0 {
		t.Errorf("must not reach upstream: %v", api.sd)
	}
}

func (f *searchAPI) QP(context.Context, int, int, int, bool) (*fhlapi.QPResponse, error) {
	return nil, errors.New("QP not expected in this test")
}

func (f *searchAPI) SC(context.Context, int, int, *int, bool) (*fhlapi.SCResponse, error) {
	return nil, errors.New("SC not expected in this test")
}

func (f *searchAPI) SCList(context.Context, bool) (*fhlapi.SCListResponse, error) {
	return nil, errors.New("SCList not expected in this test")
}

func (f *searchAPI) SSC(context.Context, string, bool) (*fhlapi.SSCResponse, error) {
	return nil, errors.New("SSC not expected in this test")
}

func (f *searchAPI) ST(context.Context, int, string, bool, bool) (*fhlapi.STResponse, error) {
	return nil, errors.New("ST not expected in this test")
}

func (f *searchAPI) AU(context.Context, int, int, int, bool) (*fhlapi.AUResponse, error) {
	return nil, errors.New("AU not expected in this test")
}

func (f *searchAPI) QSub(context.Context, int, int, *string) (*fhlapi.SubVerseResponse, error) {
	return nil, errors.New("QSub not expected in this test")
}

func (f *searchAPI) QAF(context.Context, int, int, *string) (*fhlapi.SubVerseResponse, error) {
	return nil, errors.New("QAF not expected in this test")
}

func (f *searchAPI) SESub(context.Context, string, int, int) (*fhlapi.SubSearchResponse, error) {
	return nil, errors.New("SESub not expected in this test")
}

func (f *searchAPI) SEAF(context.Context, string, int, int) (*fhlapi.SubSearchResponse, error) {
	return nil, errors.New("SEAF not expected in this test")
}

func (f *searchAPI) RT(context.Context, int, int, bool) (*fhlapi.RTResponse, error) {
	return nil, errors.New("RT not expected in this test")
}

func (f *searchAPI) Articles(context.Context, fhlapi.ArticleQuery) (*fhlapi.ArticlesResponse, error) {
	return nil, errors.New("Articles not expected in this test")
}

// pagedAPI mimics se.php's real cap: limit > 500 is refused, otherwise `total` records are served by offset.
type pagedAPI struct {
	searchAPI
	total int
	reqs  []fhlapi.SERequest
}

func (f *pagedAPI) SE(_ context.Context, r fhlapi.SERequest) (*fhlapi.SEResponse, error) {
	f.reqs = append(f.reqs, r)
	if r.Limit > 500 {
		return nil, &fhlapi.Error{Kind: fhlapi.ErrUpstream, Message: `status "Fail: record count > 500"`}
	}
	out := &fhlapi.SEResponse{Status: "success"}
	for i := r.Offset; i < f.total && len(out.Record) < r.Limit; i++ {
		out.Record = append(out.Record, fhlapi.SERecord{Chineses: "約", Engs: "Joh", Chap: 1, Sec: i + 1})
	}
	out.RecordCount = len(out.Record)
	return out, nil
}

func TestSearchBibleLimitAbove500IsPaged(t *testing.T) {
	api := &pagedAPI{total: 1200}
	got, err := callOn(t, api, "search_bible", `{"query":"愛","limit":1500}`)
	if err != nil {
		t.Fatal(err)
	}
	var res struct {
		TotalCount int           `json:"total_count"`
		Limit      int           `json:"limit"`
		Results    []interface{} `json:"results"`
	}
	if err := json.Unmarshal([]byte(got), &res); err != nil {
		t.Fatal(err)
	}
	if len(res.Results) != 1200 || res.TotalCount != 1200 || res.Limit != 1500 {
		t.Errorf("results=%d total=%d limit=%d", len(res.Results), res.TotalCount, res.Limit)
	}
	if len(api.reqs) != 3 || api.reqs[1].Offset != 500 || api.reqs[2].Offset != 1000 || api.reqs[2].Limit != 500 {
		t.Errorf("unexpected page requests: %+v", api.reqs)
	}
	if _, err := callOn(t, &pagedAPI{total: 10}, "search_bible", `{"query":"愛","limit":5001}`); err == nil {
		t.Error("limit above 5000 should be refused")
	}
}
