// Package runstate keeps the per-instance state file that lets "status" and "stop" find and authenticate
// a running "serve": its URL, a random instance id and a high-entropy control token. The file lives only
// while the server runs (removed on shutdown; a leftover one is detected as stale by probing the server).
package runstate

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// State is the content of the state file.
type State struct {
	Instance  string    `json:"instance"` // random id, also reported by the control endpoint
	PID       int       `json:"pid"`      // informational only: stop never signals a PID
	Host      string    `json:"host"`
	Port      int       `json:"port"`
	URL       string    `json:"url"` // MCP endpoint
	Version   string    `json:"version"`
	Token     string    `json:"token"` // bearer token for /_control/*
	StartedAt time.Time `json:"started_at"`
}

// Path returns the state file for a server on the given port inside the data directory.
func Path(dataDir string, port int) string {
	return filepath.Join(dataDir, "run", fmt.Sprintf("state-%d.json", port))
}

// NewRandom returns n random bytes, hex-encoded.
func NewRandom(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// Write stores s atomically (temp file + rename) with owner-only permissions where the OS supports them.
func Write(path string, s State) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".state-*.tmp")
	if err != nil {
		return err
	}
	name := tmp.Name()
	_ = tmp.Chmod(0o600)
	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		os.Remove(name)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(name)
		return err
	}
	if err := os.Rename(name, path); err != nil {
		os.Remove(name)
		return err
	}
	return nil
}

// Read loads the state file. A missing file returns (nil, nil); a corrupt one returns an error.
func Read(path string) (*State, error) {
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var s State
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, fmt.Errorf("state file %s is corrupt: %w", path, err)
	}
	if s.Instance == "" || s.Token == "" {
		return nil, fmt.Errorf("state file %s is incomplete", path)
	}
	return &s, nil
}

// Remove deletes the state file, but only if it still belongs to instance (a newer server on the same
// port may have replaced it). A missing file is not an error.
func Remove(path, instance string) error {
	s, err := Read(path)
	if err != nil {
		// A corrupt file is nobody's state; clear it.
		return os.Remove(path)
	}
	if s == nil || s.Instance != instance {
		return nil
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}
