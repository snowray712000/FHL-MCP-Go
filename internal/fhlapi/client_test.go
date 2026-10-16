package fhlapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/snowray712000/FHL-MCP-Go/internal/cache"
)

func fixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "tests", "fixtures", "upstream", name))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func newMock(t *testing.T, h http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return &Client{BaseURL: srv.URL + "/api", HTTP: srv.Client(), UserAgent: "test-agent"}
}

func TestQBSendsExpectedQuery(t *testing.T) {
	var got http.Request
	c := newMock(t, func(w http.ResponseWriter, r *http.Request) {
		got = *r.Clone(r.Context())
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write(fixture(t, "qb_john_3_16.json"))
	})
	resp, err := c.QB(context.Background(), QBRequest{BID: 43, Chapter: 3, Verse: "16", Version: "unv", Strong: false, Simplified: true})
	if err != nil {
		t.Fatal(err)
	}
	if got.URL.Path != "/api/qb.php" {
		t.Errorf("path = %q", got.URL.Path)
	}
	q := got.URL.Query()
	for k, v := range map[string]string{"bid": "43", "chap": "3", "version": "unv", "strong": "0", "sec": "16", "gb": "1"} {
		if q.Get(k) != v {
			t.Errorf("query %s = %q want %q", k, q.Get(k), v)
		}
	}
	if got.Header.Get("User-Agent") != "test-agent" {
		t.Errorf("user agent = %q", got.Header.Get("User-Agent"))
	}
	if resp.RecordCount != 1 || resp.Record[0].Sec != 16 || resp.Prev == nil || resp.Next == nil {
		t.Errorf("unexpected response %+v", resp)
	}
}

func TestQBWholeChapterOmitsSec(t *testing.T) {
	c := newMock(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Has("sec") {
			t.Errorf("sec must be absent for a whole chapter: %s", r.URL.RawQuery)
		}
		_, _ = w.Write(fixture(t, "qb_john_99_1_empty.json"))
	})
	if _, err := c.QB(context.Background(), QBRequest{BID: 43, Chapter: 3, Version: "unv"}); err != nil {
		t.Fatal(err)
	}
}

func TestQBErrors(t *testing.T) {
	cases := []struct {
		name string
		h    http.HandlerFunc
		kind error
	}{
		{"http 500", func(w http.ResponseWriter, r *http.Request) { http.Error(w, "boom", 500) }, ErrHTTP},
		{"not json", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("<html>")) }, ErrParse},
		{"status fail", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(`{"status":"Fail"}`)) }, ErrUpstream},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := newMock(t, tc.h)
			_, err := c.QB(context.Background(), QBRequest{BID: 43, Chapter: 3, Version: "unv"})
			if !errors.Is(err, tc.kind) {
				t.Fatalf("err = %v, want kind %v", err, tc.kind)
			}
		})
	}
}

func TestQBNetworkErrorAndCancel(t *testing.T) {
	c := &Client{BaseURL: "http://127.0.0.1:1/api", HTTP: http.DefaultClient}
	if _, err := c.QB(context.Background(), QBRequest{BID: 43, Chapter: 3}); !errors.Is(err, ErrNetwork) {
		t.Fatalf("err = %v, want network error", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := c.QB(ctx, QBRequest{BID: 43, Chapter: 3}); !errors.Is(err, ErrNetwork) {
		t.Fatalf("cancelled: err = %v", err)
	}
}

func TestABQueryAndParse(t *testing.T) {
	var path, gb string
	c := newMock(t, func(w http.ResponseWriter, r *http.Request) {
		path, gb = r.URL.Path, r.URL.Query().Get("gb")
		_, _ = w.Write(fixture(t, "ab.json"))
	})
	res, err := c.AB(context.Background(), true)
	if err != nil {
		t.Fatal(err)
	}
	if path != "/api/ab.php" || gb != "1" {
		t.Errorf("path=%s gb=%s, want /api/ab.php and gb=1 (D4)", path, gb)
	}
	if len(res.Record) != 86 || res.Record[0].Book != "unv" || res.Record[0].Strong != 1 || res.Record[2].OTOnly != 1 {
		t.Errorf("unexpected parse: %d records, first %+v", len(res.Record), res.Record[0])
	}
}

func TestABUpstreamFailureStatus(t *testing.T) {
	c := newMock(t, func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(`{"status":"Fail"}`)) })
	if _, err := c.AB(context.Background(), false); !errors.Is(err, ErrUpstream) {
		t.Errorf("err = %v, want ErrUpstream", err)
	}
}

func TestSEQuery(t *testing.T) {
	var q map[string]string
	c := newMock(t, func(w http.ResponseWriter, r *http.Request) {
		q = map[string]string{}
		for k := range r.URL.Query() {
			q[k] = r.URL.Query().Get(k)
		}
		_, _ = w.Write(fixture(t, "se_ai_john_range_limit3.json"))
	})
	res, err := c.SE(context.Background(), SERequest{Query: "愛", Version: "unv", Orig: 0, Range: 3, Limit: 3, RangeBID: 43, RangeEID: 43, Simplified: true})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"VERSION": "unv", "orig": "0", "q": "愛", "RANGE": "3", "offset": "0", "count_only": "0", "index_only": "0", "limit": "3", "range_bid": "43", "range_eid": "43", "gb": "1"}
	for k, v := range want {
		if q[k] != v {
			t.Errorf("param %s = %q, want %q", k, q[k], v)
		}
	}
	if len(q) != len(want) {
		t.Errorf("unexpected params: %v", q)
	}
	if res.RecordCount != 3 || len(res.Record) != 3 || res.Record[0].Engs != "John" || res.Record[0].Sec != 16 {
		t.Errorf("unexpected parse: %+v", res)
	}

	c = newMock(t, func(w http.ResponseWriter, r *http.Request) {
		if _, ok := r.URL.Query()["range_bid"]; ok {
			t.Error("range_bid must only be sent for RANGE=3")
		}
		_, _ = w.Write(fixture(t, "se_ai_limit3.json"))
	})
	if _, err := c.SE(context.Background(), SERequest{Query: "愛", Version: "unv", Limit: 3}); err != nil {
		t.Fatal(err)
	}
}

func TestSEEmptyQueryFailStatus(t *testing.T) {
	c := newMock(t, func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write(fixture(t, "se_empty_query.json")) })
	_, err := c.SE(context.Background(), SERequest{Version: "unv", Limit: 50})
	if !errors.Is(err, ErrUpstream) {
		t.Fatalf("err = %v, want ErrUpstream", err)
	}
}

func TestSDQueryAndParse(t *testing.T) {
	var n, k string
	c := newMock(t, func(w http.ResponseWriter, r *http.Request) {
		n, k = r.URL.Query().Get("N"), r.URL.Query().Get("k")
		_, _ = w.Write(fixture(t, "sd_nt_25.json"))
	})
	res, err := c.SD(context.Background(), true, 25, false)
	if err != nil {
		t.Fatal(err)
	}
	if n != "0" || k != "25" {
		t.Errorf("N=%s k=%s", n, k)
	}
	if len(res.Record) != 1 || res.Record[0].SN != "00025" || len(res.Record[0].Same) != 3 {
		t.Errorf("unexpected parse: %+v", res.Record)
	}

	c = newMock(t, func(w http.ResponseWriter, r *http.Request) {
		n = r.URL.Query().Get("N")
		_, _ = w.Write(fixture(t, "sd_nt_99999.json"))
	})
	res, err = c.SD(context.Background(), false, 430, false)
	if err != nil || n != "1" {
		t.Fatalf("err=%v N=%s (Hebrew must send N=1)", err, n)
	}
	if string(res.Record[0].Orig) != "null" || res.Record[0].SN != "" {
		t.Errorf("not-found record must keep explicit nulls: %+v", res.Record[0])
	}
}

func TestQPQueryAndParse(t *testing.T) {
	var q map[string]string
	c := newMock(t, func(w http.ResponseWriter, r *http.Request) {
		q = map[string]string{}
		for k := range r.URL.Query() {
			q[k] = r.URL.Query().Get(k)
		}
		_, _ = w.Write(fixture(t, "qp_gen_1_1.json"))
	})
	res, err := c.QP(context.Background(), 1, 1, 1, true)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"bid": "1", "chap": "1", "sec": "1", "gb": "1"}
	for k, v := range want {
		if q[k] != v {
			t.Errorf("param %s = %q, want %q", k, q[k], v)
		}
	}
	if len(q) != len(want) || res.N != 1 || len(res.Record) != 8 || string(res.Record[0].WID) != "0" {
		t.Errorf("unexpected parse: q=%v N=%d records=%d", q, res.N, len(res.Record))
	}
}

func TestQPMalformedJSONForMissingVerse(t *testing.T) {
	c := newMock(t, func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write(fixture(t, "qp_john_99_1.json")) })
	_, err := c.QP(context.Background(), 43, 99, 1, false)
	if !errors.Is(err, ErrParse) {
		t.Errorf("err = %v, want ErrParse (upstream sends invalid JSON for a verse that does not exist)", err)
	}
}

func TestSCQueries(t *testing.T) {
	var queries []string
	c := newMock(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/sc.php" {
			t.Errorf("path = %q", r.URL.Path)
		}
		queries = append(queries, r.URL.RawQuery)
		if r.URL.Query().Has("validbook") {
			_, _ = w.Write(fixture(t, "sc_list.json"))
			return
		}
		_, _ = w.Write(fixture(t, "sc_john_3_16.json"))
	})
	list, err := c.SCList(context.Background(), true)
	if err != nil || list.RecordCount != 7 || len(list.Record) != 7 || list.Record[3].ID != 4 {
		t.Fatalf("list = %+v, err %v", list, err)
	}
	v := 16
	res, err := c.SC(context.Background(), 43, 3, &v, false)
	if err != nil || res.RecordCount != 1 || res.Record[0].ComText == nil || res.Record[0].BookName == "" {
		t.Fatalf("sc = %+v, err %v", res, err)
	}
	if _, err := c.SC(context.Background(), 43, 3, nil, false); err != nil {
		t.Fatal(err)
	}
	want := []string{"gb=1&validbook=1", "bid=43&chap=3&gb=0&sec=16", "bid=43&chap=3&gb=0"}
	for i := range want {
		if queries[i] != want[i] {
			t.Errorf("query %d = %q, want %q", i, queries[i], want[i])
		}
	}
}

func TestSCWithoutRecords(t *testing.T) {
	c := newMock(t, func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write(fixture(t, "sc_john_99_1.json")) })
	v := 1
	res, err := c.SC(context.Background(), 43, 99, &v, false)
	if err != nil || res.RecordCount != 0 || len(res.Record) != 0 {
		t.Errorf("res = %+v, err %v (no record key means no commentary)", res, err)
	}
}

func TestSSCQueryAndStatuses(t *testing.T) {
	var query string
	c := newMock(t, func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.RawQuery
		switch r.URL.Query().Get("key") {
		case "愛":
			_, _ = w.Write(fixture(t, "ssc_love_fail.json"))
		case "zzqqxx":
			_, _ = w.Write(fixture(t, "ssc_none.json"))
		default:
			_, _ = w.Write(fixture(t, "ssc_nicodemus.json"))
		}
	})
	res, err := c.SSC(context.Background(), "尼哥底母", true)
	if err != nil || len(res.Record) != 5 || res.Record[2].BID != 42 || res.Record[0].Tag != 3 {
		t.Fatalf("res = %+v, err %v", res, err)
	}
	if want := "gb=1&key=%E5%B0%BC%E5%93%A5%E5%BA%95%E6%AF%8D"; query != want {
		t.Errorf("query = %q, want %q", query, want)
	}
	if res, err := c.SSC(context.Background(), "zzqqxx", false); err != nil || len(res.Record) != 0 {
		t.Errorf("no hits: %+v, %v", res, err)
	}
	if _, err := c.SSC(context.Background(), "愛", false); !errors.Is(err, ErrTooManyRecords) {
		t.Errorf("err = %v, want ErrTooManyRecords", err)
	}
}

func TestSTQueryAndParse(t *testing.T) {
	var queries []string
	c := newMock(t, func(w http.ResponseWriter, r *http.Request) {
		queries = append(queries, r.URL.RawQuery)
		if r.URL.Query().Get("count_only") == "1" {
			_, _ = w.Write(fixture(t, "st_love_count.json"))
			return
		}
		_, _ = w.Write(fixture(t, "st_faith_torrey_en.json"))
	})
	cnt, err := c.ST(context.Background(), 4, "愛", true, false)
	if err != nil || cnt.RecordCount == nil || *cnt.RecordCount != 23 || len(cnt.Record) != 0 {
		t.Fatalf("count = %+v, err %v", cnt, err)
	}
	res, err := c.ST(context.Background(), 0, "faith", false, true)
	if err != nil || len(res.Record) != 3 || res.Record[2].ID != "T0000191" || res.Record[2].Topic == nil || *res.Record[2].Topic != "Faith" {
		t.Fatalf("res = %+v, err %v", res, err)
	}
	want := []string{"N=4&count_only=1&gb=0&keyword=%E6%84%9B", "N=0&count_only=0&gb=1&keyword=faith"}
	for i := range want {
		if queries[i] != want[i] {
			t.Errorf("query %d = %q, want %q", i, queries[i], want[i])
		}
	}
}

func TestCommentaryEndpointsRejectFailureStatus(t *testing.T) {
	c := newMock(t, func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(`{"status":"Fail:something"}`)) })
	v := 1
	if _, err := c.SC(context.Background(), 43, 3, &v, false); !errors.Is(err, ErrUpstream) {
		t.Errorf("SC err = %v", err)
	}
	if _, err := c.SCList(context.Background(), false); !errors.Is(err, ErrUpstream) {
		t.Errorf("SCList err = %v", err)
	}
	if _, err := c.SSC(context.Background(), "x", false); !errors.Is(err, ErrUpstream) {
		t.Errorf("SSC err = %v", err)
	}
	if _, err := c.ST(context.Background(), 4, "x", false, false); !errors.Is(err, ErrUpstream) {
		t.Errorf("ST err = %v", err)
	}
}

func newCachedMock(t *testing.T, h http.HandlerFunc) (*Client, *int) {
	t.Helper()
	n := 0
	c := newMock(t, func(w http.ResponseWriter, r *http.Request) { n++; h(w, r) })
	c.Cache = cache.New(filepath.Join(t.TempDir(), "cache"), 0)
	return c, &n
}

func TestCacheServesRepeatedRequests(t *testing.T) {
	c, n := newCachedMock(t, func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write(fixture(t, "qb_john_3_16.json")) })
	req := QBRequest{BID: 43, Chapter: 3, Verse: "16", Version: "unv"}
	for i := 0; i < 3; i++ {
		res, err := c.QB(context.Background(), req)
		if err != nil || res.Record[0].Sec != 16 {
			t.Fatalf("QB %d: %+v, %v", i, res, err)
		}
	}
	if *n != 1 {
		t.Errorf("upstream saw %d requests, want 1", *n)
	}
	// Anything that changes the answer is a different entry: simplified, strong numbers, version, verse.
	for _, r := range []QBRequest{
		{BID: 43, Chapter: 3, Verse: "16", Version: "unv", Simplified: true},
		{BID: 43, Chapter: 3, Verse: "16", Version: "unv", Strong: true},
		{BID: 43, Chapter: 3, Verse: "16", Version: "kjv"},
		{BID: 43, Chapter: 3, Verse: "17", Version: "unv"},
	} {
		if _, err := c.QB(context.Background(), r); err != nil {
			t.Fatal(err)
		}
	}
	if *n != 5 {
		t.Errorf("upstream saw %d requests, want 5 (each variant is its own entry)", *n)
	}
}

func TestCacheSkipsFailuresAndMalformedBodies(t *testing.T) {
	body := `{"status":"Fail:record count > 500"}`
	c, n := newCachedMock(t, func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(body)) })
	for i := 0; i < 2; i++ {
		if _, err := c.SSC(context.Background(), "愛", false); !errors.Is(err, ErrTooManyRecords) {
			t.Fatalf("err = %v", err)
		}
	}
	if *n != 2 {
		t.Errorf("a failure status must not be cached (requests=%d)", *n)
	}

	c, n = newCachedMock(t, func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write(fixture(t, "qp_john_99_1.json")) })
	for i := 0; i < 2; i++ {
		if _, err := c.QP(context.Background(), 43, 99, 1, false); !errors.Is(err, ErrParse) {
			t.Fatalf("err = %v", err)
		}
	}
	if *n != 2 {
		t.Errorf("malformed JSON must not be cached (requests=%d)", *n)
	}

	c, n = newCachedMock(t, func(w http.ResponseWriter, _ *http.Request) { http.Error(w, "boom", 500) })
	for i := 0; i < 2; i++ {
		if _, err := c.AB(context.Background(), false); !errors.Is(err, ErrHTTP) {
			t.Fatalf("err = %v", err)
		}
	}
	if *n != 2 {
		t.Errorf("HTTP errors must not be cached (requests=%d)", *n)
	}
}

func TestCacheMergesConcurrentRequests(t *testing.T) {
	release := make(chan struct{})
	c, n := newCachedMock(t, func(w http.ResponseWriter, _ *http.Request) {
		<-release
		_, _ = w.Write(fixture(t, "ab.json"))
	})
	var wg sync.WaitGroup
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := c.AB(context.Background(), false); err != nil {
				t.Error(err)
			}
		}()
	}
	time.Sleep(100 * time.Millisecond)
	close(release)
	wg.Wait()
	if *n != 1 {
		t.Errorf("upstream saw %d requests, want 1", *n)
	}
}

func TestEveryEndpointHasACachePolicy(t *testing.T) {
	for _, ep := range []string{"qb.php", "ab.php", "se.php", "sd.php", "qp.php", "sc.php", "ssc.php", "st.php", "au.php", "qsub.php", "qaf.php", "sesub.php", "seaf.php", "rt.php", "json.php"} {
		if _, ok := endpointTTL[ep]; !ok {
			t.Errorf("%s has no cache policy", ep)
		}
	}
	if endpointTTL["sd.php"] != 0 || endpointTTL["se.php"] != 24*time.Hour || endpointTTL["qb.php"] != 7*24*time.Hour {
		t.Errorf("unexpected TTLs: %v", endpointTTL)
	}
}

func TestCacheDisabledByDefault(t *testing.T) {
	n := 0
	c := newMock(t, func(w http.ResponseWriter, _ *http.Request) { n++; _, _ = w.Write(fixture(t, "ab.json")) })
	for i := 0; i < 2; i++ {
		if _, err := c.AB(context.Background(), false); err != nil {
			t.Fatal(err)
		}
	}
	if n != 2 {
		t.Errorf("without a Cache every call goes upstream (requests=%d)", n)
	}
}

func TestAUQueryAndRawFields(t *testing.T) {
	var query string
	c := newMock(t, func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.RawQuery
		_, _ = w.Write(fixture(t, "au_john_3.json"))
	})
	res, err := c.AU(context.Background(), 1, 43, 3, true)
	if err != nil {
		t.Fatal(err)
	}
	if query != "bid=43&chap=3&gb=1&version=1" {
		t.Errorf("query = %q", query)
	}
	// The upstream sends numbers as strings; they must reach the output unchanged.
	if string(res.Chap) != `"3"` || string(res.PBID) != `"43"` || string(res.NChap) != `"4"` || len(res.MP3) == 0 {
		t.Errorf("raw fields = %s %s %s", res.Chap, res.PBID, res.NChap)
	}
	c = newMock(t, func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(`{"status":"Fail:x"}`)) })
	if _, err := c.AU(context.Background(), 0, 43, 3, false); !errors.Is(err, ErrUpstream) {
		t.Errorf("err = %v", err)
	}
}

func TestSubEndpointQueries(t *testing.T) {
	var path, query string
	body := "qsub_tob_1_1-3.json"
	c := newMock(t, func(w http.ResponseWriter, r *http.Request) {
		path, query = r.URL.Path, r.URL.RawQuery
		_, _ = w.Write(fixture(t, body))
	})
	verse := "1-3"
	res, err := c.QSub(context.Background(), 103, 1, &verse)
	if err != nil {
		t.Fatal(err)
	}
	if path != "/api/qsub.php" || query != "bid=103&chap=1&sec=1-3&strong=0" {
		t.Errorf("qsub request = %s?%s", path, query)
	}
	if string(res.RecordCount) != "3" || len(res.Record) != 3 || string(res.Record[0].Bid) != "103" || res.Version == nil || *res.Version != "c1933" {
		t.Errorf("qsub response = %+v", res)
	}
	// No verse means the whole chapter: sec is not sent at all.
	if _, err := c.QAF(context.Background(), 201, 1, nil); err != nil {
		t.Fatal(err)
	}
	if path != "/api/qaf.php" || query != "bid=201&chap=1&strong=0" {
		t.Errorf("qaf request = %s?%s", path, query)
	}

	body = "sesub_wisdom_off2_limit2.json"
	sr, err := c.SESub(context.Background(), "智慧", 2, 2)
	if err != nil {
		t.Fatal(err)
	}
	if path != "/api/sesub.php" || query != "limit=2&offset=2&q=%E6%99%BA%E6%85%A7" {
		t.Errorf("sesub request = %s?%s", path, query)
	}
	if string(sr.RecordCount) != "2" || len(sr.Record) != 2 {
		t.Errorf("sesub response = %+v", sr)
	}
	if _, err := c.SEAF(context.Background(), "教會", 0, 0); err != nil {
		t.Fatal(err)
	}
	if path != "/api/seaf.php" || query != "limit=0&offset=0&q=%E6%95%99%E6%9C%83" {
		t.Errorf("seaf request = %s?%s", path, query)
	}

	c = newMock(t, func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(`{"status":"Fail: Empty Query"}`)) })
	if _, err := c.SESub(context.Background(), "", 10, 0); !errors.Is(err, ErrUpstream) {
		t.Errorf("err = %v", err)
	}
}

func TestRTQueryAndNotFound(t *testing.T) {
	var query string
	body := "rt_gen_1_1.json"
	c := newMock(t, func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.RawQuery
		_, _ = w.Write(fixture(t, body))
	})
	res, err := c.RT(context.Background(), 1, 1, true)
	if err != nil {
		t.Fatal(err)
	}
	if query != "bid=1&gb=1&id=1&version=tcv" {
		t.Errorf("query = %q", query)
	}
	if res.RecordCount != 1 || len(res.Record) != 1 || string(res.Record[0].ID) != "1" || res.Engs == nil || *res.Engs != "Gen" || res.Record[0].Text == nil {
		t.Errorf("response = %+v", res)
	}
	body = "rt_999_999.json"
	res, err = c.RT(context.Background(), 999, 999, false)
	if err != nil || res.RecordCount != 0 || len(res.Record) != 0 {
		t.Errorf("not found: %v %+v", err, res)
	}
}

func TestArticlesQueryServerAndCache(t *testing.T) {
	n := 0
	var path, query string
	body := "articles_title_ai.json"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n++
		path, query = r.URL.Path, r.URL.RawQuery
		_, _ = w.Write(fixture(t, body))
	}))
	t.Cleanup(srv.Close)
	// The article search lives on another server than BaseURL; BaseURL must stay unused.
	c := &Client{BaseURL: "http://127.0.0.1:1/unused", ArticlesBaseURL: srv.URL + "/api", HTTP: srv.Client(), Cache: cache.New(t.TempDir(), 0)}
	q := ArticleQuery{Title: "愛", Column: "women3", Simplified: true}
	res, err := c.Articles(context.Background(), q)
	if err != nil {
		t.Fatal(err)
	}
	if path != "/api/json.php" || query != "gb=1&ptab=women3&title=%E6%84%9B" {
		t.Errorf("request = %s?%s", path, query)
	}
	if res.Status != 1 || res.RecordCount != 3 || len(res.Record) != 3 || string(res.Record[0].ID) != `"9247"` || len(res.Record[0].Txt) == 0 {
		t.Errorf("response = %+v", res)
	}
	// status 1 is a success that may be cached, a second identical query is answered from disk.
	if _, err := c.Articles(context.Background(), q); err != nil || n != 1 {
		t.Errorf("err=%v requests=%d, want a cache hit", err, n)
	}
	// "no data" (status 0) comes back as a normal answer and is not cached.
	body = "articles_none.json"
	for i := 0; i < 2; i++ {
		res, err = c.Articles(context.Background(), ArticleQuery{Title: "zzqqxx"})
		if err != nil || res.Status != 0 || res.Result != "no data" {
			t.Fatalf("no data: %v %+v", err, res)
		}
	}
	if n != 3 {
		t.Errorf("requests = %d, want 3 (failures are asked again)", n)
	}
}
