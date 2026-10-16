# app_versions

版本更新紀錄（對應 FHLNUI 的 `app_versions.json`）。最新的在最上面；`TODOs` 區塊永遠放第一。

- last: **0.0.0.2**
- 版次格式：`主.次.修.小`（例 `0.0.0.1`）。小改動升最後一碼；整理一批小版本時，另立一個較高版次並註明整理範圍。
- 發版時：先更新本檔，再 commit，標題為 `發版 {版次}：{摘要}`（見 `.claude/CLAUDE.md` 的 commit 慣例）。
- 目前沒有 Go 程式碼，尚無「程式內版本字串」；有了之後，`version` 指令與建置注入的版本須與 `last` 一致。

## TODOs

- 🆕 新功能: M0 來源與契約：匯出 29 個工具 schema、盤點 Prompts／Resources、確認是否納入首版 (見計畫書 §13、§15)
- 🆕 新功能: M1 最小垂直切片：Go SDK + HTTP + 經文查詢 + `serve`，Claude Code 與 Codex 同時連線
- ⏳ 待決定: 計畫書 §15 的待定項目 (完整功能範圍、Ubuntu／macOS 最低版、簽章公證、Smithery 相容)

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
