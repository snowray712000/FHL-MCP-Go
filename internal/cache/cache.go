// Package cache is a small on-disk cache for upstream JSON responses: one file per entry,
// atomic writes, TTL expiry, a disk-size cap with oldest-first cleanup, and merging of
// concurrent identical lookups. A corrupt or unreadable entry is simply a miss.
package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// FormatVersion is part of the directory name and of every key, so a change of the entry
// format never reads (or is polluted by) older files.
const FormatVersion = 1

// DefaultMaxBytes is the disk cap used when none is configured.
const DefaultMaxBytes int64 = 200 << 20

const (
	entrySuffix = ".json"
	tmpPrefix   = ".tmp-"
)

// entry is the on-disk form. Body is the upstream response as received (valid JSON).
type entry struct {
	V         int             `json:"v"`
	Key       string          `json:"key"`
	CachedAt  int64           `json:"cached_at"`
	ExpiresAt int64           `json:"expires_at"` // unix seconds; 0 = permanent
	Body      json.RawMessage `json:"body"`
}

// Cache is safe for concurrent use. The zero value is not usable; call New.
type Cache struct {
	// Now is the clock; nil means time.Now. Tests replace it.
	Now func() time.Time

	root     string // <dir>/v<FormatVersion>
	maxBytes int64

	mu     sync.Mutex // guards size, sized, flight
	size   int64
	sized  bool // size has been measured from disk
	flight map[string]*call

	hits, misses atomic.Int64
}

type call struct {
	done chan struct{}
	body []byte
	err  error
}

// New returns a cache that keeps its files under dir (created on the first write).
// maxBytes <= 0 selects DefaultMaxBytes.
func New(dir string, maxBytes int64) *Cache {
	if maxBytes <= 0 {
		maxBytes = DefaultMaxBytes
	}
	return &Cache{
		root:     filepath.Join(dir, "v"+strconv.Itoa(FormatVersion)),
		maxBytes: maxBytes,
		flight:   map[string]*call{},
	}
}

func (c *Cache) now() time.Time {
	if c.Now != nil {
		return c.Now()
	}
	return time.Now()
}

// Stats reports the lookups answered from disk and the ones that were not.
func (c *Cache) Stats() (hits, misses int64) { return c.hits.Load(), c.misses.Load() }

func (c *Cache) path(key string) string {
	sum := sha256.Sum256([]byte(key))
	return filepath.Join(c.root, hex.EncodeToString(sum[:])+entrySuffix)
}

// Get returns the stored body for key. Expired, corrupt or mismatching entries are misses
// (and are removed).
func (c *Cache) Get(key string) ([]byte, bool) {
	p := c.path(key)
	raw, err := os.ReadFile(p)
	if err != nil {
		return nil, false
	}
	var e entry
	if err := json.Unmarshal(raw, &e); err != nil || e.V != FormatVersion || e.Key != key || len(e.Body) == 0 {
		c.remove(p, int64(len(raw)))
		return nil, false
	}
	if e.ExpiresAt != 0 && c.now().Unix() >= e.ExpiresAt {
		c.remove(p, int64(len(raw)))
		return nil, false
	}
	// Best effort: a read keeps an entry young for the oldest-first cleanup.
	now := c.now()
	_ = os.Chtimes(p, now, now)
	return e.Body, true
}

// Put stores body under key. ttl <= 0 means the entry does not expire (it can still be
// removed by the size cleanup). Failures are ignored: the cache is an optimisation.
func (c *Cache) Put(key string, body []byte, ttl time.Duration) {
	e := entry{V: FormatVersion, Key: key, CachedAt: c.now().Unix(), Body: body}
	if ttl > 0 {
		e.ExpiresAt = c.now().Add(ttl).Unix()
	}
	data, err := json.Marshal(e)
	if err != nil {
		return
	}
	if err := os.MkdirAll(c.root, 0o755); err != nil {
		return
	}
	c.ensureSized()

	p := c.path(key)
	var old int64
	if st, err := os.Stat(p); err == nil {
		old = st.Size()
	}
	tmp, err := os.CreateTemp(c.root, tmpPrefix+"*")
	if err != nil {
		return
	}
	_, werr := tmp.Write(data)
	cerr := tmp.Close()
	if werr != nil || cerr != nil || os.Rename(tmp.Name(), p) != nil {
		_ = os.Remove(tmp.Name())
		return
	}
	c.mu.Lock()
	c.size += int64(len(data)) - old
	over := c.size > c.maxBytes
	c.mu.Unlock()
	if over {
		c.evict()
	}
}

// ensureSized measures the directory once (flat listing, no recursion) so the size counter
// starts right; afterwards it is maintained incrementally.
func (c *Cache) ensureSized() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.sized {
		return
	}
	c.sized = true
	c.size = 0
	des, err := os.ReadDir(c.root)
	if err != nil {
		return
	}
	for _, de := range des {
		if info, err := de.Info(); err == nil && !de.IsDir() {
			c.size += info.Size()
		}
	}
}

func (c *Cache) remove(p string, size int64) {
	if os.Remove(p) == nil {
		c.mu.Lock()
		c.size -= size
		if c.size < 0 {
			c.size = 0
		}
		c.mu.Unlock()
	}
}

// evict deletes the least recently used files until the cache is at 90% of its cap.
func (c *Cache) evict() {
	des, err := os.ReadDir(c.root)
	if err != nil {
		return
	}
	type file struct {
		path string
		size int64
		mod  time.Time
	}
	var files []file
	var total int64
	for _, de := range des {
		if de.IsDir() {
			continue
		}
		info, err := de.Info()
		if err != nil {
			continue
		}
		files = append(files, file{filepath.Join(c.root, de.Name()), info.Size(), info.ModTime()})
		total += info.Size()
	}
	sort.Slice(files, func(i, j int) bool { return files[i].mod.Before(files[j].mod) })
	target := c.maxBytes / 10 * 9
	for _, f := range files {
		if total <= target {
			break
		}
		if os.Remove(f.path) == nil {
			total -= f.size
		}
	}
	c.mu.Lock()
	c.size = total
	c.mu.Unlock()
}

// Fetch produces a body on a cache miss. store says whether the body may be cached.
type Fetch func(ctx context.Context) (body []byte, store bool, err error)

// Do returns the cached body for key or runs fetch. Concurrent Do calls for the same key share one
// fetch. A fetch that fails because its caller's context ended does not fail the callers that merged
// into it: they try again themselves.
func (c *Cache) Do(ctx context.Context, key string, ttl time.Duration, fetch Fetch) ([]byte, error) {
	for {
		if body, ok := c.Get(key); ok {
			c.hits.Add(1)
			return body, nil
		}
		c.mu.Lock()
		if fl, running := c.flight[key]; running {
			c.mu.Unlock()
			select {
			case <-fl.done:
			case <-ctx.Done():
				return nil, ctx.Err()
			}
			if fl.err != nil && (errors.Is(fl.err, context.Canceled) || errors.Is(fl.err, context.DeadlineExceeded)) && ctx.Err() == nil {
				continue
			}
			return fl.body, fl.err
		}
		fl := &call{done: make(chan struct{})}
		c.flight[key] = fl
		c.mu.Unlock()

		c.misses.Add(1)
		body, store, err := fetch(ctx)
		if err == nil && store {
			c.Put(key, body, ttl)
		}
		fl.body, fl.err = body, err
		c.mu.Lock()
		delete(c.flight, key)
		c.mu.Unlock()
		close(fl.done)
		return body, err
	}
}

// Clear deletes every cache entry of this program (all format versions under the cache directory)
// and returns how many files were removed. Entries other processes write meanwhile are simply new
// entries; nothing here needs a lock with them.
func (c *Cache) Clear() (int, error) {
	base := filepath.Dir(c.root)
	dirs, err := os.ReadDir(base)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	removed := 0
	var firstErr error
	for _, d := range dirs {
		if !d.IsDir() || !isVersionDir(d.Name()) {
			continue
		}
		dir := filepath.Join(base, d.Name())
		files, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, f := range files {
			if f.IsDir() || !(strings.HasSuffix(f.Name(), entrySuffix) || strings.HasPrefix(f.Name(), tmpPrefix)) {
				continue
			}
			if err := os.Remove(filepath.Join(dir, f.Name())); err != nil {
				if firstErr == nil && !os.IsNotExist(err) {
					firstErr = err
				}
				continue
			}
			removed++
		}
		_ = os.Remove(dir) // only succeeds when empty
	}
	c.mu.Lock()
	c.size, c.sized = 0, false
	c.mu.Unlock()
	return removed, firstErr
}

func isVersionDir(name string) bool {
	if len(name) < 2 || name[0] != 'v' {
		return false
	}
	for _, r := range name[1:] {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
