package app

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/snowray712000/FHL-MCP-Go/assets"
	"github.com/snowray712000/FHL-MCP-Go/internal/fhlapi"
	"github.com/snowray712000/FHL-MCP-Go/internal/version"
)

func fixtureFile(t *testing.T, parts ...string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(append([]string{"..", "..", "tests", "fixtures"}, parts...)...))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// fakeUpstream serves the recorded qb.php fixtures.
func fakeUpstream(t *testing.T) *fhlapi.Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		name := "qb_john_3_16.json"
		switch {
		case q.Get("chap") == "99":
			name = "qb_john_99_1_empty.json"
		case q.Get("sec") == "16-17":
			name = "qb_john_3_16_17_strong.json"
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write(fixtureFile(t, "upstream", name))
	}))
	t.Cleanup(srv.Close)
	return &fhlapi.Client{BaseURL: srv.URL + "/api", HTTP: srv.Client(), UserAgent: "test"}
}

// startServer runs the real server on a random loopback port and returns its base URL.
func startServer(t *testing.T) (base string, stop func() error) {
	t.Helper()
	cfg := Config{Host: "127.0.0.1", Port: 0, API: fakeUpstream(t)}
	ln, err := Listen(cfg)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- Serve(ctx, cfg, ln, io.Discard) }()
	t.Cleanup(cancel)
	return "http://" + ln.Addr().String(), func() error {
		cancel()
		select {
		case err := <-done:
			return err
		case <-time.After(10 * time.Second):
			t.Fatal("server did not shut down")
			return nil
		}
	}
}

func connect(t *testing.T, base string) *mcp.ClientSession {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	c := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1"}, nil)
	sess, err := c.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: base + "/mcp"}, nil)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() { _ = sess.Close() })
	return sess
}

func textOf(t *testing.T, res *mcp.CallToolResult) string {
	t.Helper()
	if len(res.Content) != 1 {
		t.Fatalf("content length = %d", len(res.Content))
	}
	tc, ok := res.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("content type %T", res.Content[0])
	}
	return tc.Text
}

func TestInitializeAdvertisesToolsOnly(t *testing.T) {
	base, _ := startServer(t)
	sess := connect(t, base)
	init := sess.InitializeResult()
	if init.ServerInfo.Name != version.Name || init.ServerInfo.Version != version.Version {
		t.Errorf("serverInfo = %+v", init.ServerInfo)
	}
	if init.ProtocolVersion != "2025-11-25" {
		t.Errorf("protocol = %s", init.ProtocolVersion)
	}
	caps := init.Capabilities
	if caps.Tools == nil || caps.Prompts != nil || caps.Resources != nil || caps.Logging != nil {
		t.Errorf("capabilities = %+v (first release advertises tools only)", caps)
	}
}

func TestToolsListEqualsFrozenContract(t *testing.T) {
	base, _ := startServer(t)
	sess := connect(t, base)
	res, err := sess.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	var want struct {
		Tools []json.RawMessage `json:"tools"`
	}
	if err := json.Unmarshal(assets.ToolsListJSON, &want); err != nil {
		t.Fatal(err)
	}
	if len(res.Tools) != 29 || len(want.Tools) != 29 {
		t.Fatalf("tools: got %d, contract %d, want 29", len(res.Tools), len(want.Tools))
	}
	byName := map[string]any{}
	for _, tool := range res.Tools {
		b, _ := json.Marshal(tool)
		var v any
		_ = json.Unmarshal(b, &v)
		byName[tool.Name] = v
	}
	for _, raw := range want.Tools {
		var v map[string]any
		_ = json.Unmarshal(raw, &v)
		name := v["name"].(string)
		got, ok := byName[name]
		if !ok {
			t.Errorf("tool %s missing", name)
			continue
		}
		gb, _ := json.Marshal(got)
		wb, _ := json.Marshal(v)
		if !bytes.Equal(gb, wb) {
			t.Errorf("tool %s differs from contract:\n got %s\nwant %s", name, gb, wb)
		}
	}
}

func TestGetBibleVerseEndToEnd(t *testing.T) {
	base, _ := startServer(t)
	sess := connect(t, base)
	res, err := sess.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "get_bible_verse",
		Arguments: map[string]any{"book": "John", "chapter": 3, "verse": "16"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatalf("unexpected error: %s", textOf(t, res))
	}
	want := string(fixtureFile(t, "golden", "get_bible_verse_john316.json"))
	if got := textOf(t, res); got != want {
		t.Errorf("text differs from Python golden:\n got %s\nwant %s", got, want)
	}
	sc, ok := res.StructuredContent.(map[string]any)
	if !ok || sc["result"] != want || len(sc) != 1 {
		t.Errorf("structuredContent = %#v, want {result: <same text>}", res.StructuredContent)
	}
}

func TestToolErrorsAreIsErrorResults(t *testing.T) {
	base, _ := startServer(t)
	sess := connect(t, base)
	cases := []struct {
		name  string
		tool  string
		args  map[string]any
		wants string
	}{
		{"unknown tool (D2)", "no_such_tool", map[string]any{}, "Error executing tool no_such_tool: Unknown tool: no_such_tool"},
		{"not implemented", "list_bible_versions", map[string]any{}, "Error executing tool list_bible_versions: Tool 'list_bible_versions' is not implemented yet in this build"},
		{"missing book", "get_bible_verse", map[string]any{"chapter": 3}, "Error executing tool get_bible_verse: Invalid parameter 'book': required parameter is missing"},
		{"unknown book (D3)", "get_bible_verse", map[string]any{"book": "不存在書卷", "chapter": 1}, "Error executing tool get_bible_verse: Invalid parameter 'book': 不存在書卷 - Book not found. Please check the spelling or use a valid book code."},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res, err := sess.CallTool(context.Background(), &mcp.CallToolParams{Name: c.tool, Arguments: c.args})
			if err != nil {
				t.Fatalf("expected an isError result, got protocol error: %v", err)
			}
			if !res.IsError || textOf(t, res) != c.wants {
				t.Errorf("isError=%v text=%q want %q", res.IsError, textOf(t, res), c.wants)
			}
			if res.StructuredContent != nil {
				t.Errorf("error result must not carry structuredContent: %#v", res.StructuredContent)
			}
		})
	}
}

func TestHostAndOriginProtection(t *testing.T) {
	base, _ := startServer(t)
	initBody := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"t","version":"1"}}}`
	post := func(host, origin string) int {
		req, _ := http.NewRequest(http.MethodPost, base+"/mcp", strings.NewReader(initBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json, text/event-stream")
		if host != "" {
			req.Host = host
		}
		if origin != "" {
			req.Header.Set("Origin", origin)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		_, _ = io.Copy(io.Discard, resp.Body)
		return resp.StatusCode
	}
	if c := post("", ""); c != 200 {
		t.Errorf("plain local request = %d, want 200", c)
	}
	if c := post("localhost:8081", ""); c != 200 {
		t.Errorf("Host localhost = %d, want 200", c)
	}
	if c := post("evil.example.com", ""); c != 403 {
		t.Errorf("foreign Host = %d, want 403", c)
	}
	if c := post("", "http://evil.example.com"); c != 403 {
		t.Errorf("foreign Origin = %d, want 403", c)
	}
}

func TestHealthAndRoot(t *testing.T) {
	base, _ := startServer(t)
	resp, err := http.Get(base + "/health")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var h map[string]string
	_ = json.NewDecoder(resp.Body).Decode(&h)
	if resp.StatusCode != 200 || h["status"] != "ok" || h["version"] != version.Version {
		t.Errorf("health = %d %v", resp.StatusCode, h)
	}
	resp2, err := http.Get(base + "/no-such-path")
	if err != nil {
		t.Fatal(err)
	}
	resp2.Body.Close()
	if resp2.StatusCode != 404 {
		t.Errorf("unknown path = %d, want 404 (no root-path MCP compat)", resp2.StatusCode)
	}
}

func TestListenRules(t *testing.T) {
	if _, err := Listen(Config{Host: "0.0.0.0", Port: 0}); err == nil {
		t.Error("non-loopback host must be rejected")
	}
	if _, err := Listen(Config{Host: "192.168.1.5", Port: 0}); err == nil {
		t.Error("LAN host must be rejected")
	}
	ln, err := Listen(Config{Host: "127.0.0.1", Port: 0})
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	port := ln.Addr().(*net.TCPAddr).Port
	if _, err := Listen(Config{Host: "127.0.0.1", Port: port}); err == nil {
		t.Error("second listen on a taken port must fail, not pick another port")
	}
}

func TestGracefulShutdown(t *testing.T) {
	base, stop := startServer(t)
	connect(t, base)
	if err := stop(); err != nil {
		t.Fatalf("shutdown error: %v", err)
	}
}
