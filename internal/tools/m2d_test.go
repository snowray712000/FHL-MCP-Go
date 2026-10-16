package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/snowray712000/FHL-MCP-Go/internal/bible"
	"github.com/snowray712000/FHL-MCP-Go/internal/fhlapi"
)

type qpCall struct {
	bid, chap, sec int
	simplified     bool
}

// wordsAPI answers qp.php (and the qb.php chapter-length lookups) from recorded fixtures.
type wordsAPI struct {
	t     *testing.T
	qp    []qpCall
	qb    []fhlapi.QBRequest
	empty bool // answer every qp.php with an empty record list
}

var qpFixtures = map[[3]int]string{
	{43, 1, 1}: "qp_john_1_1.json",
	{59, 1, 1}: "qp_james_1_1.json",
	{59, 1, 2}: "qp_james_1_2.json",
	{1, 1, 1}:  "qp_gen_1_1.json",
	{8, 1, 22}: "qp_ruth_1_22.json",
	{8, 2, 1}:  "qp_ruth_2_1.json",
}

func (f *wordsAPI) QP(_ context.Context, bid, chap, sec int, simplified bool) (*fhlapi.QPResponse, error) {
	f.qp = append(f.qp, qpCall{bid, chap, sec, simplified})
	if f.empty {
		return &fhlapi.QPResponse{Status: "success"}, nil
	}
	name, ok := qpFixtures[[3]int{bid, chap, sec}]
	if !ok {
		f.t.Fatalf("unexpected qp.php request bid=%d chap=%d sec=%d", bid, chap, sec)
	}
	var out fhlapi.QPResponse
	return &out, json.Unmarshal(fixture(f.t, "upstream", name), &out)
}

func (f *wordsAPI) QB(_ context.Context, r fhlapi.QBRequest) (*fhlapi.QBResponse, error) {
	f.qb = append(f.qb, r)
	if r.BID != 8 || r.Chapter != 1 {
		f.t.Fatalf("unexpected qb.php request %+v", r)
	}
	var out fhlapi.QBResponse
	return &out, json.Unmarshal(fixture(f.t, "upstream", "qb_ruth_1.json"), &out)
}

func (f *wordsAPI) AB(context.Context, bool) (*fhlapi.ABResponse, error) {
	return nil, errors.New("AB not expected in this test")
}

func (f *wordsAPI) SE(context.Context, fhlapi.SERequest) (*fhlapi.SEResponse, error) {
	return nil, errors.New("SE not expected in this test")
}

func (f *wordsAPI) SD(context.Context, bool, int, bool) (*fhlapi.SDResponse, error) {
	return nil, errors.New("SD not expected in this test")
}

func TestM2dGoldens(t *testing.T) {
	cases := []struct{ tool, args, golden string }{
		{"get_word_analysis", `{"book":"John","chapter":1,"verse":1}`, "get_word_analysis_john_1_1.json"},
		{"get_word_analysis", `{"book":"Gen","chapter":1,"verse":1}`, "get_word_analysis_gen_1_1.json"},
		{"get_parsing_reference", `{"book":"James","chapter":1,"verse":"1"}`, "get_parsing_reference_james_1_1.md"},
		{"get_parsing_reference", `{"book":"雅各書","chapter":1,"verse":"1-2"}`, "get_parsing_reference_james_1_1_2.md"},
		{"get_parsing_reference", `{"book":"創","chapter":1,"verse":"1"}`, "get_parsing_reference_gen_1_1.md"},
		{"get_parsing_reference", `{"book":"Ruth","chapter":1,"verse":"22","end_chapter":2,"end_verse":1}`, "get_parsing_reference_ruth_1_22_to_2_1.md"},
	}
	for _, c := range cases {
		t.Run(c.golden, func(t *testing.T) {
			got, err := callOn(t, &wordsAPI{t: t}, c.tool, c.args)
			if err != nil {
				t.Fatal(err)
			}
			if want := string(fixture(t, "golden", c.golden)); got != want {
				t.Errorf("differs from the golden (len %d vs %d)", len(got), len(want))
			}
		})
	}
}

func TestWordAnalysisRequest(t *testing.T) {
	api := &wordsAPI{t: t}
	if _, err := callOn(t, api, "get_word_analysis", `{"book":"約","chapter":"1","verse":1,"use_simplified":true}`); err != nil {
		t.Fatal(err)
	}
	if len(api.qp) != 1 || api.qp[0] != (qpCall{43, 1, 1, true}) {
		t.Errorf("qp calls = %+v, want one bid=43 call with gb=1 (D4)", api.qp)
	}
}

func TestParsingReferenceRequests(t *testing.T) {
	api := &wordsAPI{t: t}
	if _, err := callOn(t, api, "get_parsing_reference", `{"book":"Ruth","chapter":1,"verse":"22","end_chapter":2,"end_verse":1}`); err != nil {
		t.Fatal(err)
	}
	if len(api.qb) != 1 {
		t.Errorf("cross-chapter query must look up the chapter length once: %+v", api.qb)
	}
	want := []qpCall{{8, 1, 22, false}, {8, 2, 1, false}}
	if fmt.Sprint(api.qp) != fmt.Sprint(want) {
		t.Errorf("qp calls = %+v, want %+v", api.qp, want)
	}

	api = &wordsAPI{t: t}
	if _, err := callOn(t, api, "get_parsing_reference", `{"book":"James","chapter":1,"verse":"1","end_verse":2}`); err != nil {
		t.Fatal(err)
	}
	if fmt.Sprint(api.qp) != fmt.Sprint([]qpCall{{59, 1, 1, false}, {59, 1, 2, false}}) || len(api.qb) != 0 {
		t.Errorf("same-chapter end_verse range: qp=%+v qb=%+v", api.qp, api.qb)
	}
}

func TestWordToolValidation(t *testing.T) {
	cases := []struct{ name, tool, args, want string }{
		{"unknown book", "get_word_analysis", `{"book":"nope","chapter":1,"verse":1}`, "Invalid parameter 'book': nope - Book not found. Please check the spelling or use a valid book code."},
		{"apocrypha", "get_word_analysis", `{"book":"Tob","chapter":1,"verse":1}`, "Invalid parameter 'book': Tob - get_word_analysis only supports the 66 canon books (book codes 1-66)."},
		{"apocrypha parsing", "get_parsing_reference", `{"book":"101","chapter":1,"verse":"1"}`, "Invalid parameter 'book': 101 - get_parsing_reference only supports the 66 canon books (book codes 1-66)."},
		{"missing verse", "get_word_analysis", `{"book":"John","chapter":1}`, "Invalid parameter 'verse': required parameter is missing"},
		{"verse not a number", "get_parsing_reference", `{"book":"John","chapter":1,"verse":"x"}`, "Invalid parameter 'verse': x - Cannot parse verse number: 'x'"},
		{"bad range", "get_parsing_reference", `{"book":"John","chapter":1,"verse":"1-y"}`, "Invalid parameter 'verse': 1-y - Cannot parse verse range: '1-y'"},
		{"empty verse", "get_parsing_reference", `{"book":"John","chapter":1,"verse":" , "}`, "Invalid parameter 'verse':  ,  - Cannot parse the verse range"},
		{"reversed chapters", "get_parsing_reference", `{"book":"John","chapter":3,"verse":"1","end_chapter":2,"end_verse":1}`, "Invalid parameter 'end_chapter': 2 - end_chapter must be greater than or equal to chapter"},
		{"cross chapter without end_verse", "get_parsing_reference", `{"book":"John","chapter":1,"verse":"1","end_chapter":2}`, "Invalid parameter 'end_verse': required when end_chapter is in a later chapter"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			api := &wordsAPI{t: t}
			_, err := callOn(t, api, c.tool, c.args)
			if err == nil || err.Error() != c.want {
				t.Fatalf("err = %v\nwant %q", err, c.want)
			}
			if len(api.qp) != 0 || len(api.qb) != 0 {
				t.Errorf("must not reach upstream: qp=%+v qb=%+v", api.qp, api.qb)
			}
		})
	}
}

func TestParsingReferenceWithoutUpstreamData(t *testing.T) {
	_, err := callOn(t, &wordsAPI{t: t, empty: true}, "get_parsing_reference", `{"book":"John","chapter":1,"verse":"1"}`)
	if err == nil || err.Error() != "no original-language data for John 1:1" {
		t.Errorf("err = %v", err)
	}
	out, err := callOn(t, &wordsAPI{t: t, empty: true}, "get_word_analysis", `{"book":"John","chapter":99,"verse":1}`)
	if err != nil {
		t.Fatal(err)
	}
	want := "{\n  \"testament\": \"NT\",\n  \"book\": \"John\",\n  \"chapter\": 99,\n  \"verse\": 1,\n  \"original_text\": \"\",\n  \"translation\": \"\",\n  \"word_count\": 0,\n  \"words\": []\n}"
	if out != want {
		t.Errorf("empty analysis:\n got %s\nwant %s", out, want)
	}
}

func TestSuperfluousInterfaceCheck(t *testing.T) {
	var _ bible.API = (*wordsAPI)(nil)
}

func (f *wordsAPI) SC(context.Context, int, int, *int, bool) (*fhlapi.SCResponse, error) {
	return nil, errors.New("SC not expected in this test")
}

func (f *wordsAPI) SCList(context.Context, bool) (*fhlapi.SCListResponse, error) {
	return nil, errors.New("SCList not expected in this test")
}

func (f *wordsAPI) SSC(context.Context, string, bool) (*fhlapi.SSCResponse, error) {
	return nil, errors.New("SSC not expected in this test")
}

func (f *wordsAPI) ST(context.Context, int, string, bool, bool) (*fhlapi.STResponse, error) {
	return nil, errors.New("ST not expected in this test")
}

// D21 (approved): ranges over 30 verses are refused up front, before any qp.php request.
func TestParsingReferenceCap(t *testing.T) {
	cases := []struct{ name, args string }{
		{"verse range", `{"book":"James","chapter":1,"verse":"1-31"}`},
		{"cross chapter", `{"book":"Ruth","chapter":1,"verse":"1","end_chapter":2,"end_verse":20}`}, // 22 + 20 verses
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			api := &wordsAPI{t: t}
			_, err := callOn(t, api, "get_parsing_reference", c.args)
			if err == nil || !strings.HasPrefix(err.Error(), "Invalid parameter 'verse': ") || !strings.Contains(err.Error(), "at most 30 verses per request") {
				t.Fatalf("err = %v", err)
			}
			if len(api.qp) != 0 {
				t.Errorf("qp.php must not be called: %+v", api.qp)
			}
		})
	}
}

func (f *wordsAPI) AU(context.Context, int, int, int, bool) (*fhlapi.AUResponse, error) {
	return nil, errors.New("AU not expected in this test")
}
