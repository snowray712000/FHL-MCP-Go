package bible

import "testing"

func TestResolveBookID(t *testing.T) {
	cases := []struct {
		in   string
		want int
		ok   bool
	}{
		{"John", 43, true},
		{"john", 43, true},
		{"JOHN", 43, true},
		{"約", 43, true},
		{"約翰福音", 43, true},
		{"Philem", 57, true},
		{"Philemon", 57, true},
		{"門", 57, true},
		{"1Cor", 46, true},
		{"1 Corinthians", 46, true},
		{"林前", 46, true},
		{"Ps", 19, true},
		{"詩", 19, true},
		{"Titus", 56, true},
		{"提多書", 56, true},
		{"Tob", 101, true},
		{"多俾亞傳", 101, true},
		{"多", 56, true}, // Python resolves this to 101 (later table entry wins); Go keeps the canon book
		{"43", 43, true},
		{"101", 101, true},
		{"113", 113, true}, // inside the accepted 101-115 range even without a table entry
		{"0", 0, false},
		{"67", 0, false},
		{"Phm", 0, false}, // not a Python alias either (D11 handles aliases in M2)
		{" John", 0, false},
		{"", 0, false},
		{"不存在書卷", 0, false},
	}
	for _, c := range cases {
		got, ok := ResolveBookID(c.in)
		if ok != c.ok || got != c.want {
			t.Errorf("ResolveBookID(%q) = %d,%v want %d,%v", c.in, got, ok, c.want, c.ok)
		}
	}
}

func TestBookTableSize(t *testing.T) {
	if len(books) != 86 {
		t.Fatalf("book table has %d entries, want 86", len(books))
	}
}
