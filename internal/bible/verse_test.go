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

// D17 (approved): Titus is Titus (56), never Tobit (101), however it is spelled; Tobit stays reachable.
func TestTitusIsNotTobit(t *testing.T) {
	for book, bid := range map[string]int{"Titus": 56, "提多書": 56, "多": 56, "56": 56, "Tob": 101, "多俾亞傳": 101, "101": 101} {
		api := &fakeAPI{file: "qb_john_3_16.json"}
		if _, err := (&Service{API: api}).GetVerse(context.Background(), VerseParams{Book: book, Chapter: 1, Version: "unv"}); err != nil {
			t.Fatalf("%s: %v", book, err)
		}
		if api.got.BID != bid {
			t.Errorf("%s -> bid %d, want %d", book, api.got.BID, bid)
		}
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

func (f *fakeAPI) AU(context.Context, int, int, int, bool) (*fhlapi.AUResponse, error) {
	return nil, errors.New("AU not expected in this test")
}

func (f *fakeAPI) QSub(context.Context, int, int, *string) (*fhlapi.SubVerseResponse, error) {
	return nil, errors.New("QSub not expected in this test")
}

func (f *fakeAPI) QAF(context.Context, int, int, *string) (*fhlapi.SubVerseResponse, error) {
	return nil, errors.New("QAF not expected in this test")
}

func (f *fakeAPI) SESub(context.Context, string, int, int) (*fhlapi.SubSearchResponse, error) {
	return nil, errors.New("SESub not expected in this test")
}

func (f *fakeAPI) SEAF(context.Context, string, int, int) (*fhlapi.SubSearchResponse, error) {
	return nil, errors.New("SEAF not expected in this test")
}

func (f *fakeAPI) RT(context.Context, int, int, bool) (*fhlapi.RTResponse, error) {
	return nil, errors.New("RT not expected in this test")
}

func (f *fakeAPI) Articles(context.Context, fhlapi.ArticleQuery) (*fhlapi.ArticlesResponse, error) {
	return nil, errors.New("Articles not expected in this test")
}

func TestMergedVersePlaceholder(t *testing.T) {
	svc := &Service{API: &fakeAPI{file: "qb_col_2_20_21_merged.json"}}
	res, err := svc.GetVerse(context.Background(), VerseParams{Book: "西", Chapter: 2, Verse: "20-21"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Verses[0].MergedIntoPrevious || res.Verses[0].Text == mergedPlaceholder {
		t.Fatalf("verse 20 must stay normal: %+v", res.Verses[0])
	}
	if v := res.Verses[1]; !v.MergedIntoPrevious || v.Text != "（併入上節）" || v.Verse != 21 {
		t.Fatalf("verse 21 not marked: %+v", v)
	}
	b, _ := json.Marshal(res.Verses[0])
	if string(b) != `{"book":"歌","book_eng":"Col","chapter":2,"verse":20,"text":"既然與基督一同死"}` {
		t.Fatalf("normal verse JSON changed: %s", b)
	}
}

func TestRTEngs(t *testing.T) {
	want := map[int]string{1: "Gen", 9: "1 Sam", 11: "1 Kin", 14: "2 Chr", 17: "Esth", 19: "Ps", 28: "Hos", 34: "Nah", 59: "James", 62: "1 John", 66: "Rev"}
	for id, w := range want {
		if got, ok := rtEngs(id); !ok || got != w {
			t.Errorf("rtEngs(%d) = %q, %v; want %q", id, got, ok, w)
		}
	}
	for _, id := range []int{0, 67, 101, -1} {
		if _, ok := rtEngs(id); ok {
			t.Errorf("rtEngs(%d) should fail", id)
		}
	}
}
