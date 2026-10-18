package hanzi

import "testing"

func TestForMode(t *testing.T) {
	cases := []struct {
		in   string
		simp bool
		want string
	}{
		{"神愛世人", true, "神爱世人"},
		{"神爱世人", false, "神愛世人"},
		{"神爱世人", true, "神爱世人"},
		{"神愛世人", false, "神愛世人"},
		{"G25 love", true, "G25 love"},
		{"", true, ""},
	}
	for _, c := range cases {
		if got := ForMode(c.in, c.simp); got != c.want {
			t.Errorf("ForMode(%q,%v)=%q want %q", c.in, c.simp, got, c.want)
		}
	}
}
