package tools

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/snowray712000/FHL-MCP-Go/internal/bible"
	"github.com/snowray712000/FHL-MCP-Go/internal/fhlapi"
)

type fakeAPI struct{ got fhlapi.QBRequest }

func (f *fakeAPI) QB(_ context.Context, r fhlapi.QBRequest) (*fhlapi.QBResponse, error) {
	f.got = r
	b, err := os.ReadFile(filepath.Join("..", "..", "tests", "fixtures", "upstream", "qb_john_3_16.json"))
	if err != nil {
		return nil, err
	}
	var out fhlapi.QBResponse
	return &out, json.Unmarshal(b, &out)
}

func call(t *testing.T, args string) (string, error, *fakeAPI) {
	t.Helper()
	api := &fakeAPI{}
	r := NewRegistry(&bible.Service{API: api})
	out, err := r.Call(context.Background(), "get_bible_verse", json.RawMessage(args))
	return out, err, api
}

func TestGetBibleVerseArgumentCoercion(t *testing.T) {
	cases := []struct {
		name, args string
		want       fhlapi.QBRequest
	}{
		{"plain", `{"book":"John","chapter":3,"verse":"16"}`, fhlapi.QBRequest{BID: 43, Chapter: 3, Verse: "16", Version: "unv"}},
		{"chapter as string (D6)", `{"book":"John","chapter":"3","verse":"16"}`, fhlapi.QBRequest{BID: 43, Chapter: 3, Verse: "16", Version: "unv"}},
		{"chapter 3.0", `{"book":"John","chapter":3.0}`, fhlapi.QBRequest{BID: 43, Chapter: 3, Version: "unv"}},
		{"null optionals (D5)", `{"book":"John","chapter":3,"verse":null,"version":null,"include_strong":null,"use_simplified":null}`, fhlapi.QBRequest{BID: 43, Chapter: 3, Version: "unv"}},
		{"flags", `{"book":"約","chapter":3,"include_strong":true,"use_simplified":true,"version":"kjv"}`, fhlapi.QBRequest{BID: 43, Chapter: 3, Version: "kjv", Strong: true, Simplified: true}},
		{"extra args ignored", `{"book":"John","chapter":3,"bogus":1}`, fhlapi.QBRequest{BID: 43, Chapter: 3, Version: "unv"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out, err, api := call(t, c.args)
			if err != nil {
				t.Fatal(err)
			}
			if api.got != c.want {
				t.Errorf("request = %+v want %+v", api.got, c.want)
			}
			if !strings.HasPrefix(out, "{\n  \"version\": \"unv\"") {
				t.Errorf("unexpected output: %.60q", out)
			}
		})
	}
}

func TestGetBibleVerseValidationErrors(t *testing.T) {
	cases := []struct{ name, args, want string }{
		{"missing book", `{"chapter":3}`, "Invalid parameter 'book': required parameter is missing"},
		{"missing chapter", `{"book":"John"}`, "Invalid parameter 'chapter': required parameter is missing"},
		{"chapter abc", `{"book":"John","chapter":"abc"}`, `Invalid parameter 'chapter': "abc" - must be an integer`},
		{"chapter fraction", `{"book":"John","chapter":3.5}`, "Invalid parameter 'chapter': 3.5 - must be an integer"},
		{"chapter bool", `{"book":"John","chapter":true}`, "Invalid parameter 'chapter': true - must be an integer"},
		{"book number", `{"book":43,"chapter":3}`, "Invalid parameter 'book': 43 - must be a string"},
		{"flag junk", `{"book":"John","chapter":3,"include_strong":"maybe"}`, `Invalid parameter 'include_strong': "maybe" - must be true or false`},
		{"unknown book", `{"book":"不存在書卷","chapter":1,"verse":"1"}`, "Invalid parameter 'book': 不存在書卷 - Book not found. Please check the spelling or use a valid book code."},
		{"args not object", `[1]`, "Invalid parameter 'arguments': must be a JSON object"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err, api := call(t, c.args)
			if err == nil || err.Error() != c.want {
				t.Fatalf("err = %v, want %q", err, c.want)
			}
			if api.got != (fhlapi.QBRequest{}) {
				t.Errorf("upstream must not be called on invalid input, got %+v", api.got)
			}
		})
	}
}

func TestNotImplemented(t *testing.T) {
	r := NewRegistry(&bible.Service{API: &fakeAPI{}})
	if r.Implemented("list_commentaries") {
		t.Fatal("list_commentaries should not be implemented yet")
	}
	if _, err := r.Call(context.Background(), "list_commentaries", nil); !errors.Is(err, ErrNotImplemented) {
		t.Fatalf("err = %v", err)
	}
}

func (f *fakeAPI) AB(context.Context, bool) (*fhlapi.ABResponse, error) {
	return nil, errors.New("AB not expected in this test")
}
