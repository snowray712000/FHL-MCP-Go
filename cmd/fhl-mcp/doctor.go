package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/snowray712000/FHL-MCP-Go/internal/app"
	"github.com/snowray712000/FHL-MCP-Go/internal/clients"
	"github.com/snowray712000/FHL-MCP-Go/internal/fhlapi"
	"github.com/snowray712000/FHL-MCP-Go/internal/version"
)

const expectedTools = 29

type doctorReport struct {
	w                io.Writer
	okN, warnN, badN int
}

func (d *doctorReport) section(title string) { fmt.Fprintf(d.w, "\n[%s]\n", title) }
func (d *doctorReport) ok(f string, a ...any) {
	d.okN++
	fmt.Fprintf(d.w, "  ✔ %s\n", fmt.Sprintf(f, a...))
}
func (d *doctorReport) warn(f string, a ...any) {
	d.warnN++
	fmt.Fprintf(d.w, "  ⚠ %s\n", fmt.Sprintf(f, a...))
}
func (d *doctorReport) bad(f string, a ...any) {
	d.badN++
	fmt.Fprintf(d.w, "  ✘ %s\n", fmt.Sprintf(f, a...))
}
func (d *doctorReport) hint(f string, a ...any) {
	fmt.Fprintf(d.w, "      → %s\n", fmt.Sprintf(f, a...))
}

func doctorCmd(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("doctor", flag.ContinueOnError)
	fs.SetOutput(stderr)
	t := addTarget(fs)
	offline := fs.Bool("offline", false, "skip the check of the FHL website")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	d := &doctorReport{w: stdout}
	ctx := context.Background()
	url := app.MCPURL(*t.host, *t.port)

	d.section("環境")
	d.ok("%s %s，%s/%s，%s", version.Name, version.Version, runtime.GOOS, runtime.GOARCH, runtime.Version())
	if runtime.GOOS == "linux" {
		if b, err := os.ReadFile("/proc/version"); err == nil && strings.Contains(strings.ToLower(string(b)), "microsoft") {
			d.warn("目前在 WSL 內執行：這裡的 127.0.0.1 與 Windows 端的 127.0.0.1 不一定相通；客戶端與服務要在同一個環境，登錄也要分別在各環境做一次")
		}
	}

	d.section("設定")
	dd := ""
	if err := app.ValidateHost(*t.host); err != nil {
		d.bad("%v", err)
	} else if *t.port <= 0 || *t.port > 65535 {
		d.bad("埠 %d 無效", *t.port)
	} else {
		d.ok("監聽位址 %s (僅限本機)", app.HostPort(*t.host, *t.port))
	}
	if v, err := app.ResolveDataDir(*t.dataDir); err != nil {
		d.bad("找不到資料目錄：%v", err)
	} else {
		dd = v
		if err := checkWritable(dd); err != nil {
			d.bad("資料目錄無法寫入 %s：%v", dd, err)
			d.hint("用 --data-dir 指定可寫入的位置")
		} else {
			d.ok("資料目錄 %s 可寫入", dd)
		}
	}
	if _, err := app.ResolveCacheMaxBytes(-1); err != nil {
		d.bad("%v", err)
	} else {
		d.ok("快取設定正常")
	}

	d.section("服務")
	if dd != "" && app.ValidateHost(*t.host) == nil && *t.port > 0 && *t.port <= 65535 {
		pctx, cancel := context.WithTimeout(ctx, 20*time.Second)
		st := app.Probe(pctx, dd, *t.host, *t.port)
		switch st.Phase {
		case app.PhaseRunning, app.PhaseRunningUnverified:
			if st.Phase == app.PhaseRunning {
				d.ok("服務執行中 %s (版本 %s)", st.URL, st.Version)
			} else {
				d.warn("有 FHL-MCP-Go 在 %s 執行 (版本 %s)，但無法驗證身分，fhl-mcp stop 停不了它", st.URL, st.Version)
				d.hint("可能用了不同的 --data-dir，或是舊版")
			}
			if st.Version != "" && st.Version != version.Version {
				d.warn("執行中的服務版本 (%s) 與這個程式 (%s) 不同；升級後請先 stop 再重新 serve", st.Version, version.Version)
			}
			if res, err := app.Smoke(pctx, url); err != nil {
				d.bad("MCP 連線測試失敗：%v", err)
			} else {
				d.ok("MCP 連線測試通過 (協定 %s，%d 個工具，get_book_list 呼叫成功)", res.ProtocolVersion, res.ToolCount)
				if res.ToolCount != expectedTools {
					d.warn("工具數量是 %d，預期 %d", res.ToolCount, expectedTools)
				}
			}
		case app.PhasePortInUse:
			d.bad("埠被其他程式占用，本工具無法啟動：%s", st.Detail)
			d.hint("關掉占用的程式，或用 --port 換埠 (客戶端登錄要同步，重新執行 fhl-mcp setup --port N)")
		default:
			d.warn("服務未執行 (需要時用 fhl-mcp serve 啟動；沒有開機自動啟動是設計如此)")
			if st.StaleState {
				d.hint("有上次異常結束留下的狀態檔，下次 serve 會自動清除")
			}
		}
		cancel()
	} else {
		d.warn("設定有問題，略過服務檢查")
	}

	d.section("信望愛網站 (上游)")
	if *offline {
		d.warn("已指定 --offline，略過")
	} else {
		uctx, cancel := context.WithTimeout(ctx, 20*time.Second)
		if _, err := fhlapi.NewClient().AB(uctx, false); err != nil {
			d.warn("連不到信望愛網站的 API：%v (服務本身不受影響，但查詢類工具會失敗)", err)
		} else {
			d.ok("信望愛網站 API 可連線 (版本清單 ab.php)")
		}
		cancel()
	}

	d.section("客戶端登錄")
	for _, id := range clients.All {
		c := clients.New(id, clientRunner)
		_, reg, err := c.Inspect(ctx)
		switch {
		case err != nil && (errors.Is(err, exec.ErrNotFound) || errors.Is(err, os.ErrNotExist)):
			d.warn("%s：找不到 %s 命令列工具 (不在 PATH)；沒裝可略過，要手動連接：%s", c.Title, c.Binary, c.Manual(url))
			d.hint("%s", c.InstallHint())
		case err != nil:
			d.bad("%s：讀取登錄狀態失敗：%v", c.Title, err)
		case !reg.Found:
			d.warn("%s：尚未登錄 %q", c.Title, clients.ServerName)
			d.hint("執行 fhl-mcp setup --client %s", c.ID)
		case clients.SameURL(reg.URL, url):
			d.ok("%s：已登錄 %s (%s)", c.Title, reg.URL, scopeNote(reg.Scope))
		default:
			d.bad("%s：登錄的網址是 %s，與目前設定 %s 不一致", c.Title, orNone(reg.URL), url)
			d.hint("若是換過埠：fhl-mcp disconnect --client %s 之後重新 fhl-mcp setup", c.ID)
		}
	}

	fmt.Fprintf(stdout, "\n結果：%d 項正常，%d 項需留意，%d 項有問題\n", d.okN, d.warnN, d.badN)
	if d.badN > 0 {
		return 1
	}
	return 0
}

func checkWritable(dir string) error {
	if err := os.MkdirAll(filepath.Join(dir, "run"), 0o700); err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Join(dir, "run"), ".doctor-*.tmp")
	if err != nil {
		return err
	}
	name := f.Name()
	f.Close()
	return os.Remove(name)
}
