package app

import "testing"

func TestResolveSimplified(t *testing.T) {
	yes, no := true, false
	cases := []struct {
		env  string
		flag *bool
		want bool
		bad  bool
	}{
		{"", nil, false, false},
		{"1", nil, true, false},
		{"TRUE", nil, true, false},
		{"off", nil, false, false},
		{"1", &no, false, false}, // flag wins
		{"", &yes, true, false},
		{"maybe", nil, false, true},
	}
	for _, c := range cases {
		t.Setenv(EnvSimplified, c.env)
		got, err := ResolveSimplified(c.flag)
		if (err != nil) != c.bad || got != c.want {
			t.Errorf("env=%q flag=%v: got %v,%v", c.env, c.flag, got, err)
		}
	}
}
