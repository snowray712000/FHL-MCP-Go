package main

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAuditRejectsLocalPathsAndStrayFiles(t *testing.T) {
	root := filepath.Join(t.TempDir(), "secret-checkout")
	good := []file{{"fhl-mcp", []byte("binary"), 0o755}, {"LICENSE", []byte("text"), 0o644}}
	if err := audit(root, good); err != nil {
		t.Fatalf("clean files rejected: %v", err)
	}
	leak := append([]file{{"fhl-mcp", []byte("x " + root + " y"), 0o755}}, good[1:]...)
	if err := audit(root, leak); err == nil || !strings.Contains(err.Error(), "local path") {
		t.Fatalf("leaked repo path not caught: %v", err)
	}
	stray := append([]file{{"state-8081.json", []byte("{}"), 0o644}}, good...)
	if err := audit(root, stray); err == nil || !strings.Contains(err.Error(), "unexpected file") {
		t.Fatalf("stray file not caught: %v", err)
	}
}

func TestArchivesKeepModesAndAreDeterministic(t *testing.T) {
	files := []file{{"fhl-mcp", []byte("bin"), 0o755}, {"LICENSE", []byte("lic"), 0o644}}
	when := time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)
	dir := t.TempDir()
	read := func(name string) []byte {
		b, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	for _, n := range []string{"a.tar.gz", "b.tar.gz"} {
		if err := writeTarGz(filepath.Join(dir, n), "top", files, when); err != nil {
			t.Fatal(err)
		}
	}
	if !bytes.Equal(read("a.tar.gz"), read("b.tar.gz")) {
		t.Fatal("tar.gz output differs between runs")
	}
	gz, err := gzip.NewReader(bytes.NewReader(read("a.tar.gz")))
	if err != nil {
		t.Fatal(err)
	}
	tr := tar.NewReader(gz)
	modes := map[string]int64{}
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		modes[h.Name] = h.Mode
	}
	if modes["top/fhl-mcp"] != 0o755 || modes["top/LICENSE"] != 0o644 {
		t.Fatalf("modes lost: %v", modes)
	}

	if err := writeZip(filepath.Join(dir, "a.zip"), "top", files, when); err != nil {
		t.Fatal(err)
	}
	zr, err := zip.OpenReader(filepath.Join(dir, "a.zip"))
	if err != nil {
		t.Fatal(err)
	}
	defer zr.Close()
	if len(zr.File) != 2 || zr.File[0].Name != "top/fhl-mcp" {
		t.Fatalf("zip entries: %v", zr.File)
	}
}

func TestNormLF(t *testing.T) {
	if got := normLF([]byte("a\r\nb\n")); string(got) != "a\nb\n" {
		t.Fatalf("got %q", got)
	}
}
