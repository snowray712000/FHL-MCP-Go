package bible

import "testing"

func TestCleanTWCB(t *testing.T) {
	in := "μέν 連接詞\r\n肯定（#可9:12;羅7:25|，在#羅16:19|他們）。\r\n<div class=\"idt\">一、連用。\r\n<div class=\"idt\">1. 「<span class=\"exp\">但</span>」，#太3:11|。</div>\r\n</div>\r\n\r\n\r\n\r\n尾"
	want := "μέν 連接詞\n肯定（可9:12;羅7:25，在羅16:19他們）。\n\n一、連用。\n\n1. 「但」，太3:11。\n\n尾"
	if got := cleanTWCB(in); got != want {
		t.Errorf("cleanTWCB:\n got %q\nwant %q", got, want)
	}
	if got := cleanTWCB(" \r\n "); got != "" {
		t.Errorf("blank = %q", got)
	}
}
