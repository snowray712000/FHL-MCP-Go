// Package clients registers this server with MCP clients (Claude Code, Codex) through their official CLIs.
//
// It never edits a client's config file itself: registration, lookup and removal all go through
// "<cli> mcp add|get|remove". The CLI is reached through a Runner so tests can fake it.
package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// ServerName is the name the server is registered under in every client.
const ServerName = "fhl"

// Runner executes client CLIs.
type Runner interface {
	// LookPath finds the CLI executable (PATH, PATHEXT on Windows).
	LookPath(name string) (string, error)
	// Run executes path with args. exit is the process exit code; err is non-nil only when the process
	// could not be run at all (not for a non-zero exit).
	Run(ctx context.Context, path string, args ...string) (stdout, stderr string, exit int, err error)
}

// ExecRunner is the production Runner.
type ExecRunner struct{}

func (ExecRunner) LookPath(name string) (string, error) { return exec.LookPath(name) }

func (ExecRunner) Run(ctx context.Context, path string, args ...string) (string, string, int, error) {
	cmd := exec.CommandContext(ctx, path, args...)
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	err := cmd.Run()
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return out.String(), errb.String(), ee.ExitCode(), nil
	}
	return out.String(), errb.String(), 0, err
}

// IDs of the supported clients.
const (
	Claude = "claude"
	Codex  = "codex"
)

// All lists the supported client IDs in display order.
var All = []string{Claude, Codex}

// Client is one MCP client reachable through its CLI.
type Client struct {
	ID     string // Claude or Codex
	Title  string // display name
	Binary string // CLI executable name
	R      Runner
}

// New returns the client with the given ID, or nil.
func New(id string, r Runner) *Client {
	switch id {
	case Claude:
		return &Client{ID: Claude, Title: "Claude Code", Binary: "claude", R: r}
	case Codex:
		return &Client{ID: Codex, Title: "Codex", Binary: "codex", R: r}
	}
	return nil
}

// Registration is what the client currently has under ServerName.
type Registration struct {
	Found bool
	URL   string // "" when the entry is not an HTTP server
	Scope string // as reported by the client ("" if unknown)
}

// Manual returns the exact command a user can run themselves.
func (c *Client) Manual(mcpURL string) string {
	if c.ID == Claude {
		return fmt.Sprintf("claude mcp add --transport http --scope user %s %s", ServerName, mcpURL)
	}
	return fmt.Sprintf("codex mcp add %s --url %s", ServerName, mcpURL)
}

// Detect locates the CLI.
func (c *Client) Detect() (string, error) { return c.R.LookPath(c.Binary) }

var notFoundRe = regexp.MustCompile(`(?i)no mcp server|not found|does not exist|unknown server`)

// Get reads the current registration. path is the CLI path from Detect.
func (c *Client) Get(ctx context.Context, path string) (Registration, error) {
	args := []string{"mcp", "get", ServerName}
	if c.ID == Codex {
		args = append(args, "--json")
	}
	out, errOut, exit, err := c.R.Run(ctx, path, args...)
	if err != nil {
		return Registration{}, fmt.Errorf("cannot run %s: %w", c.Binary, err)
	}
	if exit != 0 {
		if notFoundRe.MatchString(out + "\n" + errOut) {
			return Registration{}, nil
		}
		return Registration{}, fmt.Errorf("%s mcp get exited with %d: %s", c.Binary, exit, strings.TrimSpace(firstNonEmpty(errOut, out)))
	}
	reg := Registration{Found: true}
	if c.ID == Claude {
		reg.URL, reg.Scope = parseClaudeGet(out)
	} else {
		reg.URL = parseCodexGet(out)
	}
	return reg, nil
}

var (
	claudeURLRe   = regexp.MustCompile(`(?m)^\s*URL:\s*(\S+)\s*$`)
	claudeScopeRe = regexp.MustCompile(`(?m)^\s*Scope:\s*(.+?)\s*$`)
	anyURLRe      = regexp.MustCompile(`https?://[^\s"',)]+`)
)

func parseClaudeGet(out string) (urlStr, scope string) {
	if m := claudeURLRe.FindStringSubmatch(out); m != nil {
		urlStr = m[1]
	}
	if m := claudeScopeRe.FindStringSubmatch(out); m != nil {
		scope = m[1]
	}
	return
}

// parseCodexGet takes the "url" of "codex mcp get <name> --json" (any nesting), else the first URL in the text.
func parseCodexGet(out string) string {
	var v any
	if json.Unmarshal([]byte(out), &v) == nil {
		if u := findURL(v); u != "" {
			return u
		}
		return ""
	}
	return anyURLRe.FindString(out)
}

func findURL(v any) string {
	switch t := v.(type) {
	case map[string]any:
		if s, ok := t["url"].(string); ok && s != "" {
			return s
		}
		for _, x := range t {
			if u := findURL(x); u != "" {
				return u
			}
		}
	case []any:
		for _, x := range t {
			if u := findURL(x); u != "" {
				return u
			}
		}
	}
	return ""
}

func firstNonEmpty(a ...string) string {
	for _, s := range a {
		if strings.TrimSpace(s) != "" {
			return s
		}
	}
	return ""
}

// SameURL compares two endpoint URLs ignoring case of scheme/host and a trailing slash.
func SameURL(a, b string) bool {
	norm := func(s string) string {
		u, err := url.Parse(strings.TrimSpace(s))
		if err != nil {
			return s
		}
		return strings.ToLower(u.Scheme) + "://" + strings.ToLower(u.Host) + strings.TrimRight(u.Path, "/")
	}
	return norm(a) == norm(b)
}

// LooksLikeThisServer reports whether u is a loopback /mcp endpoint, i.e. something this tool could have
// registered (possibly on another port). Disconnect only removes such entries.
func LooksLikeThisServer(u string) bool {
	p, err := url.Parse(u)
	if err != nil || p.Scheme != "http" {
		return false
	}
	h := p.Hostname()
	return (h == "127.0.0.1" || h == "localhost" || h == "::1") && strings.TrimRight(p.Path, "/") == "/mcp"
}

func (c *Client) add(ctx context.Context, path, mcpURL string) error {
	args := []string{"mcp", "add", ServerName, "--url", mcpURL}
	if c.ID == Claude {
		args = []string{"mcp", "add", "--transport", "http", "--scope", "user", ServerName, mcpURL}
	}
	return c.mutate(ctx, path, "add", args)
}

func (c *Client) remove(ctx context.Context, path string, reg Registration) error {
	args := []string{"mcp", "remove", ServerName}
	if c.ID == Claude {
		switch {
		case strings.HasPrefix(strings.ToLower(reg.Scope), "user"):
			args = append(args, "--scope", "user")
		case strings.HasPrefix(strings.ToLower(reg.Scope), "local"):
			args = append(args, "--scope", "local")
		case strings.HasPrefix(strings.ToLower(reg.Scope), "project"):
			args = append(args, "--scope", "project")
		}
	}
	return c.mutate(ctx, path, "remove", args)
}

func (c *Client) mutate(ctx context.Context, path, what string, args []string) error {
	out, errOut, exit, err := c.R.Run(ctx, path, args...)
	if err != nil {
		return fmt.Errorf("cannot run %s: %w", c.Binary, err)
	}
	if exit != 0 {
		return fmt.Errorf("%s mcp %s exited with %d: %s", c.Binary, what, exit, strings.TrimSpace(firstNonEmpty(errOut, out)))
	}
	return nil
}

// State is the result of a setup/disconnect step for one client.
type State string

const (
	CLIMissing State = "cli_missing" // CLI not on PATH; nothing was changed
	Added      State = "added"
	Already    State = "already" // same name and URL were already registered
	Conflict   State = "conflict"
	Removed    State = "removed"
	NotPresent State = "not_registered"
	Refused    State = "refused" // disconnect found an entry that is not ours
	Failed     State = "failed"
)

// Outcome is reported per client, so a failure in one never hides another's result.
type Outcome struct {
	Client      *Client
	State       State
	CLIPath     string
	ExistingURL string // set for Conflict/Refused/Already
	Scope       string
	Err         error
}

// OK reports whether the outcome leaves the client in the wanted state.
func (o Outcome) OK() bool {
	switch o.State {
	case Added, Already, Removed, NotPresent:
		return true
	}
	return false
}

func withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, 90*time.Second)
}

// Setup registers mcpURL as a user-level HTTP server. Same name+URL is success; same name with another
// URL is a conflict that is reported, never overwritten.
func (c *Client) Setup(ctx context.Context, mcpURL string) Outcome {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	path, err := c.Detect()
	if err != nil {
		return Outcome{Client: c, State: CLIMissing}
	}
	o := Outcome{Client: c, CLIPath: path}
	reg, err := c.Get(ctx, path)
	if err != nil {
		o.State, o.Err = Failed, err
		return o
	}
	if reg.Found {
		o.ExistingURL, o.Scope = reg.URL, reg.Scope
		if SameURL(reg.URL, mcpURL) {
			o.State = Already
		} else {
			o.State = Conflict
		}
		return o
	}
	if err := c.add(ctx, path, mcpURL); err != nil {
		o.State, o.Err = Failed, err
		return o
	}
	// Do not trust the exit code alone: read it back.
	got, err := c.Get(ctx, path)
	if err != nil || !got.Found || !SameURL(got.URL, mcpURL) {
		o.State = Failed
		o.Err = fmt.Errorf("%s reported success but the registration cannot be read back (found=%v url=%q err=%v)", c.Binary, got.Found, got.URL, err)
		return o
	}
	o.State, o.Scope = Added, got.Scope
	return o
}

// Disconnect removes this tool's registration only: an entry named ServerName that is not a loopback
// /mcp endpoint belongs to someone else and is left alone (Refused).
func (c *Client) Disconnect(ctx context.Context) Outcome {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	path, err := c.Detect()
	if err != nil {
		return Outcome{Client: c, State: CLIMissing}
	}
	o := Outcome{Client: c, CLIPath: path}
	reg, err := c.Get(ctx, path)
	if err != nil {
		o.State, o.Err = Failed, err
		return o
	}
	if !reg.Found {
		o.State = NotPresent
		return o
	}
	o.ExistingURL, o.Scope = reg.URL, reg.Scope
	if !LooksLikeThisServer(reg.URL) {
		o.State = Refused
		return o
	}
	if err := c.remove(ctx, path, reg); err != nil {
		o.State, o.Err = Failed, err
		return o
	}
	if again, err := c.Get(ctx, path); err == nil && again.Found && again.URL == reg.URL {
		o.State, o.Err = Failed, fmt.Errorf("%s reported success but the entry is still there", c.Binary)
		return o
	}
	o.State = Removed
	return o
}

// Inspect reports the current registration without changing anything (used by doctor).
func (c *Client) Inspect(ctx context.Context) (path string, reg Registration, err error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	path, err = c.Detect()
	if err != nil {
		return "", Registration{}, err
	}
	reg, err = c.Get(ctx, path)
	return path, reg, err
}
