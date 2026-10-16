package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCacheClearCommand(t *testing.T) {
	data := t.TempDir()
	dir := filepath.Join(data, "cache", "v1")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, n := range []string{"a.json", "b.json"} {
		if err := os.WriteFile(filepath.Join(dir, n), []byte("{}"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	var out, errb bytes.Buffer
	if code := run([]string{"cache", "clear", "--data-dir", data}, &out, &errb); code != 0 {
		t.Fatalf("exit %d: %s", code, errb.String())
	}
	if !strings.Contains(out.String(), "已清除 2 個快取檔案") {
		t.Errorf("output = %q", out.String())
	}
	if des, _ := os.ReadDir(dir); len(des) != 0 {
		t.Errorf("files left: %v", des)
	}
	if code := run([]string{"cache"}, &out, &errb); code != 2 {
		t.Errorf("missing subcommand: exit %d", code)
	}
	if code := run([]string{"cache", "clear", "--bogus"}, &out, &errb); code != 2 {
		t.Errorf("bad flag: exit %d", code)
	}
}
