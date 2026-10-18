package bible

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/snowray712000/FHL-MCP-Go/internal/fhlapi"
)

// uiabvAPI adds the optional uiabv.php list on top of the plain fake (whose AB always fails).
type uiabvAPI struct {
	fakeAPI
	fail bool
}

func (f *uiabvAPI) AB(context.Context, bool) (*fhlapi.ABResponse, error) {
	return &fhlapi.ABResponse{Status: "success", Record: []fhlapi.ABRecord{{Book: "unv", CName: "ab", Strong: 1}}}, nil
}

func (f *uiabvAPI) UIABV(context.Context, bool) (*fhlapi.ABResponse, error) {
	if f.fail {
		return nil, errors.New("down")
	}
	b, err := os.ReadFile(filepath.Join("..", "..", "tests", "fixtures", "upstream", "uiabv.json"))
	if err != nil {
		return nil, err
	}
	var out fhlapi.ABResponse
	return &out, json.Unmarshal(b, &out)
}

func TestListVersionsPrefersUIABVAndFallsBack(t *testing.T) {
	res, err := (&Service{API: &uiabvAPI{}}).ListVersions(context.Background(), false)
	if err != nil {
		t.Fatal(err)
	}
	by := map[string]VersionItem{}
	for _, v := range res.Versions {
		by[v.Code] = v
	}
	if res.TotalCount != 6 || by["esv"].Code == "" || by["lxx"].Testament != "ot_only" || by["gnt6"].Testament != "nt_only" || !by["unv"].CanDownload || by["unv"].Version == "" {
		t.Errorf("uiabv list = %+v", res.Versions)
	}
	res, err = (&Service{API: &uiabvAPI{fail: true}}).ListVersions(context.Background(), false)
	if err != nil || res.TotalCount != 1 || res.Versions[0].Name != "ab" {
		t.Errorf("fallback to ab.php: %v %+v", err, res)
	}
}

// origAPI serves a tiny pre-built original-text file: Gen 1:1-3 and Gen 2:1, ASCII only.
type origAPI struct {
	fakeAPI
	calls int
}

func (f *origAPI) OrigData(_ context.Context, name string) ([]byte, error) {
	f.calls++
	if name == "bhs" {
		return nil, errors.New("down")
	}
	return []byte(`{"col":["book","chap","sec","text"],"data":[[1,1,1,"In<WG1> be<WG2>"],[1,1,2,"two<WG3>"],[1,1,3,"three<WG4>"],[1,2,1,"next<WG5>"]]}`), nil
}

func (f *origAPI) QB(context.Context, fhlapi.QBRequest) (*fhlapi.QBResponse, error) {
	return &fhlapi.QBResponse{Version: "upstream", RecordCount: 0}, nil
}

func TestGetVerseOrigStrongsOffline(t *testing.T) {
	api := &origAPI{}
	svc := &Service{API: api}
	res, err := svc.GetVerse(context.Background(), VerseParams{Book: "Gen", Chapter: 1, Verse: "2-3", Version: "lxx", IncludeStrong: true})
	if err != nil {
		t.Fatal(err)
	}
	if res.Version != "lxx" || res.VersionName != "七十士譯本" || res.RecordCount != 2 || res.Verses[0].Text != "two<WG3>" || res.Verses[1].Verse != 3 || res.Verses[0].Book != "創" || res.Verses[0].BookEng != "Gen" {
		t.Errorf("result = %+v", res)
	}
	if n := res.Navigation; n == nil || n.Prev.Verse != 1 || n.Next == nil || n.Next.Chapter != 2 {
		t.Errorf("navigation = %+v", n)
	}
	whole, _ := svc.GetVerse(context.Background(), VerseParams{Book: "Gen", Chapter: 1, Version: "lxx", IncludeStrong: true})
	if whole.RecordCount != 3 || api.calls != 1 {
		t.Errorf("whole chapter = %d records, file loaded %d times", whole.RecordCount, api.calls)
	}
	// without include_strong, other versions, and a failed download all go to the upstream API as before
	for _, p := range []VerseParams{
		{Book: "Gen", Chapter: 1, Version: "lxx"},
		{Book: "Gen", Chapter: 1, Version: "unv", IncludeStrong: true},
		{Book: "Gen", Chapter: 1, Version: "bhs", IncludeStrong: true},
	} {
		if r, err := svc.GetVerse(context.Background(), p); err != nil || r.Version != "upstream" {
			t.Errorf("%+v: %v %+v", p, err, r)
		}
	}
}

func TestSearchBibleByNumberInOrigVersions(t *testing.T) {
	svc := &Service{API: &origAPI{}} // rows: Gen 1:1 {1,2}, 1:2 {3}, 1:3 {4}, 2:1 {5}; SE is never asked
	ctx := context.Background()
	search := func(q, typ, scope string) *SearchResult {
		t.Helper()
		r, err := svc.SearchBible(ctx, SearchParams{Query: q, SearchType: typ, Scope: scope, Version: "lxx", Limit: 10})
		if err != nil {
			t.Fatalf("%s: %v", q, err)
		}
		return r
	}
	if r := search("3", "greek_number", "all"); r.TotalCount != 1 || r.Results[0].Verse != 2 || r.Results[0].Text != "two<WG3>" {
		t.Errorf("single number = %+v", r)
	}
	if r := search("G3 or G4", "greek_number", "all"); r.TotalCount != 2 || r.Results[1].Verse != 3 {
		t.Errorf("or = %+v", r)
	}
	if r := search("G1-G2", "greek_number", "all"); r.TotalCount != 1 {
		t.Errorf("range = %+v", r)
	}
	if r := search("3", "hebrew_number", "all"); r.TotalCount != 0 {
		t.Errorf("hebrew number in lxx = %+v", r)
	}
	if r := search("3", "greek_number", "nt"); r.TotalCount != 0 {
		t.Errorf("scope nt excludes Genesis = %+v", r)
	}
	one, two := 1, 1
	adv, err := svc.SearchAdvanced(ctx, AdvancedParams{Query: "5", SearchType: "greek_number", RangeStart: &one, RangeEnd: &two, Version: "lxx", Limit: 10})
	if err != nil || adv.(*AdvancedResult).TotalCount != 1 {
		t.Errorf("advanced = %+v %v", adv, err)
	}
}

// hebAPI serves a tiny bhs file; the Hebrew is written with \u escapes (editors normalise the points).
type hebAPI struct{ fakeAPI }

func (f *hebAPI) OrigData(_ context.Context, name string) ([]byte, error) {
	if name != "bhs" {
		return nil, errors.New("only bhs")
	}
	bereshit := `בְּרֵאשִׁית<WH7225> בָּרָא<WH1254> אֱלֹהִים<WH430>׃`
	erets := `וְהָאָרֶץ<WH776>׃`
	exodus := `אֱלֹהֵים<WH430>־הָאָרֶץ`
	return []byte(`{"col":["book","chap","sec","text"],"data":[[1,1,1,"` + bereshit + `"],[1,1,2,"` + erets + `"],[2,1,1,"` + exodus + `"]]}`), nil
}

func TestSearchBibleHebrewKeywordInBHS(t *testing.T) {
	svc := &Service{API: &hebAPI{}}
	ctx := context.Background()
	verses := func(q, scope string) []int {
		t.Helper()
		r, err := svc.SearchBible(ctx, SearchParams{Query: q, SearchType: "keyword", Scope: scope, Version: "bhs", Limit: 10})
		if err != nil {
			t.Fatalf("%q: %v", q, err)
		}
		var out []int
		for _, v := range r.Results {
			out = append(out, v.Chapter*100+v.Verse)
			if strings.Contains(v.Text, "<WH") {
				t.Errorf("numbers must not leak into the text: %q", v.Text)
			}
		}
		if r.TotalCount != len(out) {
			t.Errorf("total %d, results %d", r.TotalCount, len(out))
		}
		return out
	}
	elohim := "אלהים"
	if got := verses(elohim, "all"); len(got) != 2 { // bare consonants: Gen 1:1 and Ex 1:1
		t.Errorf("bare = %v", got)
	}
	if got := verses("אֱלֹהִים", "all"); len(got) != 1 || got[0] != 101 { // with points: hiriq only
		t.Errorf("with points = %v", got)
	}
	if got := verses("אלהימ", "all"); len(got) != 2 { // final mem == mem
		t.Errorf("final form = %v", got)
	}
	if got := verses("ברא "+elohim, "all"); len(got) != 1 { // every word must occur
		t.Errorf("two words = %v", got)
	}
	if got := verses("ראש", "all"); len(got) != 1 { // part of a word
		t.Errorf("substring = %v", got)
	}
	if got := verses("הארץ", "all"); len(got) != 2 { // after a maqaf, and in a word with prefix vav is not whole-word but substring
		t.Errorf("maqaf = %v", got)
	}
	if got := verses("בּרא", "all"); len(got) != 1 { // dagesh typed: must have it (bet in Gen 1:1 has dagesh)
		t.Errorf("dagesh = %v", got)
	}
	if got := verses(elohim, "nt"); len(got) != 0 {
		t.Errorf("nt scope = %v", got)
	}
	// a query without Hebrew letters (FHL's old ASCII code) still goes to se.php
	if _, err := svc.SearchBible(ctx, SearchParams{Query: "abc", SearchType: "keyword", Scope: "all", Version: "bhs", Limit: 10}); err == nil {
		t.Error("ASCII query must reach the upstream API")
	}
}
