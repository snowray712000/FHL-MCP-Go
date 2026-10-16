// Package app wires configuration, the upstream client, the MCP server and the HTTP listener.
package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/snowray712000/FHL-MCP-Go/internal/bible"
	"github.com/snowray712000/FHL-MCP-Go/internal/fhlapi"
	"github.com/snowray712000/FHL-MCP-Go/internal/mcpserver"
	"github.com/snowray712000/FHL-MCP-Go/internal/tools"
	"github.com/snowray712000/FHL-MCP-Go/internal/transport"
	"github.com/snowray712000/FHL-MCP-Go/internal/version"
)

// DefaultPort is the fixed port the client registration points at.
const DefaultPort = 8081

// Config is the serve configuration (M1: flags only; config file and FHL_ env vars come in M3).
type Config struct {
	Host string // must be a loopback address
	Port int
	// API is the upstream client; nil means the production client. Tests inject a mock-backed one.
	API *fhlapi.Client
	// Logger receives structured logs; nil discards them.
	Logger *slog.Logger
}

// ValidateHost accepts only loopback hosts: the first release is local-only.
func ValidateHost(host string) error {
	if host == "localhost" {
		return nil
	}
	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() {
		return fmt.Errorf("host %q is not a loopback address; only 127.0.0.1, ::1 and localhost are allowed", host)
	}
	return nil
}

// Listen validates the config and binds the port. It fails (never picks another port) if the port is taken.
func Listen(cfg Config) (net.Listener, error) {
	if err := ValidateHost(cfg.Host); err != nil {
		return nil, err
	}
	if cfg.Port < 0 || cfg.Port > 65535 {
		return nil, fmt.Errorf("port %d is out of range", cfg.Port)
	}
	addr := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("cannot listen on %s (is another program, or another copy of this server, using the port?): %w", addr, err)
	}
	return ln, nil
}

// Serve runs the HTTP MCP server on ln until ctx is cancelled, then shuts down gracefully.
func Serve(ctx context.Context, cfg Config, ln net.Listener, out io.Writer) error {
	log := cfg.Logger
	if log == nil {
		log = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	api := cfg.API
	if api == nil {
		api = fhlapi.NewClient()
	}

	registry := tools.NewRegistry(&bible.Service{API: api})
	srv, err := mcpserver.New(mcpserver.Options{Name: version.Name, Version: version.Version, Tools: registry, Logger: log})
	if err != nil {
		return err
	}

	httpSrv := &http.Server{
		Handler:           transport.NewHandler(srv, version.Name, version.Version),
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       2 * time.Minute,
		// No WriteTimeout / ReadTimeout: MCP uses long-lived streams.
		BaseContext: func(net.Listener) context.Context { return ctx },
	}

	url := "http://" + ln.Addr().String() + "/mcp"
	fmt.Fprintf(out, "%s %s\nMCP endpoint: %s\nPress Ctrl+C to stop.\n", version.Name, version.Version, url)
	log.Info("serving", "url", url)

	errCh := make(chan error, 1)
	go func() { errCh <- httpSrv.Serve(ln) }()

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
	}

	fmt.Fprintln(out, "Shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpSrv.Shutdown(shutdownCtx); err != nil {
		_ = httpSrv.Close()
		return fmt.Errorf("shutdown: %w", err)
	}
	return nil
}
