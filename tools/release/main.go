// Command release builds the distributable archives (plan §11): one archive per
// target with the binary, double-click scripts, quick start, LICENSE and
// third-party notices, plus SHA256SUMS.txt, then audits what it produced.
//
// Run from the repository root:
//
//	go run ./tools/release                 # all three targets into dist/
//	go run ./tools/release -targets windows/amd64
//
// Builds are CGO_ENABLED=0 and -trimpath, and archive contents/mtimes are derived
// from the git commit, so the same commit gives the same files.
package main

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"
)

const (
	mainPkg    = "./cmd/fhl-mcp"
	versionPkg = "github.com/snowray712000/FHL-MCP-Go/internal/version"
)

type target struct{ goos, goarch string }

func (t target) String() string { return t.goos + "/" + t.goarch }

type file struct {
	name string // path inside the archive's top folder
	data []byte
	mode int64
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "release:", err)
		os.Exit(1)
	}
}

func run() error {
	out := flag.String("out", "dist", "output directory (created if missing; existing release files are replaced)")
	targetsFlag := flag.String("targets", "windows/amd64,darwin/arm64,linux/amd64", "comma-separated GOOS/GOARCH list")
	allowDirty := flag.Bool("allow-dirty", false, "allow building with uncommitted changes (marks the commit as -dirty)")
	flag.Parse()

	root, err := os.Getwd()
	if err != nil {
		return err
	}
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		return fmt.Errorf("run from the repository root (no go.mod in %s)", root)
	}
	ver, err := checkVersion(root)
	if err != nil {
		return err
	}
	commit, when, dirty, err := gitInfo(root)
	if err != nil {
		return err
	}
	if dirty && !*allowDirty {
		return fmt.Errorf("working tree has uncommitted changes; commit first or pass -allow-dirty (the build is then stamped %s-dirty)", commit)
	}
	stamp := commit
	if dirty {
		stamp += "-dirty"
	}
	date := when.UTC().Format("2006-01-02")

	var targets []target
	for _, s := range strings.Split(*targetsFlag, ",") {
		os_, arch, ok := strings.Cut(strings.TrimSpace(s), "/")
		if !ok {
			return fmt.Errorf("bad target %q (want GOOS/GOARCH)", s)
		}
		targets = append(targets, target{os_, arch})
	}

	outDir := filepath.Join(root, *out)
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	staging, err := os.MkdirTemp("", "fhl-release-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(staging)

	var sums []string
	for _, t := range targets {
		fmt.Printf("== %s %s (%s)\n", ver, t, stamp)
		top := fmt.Sprintf("fhl-mcp-go_%s_%s_%s", ver, t.goos, t.goarch)
		files, err := buildFiles(root, staging, t, top, ver, stamp, date)
		if err != nil {
			return fmt.Errorf("%s: %w", t, err)
		}
		archive := top + ".tar.gz"
		if t.goos == "windows" {
			archive = top + ".zip"
		}
		path := filepath.Join(outDir, archive)
		if t.goos == "windows" {
			err = writeZip(path, top, files, when)
		} else {
			err = writeTarGz(path, top, files, when)
		}
		if err != nil {
			return fmt.Errorf("%s: %w", t, err)
		}
		if err := audit(root, files); err != nil {
			return fmt.Errorf("%s audit: %w", t, err)
		}
		sum, err := fileSHA256(path)
		if err != nil {
			return err
		}
		sums = append(sums, fmt.Sprintf("%s  %s", sum, archive))
		fmt.Printf("   %s  %s\n", sum, archive)
	}
	sort.Strings(sums)
	return os.WriteFile(filepath.Join(outDir, "SHA256SUMS.txt"), []byte(strings.Join(sums, "\n")+"\n"), 0o644)
}

// checkVersion reads the version constant and requires it to match app_versions.md.
func checkVersion(root string) (string, error) {
	b, err := os.ReadFile(filepath.Join(root, "internal", "version", "version.go"))
	if err != nil {
		return "", err
	}
	m := regexp.MustCompile(`const Version = "([0-9]+(?:\.[0-9]+){3})"`).FindSubmatch(b)
	if m == nil {
		return "", fmt.Errorf("cannot find the version constant in internal/version/version.go")
	}
	ver := string(m[1])
	log, err := os.ReadFile(filepath.Join(root, "app_versions.md"))
	if err != nil {
		return "", err
	}
	if !bytes.Contains(log, []byte("- last: **"+ver+"**")) {
		return "", fmt.Errorf("app_versions.md `last` is not %s", ver)
	}
	return ver, nil
}

func gitInfo(root string) (commit string, when time.Time, dirty bool, err error) {
	g := func(args ...string) (string, error) {
		c := exec.Command("git", args...)
		c.Dir = root
		b, err := c.Output()
		return strings.TrimSpace(string(b)), err
	}
	if commit, err = g("rev-parse", "--short=7", "HEAD"); err != nil {
		return "", time.Time{}, false, fmt.Errorf("git rev-parse: %w", err)
	}
	ts, err := g("show", "-s", "--format=%cI", "HEAD")
	if err != nil {
		return "", time.Time{}, false, fmt.Errorf("git show: %w", err)
	}
	if when, err = time.Parse(time.RFC3339, ts); err != nil {
		return "", time.Time{}, false, err
	}
	st, err := g("status", "--porcelain")
	if err != nil {
		return "", time.Time{}, false, fmt.Errorf("git status: %w", err)
	}
	return commit, when, st != "", nil
}

// buildFiles compiles the binary for t and returns everything that goes in the archive.
func buildFiles(root, staging string, t target, top, ver, stamp, date string) ([]file, error) {
	exe := "fhl-mcp"
	if t.goos == "windows" {
		exe += ".exe"
	}
	binPath := filepath.Join(staging, top+"_"+exe)
	ldflags := fmt.Sprintf("-s -w -X %s.Commit=%s -X %s.Date=%s", versionPkg, stamp, versionPkg, date)
	env := append(os.Environ(), "CGO_ENABLED=0", "GOOS="+t.goos, "GOARCH="+t.goarch, "GOFLAGS=-mod=readonly")
	if err := runCmd(root, env, "go", "build", "-trimpath", "-buildvcs=false", "-ldflags", ldflags, "-o", binPath, mainPkg); err != nil {
		return nil, err
	}
	bin, err := os.ReadFile(binPath)
	if err != nil {
		return nil, err
	}

	// Smoke test when the binary can run on this machine.
	if t.goos == runtime.GOOS && t.goarch == runtime.GOARCH {
		got, err := exec.Command(binPath, "version").Output()
		if err != nil {
			return nil, fmt.Errorf("running the built binary: %w", err)
		}
		for _, want := range []string{ver, stamp, date} {
			if !strings.Contains(string(got), want) {
				return nil, fmt.Errorf("`version` output %q lacks %q", strings.TrimSpace(string(got)), want)
			}
		}
		fmt.Printf("   %s", got)
	}

	read := func(rel string) ([]byte, error) { return os.ReadFile(filepath.Join(root, filepath.FromSlash(rel))) }
	license, err := read("LICENSE")
	if err != nil {
		return nil, err
	}
	notices, err := buildNotices(root, t)
	if err != nil {
		return nil, err
	}
	quick, err := read("doc/快速開始.md")
	if err != nil {
		return nil, err
	}
	files := []file{
		{exe, bin, 0o755},
		{"LICENSE", normLF(license), 0o644},
		{"THIRD_PARTY_NOTICES.md", normLF(notices), 0o644},
		{"快速開始.md", normLF(quick), 0o644},
	}
	var scripts []string
	var crlf bool
	switch t.goos {
	case "windows":
		scripts, crlf = []string{"啟動-FHL-MCP.cmd", "首次設定.cmd"}, true
	case "darwin":
		scripts = []string{"Start-FHL-MCP.command", "Setup.command"}
	}
	for _, s := range scripts {
		b, err := read("scripts/" + s)
		if err != nil {
			return nil, err
		}
		b = normLF(b)
		mode := int64(0o755)
		if crlf {
			b = bytes.ReplaceAll(b, []byte("\n"), []byte("\r\n"))
			mode = 0o644
		}
		files = append(files, file{s, b, mode})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].name < files[j].name })
	return files, nil
}

func normLF(b []byte) []byte { return bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n")) }

// buildNotices is release/NOTICE-head.md plus the license text of every module
// linked into the binary for target t (taken from the module cache, so it cannot go stale).
func buildNotices(root string, t target) ([]byte, error) {
	head, err := os.ReadFile(filepath.Join(root, "release", "NOTICE-head.md"))
	if err != nil {
		return nil, err
	}
	c := exec.Command("go", "list", "-deps", "-f", "{{with .Module}}{{if not .Main}}{{.Path}}|{{.Version}}|{{.Dir}}{{end}}{{end}}", mainPkg)
	c.Dir = root
	c.Env = append(os.Environ(), "CGO_ENABLED=0", "GOOS="+t.goos, "GOARCH="+t.goarch)
	listed, err := c.Output()
	if err != nil {
		return nil, fmt.Errorf("go list: %w", err)
	}
	seen := map[string]bool{}
	var mods []string
	for _, l := range strings.Split(string(listed), "\n") {
		if l = strings.TrimSpace(l); l != "" && !seen[l] {
			seen[l] = true
			mods = append(mods, l)
		}
	}
	sort.Strings(mods)
	var b bytes.Buffer
	b.Write(bytes.TrimRight(normLF(head), "\n"))
	b.WriteString("\n")
	for _, m := range mods {
		parts := strings.Split(m, "|")
		if len(parts) != 3 || parts[2] == "" {
			return nil, fmt.Errorf("module %q has no source directory (run `go mod download`)", m)
		}
		var text []byte
		for _, n := range []string{"LICENSE", "LICENSE.md", "LICENSE.txt", "COPYING"} {
			if text, err = os.ReadFile(filepath.Join(parts[2], n)); err == nil {
				break
			}
		}
		if text == nil {
			return nil, fmt.Errorf("module %s %s has no license file", parts[0], parts[1])
		}
		fmt.Fprintf(&b, "\n### %s %s\n\n```\n%s\n```\n", parts[0], parts[1], strings.TrimRight(string(normLF(text)), "\n"))
	}
	return b.Bytes(), nil
}

func runCmd(dir string, env []string, name string, args ...string) error {
	c := exec.Command(name, args...)
	c.Dir, c.Env = dir, env
	c.Stdout, c.Stderr = os.Stdout, os.Stderr
	return c.Run()
}

func writeZip(path, top string, files []file, when time.Time) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	zw := zip.NewWriter(f)
	for _, fl := range files {
		h := &zip.FileHeader{Name: top + "/" + fl.name, Method: zip.Deflate, Modified: when}
		h.SetMode(os.FileMode(fl.mode))
		w, err := zw.CreateHeader(h)
		if err != nil {
			return err
		}
		if _, err := w.Write(fl.data); err != nil {
			return err
		}
	}
	if err := zw.Close(); err != nil {
		return err
	}
	return f.Close()
}

func writeTarGz(path, top string, files []file, when time.Time) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewWriterLevel(f, gzip.BestCompression)
	if err != nil {
		return err
	}
	tw := tar.NewWriter(gz)
	dir := &tar.Header{Typeflag: tar.TypeDir, Name: top + "/", Mode: 0o755, ModTime: when, Format: tar.FormatPAX}
	if err := tw.WriteHeader(dir); err != nil {
		return err
	}
	for _, fl := range files {
		h := &tar.Header{Typeflag: tar.TypeReg, Name: top + "/" + fl.name, Mode: fl.mode, Size: int64(len(fl.data)), ModTime: when, Format: tar.FormatPAX}
		if err := tw.WriteHeader(h); err != nil {
			return err
		}
		if _, err := tw.Write(fl.data); err != nil {
			return err
		}
	}
	if err := tw.Close(); err != nil {
		return err
	}
	if err := gz.Close(); err != nil {
		return err
	}
	return f.Close()
}

// audit checks the plan §11 item 5 rule: nothing machine- or user-specific ships.
// Built-in file lists are tiny and fixed, so the check is "no unexpected files" plus
// "no local path / home directory / user name strings in any file".
func audit(root string, files []file) error {
	allowed := regexp.MustCompile(`^(fhl-mcp(\.exe)?|LICENSE|THIRD_PARTY_NOTICES\.md|快速開始\.md|啟動-FHL-MCP\.cmd|首次設定\.cmd|Start-FHL-MCP\.command|Setup\.command)$`)
	forbidden := map[string]bool{}
	add := func(s string) {
		if len(s) >= 4 {
			forbidden[s] = true
			forbidden[filepath.ToSlash(s)] = true
			forbidden[strings.ReplaceAll(s, "/", `\`)] = true
		}
	}
	add(root)
	if h, err := os.UserHomeDir(); err == nil {
		add(h)
	}
	if h := os.Getenv("TEMP"); h != "" {
		add(h)
	}
	for _, fl := range files {
		if !allowed.MatchString(fl.name) {
			return fmt.Errorf("unexpected file %q in the archive", fl.name)
		}
		for bad := range forbidden {
			if bytes.Contains(fl.data, []byte(bad)) {
				return fmt.Errorf("%s contains the local path %q", fl.name, bad)
			}
		}
	}
	return nil
}

func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
