# app_versions

版本更新紀錄（對應 FHLNUI 的 `app_versions.json`）。最新的在最上面；`TODOs` 區塊永遠放第一。

- last: **0.0.0.6**
- 版次格式：`主.次.修.小`（例 `0.0.0.1`）。小改動升最後一碼；整理一批小版本時，另立一個較高版次並註明整理範圍。
- 發版時：先更新本檔，再 commit，標題為 `發版 {版次}：{摘要}`（見 `.claude/CLAUDE.md` 的 commit 慣例）。
- 目前沒有 Go 程式碼，尚無「程式內版本字串」；有了之後，`version` 指令與建置注入的版本須與 `last` 一致。

## TODOs

- 🆕 新功能: M1 最小垂直切片：鎖定 Go 與 `go-sdk` 版本、HTTP `/mcp`、經文查詢、`serve`，Claude Code 與 Codex 同時連線
- 🆕 新功能: M1 早期試驗：(a) 未知工具回 `isError` 結果 (D2)、(b) text + `structuredContent{result}` + `outputSchema` (D1)、(c) Host／Origin 驗證、(d) 所選 Go 工具鏈對 macOS 12 的支援
- ⏳ 待決定: macOS 最低版待 M1 選定 Go 工具鏈並實機驗證後鎖定

## 0.0.0.6

- ✨ 新增: M0 收尾：計畫書 §15 全部定案 — 專案名稱 `FHL-MCP-Go`；首版只交付 29 Tools；平台 Ubuntu 22.04/24.04 x64 + macOS 12+ arm64 (最終 macOS 最低版 M1 鎖定) + Windows x64；首版 beta 不簽章、不公證；不做 Smithery 與根路徑 `/` 相容 (只提供 `/mcp`)
- ✨ 新增: `.claude/CLAUDE.md` 同步 §15 定案與 D2–D12 核准；`code/ai-log/261016l` 記錄決策與 M1 試驗清單
- ℹ️ 說明: M0 退出條件全部滿足 (來源 manifest、契約基準、Prompts／Resources 盤點、工具矩陣與差異清單、範圍與平台定案)；仍無 Go 程式碼，下一步為 M1

## 0.0.0.5

- ✨ 新增: `doc/m0/tool-matrix.md`：29 個工具契約矩陣 (實測上游端點、輸出欄位、備註) 與原版差異清單 D1–D16
- ✨ 新增: `doc/m0/contract/tool-samples.json` (55 次呼叫、32 次上游請求的原始樣本)、`contract/upstream-ssc-samples.json` (D12 上游回應)
- ✨ 新增: 計畫書 §9「已核准的原版差異」：D2–D12 經使用者核准 (錯誤格式維持 `isError`、不認得書卷統一訊息、`use_simplified` 真正生效並納入快取 key、`null`／數字字串放寬、查無資料形狀保留、上游 `status` 非 success 回明確訊息)；D10 先照原樣
- 🐞 發現: Python 原版 `use_simplified` 在多數工具被忽略；`InvalidParameterError` 建構錯誤不只 `get_book_info`；`search_commentary` 常見詞被上游 `record count > 500` 吞成 0 筆
- ℹ️ 說明: 實測上游一律為 `bible.fhl.net/api/*.php` (文章為 `www.fhl.net/api/json.php`)，未使用 `/json/`；M1 須確認 Go SDK 能讓未知工具回 `isError` 結果；仍無 Go 程式碼

## 0.0.0.4

- ✨ 新增: `doc/m0/prompts-resources-inventory.md`：盤點 19 個 Prompts (4/3/4/5/3、45 個參數) 與 7 個 Resource URI (4 個 scheme)，並記錄 stdio `server.py` 的曝光方式與原版缺陷 (路徑未 `unquote`、`int()` 例外、`info://books` 忽略 `simplified`)
- ✨ 新增: `doc/m0/contract/prompts-inventory.json`：19 個 Prompts 的參數與渲染結果樣本 (M4 可作 golden)
- ✨ 新增: 決定「首版只交付 29 個 Tools，Prompts 與 Resources 移到下一版本 (M4)」；計畫書 §1 加入首版完成定義 (M0–M3、M5、M6)，§4、§5、§12、§13、§15、§17 與 `.claude/CLAUDE.md` 同步
- ℹ️ 說明: Prompts／Resources 只存在於 stdio 的 `server.py`，HTTP 入口為 0 項；其餘 27 個工具尚未取樣輸出；仍無 Go 程式碼

## 0.0.0.3

- ✨ 新增: M0 契約基準 `doc/m0/contract/`：Python HTTP server 實測匯出 `initialize`、`tools/list` (29 個工具含 inputSchema / outputSchema)、空的 prompts / resources 列表、3 筆 `tools/call` 樣本，附 `README.md` 整理事實
- ✨ 新增: 決定輸出格式：Go 版同時輸出 text content 與 `structuredContent{result}`，並宣告相同 `outputSchema` (計畫書 §9、`.claude/CLAUDE.md` 已同步)
- 🐞 發現: Python 原版 `get_book_info` 查無書卷時回 `InvalidParameterError.__init__() missing ... 'value'` (例外建構錯誤)，Go 版歸為明確 bug fix，錯誤格式待定
- ℹ️ 說明: HTTP 入口的 Prompts / Resources 皆為 0 項 (19 Prompts、7 類 Resources 僅存在於 stdio 的 `server.py`)；其餘 27 個工具尚未取樣輸出；仍無 Go 程式碼

## 0.0.0.2

- ✨ 新增: M0 來源 manifest `doc/m0/source-manifest.tsv`：Python 原專案 254 個檔案的 SHA-256 與 Git 狀態 (237 未動 / 7 已修改 / 10 未追蹤；HEAD `03e9cb0`)
- ✨ 新增: `doc/m0/source-manifest.md`：manifest 範圍、未追蹤檔案說明、已核對事實 (HTTP 入口 29 個工具名稱與計畫書一致)
- ℹ️ 說明: 仍無 Go 程式碼；尚未匯出 `tools/list` schema，也尚未盤點 Prompts／Resources

## 0.0.0.1

- ✨ 新增: 重寫計畫書 `doc/FHL-MCP-Go-重寫計畫書.md` (討論版 v0.1)
- ✨ 新增: `.claude/CLAUDE.md`：專案現況、ai-log 流程、架構限制、git commit 慣例 (沿用 FHLNUI)
- ✨ 新增: `code/ai-log/`：AI 工作紀錄與 `index.md` 索引
- ✨ 新增: 本檔 `app_versions.md`
- ℹ️ 說明: 尚無任何 Go 程式碼，此版僅為規劃與專案骨架
