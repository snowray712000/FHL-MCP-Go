package tools

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/snowray712000/FHL-MCP-Go/internal/fhlapi"
)

type rtCall struct {
	bid, id    int
	simplified bool
}

// footnoteAPI answers rt.php from recorded upstream fixtures and records the requests.
type footnoteAPI struct {
	t  *testing.T
	rt []rtCall
}

func (f *footnoteAPI) RT(_ context.Context, bid, id int, simplified bool) (*fhlapi.RTResponse, error) {
	f.rt = append(f.rt, rtCall{bid, id, simplified})
	name := "rt_999_999.json"
	switch {
	case bid == 1 && id == 1 && simplified:
		name = "rt_gen_1_1_gb1.json"
	case bid == 1 && id == 1:
		name = "rt_gen_1_1.json"
	}
	var out fhlapi.RTResponse
	return &out, json.Unmarshal(fixture(f.t, "upstream", name), &out)
}

func (f *footnoteAPI) QB(context.Context, fhlapi.QBRequest) (*fhlapi.QBResponse, error) {
	return nil, errors.New("QB not expected in this test")
}

func (f *footnoteAPI) AB(context.Context, bool) (*fhlapi.ABResponse, error) {
	return nil, errors.New("AB not expected in this test")
}

func (f *footnoteAPI) SE(context.Context, fhlapi.SERequest) (*fhlapi.SEResponse, error) {
	return nil, errors.New("SE not expected in this test")
}

func (f *footnoteAPI) SD(context.Context, bool, int, bool) (*fhlapi.SDResponse, error) {
	return nil, errors.New("SD not expected in this test")
}

func (f *footnoteAPI) QP(context.Context, int, int, int, bool) (*fhlapi.QPResponse, error) {
	return nil, errors.New("QP not expected in this test")
}

func (f *footnoteAPI) SC(context.Context, int, int, *int, bool) (*fhlapi.SCResponse, error) {
	return nil, errors.New("SC not expected in this test")
}

func (f *footnoteAPI) SCList(context.Context, bool) (*fhlapi.SCListResponse, error) {
	return nil, errors.New("SCList not expected in this test")
}

func (f *footnoteAPI) SSC(context.Context, string, bool) (*fhlapi.SSCResponse, error) {
	return nil, errors.New("SSC not expected in this test")
}

func (f *footnoteAPI) ST(context.Context, int, string, bool, bool) (*fhlapi.STResponse, error) {
	return nil, errors.New("ST not expected in this test")
}

func (f *footnoteAPI) AU(context.Context, int, int, int, bool) (*fhlapi.AUResponse, error) {
	return nil, errors.New("AU not expected in this test")
}

func (f *footnoteAPI) QSub(context.Context, int, int, *string) (*fhlapi.SubVerseResponse, error) {
	return nil, errors.New("QSub not expected in this test")
}

func (f *footnoteAPI) QAF(context.Context, int, int, *string) (*fhlapi.SubVerseResponse, error) {
	return nil, errors.New("QAF not expected in this test")
}

func (f *footnoteAPI) SESub(context.Context, string, int, int) (*fhlapi.SubSearchResponse, error) {
	return nil, errors.New("SESub not expected in this test")
}

func (f *footnoteAPI) SEAF(context.Context, string, int, int) (*fhlapi.SubSearchResponse, error) {
	return nil, errors.New("SEAF not expected in this test")
}

// All three goldens are live Python M0 samples (the 'not found' one is a successful answer with status "not_found").
func TestM2iGoldens(t *testing.T) {
	cases := []struct{ args, golden string }{
		{`{"book_id":1,"footnote_id":1}`, "get_bible_footnote_gen1_1.json"},
		{`{"book_id":"1","footnote_id":"1","use_simplified":false}`, "get_bible_footnote_gen1_1.json"},
		{`{"book_id":1,"footnote_id":1,"use_simplified":true}`, "get_bible_footnote_gen1_1_simplified.json"},
		{`{"book_id":999,"footnote_id":999}`, "get_bible_footnote_not_found.json"},
	}
	for _, c := range cases {
		t.Run(c.golden+" "+c.args, func(t *testing.T) {
			got, err := callOn(t, &footnoteAPI{t: t}, "get_bible_footnote", c.args)
			if err != nil {
				t.Fatal(err)
			}
			if want := string(fixture(t, "golden", c.golden)); got != want {
				t.Errorf("differs from the golden:\n got %s\nwant %s", got, want)
			}
		})
	}
}

func TestFootnoteRequests(t *testing.T) {
	api := &footnoteAPI{t: t}
	for _, a := range []string{`{"book_id":1,"footnote_id":1}`, `{"book_id":1,"footnote_id":1,"use_simplified":true}`} {
		if _, err := callOn(t, api, "get_bible_footnote", a); err != nil {
			t.Fatal(err)
		}
	}
	if len(api.rt) != 2 || api.rt[0] != (rtCall{1, 1, false}) || api.rt[1] != (rtCall{1, 1, true}) {
		t.Errorf("rt calls = %+v", api.rt)
	}
}

func TestFootnoteValidation(t *testing.T) {
	cases := []struct{ name, args, want string }{
		{"missing book_id", `{"footnote_id":1}`, "Invalid parameter 'book_id': required parameter is missing"},
		{"missing footnote_id", `{"book_id":1}`, "Invalid parameter 'footnote_id': required parameter is missing"},
		{"not a number", `{"book_id":"John","footnote_id":1}`, "Invalid parameter 'book_id': \"John\" - must be an integer"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			api := &footnoteAPI{t: t}
			_, err := callOn(t, api, "get_bible_footnote", c.args)
			if err == nil || err.Error() != c.want {
				t.Fatalf("err = %v\nwant %q", err, c.want)
			}
			if len(api.rt) != 0 {
				t.Errorf("must not reach upstream: %+v", api.rt)
			}
		})
	}
}
