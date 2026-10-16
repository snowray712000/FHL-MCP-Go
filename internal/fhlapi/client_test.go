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
