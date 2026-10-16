package tools

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/snowray712000/FHL-MCP-Go/internal/fhlapi"
)

type subVerseCall struct {
	bid, chap int
	verse     string // "" = not sent
}

type subSearchCall struct {
	q             string
	limit, offset int
}

// scriptureAPI answers qsub / qaf / sesub / seaf.php from recorded upstream fixtures and records the requests.
type scriptureAPI struct {
	t       *testing.T
	qsub    []subVerseCall
	qaf     []subVerseCall
	sesub   []subSearchCall
	seaf    []subSearchCall
	failing bool
}

func (f *scriptureAPI) load(name string, dst any) error {
	return json.Unmarshal(fixture(f.t, "upstream", name), dst)
}

func subCall(bid, chap int, verse *string) subVerseCall {
	c := subVerseCall{bid: bid, chap: chap}
	if verse != nil {
		c.verse = *verse
	}
	return c
}

func (f *scriptureAPI) QSub(_ context.Context, bid, chap int, verse *string) (*fhlapi.SubVerseResponse, error) {
	c := subCall(bid, chap, verse)
	f.qsub = append(f.qsub, c)
	name := map[subVerseCall]string{
		{103, 1, "1"}:   "qsub_tob_1_1.json",
		{103, 1, "1-3"}: "qsub_tob_1_1-3.json",
		{101, 1, "1,3"}: "qsub_1mac_1_1_3.json",
		{103, 99, ""}:   "qsub_tob_99.json",
	}[c]
	if name == "" {
		f.t.Fatalf("unexpected qsub.php request %+v", c)
	}
	var out fhlapi.SubVerseResponse
	return &out, f.load(name, &out)
}

func (f *scriptureAPI) QAF(_ context.Context, bid, chap int, verse *string) (*fhlapi.SubVerseResponse, error) {
	c := subCall(bid, chap, verse)
	f.qaf = append(f.qaf, c)
	name := map[subVerseCall]string{
		{201, 1, "1"}:   "qaf_1clem_1_1.json",
		{212, 1, "1-2"}: "qaf_did_1_1-2.json",
		{201, 99, ""}:   "qaf_1clem_99.json",
	}[c]
	if name == "" {
		f.t.Fatalf("unexpected qaf.php request %+v", c)
	}
	var out fhlapi.SubVerseResponse
	return &out, f.load(name, &out)
}

func (f *scriptureAPI) SESub(_ context.Context, q string, limit, offset int) (*fhlapi.SubSearchResponse, error) {
	c := subSearchCall{q, limit, offset}
	f.sesub = append(f.sesub, c)
	name := map[subSearchCall]string{
		{"智慧", 3, 0}:      "sesub_wisdom_limit3.json",
		{"智慧", 2, 2}:      "sesub_wisdom_off2_limit2.json",
		{"zzqqxx", 10, 0}: "sesub_none.json",
	}[c]
	if name == "" {
		f.t.Fatalf("unexpected sesub.php request %+v", c)
	}
	var out fhlapi.SubSearchResponse
	return &out, f.load(name, &out)
}

func (f *scriptureAPI) SEAF(_ context.Context, q string, limit, offset int) (*fhlapi.SubSearchResponse, error) {
	c := subSearchCall{q, limit, offset}
	f.seaf = append(f.seaf, c)
	name := map[subSearchCall]string{
		{"教會", 3, 0}:      "seaf_church_limit3.json",
		{"zzqqxx", 10, 0}: "seaf_none.json",
	}[c]
	if name == "" {
		f.t.Fatalf("unexpected seaf.php request %+v", c)
	}
	var out fhlapi.SubSearchResponse
	return &out, f.load(name, &out)
}

func (f *scriptureAPI) calls() int { return len(f.qsub) + len(f.qaf) + len(f.sesub) + len(f.seaf) }

func (f *scriptureAPI) QB(context.Context, fhlapi.QBRequest) (*fhlapi.QBResponse, error) {
	return nil, errors.New("QB not expected in this test")
}

func (f *scriptureAPI) AB(context.Context, bool) (*fhlapi.ABResponse, error) {
	return nil, errors.New("AB not expected in this test")
}

func (f *scriptureAPI) SE(context.Context, fhlapi.SERequest) (*fhlapi.SEResponse, error) {
	return nil, errors.New("SE not expected in this test")
}

func (f *scriptureAPI) SD(context.Context, bool, int, bool) (*fhlapi.SDResponse, error) {
	return nil, errors.New("SD not expected in this test")
}

func (f *scriptureAPI) QP(context.Context, int, int, int, bool) (*fhlapi.QPResponse, error) {
	return nil, errors.New("QP not expected in this test")
}

func (f *scriptureAPI) SC(context.Context, int, int, *int, bool) (*fhlapi.SCResponse, error) {
	return nil, errors.New("SC not expected in this test")
}

func (f *scriptureAPI) SCList(context.Context, bool) (*fhlapi.SCListResponse, error) {
	return nil, errors.New("SCList not expected in this test")
}

func (f *scriptureAPI) SSC(context.Context, string, bool) (*fhlapi.SSCResponse, error) {
	return nil, errors.New("SSC not expected in this test")
}

func (f *scriptureAPI) ST(context.Context, int, string, bool, bool) (*fhlapi.STResponse, error) {
	return nil, errors.New("ST not expected in this test")
}

func (f *scriptureAPI) AU(context.Context, int, int, int, bool) (*fhlapi.AUResponse, error) {
	return nil, errors.New("AU not expected in this test")
}

// Goldens come from running the Python reference's own http_* functions (tests/oracle/scripture_oracle.py)
// on the recorded fixtures with the book table replaced by the upstream's real numbering (D24). The list
// formatter was first checked against the two live Python samples using the reference's own table.
func TestM2gGoldens(t *testing.T) {
	cases := []struct{ tool, args, golden string }{
		{"get_apocrypha_verse", `{"book":"多","chapter":1,"verse":"1"}`, "get_apocrypha_verse_tob1_1.json"},
		{"get_apocrypha_verse", `{"book":"Tobit","chapter":1,"verse":"1-3"}`, "get_apocrypha_verse_tob1_1-3.json"},
		{"get_apocrypha_verse", `{"book":"加上","chapter":"1","verse":"1,3"}`, "get_apocrypha_verse_1mac1_1_3.json"},
		{"get_apocrypha_verse", `{"book":"多俾亞傳","chapter":99}`, "get_apocrypha_verse_tob99_empty.json"},
		{"get_apocrypha_verse", `{"book":"多俾亞傳","chapter":99,"verse":null}`, "get_apocrypha_verse_tob99_empty.json"},
		{"search_apocrypha", `{"query":"智慧","limit":3}`, "search_apocrypha_wisdom.json"},
		{"search_apocrypha", `{"query":"智慧","limit":2,"offset":2}`, "search_apocrypha_wisdom_off2.json"},
		{"search_apocrypha", `{"query":"zzqqxx"}`, "search_apocrypha_none.json"},
		{"list_apocrypha_books", `{}`, "list_apocrypha_books.json"},
		{"get_apostolic_fathers_verse", `{"book":"革利免前書","chapter":1,"verse":"1"}`, "get_apostolic_fathers_verse_1clem1_1.json"},
		{"get_apostolic_fathers_verse", `{"book":"Didache","chapter":1,"verse":"1-2"}`, "get_apostolic_fathers_verse_did1_1-2.json"},
		{"get_apostolic_fathers_verse", `{"book":"1Clem","chapter":99}`, "get_apostolic_fathers_verse_1clem99_empty.json"},
		{"search_apostolic_fathers", `{"query":"教會","limit":3}`, "search_apostolic_fathers_church.json"},
		{"search_apostolic_fathers", `{"query":"zzqqxx"}`, "search_apostolic_fathers_none.json"},
		{"list_apostolic_fathers_books", `{}`, "list_apostolic_fathers_books.json"},
	}
	for _, c := range cases {
		t.Run(c.golden+" "+c.args, func(t *testing.T) {
			got, err := callOn(t, &scriptureAPI{t: t}, c.tool, c.args)
			if err != nil {
				t.Fatal(err)
			}
			if want := string(fixture(t, "golden", c.golden)); got != want {
				t.Errorf("differs from the golden:\n got %.400s\nwant %.400s", got, want)
			}
		})
	}
}

// D24: the upstream's real book numbers are sent, whichever documented name was used.
func TestScriptureBookNumbers(t *testing.T) {
	apoc := map[string]int{
		"多": 103, "Tob": 103, "tobit": 103, "多俾亞傳": 103, "103": 103, "友": 104, "Jdt": 104,
		"加上": 101, "1Mac": 101, "1 Maccabees": 101, "加下": 102, "2Mac": 102, "智": 106, "智慧篇": 106, "Wis": 106,
		"德": 105, "便西拉智訓": 105, "Sir": 105, "巴": 109, "Baruch": 109, "耶信": 110, "EpJer": 110,
		"拉上": 107, "1Es": 107, "三童歌": 112, "蘇撒拿傳": 113, "Bel": 114, "斯補": 115, "101": 101,
	}
	for name, want := range apoc {
		api := &scriptureAPI{t: t}
		_, _ = callOn(t, &failFast{api}, "get_apocrypha_verse", `{"book":"`+name+`","chapter":1}`)
		if len(api.qsub) != 1 || api.qsub[0].bid != want {
			t.Errorf("apocrypha %q -> qsub %+v, want bid %d", name, api.qsub, want)
		}
	}
	apos := map[string]int{
		"革": 201, "革利免前書": 201, "1Clem": 201, "革二": 202, "2Clem": 202, "訓": 212, "十二使徒遺訓": 212, "Did": 212,
		"巴": 213, "巴拿巴書": 213, "Barn": 213, "黑": 214, "黑馬牧人書": 214, "Herm": 214, "帕": 216, "Pap": 216,
		"坡": 210, "Polycarp": 210, "伊格那丟致羅馬人書": 206, "Ign. Rom.": 206, "205": 205, "217": 217,
	}
	for name, want := range apos {
		api := &scriptureAPI{t: t}
		_, _ = callOn(t, &failFast{api}, "get_apostolic_fathers_verse", `{"book":"`+name+`","chapter":1}`)
		if len(api.qaf) != 1 || api.qaf[0].bid != want {
			t.Errorf("apostolic %q -> qaf %+v, want bid %d", name, api.qaf, want)
		}
	}
}

// failFast answers every qsub / qaf request with an upstream failure after recording it, so a test can look
// at the request without needing a fixture for each book.
type failFast struct{ *scriptureAPI }

func (f *failFast) QSub(_ context.Context, bid, chap int, verse *string) (*fhlapi.SubVerseResponse, error) {
	f.qsub = append(f.qsub, subCall(bid, chap, verse))
	return nil, &fhlapi.Error{Kind: fhlapi.ErrUpstream, Message: "stop"}
}

func (f *failFast) QAF(_ context.Context, bid, chap int, verse *string) (*fhlapi.SubVerseResponse, error) {
	f.qaf = append(f.qaf, subCall(bid, chap, verse))
	return nil, &fhlapi.Error{Kind: fhlapi.ErrUpstream, Message: "stop"}
}

func TestScriptureRequests(t *testing.T) {
	api := &scriptureAPI{t: t}
	for _, a := range []string{
		`{"book":"多","chapter":"1","verse":" 1 "}`, // numeric string chapter (D6); spaces around the verse spec are dropped
		`{"book":"Tobit","chapter":1,"verse":"1-3"}`,
	} {
		if _, err := callOn(t, api, "get_apocrypha_verse", a); err != nil {
			t.Fatalf("%s: %v", a, err)
		}
	}
	if len(api.qsub) != 2 || api.qsub[0] != (subVerseCall{103, 1, "1"}) || api.qsub[1] != (subVerseCall{103, 1, "1-3"}) {
		t.Errorf("qsub calls = %+v", api.qsub)
	}

	api = &scriptureAPI{t: t}
	if _, err := callOn(t, api, "search_apocrypha", `{"query":"智慧","limit":"3","offset":null}`); err != nil {
		t.Fatal(err)
	}
	if len(api.sesub) != 1 || api.sesub[0] != (subSearchCall{"智慧", 3, 0}) {
		t.Errorf("sesub calls = %+v (offset null is absent, D5)", api.sesub)
	}
}

func TestScriptureValidation(t *testing.T) {
	notBook := func(rng, book string) string {
		return "Invalid parameter 'book': " + book + " - Not a book of this collection (" + rng + "). Use the matching list_*_books tool to see the valid names and ids."
	}
	cases := []struct{ name, tool, args, want string }{
		{"canon book", "get_apocrypha_verse", `{"book":"John","chapter":1}`, notBook("101-115", "John")},
		{"ambiguous Daniel additions", "get_apocrypha_verse", `{"book":"但補","chapter":1}`, notBook("101-115", "但補")},
		{"apocrypha id out of range", "get_apocrypha_verse", `{"book":"116","chapter":1}`, notBook("101-115", "116")},
		{"canon id", "get_apocrypha_verse", `{"book":"43","chapter":1}`, notBook("101-115", "43")},
		{"apocrypha book in apostolic tool", "get_apostolic_fathers_verse", `{"book":"Tob","chapter":1}`, notBook("201-217", "Tob")},
		{"ambiguous Ignatius", "get_apostolic_fathers_verse", `{"book":"伊","chapter":1}`, notBook("201-217", "伊")},
		{"empty book", "get_apostolic_fathers_verse", `{"book":"","chapter":1}`, "Invalid parameter 'book': Not a book of this collection (201-217). Use the matching list_*_books tool to see the valid names and ids."},
		{"missing chapter", "get_apocrypha_verse", `{"book":"多"}`, "Invalid parameter 'chapter': required parameter is missing"},
		{"bad verse", "get_apocrypha_verse", `{"book":"多","chapter":1,"verse":"abc"}`, "Invalid parameter 'verse': abc - Use forms such as 1, 1-5, 1,3,5 or 1-2,5,8-10."},
		{"empty verse", "get_apostolic_fathers_verse", `{"book":"革","chapter":1,"verse":""}`, "Invalid parameter 'verse': Use forms such as 1, 1-5, 1,3,5 or 1-2,5,8-10."},
		{"empty query", "search_apocrypha", `{"query":"   "}`, "Invalid parameter 'query': must not be empty"},
		{"missing query", "search_apostolic_fathers", `{}`, "Invalid parameter 'query': required parameter is missing"},
		{"bad limit", "search_apocrypha", `{"query":"智慧","limit":"x"}`, "Invalid parameter 'limit': \"x\" - must be an integer"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			api := &scriptureAPI{t: t}
			_, err := callOn(t, api, c.tool, c.args)
			if err == nil || err.Error() != c.want {
				t.Fatalf("err = %v\nwant %q", err, c.want)
			}
			if api.calls() != 0 {
				t.Errorf("must not reach upstream: %+v", api)
			}
		})
	}
}

func TestScriptureBookTables(t *testing.T) {
	for _, c := range []struct {
		tool     string
		lo, hi   int
		wantName string
	}{{"list_apocrypha_books", 101, 115, "apocrypha"}, {"list_apostolic_fathers_books", 201, 217, "apostolic_fathers"}} {
		out, err := callOn(t, &scriptureAPI{t: t}, c.tool, `{}`)
		if err != nil {
			t.Fatal(err)
		}
		var v struct {
			Status string `json:"status"`
			Count  int    `json:"book_count"`
			Books  []struct {
				ID            int      `json:"id"`
				NameZH        string   `json:"name_zh"`
				NameEN        string   `json:"name_en"`
				Abbreviations []string `json:"abbreviations"`
			} `json:"books"`
		}
		if err := json.Unmarshal([]byte(out), &v); err != nil || v.Status != "success" || v.Count != c.hi-c.lo+1 || len(v.Books) != v.Count {
			t.Fatalf("%s: %v %+v", c.tool, err, v)
		}
		seen := map[string]int{}
		for i, b := range v.Books {
			if b.ID != c.lo+i || b.NameZH == "" || b.NameEN == "" || len(b.Abbreviations) == 0 {
				t.Errorf("%s: book %d = %+v", c.tool, i, b)
			}
			for _, a := range b.Abbreviations {
				k := strings.ToLower(a)
				if prev, dup := seen[k]; dup {
					t.Errorf("%s: alias %q belongs to both %d and %d", c.tool, a, prev, b.ID)
				}
				seen[k] = b.ID
			}
		}
	}
}

func (f *scriptureAPI) RT(context.Context, int, int, bool) (*fhlapi.RTResponse, error) {
	return nil, errors.New("RT not expected in this test")
}
