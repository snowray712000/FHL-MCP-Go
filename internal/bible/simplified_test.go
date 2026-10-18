package bible

import "testing"

func TestCitationAcceptsSimplifiedBookNames(t *testing.T) {
	for in, want := range map[string]string{
		"约 3:16":          "約",
		"约翰福音3:16-18":     "約翰福音",
		"创世记 1":           "創世記",
		"罗马书 8:1; 启 22:1": "羅馬書",
		"犹 5":             "猶",
	} {
		qs, ok := ParseCitationQueries(in)
		if !ok || len(qs) == 0 {
			t.Errorf("%q not parsed", in)
			continue
		}
		if id, ok := ResolveBookID(qs[0].Book); !ok || id == 0 {
			t.Errorf("%q: book %q unresolved", in, qs[0].Book)
		}
		_ = want
	}
}

func TestScriptureSetAcceptsSimplifiedBookNames(t *testing.T) {
	if b, err := Apocrypha.resolve("友弟德传"); err != nil || b == nil {
		t.Errorf("友弟德传: %v", err)
	}
	if b, err := ApostolicFathers.resolve("黑马牧人书"); err != nil || b == nil {
		t.Errorf("黑马牧人书: %v", err)
	}
	if b, err := Apocrypha.resolve("德训篇"); err != nil || b == nil {
		t.Errorf("德训篇: %v", err)
	}
}

func TestFoldSimplifiedBookLeavesTraditionalAlone(t *testing.T) {
	if s, changed := foldSimplifiedBook("約翰福音"); changed || s != "約翰福音" {
		t.Errorf("traditional changed: %q", s)
	}
}
