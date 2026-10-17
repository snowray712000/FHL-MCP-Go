package main

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var when = time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC)

// fakeDist writes archives shaped like tools/release output plus SHA256SUMS.txt.
func fakeDist(t *testing.T, ver string) string {
	t.Helper()
	dir := t.TempDir()
	var sums []string
	for _, p := range platforms {
		top := fmt.Sprintf("fhl-mcp-go_%s_%s_%s", ver, p.goos, p.goarch)
		files := map[string]string{p.exe: "BIN-" + p.key(), "LICENSE": "lic", "THIRD_PARTY_NOTICES.md": "notices-" + p.key(), "快速開始.md": "q"}
		var name string
		var buf bytes.Buffer
		if p.goos == "windows" {
			name = top + ".zip"
			zw := zip.NewWriter(&buf)
			for n, d := range files {
				w, _ := zw.CreateHeader(&zip.FileHeader{Name: top + "/" + n, Modified: when, Method: zip.Deflate})
				w.Write([]byte(d))
			}
			zw.Close()
		} else {
			name = top + ".tar.gz"
			gz := gzip.NewWriter(&buf)
			tw := tar.NewWriter(gz)
			for n, d := range files {
				tw.WriteHeader(&tar.Header{Typeflag: tar.TypeReg, Name: top + "/" + n, Mode: 0o755, Size: int64(len(d)), ModTime: when})
				tw.Write([]byte(d))
			}
			tw.Close()
			gz.Close()
		}
		os.WriteFile(filepath.Join(dir, name), buf.Bytes(), 0o644)
		h := sha256.Sum256(buf.Bytes())
		sums = append(sums, hex.EncodeToString(h[:])+"  "+name)
	}
	os.WriteFile(filepath.Join(dir, "SHA256SUMS.txt"), []byte(strings.Join(sums, "\n")+"\n"), 0o644)
	return dir
}

type entry struct {
	data string
	mode int64
}

func readTgz(t *testing.T, p string) map[string]entry {
	t.Helper()
	f, err := os.Open(p)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	tr := tar.NewReader(gz)
	m := map[string]entry{}
	for {
		h, err := tr.Next()
		if err == io.EOF {
			return m
		}
		if err != nil {
			t.Fatal(err)
		}
		if h.Typeflag == tar.TypeReg {
			b, _ := io.ReadAll(tr)
			m[h.Name] = entry{string(b), h.Mode}
		}
	}
}

func TestBuildsMainAndPlatformPackages(t *testing.T) {
	dist := fakeDist(t, "9.9.9")
	out := filepath.Join(t.TempDir(), "npm")
	names, err := run(dist, out, "fhlnet", "9.9.9")
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 4 || names[3] != "fhlnet-fhl-mcp-9.9.9.tgz" {
		t.Fatalf("names (main package must be last): %v", names)
	}
	win := readTgz(t, filepath.Join(out, "fhlnet-fhl-mcp-win32-x64-9.9.9.tgz"))
	if win["package/bin/fhl-mcp.exe"].data != "BIN-win32-x64" || win["package/bin/fhl-mcp.exe"].mode != 0o755 {
		t.Errorf("windows binary entry: %+v", win["package/bin/fhl-mcp.exe"])
	}
	if win["package/THIRD_PARTY_NOTICES.md"].data != "notices-win32-x64" || win["package/LICENSE"].data != "lic" {
		t.Errorf("license files not taken from the archive: %v", win)
	}
	var pj pkgJSON
	if err := json.Unmarshal([]byte(win["package/package.json"].data), &pj); err != nil {
		t.Fatal(err)
	}
	if pj.Name != "@fhlnet/fhl-mcp-win32-x64" || pj.Version != "9.9.9" || pj.OS[0] != "win32" || pj.CPU[0] != "x64" || pj.PublishConfig["access"] != "public" {
		t.Errorf("platform package.json: %+v", pj)
	}
	mac := readTgz(t, filepath.Join(out, "fhlnet-fhl-mcp-darwin-arm64-9.9.9.tgz"))
	if mac["package/bin/fhl-mcp"].data != "BIN-darwin-arm64" {
		t.Errorf("darwin binary: %+v", mac)
	}
	m := readTgz(t, filepath.Join(out, "fhlnet-fhl-mcp-9.9.9.tgz"))
	var mj pkgJSON
	json.Unmarshal([]byte(m["package/package.json"].data), &mj)
	if mj.Bin["fhl-mcp"] != "bin/fhl-mcp.js" || len(mj.OptionalDependencies) != 3 || mj.OptionalDependencies["@fhlnet/fhl-mcp-linux-x64"] != "9.9.9" {
		t.Errorf("main package.json: %+v", mj)
	}
	l := m["package/bin/fhl-mcp.js"]
	if !strings.Contains(l.data, "'@fhlnet/fhl-mcp-darwin-arm64'") || strings.Contains(l.data, "@SCOPE@") || l.mode != 0o755 {
		t.Errorf("launcher not rendered: mode %o", l.mode)
	}
}

func TestOutputIsReproducible(t *testing.T) {
	dist := fakeDist(t, "9.9.9")
	a, b := t.TempDir(), t.TempDir()
	run(dist, a, "fhlnet", "9.9.9")
	run(dist, b, "fhlnet", "9.9.9")
	for _, n := range []string{"fhlnet-fhl-mcp-9.9.9.tgz", "fhlnet-fhl-mcp-linux-x64-9.9.9.tgz"} {
		x, _ := os.ReadFile(filepath.Join(a, n))
		y, _ := os.ReadFile(filepath.Join(b, n))
		if !bytes.Equal(x, y) {
			t.Errorf("%s differs between runs", n)
		}
	}
}

func TestRejectsTamperedOrMissingArchive(t *testing.T) {
	dist := fakeDist(t, "9.9.9")
	os.WriteFile(filepath.Join(dist, "fhl-mcp-go_9.9.9_linux_amd64.tar.gz"), []byte("changed"), 0o644)
	if _, err := run(dist, t.TempDir(), "fhlnet", "9.9.9"); err == nil || !strings.Contains(err.Error(), "SHA256SUMS") {
		t.Errorf("tampered archive accepted: %v", err)
	}
	if _, err := run(dist, t.TempDir(), "fhlnet", "1.0.0"); err == nil {
		t.Error("missing version accepted")
	}
}

func TestRejectsNonSemverVersion(t *testing.T) {
	if _, err := run(t.TempDir(), t.TempDir(), "fhlnet", "1.1.2.1"); err == nil || !strings.Contains(err.Error(), "semver") {
		t.Errorf("dev build version accepted: %v", err)
	}
}

func TestSemverPattern(t *testing.T) {
	for v, want := range map[string]bool{"1.1.1": true, "1.2.1": true, "1.1.3-rc.1": true, "1x1x1": false, "1.1.2.1": false, "1.1": false} {
		if semverRe.MatchString(v) != want {
			t.Errorf("semverRe(%q) = %v, want %v", v, !want, want)
		}
	}
}
