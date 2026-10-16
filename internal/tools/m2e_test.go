package tools

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/snowray712000/FHL-MCP-Go/internal/fhlapi"
)

type scCall struct {
	bid, chap  int
	verse      string // "" = not sent
	simplified bool
}

type stCall struct {
	source     int
	keyword    string
	countOnly  bool
	simplified bool
}

// cmtAPI answers sc / ssc / st.php from recorded upstream fixtures and records the requests.
type cmtAPI struct {
	t      *testing.T
	sc     []scCall
	scList []bool
	ssc    []string
	st     []stCall
}

func (f *cmtAPI) load(name string, dst any) error {
	return json.Unmarshal(fixture(f.t, "upstream", name), dst)
}

func (f *cmtAPI) SC(_ context.Context, bid, chap int, verse *int, simplified bool) (*fhlapi.SCResponse, error) {
	c := scCall{bid: bid, chap: chap, simplified: simplified}
	if verse != nil {
		c.verse = strconv.Itoa(*verse)
	}
	f.sc = append(f.sc, c)
	name := "sc_john_99_1.json"
	switch {
	case chap == 3 && verse == nil:
		name = "sc_john_3_nosec.json"
	case chap == 3 && simplified:
		name = "sc_john_3_16_gb1.json"
	case chap == 3:
		name = "sc_john_3_16.json"
	}
	var out fhlapi.SCResponse
	return &out, f.load(name, &out)
}

func (f *cmtAPI) SCList(_ context.Context, simplified bool) (*fhlapi.SCListResponse, error) {
	f.scList = append(f.scList, simplified)
	var out fhlapi.SCListResponse
	return &out, f.load("sc_list.json", &out)
}

func (f *cmtAPI) SSC(_ context.Context, key string, _ bool) (*fhlapi.SSCResponse, error) {
	f.ssc = append(f.ssc, key)
	var out fhlapi.SSCResponse
	switch key {
	case "尼哥底母":
		return &out, f.load("ssc_nicodemus.json", &out)
	case "zzqqxx":
		return &out, f.load("ssc_none.json", &out)
	case "愛":
		return nil, &fhlapi.Error{Kind: fhlapi.ErrTooManyRecords, Message: `status "Fail:record count > 500"`}
	}
	f.t.Fatalf("unexpected ssc.php key %q", key)
	return nil, nil
}

func (f *cmtAPI) ST(_ context.Context, source int, keyword string, countOnly, simplified bool) (*fhlapi.STResponse, error) {
	f.st = append(f.st, stCall{source, keyword, countOnly, simplified})
	name := ""
	switch {
	case keyword == "愛" && countOnly:
		name = "st_love_count.json"
	case keyword == "愛":
		name = "st_love_all.json"
	case keyword == "faith" && source == 0:
		name = "st_faith_torrey_en.json"
	case keyword == "zzqqxx" && countOnly:
		name = "st_none_count.json"
	case keyword == "zzqqxx":
		name = "st_none.json"
	default:
		f.t.Fatalf("unexpected st.php request %+v", f.st[len(f.st)-1])
	}
	var out fhlapi.STResponse
	return &out, f.load(name, &out)
}

func (f *cmtAPI) QB(context.Context, fhlapi.QBRequest) (*fhlapi.QBResponse, error) {
	return nil, errors.New("QB not expected in this test")
}

func (f *cmtAPI) AB(context.Context, bool) (*fhlapi.ABResponse, error) {
	return nil, errors.New("AB not expected in this test")
}

func (f *cmtAPI) SE(context.Context, fhlapi.SERequest) (*fhlapi.SEResponse, error) {
	return nil, errors.New("SE not expected in this test")
}

func (f *cmtAPI) SD(context.Context, bool, int, bool) (*fhlapi.SDResponse, error) {
	return nil, errors.New("SD not expected in this test")
}

func (f *cmtAPI) QP(context.Context, int, int, int, bool) (*fhlapi.QPResponse, error) {
	return nil, errors.New("QP not expected in this test")
}

// Goldens: list_commentaries, get_commentary_john316 and get_topic_study_love{,_count} are the live
// Python M0 samples; the rest come from running the Python reference's own tools/commentary.py on
// the recorded fixtures (tests/oracle/commentary_oracle.py).
func TestM2eGoldens(t *testing.T) {
	cases := []struct{ tool, args, golden string }{
		{"list_commentaries", `{}`, "list_commentaries.json"},
		{"get_commentary", `{"book":"John","chapter":3,"verse":16}`, "get_commentary_john316.json"},
		{"get_commentary", `{"book":"John","chapter":99,"verse":1}`, "get_commentary_john99_empty.json"},
		{"get_commentary", `{"book":"John","chapter":3}`, "get_commentary_john3_noverse.json"},
		{"get_commentary", `{"book":"John","chapter":3,"verse":null}`, "get_commentary_john3_noverse.json"},
		{"search_commentary", `{"keyword":"尼哥底母"}`, "search_commentary_nicodemus.json"},
		{"search_commentary", `{"keyword":"zzqqxx"}`, "search_commentary_none.json"},
		{"get_topic_study", `{"keyword":"愛"}`, "get_topic_study_love.json"},
		{"get_topic_study", `{"keyword":"愛","count_only":true}`, "get_topic_study_love_count.json"},
		{"get_topic_study", `{"keyword":"faith","source":"torrey_en"}`, "get_topic_study_faith_torrey_en.json"},
		{"get_topic_study", `{"keyword":"zzqqxx"}`, "get_topic_study_none.json"},
		{"get_topic_study", `{"keyword":"zzqqxx","count_only":true}`, "get_topic_study_none_count.json"},
	}
	for _, c := range cases {
		t.Run(c.golden+" "+c.args, func(t *testing.T) {
			got, err := callOn(t, &cmtAPI{t: t}, c.tool, c.args)
			if err != nil {
				t.Fatal(err)
			}
			if want := string(fixture(t, "golden", c.golden)); got != want {
				t.Errorf("differs from the golden:\n got %.300s\nwant %.300s", got, want)
			}
		})
	}
}

func TestCommentaryRequests(t *testing.T) {
	api := &cmtAPI{t: t}
	for _, a := range []string{
		`{"book":"約","chapter":"3","verse":"16"}`,
		`{"book":"43","chapter":3,"verse":16,"use_simplified":true}`,
		`{"book":"John","chapter":3}`,
	} {
		if _, err := callOn(t, api, "get_commentary", a); err != nil {
			t.Fatalf("%s: %v", a, err)
		}
	}
	want := []scCall{{43, 3, "16", false}, {43, 3, "16", true}, {43, 3, "", false}}
	if len(api.sc) != len(want) {
		t.Fatalf("sc calls = %+v", api.sc)
	}
	for i := range want {
		if api.sc[i] != want[i] {
			t.Errorf("sc call %d = %+v, want %+v", i, api.sc[i], want[i])
		}
	}

	api = &cmtAPI{t: t}
	if _, err := callOn(t, api, "list_commentaries", `{"use_simplified":true}`); err != nil {
		t.Fatal(err)
	}
	if len(api.scList) != 1 || !api.scList[0] {
		t.Errorf("sc list simplified flags = %v (D4: use_simplified is honoured)", api.scList)
	}

	api = &cmtAPI{t: t}
	for _, a := range []string{`{"keyword":"faith","source":"torrey_en"}`, `{"keyword":"zzqqxx","count_only":true,"use_simplified":true}`} {
		if _, err := callOn(t, api, "get_topic_study", a); err != nil {
			t.Fatalf("%s: %v", a, err)
		}
	}
	if len(api.st) != 2 || api.st[0] != (stCall{0, "faith", false, false}) || api.st[1] != (stCall{4, "zzqqxx", true, true}) {
		t.Errorf("st calls = %+v", api.st)
	}
}

func TestTopicSourceMapping(t *testing.T) {
	for source, n := range map[string]int{"all": 4, "torrey_en": 0, "naves_en": 1, "torrey_zh": 2, "naves_zh": 3} {
		api := &cmtAPI{t: t}
		if _, err := callOn(t, api, "get_topic_study", `{"keyword":"zzqqxx","source":"`+source+`","count_only":true}`); err != nil {
			t.Fatal(err)
		}
		if api.st[0].source != n {
			t.Errorf("source %s -> N=%d, want %d", source, api.st[0].source, n)
		}
	}
}

func TestCommentaryToolValidation(t *testing.T) {
	cases := []struct{ name, tool, args, want string }{
		{"unknown book", "get_commentary", `{"book":"nope","chapter":1,"verse":1}`, "Invalid parameter 'book': nope - Book not found. Please check the spelling or use a valid book code."},
		{"missing chapter", "get_commentary", `{"book":"John"}`, "Invalid parameter 'chapter': required parameter is missing"},
		{"empty keyword", "search_commentary", `{"keyword":"  "}`, "Invalid parameter 'keyword': must not be empty"},
		{"missing keyword", "search_commentary", `{}`, "Invalid parameter 'keyword': required parameter is missing"},
		{"topic empty keyword", "get_topic_study", `{"keyword":""}`, "Invalid parameter 'keyword': must not be empty"},
		{"topic bad source", "get_topic_study", `{"keyword":"愛","source":"x"}`, "Invalid parameter 'source': x - Must be one of: all, torrey_en, naves_en, torrey_zh, naves_zh"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			api := &cmtAPI{t: t}
			_, err := callOn(t, api, c.tool, c.args)
			if err == nil || err.Error() != c.want {
				t.Fatalf("err = %v\nwant %q", err, c.want)
			}
			if len(api.sc)+len(api.ssc)+len(api.st) != 0 {
				t.Errorf("must not reach upstream: %+v %+v %+v", api.sc, api.ssc, api.st)
			}
		})
	}

	// A verse that is not an integer is rejected by the argument parser before any lookup.
	if _, err := callOn(t, &cmtAPI{t: t}, "get_commentary", `{"book":"John","chapter":3,"verse":"x"}`); err == nil || !strings.HasPrefix(err.Error(), "Invalid parameter 'verse'") {
		t.Errorf("non-numeric verse: err = %v", err)
	}
}

// D12 (approved): ssc.php refusing a too-common keyword is an explicit error, not "0 results".
func TestSearchCommentaryTooManyRecords(t *testing.T) {
	_, err := callOn(t, &cmtAPI{t: t}, "search_commentary", `{"keyword":"愛"}`)
	if err == nil || !strings.HasPrefix(err.Error(), "Invalid parameter 'keyword': 愛 - too many results (more than 500 records)") || !strings.Contains(err.Error(), "結果超過 500 筆，請縮小關鍵字") {
		t.Errorf("err = %v", err)
	}
}

func (f *cmtAPI) AU(context.Context, int, int, int, bool) (*fhlapi.AUResponse, error) {
	return nil, errors.New("AU not expected in this test")
}

func (f *cmtAPI) QSub(context.Context, int, int, *string) (*fhlapi.SubVerseResponse, error) {
	return nil, errors.New("QSub not expected in this test")
}

func (f *cmtAPI) QAF(context.Context, int, int, *string) (*fhlapi.SubVerseResponse, error) {
	return nil, errors.New("QAF not expected in this test")
}

func (f *cmtAPI) SESub(context.Context, string, int, int) (*fhlapi.SubSearchResponse, error) {
	return nil, errors.New("SESub not expected in this test")
}

func (f *cmtAPI) SEAF(context.Context, string, int, int) (*fhlapi.SubSearchResponse, error) {
	return nil, errors.New("SEAF not expected in this test")
}

func (f *cmtAPI) RT(context.Context, int, int, bool) (*fhlapi.RTResponse, error) {
	return nil, errors.New("RT not expected in this test")
}

func (f *cmtAPI) Articles(context.Context, fhlapi.ArticleQuery) (*fhlapi.ArticlesResponse, error) {
	return nil, errors.New("Articles not expected in this test")
}
