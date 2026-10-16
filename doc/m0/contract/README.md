# M0 契約基準：Python HTTP server 實測匯出

- 匯出日期：2026-10-16
- 來源：`D:\coding\信望愛\FHL-MCP-Server` 工作目錄（HEAD `03e9cb0` + 未提交修改，見 [../source-manifest.md](../source-manifest.md)）
- 匯出方式：以 `create_http_app()` 綁 `127.0.0.1:18081` 啟動（正式入口綁 `0.0.0.0:8081`，只改綁定位址與埠，其餘中介層相同），
  用 Python `mcp` 1.29.0 的 Streamable HTTP client 呼叫；完成後已關閉伺服器。唯讀，未修改原專案。
- 檔案內容為 client 反序列化後再輸出（`by_alias`、`exclude_none`），不是 wire 上的原始 JSON-RPC bytes。

| 檔案 | 內容 |
| --- | --- |
| `initialize.json` | 協議版本、capabilities、serverInfo |
| `tools-list.json` | 完整 `tools/list`（29 項，含 description、inputSchema、outputSchema） |
| `prompts-list.json` / `resources-list.json` / `resource-templates-list.json` | 皆為空 |
| `sample-calls.json` | 3 次 `tools/call` 的原始結果樣本（成功 2、錯誤 1） |

## 事實

1. **29 個工具**，名稱與計畫書 §4 一致；沒有分頁（`nextCursor` 為空）。
2. **HTTP 入口的 Prompts／Resources／ResourceTemplates 皆為 0 項**，但 capabilities 仍宣告 `prompts`、`resources`。
   19 Prompts 與 7 類 Resources 只存在於 stdio 的 `server.py`，不是 HTTP 契約基準（與 CLAUDE.md 的說法一致）。
3. `protocolVersion` 為 `2025-11-25`；`serverInfo.name` = `FHL Bible MCP Server`，`serverInfo.version` = `1.29.0`
   （這是 mcp SDK 版本，不是產品版本 `0.1.2`）。
4. **每個工具的 `outputSchema` 都是 `{"result": string}`**（required），title 為 `{tool}Output`。
5. **成功回傳同時含 `content[0].text` 與 `structuredContent.result`，兩者是同一個字串**；字串本身是 `indent=2`、
   `ensure_ascii=false` 的 JSON 文字。→ Go 版要同時輸出 text content 與 `structuredContent: {"result": "<同字串>"}`，
   並宣告相同 outputSchema，才算相容；計畫書 §9 只提到 text，這點需補。
6. `inputSchema.properties` 為 FastMCP 從 Python 簽名自動產生：`title` 為欄位名首字大寫、`verse` 之類可選參數是
   `"default": null` 且 `type: "string"`（不是 `anyOf` null），inputSchema 的 `title` 為 `{tool}Arguments`。
7. description 是 Python docstring 原文（正體中文，含 `Args:` 縮排區塊與換行）。

## 已知原版缺陷（尚未決定是否保留）

- `get_book_info(book="不存在書卷")` 回 `isError: true`，文字為
  `Error executing tool get_book_info: InvalidParameterError.__init__() missing 1 required positional argument: 'value'`。
  這是 Python 例外類別建構錯誤，不是刻意設計的錯誤訊息；Go 版應歸類為「明確 bug fix」，另定錯誤格式。

## 未涵蓋

- 只做了 3 次呼叫；其餘 27 個工具的實際輸出格式、錯誤路徑、null/數字 vs 字串行為尚未取樣（需 fixture 化，並注意上游流量）。
- 未量測 wire 層原始 JSON-RPC；未驗證 session / 重連行為。
