// Package transport serves the MCP server over local Streamable HTTP with Host/Origin protection.
package transport

import (
	"encoding/json"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// NewHandler returns the HTTP handler: /mcp (MCP), /health (liveness, never touches upstream), / (text hint).
//
// Protection: the SDK rejects a non-localhost Host on loopback by default (DNS rebinding);
// the SDK does not check Origin, so the MCP handler is wrapped with http.CrossOriginProtection
// (browser cross-origin requests are rejected; clients that send no Origin, like CLIs, pass).
func NewHandler(srv *mcp.Server, name, version string, opts ...Option) http.Handler {
	var o options
	for _, opt := range opts {
		opt(&o)
	}
	mcpHandler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return srv }, nil)

	mux := http.NewServeMux()
	if o.control != nil {
		o.control.register(mux)
	}
	mux.Handle("/mcp", http.NewCrossOriginProtection().Handler(mcpHandler))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "name": name, "version": version})
	})
	mux.HandleFunc("/{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte(name + " " + version + "\nMCP endpoint: /mcp (Streamable HTTP)\n"))
	})
	return mux
}
