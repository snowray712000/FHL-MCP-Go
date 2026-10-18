package tools

import (
	"context"
	"encoding/json"
	"testing"
)

func TestDefaultSimplifiedInjectsOnlyWhenOmitted(t *testing.T) {
	var seen bool
	r := &Registry{handlers: map[string]handler{}, DefaultSimplified: true}
	r.handlers["probe"] = func(_ context.Context, a args) (string, error) {
		var err error
		seen, err = a.optBool("use_simplified", false)
		return "", err
	}
	for raw, want := range map[string]bool{
		`{}`:                         true,
		``:                           true,
		`{"use_simplified":null}`:    true,
		`{"use_simplified":true}`:    true,
		`{"use_simplified":false}`:   false,
		`{"use_simplified":"false"}`: false,
	} {
		if _, err := r.Call(context.Background(), "probe", json.RawMessage(raw)); err != nil {
			t.Fatalf("%q: %v", raw, err)
		}
		if seen != want {
			t.Errorf("%q: use_simplified=%v want %v", raw, seen, want)
		}
	}
	r.DefaultSimplified = false
	r.Call(context.Background(), "probe", json.RawMessage(`{}`))
	if seen {
		t.Error("default off must stay false")
	}
}
