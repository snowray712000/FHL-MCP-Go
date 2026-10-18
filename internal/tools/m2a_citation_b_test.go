package tools

import (
	"context"
	"reflect"
	"strconv"
	"testing"

	"github.com/snowray712000/FHL-MCP-Go/internal/bible"
	"github.com/snowray712000/FHL-MCP-Go/internal/fhlapi"
)

func TestCitationQueriesExpansion(t *testing.T) {
	q := func(book string, ch int, spec string) bible.CitationQuery {
		return bible.CitationQuery{Book: book, Chapter: ch, Spec: spec}
	}
	cases := []struct {
		in   string
		want []bible.CitationQuery
	}{
		{"約 3:16; 羅 8:28", []bible.CitationQuery{q("約", 3, "16"), q("羅", 8, "28")}},
		{"約3:16；羅8:28－30", []bible.CitationQuery{q("約", 3, "16"), q("羅", 8, "28-30")}},
		{"約 3:16;18:1", []bible.CitationQuery{q("約", 3, "16"), q("約", 18, "1")}},     // book inherited
		{"約 3:16,18", []bible.CitationQuery{q("約", 3, "16,18")}},                     // merged: one request
		{"約 3:16,5:1-2", []bible.CitationQuery{q("約", 3, "16"), q("約", 5, "1-2")}},   // chapter chain
		{"約 3:16-4:2", []bible.CitationQuery{q("約", 3, "16-end"), q("約", 4, "1-2")}}, // cross-chapter
		{"約 3:16-6:2", []bible.CitationQuery{q("約", 3, "16-end"), q("約", 4, ""), q("約", 5, ""), q("約", 6, "1-2")}},
		{"約 3:16-4:end", []bible.CitationQuery{q("約", 3, "16-end"), q("約", 4, "1-end")}},
		{"約 3:16-3:18", []bible.CitationQuery{q("約", 3, "16-18")}}, // same chapter
		{"約 3; 羅 8", []bible.CitationQuery{q("約", 3, ""), q("羅", 8, "")}},
		{"約 3,5", []bible.CitationQuery{q("約", 3, ""), q("約", 3, "5")}}, // FHLNUI: first number is the chapter
		{"太 5-7", nil}, // chapter ranges unsupported
		{"約 4:2-3:16", nil},
		{"3:16", nil}, // first segment needs a book
		{"約 3:16;", []bible.CitationQuery{q("約", 3, "16")}},
		{"猶 5; 約 3:16", []bible.CitationQuery{q("猶", 1, "5"), q("約", 3, "16")}},
	}
	for _, c := range cases {
		got, ok := bible.ParseCitationQueries(c.in)
		if c.want == nil {
			if ok {
				t.Errorf("%q: want rejection, got %+v", c.in, got)
			}
			continue
		}
		if !ok || !reflect.DeepEqual(got, c.want) {
			t.Errorf("%q:\n got %+v (%v)\nwant %+v", c.in, got, ok, c.want)
		}
	}
	if _, ok := bible.ParseCitationQueries("約 1:1-60:1"); ok {
		t.Error("a cross-chapter range over the request cap must be rejected")
	}
}

// chapterAPI fakes qb.php: every chapter has 5 verses.
type chapterAPI struct {
	bible.API
	got []fhlapi.QBRequest
}

func (f *chapterAPI) QB(_ context.Context, r fhlapi.QBRequest) (*fhlapi.QBResponse, error) {
	f.got = append(f.got, r)
	resp := &fhlapi.QBResponse{Status: "success", Version: "unv", VName: "和合本", Prev: &fhlapi.Ref{Chap: 1, Sec: 1}}
	for v := 1; v <= 5; v++ {
		resp.Record = append(resp.Record, fhlapi.Record{BID: r.BID, Chap: r.Chapter, Sec: v, BibleText: strconv.Itoa(r.Chapter) + ":" + strconv.Itoa(v)})
	}
	if r.Verse != "" && r.Verse != "1-end" { // explicit specs return only the asked verses (single numbers here)
		resp.Record = nil
		spec, _ := strconv.Atoi(r.Verse)
		resp.Record = append(resp.Record, fhlapi.Record{BID: r.BID, Chap: r.Chapter, Sec: spec, BibleText: strconv.Itoa(r.Chapter) + ":" + r.Verse})
	}
	resp.RecordCount = len(resp.Record)
	return resp, nil
}

func TestCitationMultiQueryResult(t *testing.T) {
	api := &chapterAPI{}
	svc := &bible.Service{API: api}
	res, err := svc.QueryCitation(context.Background(), bible.CitationParams{Citation: "約 3:4-5:2; 羅 8:3"})
	if err != nil {
		t.Fatal(err)
	}
	// 3:4-end is answered from the whole chapter 3, 4 is whole, 5:1-2 and 8:3 are sent as typed.
	wantReqs := []string{"3/", "4/", "5/1-2", "8/3"}
	if len(api.got) != len(wantReqs) {
		t.Fatalf("requests: %+v", api.got)
	}
	for i, r := range api.got {
		if got := strconv.Itoa(r.Chapter) + "/" + r.Verse; got != wantReqs[i] {
			t.Errorf("request %d = %s, want %s", i, got, wantReqs[i])
		}
	}
	// 3:4,5 (2) + 4:1..5 (5) + 5:1-2 (fake returns one record) + 8:3 (1)
	if res.RecordCount != len(res.Verses) || res.Navigation != nil || res.Verses[0].Chapter != 3 || res.Verses[0].Verse != 4 {
		t.Errorf("combined result: count=%d verses=%d nav=%v first=%+v", res.RecordCount, len(res.Verses), res.Navigation, res.Verses[0])
	}
	if last := res.Verses[len(res.Verses)-1]; last.Chapter != 8 || last.Verse != 3 {
		t.Errorf("order: last = %+v", last)
	}
}
