package cache

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func newTest(t *testing.T, max int64) (*Cache, *time.Time) {
	t.Helper()
	c := New(filepath.Join(t.TempDir(), "cache"), max)
	now := time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)
	c.Now = func() time.Time { return now }
	return c, &now
}

func files(t *testing.T, c *Cache) []string {
	t.Helper()
	des, err := os.ReadDir(c.root)
	if err != nil {
		return nil
	}
	var out []string
	for _, de := range des {
		out = append(out, de.Name())
	}
	return out
}

func TestPutGetAndTTL(t *testing.T) {
	c, now := newTest(t, 0)
	body := []byte(`{"status":"success","text":"愛 <a> & b"}`)
	c.Put("k1", body, time.Hour)
	c.Put("forever", body, 0)
	got, ok := c.Get("k1")
	if !ok || !strings.Contains(string(got), `"text"`) {
		t.Fatalf("Get = %q, %v", got, ok)
	}
	if _, ok := c.Get("other"); ok {
		t.Error("unknown key must miss")
	}
	*now = now.Add(59 * time.Minute)
	if _, ok := c.Get("k1"); !ok {
		t.Error("entry must still be valid before its TTL")
	}
	*now = now.Add(2 * time.Minute)
	if _, ok := c.Get("k1"); ok {
		t.Error("entry must expire after its TTL")
	}
	if len(files(t, c)) != 1 {
		t.Errorf("the expired file should be removed: %v", files(t, c))
	}
	*now = now.Add(24 * 365 * time.Hour)
	if _, ok := c.Get("forever"); !ok {
		t.Error("ttl 0 means no expiry")
	}
}

func TestCorruptEntryIsAMiss(t *testing.T) {
	c, _ := newTest(t, 0)
	c.Put("k", []byte(`{"a":1}`), time.Hour)
	p := c.path("k")
	for _, bad := range []string{"", "not json", `{"v":999,"key":"k","body":{"a":1}}`, `{"v":1,"key":"other","body":{"a":1}}`, `{"v":1,"key":"k"}`} {
		if err := os.WriteFile(p, []byte(bad), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, ok := c.Get("k"); ok {
			t.Errorf("%q must be a miss", bad)
		}
		if _, err := os.Stat(p); err == nil {
			t.Errorf("%q: corrupt file should be removed", bad)
		}
	}
}

func TestWriteIsAtomicAndLeavesNoTempFiles(t *testing.T) {
	c, _ := newTest(t, 0)
	for i := 0; i < 20; i++ {
		c.Put("same", []byte(`{"n":1}`), time.Hour)
	}
	if got := files(t, c); len(got) != 1 || strings.HasPrefix(got[0], tmpPrefix) {
		t.Errorf("files = %v", got)
	}
}

func TestEvictsOldestWhenOverTheCap(t *testing.T) {
	c, _ := newTest(t, 3000)
	body := []byte(`{"pad":"` + strings.Repeat("x", 400) + `"}`)
	for i := 0; i < 20; i++ {
		c.Put("key"+string(rune('a'+i)), body, 0)
		// distinct, increasing modification times so "oldest" is well defined
		_ = os.Chtimes(c.path("key"+string(rune('a'+i))), time.Unix(1000+int64(i), 0), time.Unix(1000+int64(i), 0))
	}
	var total int64
	for _, name := range files(t, c) {
		st, _ := os.Stat(filepath.Join(c.root, name))
		total += st.Size()
	}
	if total > 3000 {
		t.Errorf("cache holds %d bytes, cap is 3000", total)
	}
	if _, ok := c.Get("keya"); ok {
		t.Error("the oldest entry should have been evicted")
	}
	if _, ok := c.Get("keyt"); !ok {
		t.Error("the newest entry must survive")
	}
}

func TestSizeIsMeasuredFromDiskOnce(t *testing.T) {
	c, _ := newTest(t, 2000)
	body := []byte(`{"pad":"` + strings.Repeat("x", 600) + `"}`)
	c.Put("a", body, 0)
	c.Put("b", body, 0)
	// A second Cache over the same directory (e.g. after a restart) must account for existing files.
	c2 := New(filepath.Dir(c.root), 2000)
	c2.Put("c", body, 0)
	c2.Put("d", body, 0)
	var total int64
	for _, name := range files(t, c2) {
		st, _ := os.Stat(filepath.Join(c2.root, name))
		total += st.Size()
	}
	if total > 2000 {
		t.Errorf("cache holds %d bytes after restart, cap is 2000", total)
	}
}

func TestDoStoresOnlyWhenAllowed(t *testing.T) {
	c, _ := newTest(t, 0)
	calls := 0
	fetch := func(store bool, err error) Fetch {
		return func(context.Context) ([]byte, bool, error) { calls++; return []byte(`{"n":1}`), store, err }
	}
	ctx := context.Background()
	for i := 0; i < 2; i++ {
		if _, err := c.Do(ctx, "no-store", time.Hour, fetch(false, nil)); err != nil {
			t.Fatal(err)
		}
	}
	if calls != 2 {
		t.Errorf("a non-storable answer must be fetched again (calls=%d)", calls)
	}
	boom := errors.New("boom")
	if _, err := c.Do(ctx, "err", time.Hour, fetch(true, boom)); !errors.Is(err, boom) {
		t.Errorf("err = %v", err)
	}
	if _, ok := c.Get("err"); ok {
		t.Error("an error must not be cached")
	}
	calls = 0
	for i := 0; i < 3; i++ {
		if _, err := c.Do(ctx, "ok", time.Hour, fetch(true, nil)); err != nil {
			t.Fatal(err)
		}
	}
	if calls != 1 {
		t.Errorf("a stored answer must be reused (calls=%d)", calls)
	}
	if h, m := c.Stats(); h != 2 || m != 4 {
		t.Errorf("stats hits=%d misses=%d", h, m)
	}
}

func TestConcurrentIdenticalRequestsShareOneFetch(t *testing.T) {
	c, _ := newTest(t, 0)
	var calls atomic.Int32
	release := make(chan struct{})
	fetch := func(context.Context) ([]byte, bool, error) {
		calls.Add(1)
		<-release
		return []byte(`{"n":1}`), true, nil
	}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := c.Do(context.Background(), "k", time.Hour, fetch); err != nil {
				t.Error(err)
			}
		}()
	}
	time.Sleep(100 * time.Millisecond) // let them all queue behind the first fetch
	close(release)
	wg.Wait()
	if calls.Load() != 1 {
		t.Errorf("fetch ran %d times, want 1", calls.Load())
	}
}

func TestFollowerRetriesWhenTheLeaderIsCancelled(t *testing.T) {
	c, _ := newTest(t, 0)
	leaderStarted := make(chan struct{})
	leaderCtx, cancel := context.WithCancel(context.Background())
	var calls atomic.Int32
	fetch := func(ctx context.Context) ([]byte, bool, error) {
		if calls.Add(1) == 1 {
			close(leaderStarted)
			<-ctx.Done()
			return nil, false, ctx.Err()
		}
		return []byte(`{"n":2}`), true, nil
	}
	done := make(chan error, 1)
	go func() { _, err := c.Do(leaderCtx, "k", time.Hour, fetch); done <- err }()
	<-leaderStarted
	followerDone := make(chan []byte, 1)
	go func() {
		b, err := c.Do(context.Background(), "k", time.Hour, fetch)
		if err != nil {
			t.Error(err)
		}
		followerDone <- b
	}()
	time.Sleep(50 * time.Millisecond)
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Errorf("leader err = %v", err)
	}
	if b := <-followerDone; string(b) != `{"n":2}` {
		t.Errorf("follower got %q", b)
	}
}

func TestClear(t *testing.T) {
	c, _ := newTest(t, 0)
	c.Put("a", []byte(`{"a":1}`), 0)
	c.Put("b", []byte(`{"b":1}`), 0)
	// An unrelated file and an older format directory.
	keep := filepath.Join(c.root, "notes.txt")
	if err := os.WriteFile(keep, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	old := filepath.Join(filepath.Dir(c.root), "v0")
	if err := os.MkdirAll(old, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(old, "x.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	n, err := c.Clear()
	if err != nil || n != 3 {
		t.Errorf("Clear = %d, %v, want 3 files", n, err)
	}
	if _, err := os.Stat(keep); err != nil {
		t.Error("unrelated files must be left alone")
	}
	if _, ok := c.Get("a"); ok {
		t.Error("cleared entry must miss")
	}
	if n, err := New(filepath.Join(t.TempDir(), "nothing"), 0).Clear(); err != nil || n != 0 {
		t.Errorf("Clear on a missing directory = %d, %v", n, err)
	}
}
