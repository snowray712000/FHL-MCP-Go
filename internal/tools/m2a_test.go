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
	t   *testing.T
	got []fhlapi.QBRequest
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
	want := "{\n  \"id\": 115,\n  \"testament\": \"NT\",\n  \"names\": {\n    \"english\": {\n      \"short\": null,\n      \"full\": null\n    },\n    \"chinese\": {\n      \"short\": null,\n      \"full\": null\n    }\n  }\n}"
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
		{"約3:16", "", 0, "", false}, // D10: the space is required
		{"約 3", "", 0, "", false},
		{"約 3:16-", "", 0, "", false},
		{"", "", 0, "", false},
	}
	for _, c := range cases {
		book, ch, v, ok := bible.ParseCitation(c.in)
		if book != c.book || ch != c.chapter || v != c.verses || ok != c.ok {
			t.Errorf("ParseCitation(%q) = %q %d %q %v", c.in, book, ch, v, ok)
		}
	}
	api := &routeAPI{t: t}
	_, err := callTool(t, api, "query_verse_citation", `{"citation":"約3:16-17"}`)
	var ip *bible.InvalidParameterError
	if !errors.As(err, &ip) || ip.Name != "citation" || ip.Value != "約3:16-17" {
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
