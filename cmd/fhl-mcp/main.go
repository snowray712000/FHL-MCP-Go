// Command fhl-mcp is the FHL Bible MCP server (Go rewrite).
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/snowray712000/FHL-MCP-Go/internal/app"
	"github.com/snowray712000/FHL-MCP-Go/internal/version"
)

const usage = `用法: fhl-mcp [指令] [選項]

指令:
  serve      前景啟動 MCP 服務 (預設)
  version    顯示版本
  help       顯示本說明

serve 選項:
  --host     監聽位址，只接受本機位址 (預設 127.0.0.1)
  --port     監聽埠 (預設 8081)；被占用時直接失敗，不會改用其他埠
  --log-level debug|info|warn|error (預設 info；日誌寫到 stderr)
`

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	cmd := "serve"
	if len(args) > 0 && len(args[0]) > 0 && args[0][0] != '-' {
		cmd, args = args[0], args[1:]
	}
	switch cmd {
	case "serve":
		return serve(args, stdout, stderr)
	case "version":
		fmt.Fprintf(stdout, "%s %s (%s/%s, %s)\n", version.Name, version.Version, runtime.GOOS, runtime.GOARCH, runtime.Version())
		return 0
	case "help":
		fmt.Fprint(stdout, usage)
		return 0
	default:
		fmt.Fprintf(stderr, "未知指令 %q\n\n%s", cmd, usage)
		return 2
	}
}

func serve(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	fs.SetOutput(stderr)
	host := fs.String("host", "127.0.0.1", "listen address (loopback only)")
	port := fs.Int("port", app.DefaultPort, "listen port")
	level := fs.String("log-level", "info", "debug|info|warn|error")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(*level)); err != nil {
		fmt.Fprintf(stderr, "無效的 --log-level %q\n", *level)
		return 2
	}
	logger := slog.New(slog.NewTextHandler(stderr, &slog.HandlerOptions{Level: lvl}))

	cfg := app.Config{Host: *host, Port: *port, Logger: logger}
	ln, err := app.Listen(cfg)
	if err != nil {
		fmt.Fprintf(stderr, "無法啟動: %v\n", err)
		return 1
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := app.Serve(ctx, cfg, ln, stdout); err != nil {
		fmt.Fprintf(stderr, "服務結束: %v\n", err)
		return 1
	}
	return 0
}
