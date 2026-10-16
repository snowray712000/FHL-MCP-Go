package app

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/snowray712000/FHL-MCP-Go/internal/runstate"
	"github.com/snowray712000/FHL-MCP-Go/internal/version"
)

// startRunServer runs a server with the state file and control endpoints on a random port.
func startRunServer(t *testing.T, dataDir string) (port int, done <-chan error) {
	t.Helper()
	cfg := Config{Host: "127.0.0.1", Port: 0, API: fakeUpstream(t), Run: &RunConfig{DataDir: dataDir}}
	ln, err := Listen(cfg)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	ch := make(chan error, 1)
	go func() { ch <- Serve(ctx, cfg, ln, io.Discard) }()
	port = ln.Addr().(*net.TCPAddr).Port
	for deadline := time.Now().Add(5 * time.Second); time.Now().Before(deadline); time.Sleep(10 * time.Millisecond) {
		if st, _ := runstate.Read(runstate.Path(dataDir, port)); st != nil {
			return port, ch
		}
	}
	t.Fatal("server never published its state file")
	return
}

func TestProbeIdentifiesAndStopShutsDown(t *testing.T) {
	dd := t.TempDir()
	port, done := startRunServer(t, dd)
	ctx := context.Background()

	st := Probe(ctx, dd, "127.0.0.1", port)
	if st.Phase != PhaseRunning || st.Version != version.Version || st.PID != os.Getpid() || st.Instance == "" {
		t.Fatalf("probe = %+v", st)
	}
	if _, err := os.Stat(runstate.Path(dd, port)); err != nil {
		t.Fatalf("state file missing: %v", err)
	}

	// Same server, but looked up through another data dir: it answers /health yet cannot be verified.
	other := Probe(ctx, t.TempDir(), "127.0.0.1", port)
	if other.Phase != PhaseRunningUnverified {
		t.Errorf("other data dir: phase = %s, want %s", other.Phase, PhaseRunningUnverified)
	}
	if err := Stop(ctx, t.TempDir(), "127.0.0.1", port, time.Second); err == nil {
		t.Error("Stop with no matching state must refuse")
	}

	if err := Stop(ctx, dd, "127.0.0.1", port, 10*time.Second); err != nil {
		t.Fatalf("Stop: %v", err)
	}
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Serve returned %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("Serve did not return after stop")
	}
	if _, err := os.Stat(runstate.Path(dd, port)); !os.IsNotExist(err) {
		t.Errorf("state file should be removed after stop: %v", err)
	}
	if st := Probe(ctx, dd, "127.0.0.1", port); st.Phase != PhaseStopped || st.StaleState {
		t.Errorf("after stop: %+v", st)
	}
}

func TestControlEndpointsRequireTheToken(t *testing.T) {
	dd := t.TempDir()
	port, _ := startRunServer(t, dd)
	st, _ := runstate.Read(runstate.Path(dd, port))
	base := "http://" + net.JoinHostPort("127.0.0.1", strconv.Itoa(port))

	do := func(method, path, auth string, hdr map[string]string) int {
		req, _ := http.NewRequest(method, base+path, nil)
		if auth != "" {
			req.Header.Set("Authorization", auth)
		}
		for k, v := range hdr {
			if k == "Host" {
				req.Host = v
			} else {
				req.Header.Set(k, v)
			}
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		return resp.StatusCode
	}

	for _, path := range []string{"/_control/info", "/_control/stop"} {
		method := http.MethodGet
		if strings.HasSuffix(path, "stop") {
			method = http.MethodPost
		}
		if c := do(method, path, "", nil); c != http.StatusUnauthorized {
			t.Errorf("%s without token = %d", path, c)
		}
		if c := do(method, path, "Bearer wrong", nil); c != http.StatusUnauthorized {
			t.Errorf("%s with wrong token = %d", path, c)
		}
		if c := do(method, path, "Basic "+st.Token, nil); c != http.StatusUnauthorized {
			t.Errorf("%s with wrong scheme = %d", path, c)
		}
	}
	// Right token but a rebinding-style Host or a browser cross-origin request is still refused.
	if c := do(http.MethodGet, "/_control/info", "Bearer "+st.Token, map[string]string{"Host": "evil.example"}); c != http.StatusForbidden {
		t.Errorf("foreign Host = %d, want 403", c)
	}
	if c := do(http.MethodPost, "/_control/stop", "Bearer "+st.Token, map[string]string{"Origin": "http://evil.example", "Sec-Fetch-Site": "cross-site"}); c != http.StatusForbidden {
		t.Errorf("cross-origin stop = %d, want 403", c)
	}
	// Still up after all the refusals, and the right token works.
	if p := Probe(context.Background(), dd, "127.0.0.1", port); p.Phase != PhaseRunning {
		t.Fatalf("server stopped by an unauthorized request: %+v", p)
	}
	if c := do(http.MethodGet, "/_control/info", "Bearer "+st.Token, nil); c != http.StatusOK {
		t.Errorf("right token = %d", c)
	}
}

func TestProbeForeignProgramOnThePort(t *testing.T) {
	foreign := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("hello"))
	}))
	defer foreign.Close()
	port := foreign.Listener.Addr().(*net.TCPAddr).Port
	dd := t.TempDir()
	if st := Probe(context.Background(), dd, "127.0.0.1", port); st.Phase != PhasePortInUse {
		t.Errorf("phase = %s, want %s", st.Phase, PhasePortInUse)
	}
	// Even a state file claiming this port does not make the foreign program "ours" or stoppable.
	_ = runstate.Write(runstate.Path(dd, port), runstate.State{Instance: "i", Token: "t", Port: port})
	if err := Stop(context.Background(), dd, "127.0.0.1", port, time.Second); err == nil {
		t.Error("Stop must not act on a foreign program")
	}
}

func TestProbeStaleStateAndClear(t *testing.T) {
	// Grab a free port, then close it so nothing listens.
	ln, _ := net.Listen("tcp", "127.0.0.1:0")
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()

	dd := t.TempDir()
	_ = runstate.Write(runstate.Path(dd, port), runstate.State{Instance: "dead", Token: "t", Port: port})
	st := Probe(context.Background(), dd, "127.0.0.1", port)
	if st.Phase != PhaseStopped || !st.StaleState {
		t.Fatalf("probe = %+v", st)
	}
	if err := ClearStale(st); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(st.StatePath); !os.IsNotExist(err) {
		t.Error("stale state was not cleared")
	}
}

func TestSmoke(t *testing.T) {
	base, _ := startServer(t)
	res, err := Smoke(context.Background(), base+"/mcp")
	if err != nil {
		t.Fatal(err)
	}
	if res.ToolCount != 29 || res.ServerName != version.Name {
		t.Errorf("smoke = %+v", res)
	}
	if _, err := Smoke(context.Background(), "http://127.0.0.1:1/mcp"); err == nil {
		t.Error("smoke against a closed port should fail")
	}
}
