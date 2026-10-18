package tools

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/snowray712000/FHL-MCP-Go/internal/bible"
	"github.com/snowray712000/FHL-MCP-Go/internal/fhlapi"
)

func fixture(t *testing.T, kind, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "tests", "fixtures", kind, name))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// routeAPI answers qb.php requests from the recorded upstream fixtures.
type routeAPI struct {
	t     *testing.T
	got   []fhlapi.QBRequest
	abGot []bool // simplified flag of every ab.php request
}

func (f *routeAPI) AB(_ context.Context, simplified bool) (*fhlapi.ABResponse, error) {
	f.abGot = append(f.abGot, simplified)
	var out fhlapi.ABResponse
	return &out, json.Unmarshal(fixture(f.t, "upstream", "ab.json"), &out)
}

func (f *routeAPI) QB(_ context.Context, r fhlapi.QBRequest) (*fhlapi.QBResponse, error) {
	f.got = append(f.got, r)
	name := ""
	switch {
	case r.BID == 57 && r.Chapter == 1 && r.Verse == "":
		name = "qb_philem_1.json"
	case r.BID == 43 && r.Chapter == 3 && r.Verse == "16-17" && !r.Strong:
		name = "qb_john_3_16_17.json"
	default:
		f.t.Fatalf("unexpected upstream request %+v", r)
	}
	var out fhlapi.QBResponse
	return &out, json.Unmarshal(fixture(f.t, "upstream", name), &out)
}

func callTool(t *testing.T, api *routeAPI, name, args string) (string, error) {
	t.Helper()
	r := NewRegistry(&bible.Service{API: api})
	return r.Call(context.Background(), name, json.RawMessage(args))
}

func TestM2aGoldens(t *testing.T) {
	cases := []struct{ tool, args, golden string }{
		{"get_book_list", `{"testament":"NT"}`, "get_book_list_nt.json"},
		{"get_book_info", `{"book":"約"}`, "get_book_info_john.json"},
		{"get_book_info", `{"book":"門"}`, "get_book_info_men.json"},
		{"get_book_info", `{"book":"Philemon"}`, "get_book_info_philemon.json"},
		{"get_bible_chapter", `{"book":"門","chapter":1}`, "get_bible_chapter_philem1.json"},
		{"query_verse_citation", `{"citation":"約 3:16-17"}`, "query_verse_citation_john316_17.json"},
		{"list_bible_versions", `{}`, "list_bible_versions.json"},
		{"search_available_versions", `{"has_strongs":true}`, "search_available_versions_strongs.json"},
	}
	for _, c := range cases {
		t.Run(c.golden, func(t *testing.T) {
			api := &routeAPI{t: t}
			got, err := callTool(t, api, c.tool, c.args)
			if err != nil {
				t.Fatal(err)
			}
			if want := string(fixture(t, "golden", c.golden)); got != want {
				t.Errorf("differs from Python golden:\n got %s\nwant %s", got, want)
			}
		})
	}
}

func TestBookListVariants(t *testing.T) {
	api := &routeAPI{t: t}
	count := func(args string) (int, string) {
		out, err := callTool(t, api, "get_book_list", args)
		if err != nil {
			t.Fatal(err)
		}
		var v struct {
			Total     int    `json:"total_count"`
			Testament string `json:"testament"`
		}
		_ = json.Unmarshal([]byte(out), &v)
		return v.Total, v.Testament
	}
	if n, label := count(`{}`); n != 86 || label != "all" {
		t.Errorf("default: %d %q", n, label)
	}
	all, _ := count(`{}`)
	if n, label := count(`{"testament":"ot"}`); n != 39 || label != "ot" {
		t.Errorf("ot: %d %q (Python filters case-insensitively and echoes the input)", n, label)
	}
	if n, label := count(`{"testament":"NT"}`); n != 27 || label != "NT" {
		t.Errorf("NT: %d %q", n, label)
	}
	if n, label := count(`{"testament":"whatever"}`); n != all || label != "whatever" {
		t.Errorf("unknown testament must list everything: %d %q", n, label)
	}
	if n, label := count(`{"testament":null}`); n != all || label != "all" {
		t.Errorf("null testament (D5): %d %q", n, label)
	}
}

func TestBookInfoEdgeCases(t *testing.T) {
	api := &routeAPI{t: t}
	_, err := callTool(t, api, "get_book_info", `{"book":"Phm"}`)
	var ip *bible.InvalidParameterError
	if !errors.As(err, &ip) || ip.Name != "book" {
		t.Errorf("Phm must be an unknown book (D3 unified message), got %v", err)
	}
	out, err := callTool(t, api, "get_book_info", `{"book":"提多書"}`)
	if err != nil {
		t.Fatal(err)
	}
	var v struct {
		ID int `json:"id"`
	}
	_ = json.Unmarshal([]byte(out), &v)
	if v.ID != 56 {
		t.Errorf("提多書 -> %d, want 56 (D17)", v.ID)
	}
	// Valid code without a table entry: names are null, like Python.
	out, err = callTool(t, api, "get_book_info", `{"book":"115"}`)
	if err != nil {
		t.Fatal(err)
	}
	want := "{\n  \"id\": 115,\n  \"testament\": \"Apocrypha\",\n  \"names\": {\n    \"english\": {\n      \"short\": null,\n      \"full\": null\n    },\n    \"chinese\": {\n      \"short\": null,\n      \"full\": null\n    }\n  }\n}"
	if out != want {
		t.Errorf("code 115:\n got %s\nwant %s", out, want)
	}
}

func TestCitationParsing(t *testing.T) {
	cases := []struct {
		in      string
		book    string
		chapter int
		verses  string
		ok      bool
	}{
		{"約 3:16", "約", 3, "16", true},
		{"  約 3:16-17  ", "約", 3, "16-17", true},
		{"John 3:16,18", "John", 3, "16,18", true},
		{"1 John 4:8", "1 John", 4, "8", true},
		// relaxed (was D10: a space after the book was required)
		{"約3:16", "約", 3, "16", true},
		{"約3:16-17", "約", 3, "16-17", true},
		{"1John 4:8", "1John", 4, "8", true},
		{"John 3：16－18，20", "John", 3, "16-18,20", true},
		{"約３：１６～１８", "約", 3, "16-18", true},
		{"約 3:16-End", "約", 3, "16-end", true},
		{"John 3:16-e.", "John", 3, "16-e", true},
		{"#約 3:16|", "約", 3, "16", true},
		{"路1．23", "路", 1, "23", true},
		{"約 3:18-16", "約", 3, "18-16", true},
		{"約 3", "約", 3, "", true},
		{"約3", "約", 3, "", true},
		{"Gen. 1", "Gen", 1, "", true},
		{"猶", "猶", 1, "", true},
		{"猶大書", "猶大書", 1, "", true},
		{"Jude 5", "Jude", 1, "5", true},
		{"猶 5-7", "猶", 1, "5-7", true},
		{"猶 23-end", "猶", 1, "23-end", true},
		{"猶1:5", "猶", 1, "5", true},
		{"約 3:16-", "", 0, "", false},
		{"約", "", 0, "", false},
		{"31", "", 0, "", false},
		{"約 3-4", "", 0, "", false},
		{"", "", 0, "", false},
	}
	for _, c := range cases {
		book, ch, v, ok := bible.ParseCitation(c.in)
		if book != c.book || ch != c.chapter || v != c.verses || ok != c.ok {
			t.Errorf("ParseCitation(%q) = %q %d %q %v", c.in, book, ch, v, ok)
		}
	}
	api := &routeAPI{t: t}
	_, err := callTool(t, api, "query_verse_citation", `{"citation":"約 3:16-"}`)
	var ip *bible.InvalidParameterError
	if !errors.As(err, &ip) || ip.Name != "citation" || ip.Value != "約 3:16-" {
		t.Errorf("want InvalidParameterError on citation, got %v", err)
	}
	if len(api.got) != 0 {
		t.Errorf("a bad citation must not reach upstream: %+v", api.got)
	}
}

func TestChapterSendsSimplifiedAndNoStrong(t *testing.T) {
	api := &routeAPI{t: t}
	if _, err := callTool(t, api, "get_bible_chapter", `{"book":"Philem","chapter":"1","use_simplified":true}`); err != nil {
		t.Fatal(err)
	}
	if len(api.got) != 1 || !api.got[0].Simplified || api.got[0].Strong || api.got[0].Verse != "" || api.got[0].Version != "unv" {
		t.Errorf("request = %+v", api.got)
	}
}

func TestSearchAvailableVersionsFilters(t *testing.T) {
	api := &routeAPI{t: t}
	type result struct {
		Total   int `json:"total_count"`
		Filters struct {
			HasStrongs *bool   `json:"has_strongs"`
			Testament  *string `json:"testament"`
		} `json:"filters"`
	}
	run := func(args string) result {
		out, err := callTool(t, api, "search_available_versions", args)
		if err != nil {
			t.Fatal(err)
		}
		var v result
		if err := json.Unmarshal([]byte(out), &v); err != nil {
			t.Fatal(err)
		}
		return v
	}
	// Fixture: 86 versions = 52 both + 32 nt_only + 2 ot_only; 3 have Strong's (all "both").
	cases := []struct {
		name, args string
		total      int
	}{
		{"no filter", `{}`, 86},
		{"null filters (D5)", `{"testament":null,"has_strongs":null}`, 86},
		{"both is not a filter (Python)", `{"testament":"both"}`, 86},
		{"empty testament is not a filter", `{"testament":""}`, 86},
		{"OT lowercase", `{"testament":"ot"}`, 54},
		{"NT", `{"testament":"NT"}`, 84},
		{"strongs", `{"has_strongs":true}`, 3},
		{"no strongs", `{"has_strongs":false}`, 83},
		{"combined", `{"has_strongs":false,"testament":"NT"}`, 81},
	}
	for _, c := range cases {
		if got := run(c.args).Total; got != c.total {
			t.Errorf("%s: total_count = %d, want %d", c.name, got, c.total)
		}
	}
	f := run(`{"testament":"ot","has_strongs":false}`).Filters
	if f.Testament == nil || *f.Testament != "ot" || f.HasStrongs == nil || *f.HasStrongs {
		t.Errorf("filters must be echoed as given: %+v", f)
	}
	f = run(`{}`).Filters
	if f.Testament != nil || f.HasStrongs != nil {
		t.Errorf("absent filters must echo null: %+v", f)
	}
}

func TestVersionListSendsSimplifiedFlag(t *testing.T) {
	api := &routeAPI{t: t}
	for _, a := range []string{`{}`, `{"use_simplified":true}`} {
		if _, err := callTool(t, api, "list_bible_versions", a); err != nil {
			t.Fatal(err)
		}
	}
	if len(api.abGot) != 2 || api.abGot[0] || !api.abGot[1] {
		t.Errorf("ab.php simplified flags = %v, want [false true] (D4)", api.abGot)
	}
}

func (f *routeAPI) SE(context.Context, fhlapi.SERequest) (*fhlapi.SEResponse, error) {
	return nil, errors.New("SE not expected in this test")
}

func (f *routeAPI) SD(context.Context, bool, int, bool) (*fhlapi.SDResponse, error) {
	return nil, errors.New("SD not expected in this test")
}

func (f *routeAPI) QP(context.Context, int, int, int, bool) (*fhlapi.QPResponse, error) {
	return nil, errors.New("QP not expected in this test")
}

func (f *routeAPI) SC(context.Context, int, int, *int, bool) (*fhlapi.SCResponse, error) {
	return nil, errors.New("SC not expected in this test")
}

func (f *routeAPI) SCList(context.Context, bool) (*fhlapi.SCListResponse, error) {
	return nil, errors.New("SCList not expected in this test")
}

func (f *routeAPI) SSC(context.Context, string, bool) (*fhlapi.SSCResponse, error) {
	return nil, errors.New("SSC not expected in this test")
}

func (f *routeAPI) ST(context.Context, int, string, bool, bool) (*fhlapi.STResponse, error) {
	return nil, errors.New("ST not expected in this test")
}

func (f *routeAPI) AU(context.Context, int, int, int, bool) (*fhlapi.AUResponse, error) {
	return nil, errors.New("AU not expected in this test")
}

func (f *routeAPI) QSub(context.Context, int, int, *string) (*fhlapi.SubVerseResponse, error) {
	return nil, errors.New("QSub not expected in this test")
}

func (f *routeAPI) QAF(context.Context, int, int, *string) (*fhlapi.SubVerseResponse, error) {
	return nil, errors.New("QAF not expected in this test")
}

func (f *routeAPI) SESub(context.Context, string, int, int) (*fhlapi.SubSearchResponse, error) {
	return nil, errors.New("SESub not expected in this test")
}

func (f *routeAPI) SEAF(context.Context, string, int, int) (*fhlapi.SubSearchResponse, error) {
	return nil, errors.New("SEAF not expected in this test")
}

func (f *routeAPI) RT(context.Context, int, int, bool) (*fhlapi.RTResponse, error) {
	return nil, errors.New("RT not expected in this test")
}

func (f *routeAPI) Articles(context.Context, fhlapi.ArticleQuery) (*fhlapi.ArticlesResponse, error) {
	return nil, errors.New("Articles not expected in this test")
}

func TestCitationRelaxedQueries(t *testing.T) {
	// "end" fetches the whole chapter and keeps the named verses; navigation is dropped.
	api := &routeAPI{t: t}
	out, err := callTool(t, api, "query_verse_citation", `{"citation":"門 23-end"}`)
	if err != nil {
		t.Fatal(err)
	}
	if len(api.got) != 1 || api.got[0].BID != 57 || api.got[0].Chapter != 1 || api.got[0].Verse != "" {
		t.Errorf("end must query the whole chapter: %+v", api.got)
	}
	var res struct {
		RecordCount int `json:"record_count"`
		Verses      []struct {
			Verse int `json:"verse"`
		} `json:"verses"`
		Navigation any `json:"navigation"`
	}
	if err := json.Unmarshal([]byte(out), &res); err != nil {
		t.Fatal(err)
	}
	if res.RecordCount != len(res.Verses) || len(res.Verses) == 0 || res.Verses[0].Verse != 23 || res.Navigation != nil {
		t.Errorf("end result: %+v", res)
	}

	// One-chapter book without a chapter, and a whole-chapter citation.
	for _, c := range []string{`{"citation":"門"}`, `{"citation":"腓利門書"}`} {
		api := &routeAPI{t: t}
		if _, err := callTool(t, api, "query_verse_citation", c); err != nil {
			t.Fatalf("%s: %v", c, err)
		}
		if len(api.got) != 1 || api.got[0].BID != 57 || api.got[0].Chapter != 1 || api.got[0].Verse != "" {
			t.Errorf("%s: %+v", c, api.got)
		}
	}
	// Fixture only has 3:16-17: relaxed spelling must reach upstream as the same normalized request.
	api = &routeAPI{t: t}
	if _, err := callTool(t, api, "query_verse_citation", `{"citation":"約３：１６－１７"}`); err != nil {
		t.Fatal(err)
	}
	if len(api.got) != 1 || api.got[0].BID != 43 || api.got[0].Chapter != 3 || api.got[0].Verse != "16-17" {
		t.Errorf("full-width citation: %+v", api.got)
	}
}
