package bible

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
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
