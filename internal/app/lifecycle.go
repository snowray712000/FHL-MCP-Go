package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/snowray712000/FHL-MCP-Go/internal/runstate"
	"github.com/snowray712000/FHL-MCP-Go/internal/transport"
	"github.com/snowray712000/FHL-MCP-Go/internal/version"
)

// ResolveDataDir returns the absolute data directory: --data-dir, else FHL_DATA_DIR, else
// <os.UserCacheDir>/fhl-mcp. The cache lives in <data>/cache and the run state in <data>/run.
func ResolveDataDir(dataDir string) (string, error) {
	if dataDir == "" {
		dataDir = os.Getenv(EnvDataDir)
	}
	if dataDir == "" {
		base, err := os.UserCacheDir()
		if err != nil {
			return "", fmt.Errorf("cannot find the user cache directory (use --data-dir): %w", err)
		}
		dataDir = filepath.Join(base, "fhl-mcp")
	}
	return filepath.Abs(dataDir)
}

// RunConfig turns on the state file and control endpoints for Serve.
type RunConfig struct {
	DataDir string // resolved data directory
}

type instance struct {
	statePath string
	id        string
	control   transport.Control
}

func (rc *RunConfig) start(ln net.Listener, url string, stop func()) (*instance, error) {
	id, err := runstate.NewRandom(8)
	if err != nil {
		return nil, err
	}
	token, err := runstate.NewRandom(32)
	if err != nil {
		return nil, err
	}
	addr := ln.Addr().(*net.TCPAddr)
	now := time.Now().UTC()
	st := runstate.State{
		Instance: id, PID: os.Getpid(), Host: addr.IP.String(), Port: addr.Port, URL: url,
		Version: version.Version, Token: token, StartedAt: now,
	}
	path := runstate.Path(rc.DataDir, addr.Port)
	if err := runstate.Write(path, st); err != nil {
		return nil, fmt.Errorf("cannot write the run state file (status/stop need it): %w", err)
	}
	return &instance{
		statePath: path, id: id,
		control: transport.Control{
			Token: token, Stop: stop,
			Info: transport.Info{Instance: id, PID: st.PID, Name: version.Name, Version: version.Version, URL: url, StartedAt: now},
		},
	}, nil
}

func (i *instance) close(log *slog.Logger) {
	if err := runstate.Remove(i.statePath, i.id); err != nil {
		log.Warn("cannot remove the run state file", "path", i.statePath, "err", err)
	}
}

// Phase is what Probe concludes about a port.
type Phase string

const (
	// PhaseRunning: our server, and the state file's instance matches it (controllable).
	PhaseRunning Phase = "running"
	// PhaseRunningUnverified: something answers /health as this product, but there is no matching state
	// (other --data-dir, an older build, or a damaged state file), so it cannot be verified or stopped here.
	PhaseRunningUnverified Phase = "running_unverified"
	// PhaseStopped: nothing listens on the port.
	PhaseStopped Phase = "stopped"
	// PhasePortInUse: something that is not this product holds the port.
	PhasePortInUse Phase = "port_in_use"
)

// Status is the result of Probe.
type Status struct {
	Phase     Phase      `json:"state"`
	URL       string     `json:"url"`
	Version   string     `json:"version,omitempty"`
	PID       int        `json:"pid,omitempty"`
	Instance  string     `json:"instance,omitempty"`
	StartedAt *time.Time `json:"started_at,omitempty"`
	StatePath string     `json:"state_file"`
	// StaleState is true when a state file exists but its server is gone.
	StaleState bool   `json:"stale_state,omitempty"`
	Detail     string `json:"detail,omitempty"`

	state *runstate.State
}

var probeClient = &http.Client{Timeout: 3 * time.Second}

// Probe inspects host:port. It identifies the process by its /health reply and, when a state file
// exists, by the authenticated /_control/info instance id; an answering port alone is never enough.
func Probe(ctx context.Context, dataDir, host string, port int) Status {
	statePath := runstate.Path(dataDir, port)
	base := "http://" + net.JoinHostPort(host, strconv.Itoa(port))
	s := Status{Phase: PhaseStopped, URL: base + "/mcp", StatePath: statePath}

	// Health first, state second: Serve writes the state file before it starts answering, so a server
	// that answers always has its state file already.
	ours, reachable, detail := healthIsOurs(ctx, base)
	st, err := runstate.Read(statePath)
	if err != nil {
		s.Detail = err.Error()
	}
	s.state = st

	switch {
	case ours:
		s.Phase = PhaseRunningUnverified
		if detail != "" {
			s.Version = detail
		}
		if st != nil {
			if info, ok := fetchInfo(ctx, base, st.Token); ok && info.Instance == st.Instance {
				s.Phase = PhaseRunning
				s.Version, s.PID, s.Instance = info.Version, info.PID, info.Instance
				t := info.StartedAt
				s.StartedAt = &t
				s.URL = info.URL
				return s
			}
			s.StaleState = true
			s.Detail = "state file does not belong to the server on this port"
		}
	case reachable:
		s.Phase = PhasePortInUse
		s.Detail = detail
		s.StaleState = st != nil
	default:
		s.StaleState = st != nil
	}
	return s
}

// healthIsOurs reports whether GET /health answers as this product. reachable means something accepted
// the connection; detail is the reported version (ours) or a short reason (not ours).
func healthIsOurs(ctx context.Context, base string) (ours, reachable bool, detail string) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, base+"/health", nil)
	resp, err := probeClient.Do(req)
	if err != nil {
		// A refused connection means nothing listens; any other failure (timeout, garbage) means something does.
		if isRefused(err) {
			return false, false, ""
		}
		return false, true, "the port is held by a program that does not speak HTTP normally (" + err.Error() + ")"
	}
	defer resp.Body.Close()
	var h struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<16)).Decode(&h); err != nil || h.Name != version.Name {
		return false, true, "the port is held by another program"
	}
	return true, true, h.Version
}

func isRefused(err error) bool {
	var op *net.OpError
	if errors.As(err, &op) && op.Op == "dial" {
		// Timeouts on dial are not "refused".
		var to interface{ Timeout() bool }
		if errors.As(err, &to) && to.Timeout() {
			return false
		}
		return true
	}
	return false
}

func fetchInfo(ctx context.Context, base, token string) (transport.Info, bool) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, base+"/_control/info", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := probeClient.Do(req)
	if err != nil {
		return transport.Info{}, false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return transport.Info{}, false
	}
	var info transport.Info
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<16)).Decode(&info); err != nil {
		return transport.Info{}, false
	}
	return info, true
}

// Stop asks a verified (PhaseRunning) server to shut down and waits until it is gone. It never signals a PID.
func Stop(ctx context.Context, dataDir, host string, port int, wait time.Duration) error {
	s := Probe(ctx, dataDir, host, port)
	if s.Phase != PhaseRunning || s.state == nil {
		return fmt.Errorf("no verified server to stop (state: %s)", s.Phase)
	}
	base := "http://" + net.JoinHostPort(host, strconv.Itoa(port))
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, base+"/_control/stop", nil)
	req.Header.Set("Authorization", "Bearer "+s.state.Token)
	resp, err := probeClient.Do(req)
	if err != nil {
		return fmt.Errorf("stop request failed: %w", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("the server refused the stop request (HTTP %d)", resp.StatusCode)
	}
	deadline := time.Now().Add(wait)
	for time.Now().Before(deadline) {
		if _, reachable, _ := healthIsOurs(ctx, base); !reachable {
			_ = runstate.Remove(s.StatePath, s.state.Instance)
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}
	}
	return fmt.Errorf("the server accepted the stop request but is still answering after %s", wait)
}

// ClearStale removes a leftover state file whose server is gone.
func ClearStale(s Status) error {
	if !s.StaleState {
		return nil
	}
	if err := os.Remove(s.StatePath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

// HostPort normalizes a --host/--port pair for display.
func HostPort(host string, port int) string { return net.JoinHostPort(host, strconv.Itoa(port)) }

// MCPURL is the endpoint clients register.
func MCPURL(host string, port int) string {
	return "http://" + strings.TrimSpace(HostPort(host, port)) + "/mcp"
}
