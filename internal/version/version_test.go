package version

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The binary version must match `last` in app_versions.md (project rule).
func TestMatchesAppVersions(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "..", "app_versions.md"))
	if err != nil {
		t.Skipf("app_versions.md not available: %v", err)
	}
	want := "- last: **" + Version + "**"
	if !strings.Contains(string(b), want) {
		t.Fatalf("app_versions.md does not contain %q", want)
	}
}
