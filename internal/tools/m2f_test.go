package tools

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/snowray712000/FHL-MCP-Go/internal/fhlapi"
)

type auCall struct {
	version, bid, chap int
	simplified         bool
}

// audioAPI answers au.php (and the qb.php chapter text) from recorded upstream fixtures.
type audioAPI struct {
	t  *testing.T
	au []auCall
	qb []fhlapi.QBRequest
}

var auFixtures = map[[3]int]string{
	{0, 43, 3}:  "au_john_3.json",
	{0, 57, 1}:  "au_phm_1.json",
	{0, 1, 1}:   "au_gen_1.json",
	{0, 66, 22}: "au_rev_22.json",
	{1, 43, 3}:  "au_john_3_taiwanese.json",
	{7, 43, 3}:  "au_john_3_hebrew.json",
	{0, 43, 99}: "au_john_99.json",
	{0, 101, 1}: "au_tob_1.json",
}

func (f *audioAPI) AU(_ context.Context, version, bid, chap int, simplified bool) (*fhlapi.AUResponse, error) {
	f.au = append(f.au, auCall{version, bid, chap, simplified})
	name, ok := auFixtures[[3]int{version, bid, chap}]
	if !ok {
		f.t.Fatalf("unexpected au.php request version=%d bid=%d chap=%d", version, bid, chap)
	}
	if simplified && version == 0 && bid == 43 && chap == 3 {
		name = "au_john_3_gb1.json"
	}
	var out fhlapi.AUResponse
	return &out, json.Unmarshal(fixture(f.t, "upstream", name), &out)
}

func (f *audioAPI) QB(_ context.Context, r fhlapi.QBRequest) (*fhlapi.QBResponse, error) {
	f.qb = append(f.qb, r)
	if r.BID != 57 || r.Chapter != 1 {
		f.t.Fatalf("unexpected qb.php request %+v", r)
	}
	var out fhlapi.QBResponse
	return &out, json.Unmarshal(fixture(f.t, "upstream", "qb_philem_1.json"), &out)
}

func (f *audioAPI) AB(context.Context, bool) (*fhlapi.ABResponse, error) {
	return nil, errors.New("AB not expected in this test")
}

func (f *audioAPI) SE(context.Context, fhlapi.SERequest) (*fhlapi.SEResponse, error) {
	return nil, errors.New("SE not expected in this test")
}

func (f *audioAPI) SD(context.Context, bool, int, bool) (*fhlapi.SDResponse, error) {
	return nil, errors.New("SD not expected in this test")
}

func (f *audioAPI) QP(context.Context, int, int, int, bool) (*fhlapi.QPResponse, error) {
	return nil, errors.New("QP not expected in this test")
}

func (f *audioAPI) SC(context.Context, int, int, *int, bool) (*fhlapi.SCResponse, error) {
	return nil, errors.New("SC not expected in this test")
}

func (f *audioAPI) SCList(context.Context, bool) (*fhlapi.SCListResponse, error) {
	return nil, errors.New("SCList not expected in this test")
}

func (f *audioAPI) SSC(context.Context, string, bool) (*fhlapi.SSCResponse, error) {
	return nil, errors.New("SSC not expected in this test")
}

func (f *audioAPI) ST(context.Context, int, string, bool, bool) (*fhlapi.STResponse, error) {
	return nil, errors.New("ST not expected in this test")
}

// Goldens: list_audio_versions, get_audio_bible_john3 and get_audio_chapter_with_text_phm1 are the
// live Python M0 samples; the other get_audio_bible outputs come from running the Python reference's
// own tools/audio.py on the recorded fixtures (tests/oracle/audio_oracle.py).
func TestM2fGoldens(t *testing.T) {
	cases := []struct{ tool, args, golden string }{
		{"list_audio_versions", `{}`, "list_audio_versions.json"},
		{"get_audio_bible", `{"book":"John","chapter":3}`, "get_audio_bible_john3.json"},
		{"get_audio_bible", `{"book":"約翰福音","chapter":"3","audio_version":"unv"}`, "get_audio_bible_john3.json"},
		{"get_audio_bible", `{"book":"Gen","chapter":1}`, "get_audio_bible_gen1.json"},
		{"get_audio_bible", `{"book":"Rev","chapter":22}`, "get_audio_bible_rev22.json"},
		{"get_audio_bible", `{"book":"John","chapter":3,"audio_version":"taiwanese"}`, "get_audio_bible_john3_taiwanese.json"},
		{"get_audio_bible", `{"book":"約","chapter":3,"audio_version":"hebrew"}`, "get_audio_bible_john3_hebrew.json"},
		{"get_audio_bible", `{"book":"John","chapter":99}`, "get_audio_bible_john99.json"},
		{"get_audio_bible", `{"book":"101","chapter":1}`, "get_audio_bible_tob1.json"},
		{"get_audio_chapter_with_text", `{"book":"門","chapter":1}`, "get_audio_chapter_with_text_phm1.json"},
	}
	for _, c := range cases {
		t.Run(c.golden+" "+c.args, func(t *testing.T) {
			got, err := callOn(t, &audioAPI{t: t}, c.tool, c.args)
			if err != nil {
				t.Fatal(err)
			}
			if want := string(fixture(t, "golden", c.golden)); got != want {
				t.Errorf("differs from the golden:\n got %.400s\nwant %.400s", got, want)
			}
		})
	}
}

func TestAudioRequests(t *testing.T) {
	api := &audioAPI{t: t}
	for _, a := range []string{
		`{"book":"John","chapter":3,"use_simplified":true}`, // get_audio_bible has no such parameter: ignored
		`{"book":"John","chapter":3,"audio_version":"hebrew"}`,
	} {
		if _, err := callOn(t, api, "get_audio_bible", a); err != nil {
			t.Fatalf("%s: %v", a, err)
		}
	}
	if len(api.au) != 2 || api.au[0] != (auCall{0, 43, 3, false}) || api.au[1] != (auCall{7, 43, 3, false}) {
		t.Errorf("au calls = %+v (get_audio_bible is always a traditional request)", api.au)
	}

	api = &audioAPI{t: t}
	if _, err := callOn(t, api, "get_audio_chapter_with_text", `{"book":"Philemon","chapter":1,"audio_version":"unv","text_version":"kjv","use_simplified":true}`); err != nil {
		t.Fatal(err)
	}
	if len(api.au) != 1 || api.au[0] != (auCall{0, 57, 1, true}) {
		t.Errorf("au calls = %+v, want one simplified request (D4)", api.au)
	}
	if len(api.qb) != 1 || api.qb[0] != (fhlapi.QBRequest{BID: 57, Chapter: 1, Version: "kjv", Simplified: true}) {
		t.Errorf("qb calls = %+v", api.qb)
	}
}

func TestAudioValidation(t *testing.T) {
	versions := "unv, taiwanese, hakka, cantonese, tcv, taiwanese_nt, red_bible, hebrew, fuzhou, greek, spring_taiwanese, spring_unv, netbible_chinese, pct, tsou, taiwanese_south, modern_taiwanese, modern_hakka, tao"
	cases := []struct{ name, tool, args, want string }{
		{"unknown book", "get_audio_bible", `{"book":"nope","chapter":1}`, "Invalid parameter 'book': nope - Book not found. Please check the spelling or use a valid book code."},
		{"unknown book with text", "get_audio_chapter_with_text", `{"book":"Phm","chapter":1}`, "Invalid parameter 'book': Phm - Book not found. Please check the spelling or use a valid book code."},
		{"bad audio version", "get_audio_bible", `{"book":"John","chapter":3,"audio_version":"x"}`, "Invalid parameter 'audio_version': x - Must be one of: " + versions},
		{"version is checked before the book", "get_audio_bible", `{"book":"nope","chapter":3,"audio_version":"x"}`, "Invalid parameter 'audio_version': x - Must be one of: " + versions},
		{"missing chapter", "get_audio_bible", `{"book":"John"}`, "Invalid parameter 'chapter': required parameter is missing"},
		{"missing book", "get_audio_chapter_with_text", `{"chapter":1}`, "Invalid parameter 'book': required parameter is missing"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			api := &audioAPI{t: t}
			_, err := callOn(t, api, c.tool, c.args)
			if err == nil || err.Error() != c.want {
				t.Fatalf("err = %v\nwant %q", err, c.want)
			}
			if len(api.au)+len(api.qb) != 0 {
				t.Errorf("must not reach upstream: %+v %+v", api.au, api.qb)
			}
		})
	}
}

func TestAudioVersionTable(t *testing.T) {
	out, err := callOn(t, &audioAPI{t: t}, "list_audio_versions", `{"ignored":1}`)
	if err != nil {
		t.Fatal(err)
	}
	var v struct {
		Total    int `json:"total_count"`
		Versions []struct {
			Code string `json:"code"`
			ID   int    `json:"id"`
		} `json:"versions"`
	}
	if err := json.Unmarshal([]byte(out), &v); err != nil || v.Total != 19 || len(v.Versions) != 19 {
		t.Fatalf("%v %+v", err, v)
	}
	last := -1
	for _, x := range v.Versions {
		if x.ID <= last || x.ID == 16 || strings.TrimSpace(x.Code) == "" {
			t.Errorf("versions must be sorted by id without id 16: %+v", v.Versions)
			break
		}
		last = x.ID
	}
}

func (f *audioAPI) QSub(context.Context, int, int, *string) (*fhlapi.SubVerseResponse, error) {
	return nil, errors.New("QSub not expected in this test")
}

func (f *audioAPI) QAF(context.Context, int, int, *string) (*fhlapi.SubVerseResponse, error) {
	return nil, errors.New("QAF not expected in this test")
}

func (f *audioAPI) SESub(context.Context, string, int, int) (*fhlapi.SubSearchResponse, error) {
	return nil, errors.New("SESub not expected in this test")
}

func (f *audioAPI) SEAF(context.Context, string, int, int) (*fhlapi.SubSearchResponse, error) {
	return nil, errors.New("SEAF not expected in this test")
}

func (f *audioAPI) RT(context.Context, int, int, bool) (*fhlapi.RTResponse, error) {
	return nil, errors.New("RT not expected in this test")
}

func (f *audioAPI) Articles(context.Context, fhlapi.ArticleQuery) (*fhlapi.ArticlesResponse, error) {
	return nil, errors.New("Articles not expected in this test")
}
