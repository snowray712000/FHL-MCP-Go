package app

import (
	"path/filepath"
	"testing"
)

func TestResolveCacheDir(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(EnvDataDir, "")
	got, err := ResolveCacheDir(dir)
	if err != nil || got != filepath.Join(dir, "cache") {
		t.Errorf("flag: %q, %v", got, err)
	}
	other := t.TempDir()
	t.Setenv(EnvDataDir, other)
	if got, _ := ResolveCacheDir(""); got != filepath.Join(other, "cache") {
		t.Errorf("env: %q", got)
	}
	if got, _ := ResolveCacheDir(dir); got != filepath.Join(dir, "cache") {
		t.Errorf("flag must win over the environment: %q", got)
	}
	t.Setenv(EnvDataDir, "")
	got, err = ResolveCacheDir("")
	if err != nil || filepath.Base(got) != "cache" || filepath.Base(filepath.Dir(got)) != "fhl-mcp" || !filepath.IsAbs(got) {
		t.Errorf("default: %q, %v", got, err)
	}
}

func TestResolveCacheMaxBytes(t *testing.T) {
	t.Setenv(EnvCacheMaxMB, "")
	if n, err := ResolveCacheMaxBytes(-1); err != nil || n != 200<<20 {
		t.Errorf("default: %d, %v", n, err)
	}
	if n, _ := ResolveCacheMaxBytes(5); n != 5<<20 {
		t.Errorf("flag: %d", n)
	}
	if n, _ := ResolveCacheMaxBytes(0); n != 0 {
		t.Errorf("0 must disable the cache: %d", n)
	}
	t.Setenv(EnvCacheMaxMB, "50")
	if n, _ := ResolveCacheMaxBytes(-1); n != 50<<20 {
		t.Errorf("env: %d", n)
	}
	if n, _ := ResolveCacheMaxBytes(7); n != 7<<20 {
		t.Errorf("flag must win over the environment: %d", n)
	}
	t.Setenv(EnvCacheMaxMB, "abc")
	if _, err := ResolveCacheMaxBytes(-1); err == nil {
		t.Error("a bad value must be reported, not ignored")
	}
}
