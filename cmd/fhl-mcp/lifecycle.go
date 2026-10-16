package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"time"

	"github.com/snowray712000/FHL-MCP-Go/internal/app"
)

// target is the common --host/--port/--data-dir trio of status, stop, setup, disconnect and doctor.
type target struct {
	host    *string
	port    *int
	dataDir *string
}

func addTarget(fs *flag.FlagSet) target {
	return target{
		host:    fs.String("host", "127.0.0.1", "server address (loopback only)"),
		port:    fs.Int("port", app.DefaultPort, "server port"),
		dataDir: fs.String("data-dir", "", "data directory (default: user cache dir; env FHL_DATA_DIR)"),
	}
}

// resolve validates the trio and returns the absolute data directory.
func (t target) resolve(stderr io.Writer) (string, bool) {
	if err := app.ValidateHost(*t.host); err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return "", false
	}
	if *t.port <= 0 || *t.port > 65535 {
		fmt.Fprintf(stderr, "無效的 --port %d\n", *t.port)
		return "", false
	}
	dd, err := app.ResolveDataDir(*t.dataDir)
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return "", false
	}
	return dd, true
}

func statusCmd(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("status", flag.ContinueOnError)
	fs.SetOutput(stderr)
	t := addTarget(fs)
	asJSON := fs.Bool("json", false, "output JSON")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	dd, ok := t.resolve(stderr)
	if !ok {
		return 2
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	st := app.Probe(ctx, dd, *t.host, *t.port)

	if *asJSON {
		b, _ := json.MarshalIndent(st, "", "  ")
		fmt.Fprintln(stdout, string(b))
	} else {
		printStatus(stdout, st)
	}
	if st.Phase == app.PhaseRunning || st.Phase == app.PhaseRunningUnverified {
		return 0
	}
	return 1
}

func printStatus(w io.Writer, st app.Status) {
	switch st.Phase {
	case app.PhaseRunning:
		fmt.Fprintf(w, "狀態：執行中\n網址：%s\n版本：%s\n程序：PID %d (僅供參考，停止不靠 PID)\n", st.URL, st.Version, st.PID)
		if st.StartedAt != nil {
			fmt.Fprintf(w, "啟動：%s (已執行 %s)\n", st.StartedAt.Local().Format("2006-01-02 15:04:05"), time.Since(*st.StartedAt).Round(time.Second))
		}
		fmt.Fprintln(w, "停止請執行：fhl-mcp stop")
	case app.PhaseRunningUnverified:
		fmt.Fprintf(w, "狀態：有一個 FHL-MCP-Go 在執行，但無法驗證身分\n網址：%s\n版本：%s\n", st.URL, st.Version)
		fmt.Fprintln(w, "可能是用不同的 --data-dir 啟動，或是舊版 (沒有控制通道)。請加上相同的 --data-dir 再試，或到執行它的視窗按 Ctrl+C。")
	case app.PhasePortInUse:
		fmt.Fprintf(w, "狀態：未執行，而且該埠被其他程式占用\n網址：%s\n原因：%s\n", st.URL, st.Detail)
		fmt.Fprintln(w, "本工具不會自動改用其他埠；請關掉占用的程式，或用 --port 另選埠 (客戶端登錄也要同步，見 fhl-mcp setup)。")
	default:
		fmt.Fprintf(w, "狀態：未執行\n網址：%s\n啟動請執行：fhl-mcp serve\n", st.URL)
		if st.StaleState {
			fmt.Fprintln(w, "(發現上次異常結束留下的狀態檔，下次啟動時會自動清除)")
		}
	}
}

func stopCmd(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("stop", flag.ContinueOnError)
	fs.SetOutput(stderr)
	t := addTarget(fs)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	dd, ok := t.resolve(stderr)
	if !ok {
		return 2
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	st := app.Probe(ctx, dd, *t.host, *t.port)
	switch st.Phase {
	case app.PhaseStopped:
		if err := app.ClearStale(st); err != nil {
			fmt.Fprintf(stderr, "無法清除過期的狀態檔 %s: %v\n", st.StatePath, err)
		}
		fmt.Fprintf(stdout, "沒有執行中的服務 (%s)。\n", st.URL)
		return 0
	case app.PhasePortInUse:
		fmt.Fprintf(stderr, "該埠被其他程式占用，不是本工具，未做任何動作 (%s)。\n", st.URL)
		return 1
	case app.PhaseRunningUnverified:
		fmt.Fprintf(stderr, "有一個 FHL-MCP-Go 在 %s 執行，但找不到對應的控制資訊，無法驗證身分，所以不會停止它。\n可能是用不同的 --data-dir 啟動，或是舊版；請加上相同的 --data-dir，或到執行它的視窗按 Ctrl+C。\n", st.URL)
		return 1
	}
	if err := app.Stop(ctx, dd, *t.host, *t.port, 10*time.Second); err != nil {
		fmt.Fprintf(stderr, "停止失敗: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "已停止 (%s)。\n", st.URL)
	return 0
}
