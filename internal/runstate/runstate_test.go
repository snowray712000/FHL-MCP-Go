package runstate

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func sample() State {
	return State{Instance: "abc", PID: 1, Host: "127.0.0.1", Port: 8081, URL: "http://127.0.0.1:8081/mcp",
		Version: "1", Token: "tok", StartedAt: time.Now().UTC().Truncate(time.Second)}
}

func TestWriteReadRoundTrip(t *testing.T) {
	p := Path(t.TempDir(), 8081)
	if got, err := Read(p); got != nil || err != nil {
		t.Fatalf("missing file: %v %v", got, err)
	}
	s := sample()
	if err := Write(p, s); err != nil {
		t.Fatal(err)
	}
	got, err := Read(p)
	if err != nil || got == nil || *got != s {
		t.Fatalf("got %+v err %v want %+v", got, err, s)
	}
	// No temp files are left behind.
	if des, _ := os.ReadDir(filepath.Dir(p)); len(des) != 1 {
		t.Errorf("run dir has %d entries, want 1", len(des))
	}
}

func TestRemoveOnlyOwnInstance(t *testing.T) {
	p := Path(t.TempDir(), 1)
	if err := Write(p, sample()); err != nil {
		t.Fatal(err)
	}
	if err := Remove(p, "someone-else"); err != nil {
		t.Fatal(err)
	}
	if got, _ := Read(p); got == nil {
		t.Fatal("a different instance's state was removed")
	}
	if err := Remove(p, "abc"); err != nil {
		t.Fatal(err)
	}
	if got, _ := Read(p); got != nil {
		t.Fatal("own state was not removed")
	}
	if err := Remove(p, "abc"); err != nil {
		t.Errorf("removing a missing file: %v", err)
	}
}

func TestCorruptAndIncomplete(t *testing.T) {
	p := Path(t.TempDir(), 1)
	_ = os.MkdirAll(filepath.Dir(p), 0o700)
	for _, body := range []string{"not json", `{"instance":"x"}`} {
		_ = os.WriteFile(p, []byte(body), 0o600)
		if got, err := Read(p); err == nil || got != nil {
			t.Errorf("body %q: got %+v err %v, want an error", body, got, err)
		}
	}
	if err := Remove(p, "x"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(p); !os.IsNotExist(err) {
		t.Error("a corrupt state file should be cleared by Remove")
	}
}

func TestNewRandomIsUniqueAndHex(t *testing.T) {
	a, _ := NewRandom(32)
	b, _ := NewRandom(32)
	if len(a) != 64 || a == b {
		t.Errorf("a=%q b=%q", a, b)
	}
}
