package assets

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestToolsListMatchesM0Contract(t *testing.T) {
	want, err := os.ReadFile(filepath.Join("..", "doc", "m0", "contract", "tools-list.json"))
	if err != nil {
		t.Skipf("M0 contract not available: %v", err)
	}
	norm := func(b []byte) []byte { return bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n")) }
	if !bytes.Equal(norm(ToolsListJSON), norm(want)) {
		t.Fatal("assets/tools-list.json differs from doc/m0/contract/tools-list.json; re-copy it")
	}
}
