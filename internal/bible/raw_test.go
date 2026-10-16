package bible

import (
	"encoding/json"
	"testing"
)

func TestNormUnescapesAndKeepsNumbers(t *testing.T) {
	cases := map[string]string{
		`"\u662f\u9580\u9582 <b> & \u0022q\u0022"`: `"是門閂 <b> & \"q\""`,
		`"3"`:      `"3"`,
		`103`:      `103`,
		`1.50`:     `1.50`,
		`null`:     `null`,
		`"a\/b"`:   `"a/b"`,
		`not json`: `not json`,
	}
	for in, want := range cases {
		if got := string(norm(json.RawMessage(in))); got != want {
			t.Errorf("norm(%s) = %s, want %s", in, got, want)
		}
	}
	if len(norm(nil)) != 0 {
		t.Error("norm(nil) must stay empty")
	}
}
