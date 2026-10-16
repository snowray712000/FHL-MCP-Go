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
	"github.com/snowray712000/FHL-MCP-Go/internal/cache"
	"github.com/snowray712000/FHL-MCP-Go/internal/version"
)

const usage = `用法: fhl-mcp [指令] [選項]

指令:
  serve      前景啟動 MCP 服務 (預設)；已在執行時只顯示既有網址
  setup      把本服務登錄到 Claude Code / Codex (使用者層級，只需做一次)
  status     顯示服務是否在執行 (--json 輸出 JSON)
  stop       經身分驗證停止執行中的服務
  doctor     分項診斷：設定、服務、信望愛網站、客戶端登錄
  disconnect 移除 Claude Code / Codex 裡本服務的登錄 (不動其他 MCP)
  cache clear 清除本工具的回應快取
  version    顯示版本
  help       顯示本說明

serve 選項:
  --host     監聽位址，只接受本機位址 (預設 127.0.0.1)
  --port     監聽埠 (預設 8081)；被占用時直接失敗，不會改用其他埠
  --log-level debug|info|warn|error (預設 info；日誌寫到 stderr)
  --data-dir 資料目錄 (預設為系統使用者快取目錄下的 fhl-mcp；也可用 FHL_DATA_DIR)；快取放在其下的 cache/
  --cache-max-mb 快取容量上限，單位 MB (預設 200，0 = 不使用快取；也可用 FHL_CACHE_MAX_MB)

status / stop / doctor 選項:
  --host --port --data-dir 同 serve (要與啟動時相同才找得到)
  status 另有 --json；doctor 另有 --offline (不連信望愛網站)

setup / disconnect 選項:
  --client claude|codex|all  指定客戶端 (setup 未指定且在終端機時會詢問)
  --host --port              登錄的網址 (預設 127.0.0.1:8081)

cache clear 選項:
  --data-dir 同 serve
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
	case "cache":
		return cacheCmd(args, stdout, stderr)
	case "status":
		return statusCmd(args, stdout, stderr)
	case "stop":
		return stopCmd(args, stdout, stderr)
	case "setup":
		return setupCmd(args, stdout, stderr)
	case "disconnect":
		return disconnectCmd(args, stdout, stderr)
	case "doctor":
		return doctorCmd(args, stdout, stderr)
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
	dataDir := fs.String("data-dir", "", "data directory (default: user cache dir; env FHL_DATA_DIR)")
	cacheMB := fs.Int("cache-max-mb", -1, "cache size cap in MB (default 200, 0 disables; env FHL_CACHE_MAX_MB)")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	maxBytes, err := app.ResolveCacheMaxBytes(*cacheMB)
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 2
	}
	cacheDir := ""
	if maxBytes > 0 {
		if cacheDir, err = app.ResolveCacheDir(*dataDir); err != nil {
			fmt.Fprintf(stderr, "%v\n", err)
			return 2
		}
	}

	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(*level)); err != nil {
		fmt.Fprintf(stderr, "無效的 --log-level %q\n", *level)
		return 2
	}
	logger := slog.New(slog.NewTextHandler(stderr, &slog.HandlerOptions{Level: lvl}))

	dd, err := app.ResolveDataDir(*dataDir)
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 2
	}
	if err := app.ValidateHost(*host); err != nil {
		fmt.Fprintf(stderr, "無法啟動: %v\n", err)
		return 1
	}

	// Already running: show the existing address instead of failing on the port (a port held by some
	// other program still fails below, and is never silently moved).
	if *port != 0 {
		st := app.Probe(context.Background(), dd, *host, *port)
		switch st.Phase {
		case app.PhaseRunning, app.PhaseRunningUnverified:
			fmt.Fprintf(stdout, "%s 已在執行，網址：%s\n停止請執行：fhl-mcp stop\n", version.Name, st.URL)
			return 0
		}
		if err := app.ClearStale(st); err != nil {
			logger.Warn("cannot remove a stale state file", "path", st.StatePath, "err", err)
		}
	}

	cfg := app.Config{Host: *host, Port: *port, Logger: logger, CacheDir: cacheDir, CacheMaxBytes: maxBytes, Run: &app.RunConfig{DataDir: dd}}
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

// cacheCmd implements "cache clear".
func cacheCmd(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] != "clear" {
		fmt.Fprintf(stderr, "用法: fhl-mcp cache clear [--data-dir 目錄]\n")
		return 2
	}
	fs := flag.NewFlagSet("cache clear", flag.ContinueOnError)
	fs.SetOutput(stderr)
	dataDir := fs.String("data-dir", "", "data directory (default: user cache dir; env FHL_DATA_DIR)")
	if err := fs.Parse(args[1:]); err != nil {
		return 2
	}
	dir, err := app.ResolveCacheDir(*dataDir)
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 2
	}
	n, err := cache.New(dir, 0).Clear()
	if err != nil {
		fmt.Fprintf(stderr, "清除快取時發生錯誤 (已清除 %d 個檔案): %v\n", n, err)
		return 1
	}
	fmt.Fprintf(stdout, "已清除 %d 個快取檔案 (%s)\n", n, dir)
	return 0
}
