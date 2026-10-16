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

type fakeAPI struct {
	file string
	err  error
	got  fhlapi.QBRequest
}

func (f *fakeAPI) QB(_ context.Context, r fhlapi.QBRequest) (*fhlapi.QBResponse, error) {
	f.got = r
	if f.err != nil {
		return nil, f.err
	}
	b, err := os.ReadFile(filepath.Join("..", "..", "tests", "fixtures", "upstream", f.file))
	if err != nil {
		return nil, err
	}
	var out fhlapi.QBResponse
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func golden(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "tests", "fixtures", "golden", name))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// The goldens are the exact strings the Python server returned in M0 (doc/m0/contract/tool-samples.json).
func TestGetVerseMatchesPythonGolden(t *testing.T) {
	cases := []struct {
		name, upstream, golden string
		p                      VerseParams
	}{
		{"john316", "qb_john_3_16.json", "get_bible_verse_john316.json", VerseParams{Book: "John", Chapter: 3, Verse: "16"}},
		{"john316-17 strong", "qb_john_3_16_17_strong.json", "get_bible_verse_john316_17_strong.json", VerseParams{Book: "約", Chapter: 3, Verse: "16-17", IncludeStrong: true}},
		{"empty chapter", "qb_john_99_1_empty.json", "get_bible_verse_john99_empty.json", VerseParams{Book: "John", Chapter: 99, Verse: "1"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			svc := &Service{API: &fakeAPI{file: c.upstream}}
			res, err := svc.GetVerse(context.Background(), c.p)
			if err != nil {
				t.Fatal(err)
			}
			got, err := MarshalIndent2(res)
			if err != nil {
				t.Fatal(err)
			}
			if want := golden(t, c.golden); got != want {
				t.Errorf("output differs from Python golden.\n got: %s\nwant: %s", got, want)
			}
		})
	}
}

func TestGetVerseRequestMapping(t *testing.T) {
	api := &fakeAPI{file: "qb_john_3_16.json"}
	svc := &Service{API: api}
	if _, err := svc.GetVerse(context.Background(), VerseParams{Book: "Titus", Chapter: 1, IncludeStrong: true, UseSimplified: true}); err != nil {
		t.Fatal(err)
	}
	want := fhlapi.QBRequest{BID: 56, Chapter: 1, Version: "unv", Strong: true, Simplified: true}
	if api.got != want {
		t.Errorf("request = %+v, want %+v", api.got, want)
	}
}

func TestGetVerseUnknownBook(t *testing.T) {
	svc := &Service{API: &fakeAPI{file: "qb_john_3_16.json"}}
	_, err := svc.GetVerse(context.Background(), VerseParams{Book: "不存在書卷", Chapter: 1, Verse: "1"})
	var ip *InvalidParameterError
	if !errors.As(err, &ip) {
		t.Fatalf("err = %v", err)
	}
	want := "Invalid parameter 'book': 不存在書卷 - Book not found. Please check the spelling or use a valid book code."
	if err.Error() != want {
		t.Errorf("message = %q", err.Error())
	}
}

func TestGetVerseUpstreamErrorPassesThrough(t *testing.T) {
	boom := &fhlapi.Error{Kind: fhlapi.ErrNetwork, Message: "down"}
	svc := &Service{API: &fakeAPI{err: boom}}
	if _, err := svc.GetVerse(context.Background(), VerseParams{Book: "John", Chapter: 3}); !errors.Is(err, fhlapi.ErrNetwork) {
		t.Fatalf("err = %v", err)
	}
}

func (f *fakeAPI) AB(context.Context, bool) (*fhlapi.ABResponse, error) {
	return nil, errors.New("AB not expected in this test")
}

func (f *fakeAPI) SE(context.Context, fhlapi.SERequest) (*fhlapi.SEResponse, error) {
	return nil, errors.New("SE not expected in this test")
}

func (f *fakeAPI) SD(context.Context, bool, int, bool) (*fhlapi.SDResponse, error) {
	return nil, errors.New("SD not expected in this test")
}

func (f *fakeAPI) QP(context.Context, int, int, int, bool) (*fhlapi.QPResponse, error) {
	return nil, errors.New("QP not expected in this test")
}

func (f *fakeAPI) SC(context.Context, int, int, *int, bool) (*fhlapi.SCResponse, error) {
	return nil, errors.New("SC not expected in this test")
}

func (f *fakeAPI) SCList(context.Context, bool) (*fhlapi.SCListResponse, error) {
	return nil, errors.New("SCList not expected in this test")
}

func (f *fakeAPI) SSC(context.Context, string, bool) (*fhlapi.SSCResponse, error) {
	return nil, errors.New("SSC not expected in this test")
}

func (f *fakeAPI) ST(context.Context, int, string, bool, bool) (*fhlapi.STResponse, error) {
	return nil, errors.New("ST not expected in this test")
}
