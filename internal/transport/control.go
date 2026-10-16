package transport

import (
	"crypto/subtle"
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"time"
)

// Control enables the authenticated local control endpoints used by "status" and "stop".
//
// There is no unauthenticated stop: every /_control/* request needs "Authorization: Bearer <Token>"
// (constant-time compared), a loopback Host header (DNS-rebinding guard) and no browser cross-origin
// context. The token only exists in the owner-only state file written by "serve".
type Control struct {
	Token string
	Info  Info
	// Stop requests a graceful shutdown of the server; it must not block.
	Stop func()
}

// Info is what an authenticated GET /_control/info returns; "status" compares Instance with the state file.
type Info struct {
	Instance  string    `json:"instance"`
	PID       int       `json:"pid"`
	Name      string    `json:"name"`
	Version   string    `json:"version"`
	URL       string    `json:"url"`
	StartedAt time.Time `json:"started_at"`
}

// Option customizes NewHandler.
type Option func(*options)

type options struct{ control *Control }

// WithControl enables /_control/info and /_control/stop.
func WithControl(c Control) Option { return func(o *options) { o.control = &c } }

func (c *Control) handler(h http.HandlerFunc) http.Handler {
	guarded := func(w http.ResponseWriter, r *http.Request) {
		if !loopbackHost(r.Host) {
			http.Error(w, "forbidden host", http.StatusForbidden)
			return
		}
		got, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !ok || c.Token == "" || subtle.ConstantTimeCompare([]byte(got), []byte(c.Token)) != 1 {
			w.Header().Set("WWW-Authenticate", "Bearer")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		h(w, r)
	}
	return http.NewCrossOriginProtection().Handler(http.HandlerFunc(guarded))
}

func (c *Control) register(mux *http.ServeMux) {
	mux.Handle("GET /_control/info", c.handler(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, c.Info)
	}))
	mux.Handle("POST /_control/stop", c.handler(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{"stopping": true, "instance": c.Info.Instance})
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		if c.Stop != nil {
			c.Stop()
		}
	}))
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

// loopbackHost reports whether a Host header names a loopback address.
func loopbackHost(hostport string) bool {
	h := hostport
	if host, _, err := net.SplitHostPort(hostport); err == nil {
		h = host
	}
	h = strings.TrimSuffix(strings.TrimPrefix(h, "["), "]")
	if strings.EqualFold(h, "localhost") {
		return true
	}
	ip := net.ParseIP(h)
	return ip != nil && ip.IsLoopback()
}
