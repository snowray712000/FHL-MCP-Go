package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/snowray712000/FHL-MCP-Go/internal/app"
	"github.com/snowray712000/FHL-MCP-Go/internal/clients"
)

// Seams for tests.
var (
	clientRunner    clients.Runner = clients.ExecRunner{}
	stdinReader     io.Reader      = os.Stdin
	stdinIsTerminal                = func() bool {
		fi, err := os.Stdin.Stat()
		return err == nil && fi.Mode()&os.ModeCharDevice != 0
	}
)

// parseClients turns --client into client IDs. ok=false means the value is invalid.
func parseClients(v string) (ids []string, explicit bool, ok bool) {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "":
		return nil, false, true
	case "all", "both":
		return clients.All, false, true
	case clients.Claude, "claude-code":
		return []string{clients.Claude}, true, true
	case clients.Codex:
		return []string{clients.Codex}, true, true
	}
	return nil, false, false
}

// askClients prompts for the client(s) to connect; empty input means both.
func askClients(in io.Reader, out io.Writer) []string {
	fmt.Fprint(out, "要連接哪個 AI 工具？\n  1) Claude Code\n  2) Codex\n  3) 兩者 (預設)\n請輸入 1、2 或 3：")
	line, _ := bufio.NewReader(in).ReadString('\n')
	switch strings.TrimSpace(line) {
	case "1":
		return []string{clients.Claude}
	case "2":
		return []string{clients.Codex}
	}
	return clients.All
}

func setupCmd(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("setup", flag.ContinueOnError)
	fs.SetOutput(stderr)
	t := addTarget(fs)
	clientFlag := fs.String("client", "", "claude | codex | all")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if _, ok := t.resolve(stderr); !ok {
		return 2
	}
	ids, explicit, ok := parseClients(*clientFlag)
	if !ok {
		fmt.Fprintf(stderr, "無效的 --client %q (可用 claude、codex、all)\n", *clientFlag)
		return 2
	}
	if ids == nil {
		if stdinIsTerminal() {
			ids = askClients(stdinReader, stdout)
		} else {
			ids = clients.All
		}
	}

	url := app.MCPURL(*t.host, *t.port)
	ctx := context.Background()
	fmt.Fprintf(stdout, "登錄網址：%s (使用者層級，名稱 %q)\n\n", url, clients.ServerName)

	var outcomes []clients.Outcome
	for _, id := range ids {
		o := clients.New(id, clientRunner).Setup(ctx, url)
		outcomes = append(outcomes, o)
		printOutcome(stdout, o, url, true)
	}

	okCount, bad := 0, false
	for _, o := range outcomes {
		switch {
		case o.OK():
			okCount++
		case o.State == clients.CLIMissing && !explicit:
			// "all": a client that is not installed is not a failure by itself.
		default:
			bad = true
		}
	}
	if okCount == 0 {
		bad = true
	}

	if okCount > 0 {
		dd, _ := app.ResolveDataDir(*t.dataDir)
		sctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		running := false
		if st := app.Probe(sctx, dd, *t.host, *t.port); st.Phase == app.PhaseRunning || st.Phase == app.PhaseRunningUnverified {
			running = true
			if res, err := app.Smoke(sctx, url); err != nil {
				fmt.Fprintf(stdout, "\n⚠ 服務有在執行，但連線測試失敗：%v\n", err)
				bad = true
			} else {
				fmt.Fprintf(stdout, "\n✔ 連線測試通過：%s %s，%d 個工具可用。\n", res.ServerName, res.ServerVersion, res.ToolCount)
			}
		} else {
			fmt.Fprintln(stdout, "\n服務目前沒有執行，尚未做連線測試。")
		}
		if running {
			fmt.Fprintln(stdout, "下一步：在客戶端重新載入 MCP 或開啟新的工作階段即可。可用 fhl-mcp doctor 檢查。")
		} else {
			fmt.Fprintln(stdout, "下一步：執行 fhl-mcp serve 啟動服務，然後在客戶端重新載入 MCP 或開啟新的工作階段。可用 fhl-mcp doctor 檢查。")
		}
	}
	if bad {
		return 1
	}
	return 0
}

func disconnectCmd(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("disconnect", flag.ContinueOnError)
	fs.SetOutput(stderr)
	t := addTarget(fs)
	clientFlag := fs.String("client", "all", "claude | codex | all")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if _, ok := t.resolve(stderr); !ok {
		return 2
	}
	ids, explicit, ok := parseClients(*clientFlag)
	if !ok || ids == nil {
		fmt.Fprintf(stderr, "無效的 --client %q (可用 claude、codex、all)\n", *clientFlag)
		return 2
	}
	url := app.MCPURL(*t.host, *t.port)
	bad := false
	for _, id := range ids {
		o := clients.New(id, clientRunner).Disconnect(context.Background())
		printOutcome(stdout, o, url, false)
		if !o.OK() && !(o.State == clients.CLIMissing && !explicit) {
			bad = true
		}
	}
	if bad {
		return 1
	}
	return 0
}

// printOutcome prints one client's result in plain 正體中文, with the manual command when needed.
func printOutcome(w io.Writer, o clients.Outcome, url string, setup bool) {
	name := o.Client.Title
	switch o.State {
	case clients.Added:
		fmt.Fprintf(w, "✔ %s：已登錄 (%s)\n", name, scopeNote(o.Scope))
	case clients.Already:
		fmt.Fprintf(w, "✔ %s：已經登錄過相同網址，不需變更 (%s)\n", name, scopeNote(o.Scope))
	case clients.Conflict:
		fmt.Fprintf(w, "✘ %s：已有名為 %q 的 MCP，但網址不同 (%s)，為避免誤蓋，沒有變更。\n", name, clients.ServerName, orNone(o.ExistingURL))
		fmt.Fprintf(w, "  若那是舊的本工具設定，請先執行 fhl-mcp disconnect --client %s，再重新 setup；若是別的伺服器，請自行處理或改名。\n", o.Client.ID)
	case clients.CLIMissing:
		if setup {
			fmt.Fprintf(w, "－ %s：找不到 %s 命令列工具 (不在 PATH)，沒有變更。要手動連接請執行：\n    %s\n", name, o.Client.Binary, o.Client.Manual(url))
		} else {
			fmt.Fprintf(w, "－ %s：找不到 %s 命令列工具，略過。\n", name, o.Client.Binary)
		}
	case clients.Removed:
		fmt.Fprintf(w, "✔ %s：已移除本服務的登錄 (%s)\n", name, orNone(o.ExistingURL))
	case clients.NotPresent:
		fmt.Fprintf(w, "－ %s：沒有名為 %q 的登錄，不需移除。\n", name, clients.ServerName)
	case clients.Refused:
		fmt.Fprintf(w, "✘ %s：名為 %q 的 MCP 網址是 %s，看起來不是本工具的本機服務，沒有移除。\n", name, clients.ServerName, orNone(o.ExistingURL))
	default:
		fmt.Fprintf(w, "✘ %s：失敗：%v\n", name, o.Err)
		if setup {
			fmt.Fprintf(w, "  可手動執行：%s\n", o.Client.Manual(url))
		}
	}
}

func scopeNote(s string) string {
	if s == "" {
		return "使用者層級"
	}
	return s
}

func orNone(s string) string {
	if s == "" {
		return "非 HTTP 類型"
	}
	return s
}
