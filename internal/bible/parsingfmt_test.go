package bible

import "testing"

// Expected values come from running the Python reference's utils/parsing_format.py on the same inputs.
func TestOriginalTextSectionMatchesPython(t *testing.T) {
	cases := []struct {
		in      string
		greek   string // address (43,1,1), tp 2
		hebrew  string // address (1,1,1), tp 3
		wordIDs int
	}{
		{"a + b c + d e + f", "c1v1w1 `a`  c1v1w6 `d` c1v1w7 `e`  c1v1w9 `f`", "⁧b1c1v1w1 `a`  b1c1v1w6 `d` b1c1v1w7 `e`  b1c1v1w9 `f`⁩", 9},
		{"x + y + z + q", "c1v1w1 `x`  c1v1w5 `z`  c1v1w7 `q`", "⁧b1c1v1w1 `x`  b1c1v1w5 `z`  b1c1v1w7 `q`⁩", 7},
		{"no markers, here. ok\nnew line",
			"c1v1w1 `no` c1v1w2 `markers`, c1v1w3 `here`. c1v1w4 `ok` ↩ c1v1w5 `new` c1v1w6 `line`",
			"⁧b1c1v1w1 `no` b1c1v1w2 `markers`, b1c1v1w3 `here`. b1c1v1w4 `ok` ↪ b1c1v1w5 `new` b1c1v1w6 `line`⁩", 6},
		{"+ w1 + u1 + tail", " c1v1w4 `u1`  c1v1w6 `tail`", "⁧ b1c1v1w4 `u1`  b1c1v1w6 `tail`⁩", 6},
		{"α־β γ", "c1v1w1 `α`־c1v1w2 `β` c1v1w3 `γ`", "⁧b1c1v1w1 `α` ־ b1c1v1w2 `β` b1c1v1w3 `γ`⁩", 3},
	}
	for _, c := range cases {
		words := bindWords(c.in)
		last := 0
		for _, w := range words {
			if w.wid > last {
				last = w.wid
			}
		}
		if last != c.wordIDs {
			t.Errorf("%q: highest wid = %d, want %d", c.in, last, c.wordIDs)
		}
		if got := buildOriginalTextSection(words, address{43, 1, 1}, 2, false); got != c.greek {
			t.Errorf("%q greek:\n got %q\nwant %q", c.in, got, c.greek)
		}
		if got := buildOriginalTextSection(words, address{1, 1, 1}, 3, true); got != c.hebrew {
			t.Errorf("%q hebrew:\n got %q\nwant %q", c.in, got, c.hebrew)
		}
	}
}

func TestAddressTPAndTokenIDs(t *testing.T) {
	tp := func(a ...address) int { return computeAddressTP(a) }
	if got := []int{
		tp(address{1, 1, 1}, address{1, 1, 2}), tp(address{1, 1, 1}, address{1, 2, 1}),
		tp(address{1, 1, 1}, address{2, 1, 1}), tp(address{1, 1, 1}),
	}; got[0] != 1 || got[1] != 2 || got[2] != 3 || got[3] != 0 {
		t.Errorf("tp = %v, want [1 2 3 0]", got)
	}
	a := address{1, 2, 3}
	if got := []string{formatBCVW(a, -1, 3), formatBCVW(a, 4, 2), formatBCVW(a, 4, 1), formatBCVW(a, 4, 0)}; got[0] != "b1c2v3" || got[1] != "c2v3w4" || got[2] != "v3w4" || got[3] != "w4" {
		t.Errorf("token ids = %v", got)
	}
}

func TestTranslationAndOTNormalization(t *testing.T) {
	if got := formatTranslation("a\r\nb\nc"); got != "a ↩ b ↩ c" {
		t.Errorf("formatTranslation = %q", got)
	}
	if got := normalizeOTWordText("l1\r\nl2\nl3"); got != "l3\nl2\nl1" {
		t.Errorf("normalizeOTWordText = %q", got)
	}
}
