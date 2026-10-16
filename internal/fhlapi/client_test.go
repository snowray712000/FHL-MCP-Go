package fhlapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
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
