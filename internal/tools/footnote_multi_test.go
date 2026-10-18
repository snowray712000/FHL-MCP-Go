package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/snowray712000/FHL-MCP-Go/internal/fhlapi"
)

// multiFootnoteAPI adds the optional RTMulti call to footnoteAPI and records the queries.
type multiFootnoteAPI struct {
	footnoteAPI
	calls []fhlapi.RTQuery
	file  string
}

func (f *multiFootnoteAPI) RTMulti(_ context.Context, q fhlapi.RTQuery) (*fhlapi.RTResponse, error) {
	f.calls = append(f.calls, q)
	var out fhlapi.RTResponse
	return &out, json.Unmarshal(fixture(f.t, "upstream", f.file), &out)
}

func TestFootnoteOtherVersionList(t *testing.T) {
	api := &multiFootnoteAPI{footnoteAPI: footnoteAPI{t: t}, file: "rtnui_hab_2_lcc.json"}
	got, err := callOn(t, api, "get_bible_footnote", `{"book_id":35,"version":"LCC","chapter":2}`)
	if err != nil {
		t.Fatal(err)
	}
	if len(api.calls) != 1 || api.calls[0] != (fhlapi.RTQuery{Engs: "Hab", Version: "lcc", Chap: 2, ID: -1}) {
		t.Fatalf("query = %+v", api.calls)
	}
	var res struct {
		QueryType, Version, Scope string
		Chapter, RecordCount      int
		Footnotes                 []struct {
			FootnoteID int `json:"footnote_id"`
			Text       string
		}
	}
	// the tool output uses snake_case keys
	got = strings.NewReplacer("query_type", "querytype", "record_count", "recordcount").Replace(got)
	if err := json.Unmarshal([]byte(got), &res); err != nil {
		t.Fatal(err)
	}
	if res.Version != "lcc" || res.Scope != "chapter" || res.Chapter != 2 || res.RecordCount != 11 || len(res.Footnotes) != 11 || res.Footnotes[0].FootnoteID != 1 || res.Footnotes[0].Text == "" {
		t.Errorf("result = %+v", res)
	}
}

func TestFootnoteOtherVersionSingleAndBookScope(t *testing.T) {
	api := &multiFootnoteAPI{footnoteAPI: footnoteAPI{t: t}, file: "rt_hab_2_lcc_1.json"}
	got, err := callOn(t, api, "get_bible_footnote", `{"book_id":35,"footnote_id":1,"version":"lcc","chapter":2,"use_simplified":true}`)
	if err != nil {
		t.Fatal(err)
	}
	if api.calls[0] != (fhlapi.RTQuery{Engs: "Hab", Version: "lcc", Chap: 2, ID: 1, Simplified: true}) {
		t.Errorf("query = %+v", api.calls[0])
	}
	if !strings.Contains(got, `"chapter": 2`) || !strings.Contains(got, `"footnote_id": 1`) || strings.Contains(got, "not_found") {
		t.Errorf("single result = %s", got)
	}
	// cnet numbers footnotes per book: the chapter is not sent, the scope is the book
	api = &multiFootnoteAPI{footnoteAPI: footnoteAPI{t: t}, file: "rtnui_philem_cnet.json"}
	got, err = callOn(t, api, "get_bible_footnote", `{"book_id":57,"version":"cnet","chapter":1}`)
	if err != nil {
		t.Fatal(err)
	}
	if api.calls[0] != (fhlapi.RTQuery{Engs: "Philem", Version: "cnet", ID: -1}) {
		t.Errorf("query = %+v", api.calls[0])
	}
	if !strings.Contains(got, `"scope": "book"`) || !strings.Contains(got, `"record_count": 20`) || strings.Contains(got, `"chapter"`) {
		t.Errorf("book result = %s", got)
	}
}

func TestFootnoteClassicStaysOnOldPath(t *testing.T) {
	api := &multiFootnoteAPI{footnoteAPI: footnoteAPI{t: t}, file: "rtnui_fail.json"}
	for _, a := range []string{`{"book_id":1,"footnote_id":1}`, `{"book_id":1,"footnote_id":1,"version":"TCV","chapter":3}`} {
		if _, err := callOn(t, api, "get_bible_footnote", a); err != nil {
			t.Fatal(err)
		}
	}
	if len(api.calls) != 0 || len(api.rt) != 2 {
		t.Errorf("calls = %+v, rt = %+v", api.calls, api.rt)
	}
	// TCV without footnote_id lists the book
	api.file = "rtnui_philem_cnet.json"
	if _, err := callOn(t, api, "get_bible_footnote", `{"book_id":57}`); err != nil {
		t.Fatal(err)
	}
	if len(api.calls) != 1 || api.calls[0].Version != "tcv" || api.calls[0].ID != -1 {
		t.Errorf("calls = %+v", api.calls)
	}
}

func TestFootnoteUnsupportedClient(t *testing.T) {
	_, err := callOn(t, &footnoteAPI{t: t}, "get_bible_footnote", `{"book_id":57,"version":"cnet"}`)
	if err == nil || !strings.Contains(err.Error(), "not supported") {
		t.Fatalf("err = %v", err)
	}
}
