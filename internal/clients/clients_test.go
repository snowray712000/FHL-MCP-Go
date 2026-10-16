package clients

import (
	"context"
	"os/exec"
	"strings"
	"testing"
)

// fakeCLI emulates "claude mcp" / "codex mcp" with in-memory state.
type fakeCLI struct {
	kind      string // claude | codex
	installed bool
	entry     *Registration // what is registered under ServerName
	calls     [][]string
	failAdd   bool
	noPersist bool // add "succeeds" but nothing is stored
}

func (f *fakeCLI) LookPath(name string) (string, error) {
	if !f.installed {
		return "", exec.ErrNotFound
	}
	return "/bin/" + name, nil
}

func (f *fakeCLI) Run(_ context.Context, _ string, args ...string) (string, string, int, error) {
	f.calls = append(f.calls, args)
	switch args[1] {
	case "get":
		if f.entry == nil {
			return "", `No MCP server named "fhl". Configured servers: none`, 1, nil
		}
		if f.kind == Claude {
			return "fhl:\n  Scope: " + f.entry.Scope + "\n  Status: ✔ Connected\n  Type: http\n  URL: " + f.entry.URL + "\n", "", 0, nil
		}
		return `{"name":"fhl","enabled":true,"transport":{"type":"streamable_http","url":"` + f.entry.URL + `"}}`, "", 0, nil
	case "add":
		if f.failAdd {
			return "", "boom", 1, nil
		}
		if !f.noPersist {
			u := args[len(args)-1]
			f.entry = &Registration{Found: true, URL: u, Scope: "User config (available in all your projects)"}
		}
		return "added", "", 0, nil
	case "remove":
		f.entry = nil
		return "removed", "", 0, nil
	}
	return "", "unknown", 2, nil
}

func newClaude(f *fakeCLI) *Client { f.kind = Claude; return New(Claude, f) }
func newCodex(f *fakeCLI) *Client  { f.kind = Codex; return New(Codex, f) }

const mcpURL = "http://127.0.0.1:8081/mcp"

func TestSetupStates(t *testing.T) {
	for _, mk := range []func(*fakeCLI) *Client{newClaude, newCodex} {
		// not installed
		f := &fakeCLI{}
		c := mk(f)
		if o := c.Setup(context.Background(), mcpURL); o.State != CLIMissing || len(f.calls) != 0 {
			t.Errorf("%s missing: %+v calls %v", c.ID, o, f.calls)
		}
		// fresh add, read back
		f = &fakeCLI{installed: true}
		c = mk(f)
		if o := c.Setup(context.Background(), mcpURL); o.State != Added || !o.OK() || f.entry == nil || f.entry.URL != mcpURL {
			t.Errorf("%s add: %+v entry %+v", c.ID, o, f.entry)
		}
		// rerun is idempotent and does not call add again
		n := len(f.calls)
		if o := c.Setup(context.Background(), mcpURL); o.State != Already {
			t.Errorf("%s rerun: %+v", c.ID, o)
		}
		for _, call := range f.calls[n:] {
			if call[1] != "get" {
				t.Errorf("%s rerun made a %q call", c.ID, call[1])
			}
		}
		// same name, other mcpURL: conflict, never overwritten
		f = &fakeCLI{installed: true, entry: &Registration{Found: true, URL: "http://example.com/mcp", Scope: "User config"}}
		c = mk(f)
		o := c.Setup(context.Background(), mcpURL)
		if o.State != Conflict || o.ExistingURL != "http://example.com/mcp" || f.entry.URL != "http://example.com/mcp" {
			t.Errorf("%s conflict: %+v entry %+v", c.ID, o, f.entry)
		}
		for _, call := range f.calls {
			if call[1] == "add" || call[1] == "remove" {
				t.Errorf("%s conflict still mutated: %v", c.ID, call)
			}
		}
		// add fails
		f = &fakeCLI{installed: true, failAdd: true}
		if o := mk(f).Setup(context.Background(), mcpURL); o.State != Failed || o.Err == nil || !strings.Contains(o.Err.Error(), "boom") {
			t.Errorf("%s failed add: %+v", c.ID, o)
		}
		// add "succeeds" but cannot be read back
		f = &fakeCLI{installed: true, noPersist: true}
		if o := mk(f).Setup(context.Background(), mcpURL); o.State != Failed {
			t.Errorf("%s unverified add: %+v", c.ID, o)
		}
	}
}

func TestSetupClaudeUsesUserScopeHTTP(t *testing.T) {
	f := &fakeCLI{installed: true}
	newClaude(f).Setup(context.Background(), mcpURL)
	var add []string
	for _, c := range f.calls {
		if c[1] == "add" {
			add = c
		}
	}
	if got := strings.Join(add, " "); got != "mcp add --transport http --scope user fhl "+mcpURL {
		t.Errorf("add args = %q", got)
	}
	f = &fakeCLI{installed: true}
	newCodex(f).Setup(context.Background(), mcpURL)
	for _, c := range f.calls {
		if c[1] == "add" && strings.Join(c, " ") != "mcp add fhl --url "+mcpURL {
			t.Errorf("codex add args = %v", c)
		}
	}
}

func TestDisconnect(t *testing.T) {
	ours := &Registration{Found: true, URL: "http://127.0.0.1:9000/mcp", Scope: "User config (available in all your projects)"}
	f := &fakeCLI{installed: true, entry: ours}
	if o := newClaude(f).Disconnect(context.Background()); o.State != Removed || f.entry != nil {
		t.Errorf("remove ours: %+v", o)
	}
	var rm []string
	for _, c := range f.calls {
		if c[1] == "remove" {
			rm = c
		}
	}
	if got := strings.Join(rm, " "); got != "mcp remove fhl --scope user" {
		t.Errorf("remove args = %q", got)
	}

	// Someone else's "fhl" server is left alone.
	f = &fakeCLI{installed: true, entry: &Registration{Found: true, URL: "https://api.example.com/mcp", Scope: "User config"}}
	if o := newClaude(f).Disconnect(context.Background()); o.State != Refused || f.entry == nil {
		t.Errorf("foreign entry: %+v", o)
	}
	// Loopback but not /mcp is not ours either.
	f = &fakeCLI{installed: true, entry: &Registration{Found: true, URL: "http://127.0.0.1:3000/other", Scope: "User config"}}
	if o := newClaude(f).Disconnect(context.Background()); o.State != Refused {
		t.Errorf("loopback non-mcp: %+v", o)
	}
	if o := newCodex(&fakeCLI{installed: true}).Disconnect(context.Background()); o.State != NotPresent || !o.OK() {
		t.Errorf("not registered: %+v", o)
	}
	if o := newCodex(&fakeCLI{}).Disconnect(context.Background()); o.State != CLIMissing {
		t.Errorf("cli missing: %+v", o)
	}
}

func TestParsers(t *testing.T) {
	// A real "claude mcp get fhl" sample.
	out := "fhl:\n  Scope: User config (available in all your projects)\n  Status: ✔ Connected\n  Type: http\n  URL: http://127.0.0.1:8081/mcp\n\nTo remove this server, run: claude mcp remove fhl -s user\n"
	if u, s := parseClaudeGet(out); u != mcpURL || !strings.HasPrefix(s, "User config") {
		t.Errorf("claude parse = %q %q", u, s)
	}
	if u, _ := parseClaudeGet("fhl:\n  Type: stdio\n  Command: x\n"); u != "" {
		t.Errorf("stdio entry should have no URL, got %q", u)
	}
	if u := parseCodexGet(`{"name":"fhl","transport":{"type":"streamable_http","url":"` + mcpURL + `"}}`); u != mcpURL {
		t.Errorf("codex json = %q", u)
	}
	if u := parseCodexGet("fhl\n  enabled: true\n  transport: streamable_http\n  url: " + mcpURL + "\n"); u != mcpURL {
		t.Errorf("codex text = %q", u)
	}
}

func TestURLHelpers(t *testing.T) {
	if !SameURL("http://127.0.0.1:8081/mcp/", "HTTP://127.0.0.1:8081/mcp") || SameURL(mcpURL, "http://localhost:8081/mcp") || SameURL(mcpURL, "http://127.0.0.1:8082/mcp") {
		t.Error("SameURL")
	}
	for u, want := range map[string]bool{
		mcpURL: true, "http://localhost:1/mcp": true, "http://[::1]:5/mcp": true,
		"https://127.0.0.1/mcp": false, "http://example.com/mcp": false, "http://127.0.0.1:1/x": false, "": false,
	} {
		if got := LooksLikeThisServer(u); got != want {
			t.Errorf("LooksLikeThisServer(%q) = %v", u, got)
		}
	}
}

func TestGetOtherErrorsAreNotTreatedAsNotRegistered(t *testing.T) {
	c := New(Claude, errRunner{})
	if _, err := c.Get(context.Background(), "claude"); err == nil {
		t.Error("a CLI that crashes must be an error, not 'not registered'")
	}
}

type errRunner struct{}

func (errRunner) LookPath(string) (string, error) { return "claude", nil }
func (errRunner) Run(context.Context, string, ...string) (string, string, int, error) {
	return "", "Error: config file is corrupt", 1, nil
}
