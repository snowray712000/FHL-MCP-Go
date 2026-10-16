package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/snowray712000/FHL-MCP-Go/internal/app"
	"github.com/snowray712000/FHL-MCP-Go/internal/fhlapi"
	"github.com/snowray712000/FHL-MCP-Go/internal/runstate"
)

// startServer runs a real server (state file + control endpoints) on a random port.
func startServer(t *testing.T) (dataDir string, port int, done <-chan error) {
	t.Helper()
	dataDir = t.TempDir()
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	t.Cleanup(up.Close)
	cfg := app.Config{Host: "127.0.0.1", Port: 0, API: &fhlapi.Client{BaseURL: up.URL + "/api", HTTP: up.Client(), UserAgent: "t"}, Run: &app.RunConfig{DataDir: dataDir}}
	ln, err := app.Listen(cfg)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	ch := make(chan error, 1)
	go func() { ch <- app.Serve(ctx, cfg, ln, io.Discard) }()
	port = ln.Addr().(*net.TCPAddr).Port
	for deadline := time.Now().Add(5 * time.Second); time.Now().Before(deadline); time.Sleep(10 * time.Millisecond) {
		if st, _ := runstate.Read(runstate.Path(dataDir, port)); st != nil {
			return dataDir, port, ch
		}
	}
	t.Fatal("server did not publish its state")
	return
}

func freePort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	return ln.Addr().(*net.TCPAddr).Port
}

func sh(t *testing.T, args ...string) (code int, out, errOut string) {
	t.Helper()
	var o, e bytes.Buffer
	code = run(args, &o, &e)
	return code, o.String(), e.String()
}

func TestStatusStopCycle(t *testing.T) {
	dd, port, done := startServer(t)
	p := strconv.Itoa(port)

	code, out, _ := sh(t, "status", "--port", p, "--data-dir", dd)
	if code != 0 || !strings.Contains(out, "執行中") || !strings.Contains(out, "http://127.0.0.1:"+p+"/mcp") {
		t.Fatalf("status: %d %q", code, out)
	}
	code, out, _ = sh(t, "status", "--port", p, "--data-dir", dd, "--json")
	var js map[string]any
	if code != 0 || json.Unmarshal([]byte(out), &js) != nil || js["state"] != "running" {
		t.Fatalf("status --json: %d %q", code, out)
	}
	if strings.Contains(out, "token") {
		t.Error("status --json must not expose the control token")
	}

	// Starting again reports the running instance instead of failing.
	var serr string
	code, out, serr = sh(t, "serve", "--port", p, "--data-dir", dd)
	if code != 0 || !strings.Contains(out, "已在執行") {
		t.Errorf("second serve: %d %q %q", code, out, serr)
	}

	// A different data dir cannot stop it.
	code, _, errOut := sh(t, "stop", "--port", p, "--data-dir", t.TempDir())
	if code != 1 || !strings.Contains(errOut, "無法驗證身分") {
		t.Errorf("stop with wrong data dir: %d %q", code, errOut)
	}
	if st := app.Probe(context.Background(), dd, "127.0.0.1", port); st.Phase != app.PhaseRunning {
		t.Fatalf("server should still run: %+v", st)
	}

	code, out, errOut = sh(t, "stop", "--port", p, "--data-dir", dd)
	if code != 0 || !strings.Contains(out, "已停止") {
		t.Fatalf("stop: %d %q %q", code, out, errOut)
	}
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("serve ended with %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("server still running")
	}

	code, out, _ = sh(t, "status", "--port", p, "--data-dir", dd)
	if code != 1 || !strings.Contains(out, "未執行") {
		t.Errorf("status after stop: %d %q", code, out)
	}
	code, out, _ = sh(t, "stop", "--port", p, "--data-dir", dd)
	if code != 0 || !strings.Contains(out, "沒有執行中的服務") {
		t.Errorf("stop when stopped: %d %q", code, out)
	}
}

func TestServeFailsWhenPortHeldByAnotherProgram(t *testing.T) {
	foreign := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer foreign.Close()
	p := strconv.Itoa(foreign.Listener.Addr().(*net.TCPAddr).Port)
	dd := t.TempDir()
	code, _, errOut := sh(t, "serve", "--port", p, "--data-dir", dd)
	if code != 1 || !strings.Contains(errOut, "無法啟動") {
		t.Errorf("serve on a taken port: %d %q", code, errOut)
	}
	code, out, _ := sh(t, "status", "--port", p, "--data-dir", dd)
	if code != 1 || !strings.Contains(out, "被其他程式占用") {
		t.Errorf("status: %d %q", code, out)
	}
	code, _, errOut = sh(t, "stop", "--port", p, "--data-dir", dd)
	if code != 1 || !strings.Contains(errOut, "不是本工具") {
		t.Errorf("stop: %d %q", code, errOut)
	}
}

func TestStopClearsStaleState(t *testing.T) {
	port := freePort(t)
	dd := t.TempDir()
	path := runstate.Path(dd, port)
	_ = runstate.Write(path, runstate.State{Instance: "dead", Token: "t", Port: port})
	code, out, _ := sh(t, "stop", "--port", strconv.Itoa(port), "--data-dir", dd)
	if code != 0 || !strings.Contains(out, "沒有執行中的服務") {
		t.Errorf("stop: %d %q", code, out)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("stale state file left behind")
	}
}

func TestLifecycleArgumentErrors(t *testing.T) {
	for _, args := range [][]string{
		{"status", "--host", "0.0.0.0"},
		{"stop", "--port", "0"},
		{"setup", "--client", "vim"},
		{"disconnect", "--client", "vim"},
		{"status", "--bogus"},
	} {
		if code, _, _ := sh(t, args...); code != 2 {
			t.Errorf("%v: exit %d, want 2", args, code)
		}
	}
}

// fakeCLI emulates the claude / codex CLIs (one registered entry each).
type fakeCLI struct {
	installed map[string]bool
	entries   map[string]string // binary -> registered URL
	adds      map[string]int
}

func newFake(installed ...string) *fakeCLI {
	f := &fakeCLI{installed: map[string]bool{}, entries: map[string]string{}, adds: map[string]int{}}
	for _, b := range installed {
		f.installed[b] = true
	}
	return f
}

func (f *fakeCLI) LookPath(name string) (string, error) {
	if !f.installed[name] {
		return "", exec.ErrNotFound
	}
	return name, nil
}

func (f *fakeCLI) Run(_ context.Context, bin string, args ...string) (string, string, int, error) {
	switch args[1] {
	case "get":
		u, ok := f.entries[bin]
		if !ok {
			return "", `No MCP server named "fhl".`, 1, nil
		}
		if bin == "claude" {
			return "fhl:\n  Scope: User config (available in all your projects)\n  Type: http\n  URL: " + u + "\n", "", 0, nil
		}
		return `{"transport":{"type":"streamable_http","url":"` + u + `"}}`, "", 0, nil
	case "add":
		f.adds[bin]++
		f.entries[bin] = args[len(args)-1]
		return "", "", 0, nil
	case "remove":
		delete(f.entries, bin)
		return "", "", 0, nil
	}
	return "", "", 2, nil
}

func useFake(t *testing.T, f *fakeCLI) {
	t.Helper()
	old := clientRunner
	clientRunner = f
	t.Cleanup(func() { clientRunner = old })
}

func TestSetupAndDisconnect(t *testing.T) {
	dd, port, _ := startServer(t)
	p := strconv.Itoa(port)
	url := "http://127.0.0.1:" + p + "/mcp"
	f := newFake("claude", "codex")
	useFake(t, f)

	code, out, errOut := sh(t, "setup", "--client", "all", "--port", p, "--data-dir", dd)
	if code != 0 || f.entries["claude"] != url || f.entries["codex"] != url {
		t.Fatalf("setup: %d out=%q err=%q entries=%v", code, out, errOut, f.entries)
	}
	if !strings.Contains(out, "連線測試通過") || !strings.Contains(out, "29 個工具") {
		t.Errorf("setup should smoke-test the running server: %q", out)
	}

	// Re-running changes nothing.
	code, out, _ = sh(t, "setup", "--port", p, "--data-dir", dd) // not a terminal in tests -> all
	if code != 0 || f.adds["claude"] != 1 || f.adds["codex"] != 1 || !strings.Contains(out, "不需變更") {
		t.Errorf("rerun: %d adds=%v %q", code, f.adds, out)
	}

	// Disconnect removes only this tool's entries.
	code, out, _ = sh(t, "disconnect", "--client", "claude", "--port", p)
	if code != 0 || len(f.entries) != 1 || f.entries["codex"] == "" || !strings.Contains(out, "已移除") {
		t.Errorf("disconnect claude: %d %v %q", code, f.entries, out)
	}
	f.entries["codex"] = "https://example.com/mcp"
	code, out, _ = sh(t, "disconnect", "--port", p)
	if code != 1 || f.entries["codex"] != "https://example.com/mcp" || !strings.Contains(out, "沒有移除") {
		t.Errorf("disconnect foreign: %d %v %q", code, f.entries, out)
	}
}

func TestSetupReportsEachClientAndExitCodes(t *testing.T) {
	port := freePort(t)
	p := strconv.Itoa(port)
	dd := t.TempDir()

	// Only claude installed: "all" succeeds (codex missing is informational, with the manual command).
	f := newFake("claude")
	useFake(t, f)
	code, out, _ := sh(t, "setup", "--client", "all", "--port", p, "--data-dir", dd)
	if code != 0 || !strings.Contains(out, "codex mcp add fhl --url http://127.0.0.1:"+p+"/mcp") || !strings.Contains(out, "服務目前沒有執行") {
		t.Errorf("claude only: %d %q", code, out)
	}
	// Explicitly asking for the missing client is a failure.
	if code, _, _ := sh(t, "setup", "--client", "codex", "--port", p, "--data-dir", dd); code != 1 {
		t.Errorf("explicit missing client: exit %d, want 1", code)
	}
	// Nothing installed at all.
	useFake(t, newFake())
	if code, _, _ := sh(t, "setup", "--port", p, "--data-dir", dd); code != 1 {
		t.Errorf("no clients: exit %d, want 1", code)
	}
	// Conflict is reported, not overwritten; the other client is still handled.
	f = newFake("claude", "codex")
	f.entries["claude"] = "http://127.0.0.1:9/mcp"
	useFake(t, f)
	code, out, _ = sh(t, "setup", "--port", p, "--data-dir", dd)
	if code != 1 || f.entries["claude"] != "http://127.0.0.1:9/mcp" || f.entries["codex"] == "" || !strings.Contains(out, "網址不同") {
		t.Errorf("conflict: %d %v %q", code, f.entries, out)
	}
}

func TestSetupAsksWhenInteractive(t *testing.T) {
	f := newFake("claude", "codex")
	useFake(t, f)
	oldIn, oldTTY := stdinReader, stdinIsTerminal
	stdinReader, stdinIsTerminal = strings.NewReader("2\n"), func() bool { return true }
	defer func() { stdinReader, stdinIsTerminal = oldIn, oldTTY }()

	port := freePort(t)
	code, out, _ := sh(t, "setup", "--port", strconv.Itoa(port), "--data-dir", t.TempDir())
	if code != 0 || f.entries["codex"] == "" || f.entries["claude"] != "" || !strings.Contains(out, "要連接哪個 AI 工具") {
		t.Errorf("interactive: %d %v %q", code, f.entries, out)
	}
}

func TestDoctor(t *testing.T) {
	dd, port, _ := startServer(t)
	p := strconv.Itoa(port)
	url := "http://127.0.0.1:" + p + "/mcp"
	f := newFake("claude")
	f.entries["claude"] = url
	useFake(t, f)

	code, out, _ := sh(t, "doctor", "--offline", "--port", p, "--data-dir", dd)
	for _, want := range []string{"[環境]", "[設定]", "[服務]", "[信望愛網站 (上游)]", "[客戶端登錄]",
		"MCP 連線測試通過", "Claude Code：已登錄", "Codex：找不到 codex", "已指定 --offline"} {
		if !strings.Contains(out, want) {
			t.Errorf("doctor output lacks %q:\n%s", want, out)
		}
	}
	if code != 0 {
		t.Errorf("exit %d, want 0 (warnings only):\n%s", code, out)
	}

	// A registration pointing at another port is a problem with a way out.
	f.entries["claude"] = "http://127.0.0.1:9999/mcp"
	code, out, _ = sh(t, "doctor", "--offline", "--port", p, "--data-dir", dd)
	if code != 1 || !strings.Contains(out, "不一致") || !strings.Contains(out, "disconnect") {
		t.Errorf("mismatch: %d\n%s", code, out)
	}
}

func TestDoctorPortHeldByAnotherProgram(t *testing.T) {
	foreign := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer foreign.Close()
	useFake(t, newFake())
	p := strconv.Itoa(foreign.Listener.Addr().(*net.TCPAddr).Port)
	code, out, _ := sh(t, "doctor", "--offline", "--port", p, "--data-dir", t.TempDir())
	if code != 1 || !strings.Contains(out, "埠被其他程式占用") {
		t.Errorf("doctor: %d\n%s", code, out)
	}
}
