# FHL-MCP-Go

[信望愛（FHL）](https://www.fhl.net/) 聖經查經 MCP 伺服器的 Go 版本。啟動後，Claude Code、Codex 等 AI 工具就能直接查經文、搜尋、查原文字典、註釋、有聲聖經、次經與使徒教父著作等。

- 單一執行檔，不需要安裝 Python 或其他執行環境（`CGO_ENABLED=0` 純 Go 建置）。
- 在本機開一個 Streamable HTTP MCP 端點：`http://127.0.0.1:8081/mcp`，只接受本機連線。
- 29 個工具，工具名稱與參數沿用原 Python 版 FHL-MCP-Server 的 HTTP 介面。（Prompts 與 Resources 之後的版本再提供。）
- 支援平台：Windows x64、macOS 13+（Apple 晶片）、Ubuntu 22.04／24.04 x64。

> 目前為 beta：發佈包沒有數位簽章與 Apple 公證，第一次執行時系統可能跳出警告。

## 安裝

### 用 npm

```sh
npm i -g @fhlnet/fhl-mcp
fhl-mcp version
```

### 從 GitHub Release 下載

到 [Releases](https://github.com/snowray712000/FHL-MCP-Go/releases) 下載對應你電腦的壓縮檔（`…_windows_amd64.zip`、`…_darwin_arm64.tar.gz`、`…_linux_amd64.tar.gz`），解壓縮後使用。可用同處的 `SHA256SUMS.txt` 核對檔案。

## 快速使用

```sh
fhl-mcp setup    # 做一次：連接 Claude Code 或 Codex
fhl-mcp serve    # 要查經時啟動；關掉視窗或按 Ctrl+C 即停
```

然後在 Claude Code／Codex 重新載入 MCP（或開新的工作階段），就可以問：「幫我查約翰福音 3:16」、「搜尋含『恩典』的經文」。

Windows 與 Mac 也可以雙擊發佈包內的 `首次設定.cmd`／`Setup.command` 與 `啟動-FHL-MCP.cmd`／`Start-FHL-MCP.command`。

常用指令：`fhl-mcp status`、`stop`、`doctor`、`disconnect`、`cache clear`。完整說明見 [doc/快速開始.md](doc/快速開始.md)。

## 提供的工具

| 類別 | 工具 |
| --- | --- |
| 經文 | `get_bible_verse`、`get_bible_chapter`、`query_verse_citation`、`get_bible_footnote` |
| 書卷與譯本 | `get_book_list`、`get_book_info`、`list_bible_versions`、`search_available_versions` |
| 搜尋 | `search_bible`、`search_bible_advanced` |
| 原文與字典 | `lookup_strongs`、`search_strongs_occurrences`、`get_word_analysis`、`get_parsing_reference` |
| 註釋與主題 | `get_commentary`、`list_commentaries`、`search_commentary`、`get_topic_study` |
| 有聲聖經 | `get_audio_bible`、`get_audio_chapter_with_text`、`list_audio_versions` |
| 次經 | `get_apocrypha_verse`、`search_apocrypha`、`list_apocrypha_books` |
| 使徒教父 | `get_apostolic_fathers_verse`、`search_apostolic_fathers`、`list_apostolic_fathers_books` |
| 信望愛文章 | `search_fhl_articles`、`list_fhl_article_columns` |

## 從原始碼建置

需要 Go 1.27.x。

```sh
go test ./...
go vet ./...
CGO_ENABLED=0 go build ./cmd/fhl-mcp
```

預設測試使用 fixtures，不會連到信望愛網站。

## 文件

- [doc/快速開始.md](doc/快速開始.md)：安裝、連接與日常指令
- [app_versions.md](app_versions.md)：版本更新紀錄
- [doc/FHL-MCP-Go-重寫計畫書.md](doc/FHL-MCP-Go-重寫計畫書.md)：重寫計畫與設計決定
- [doc/次經-整合指南.md](doc/次經-整合指南.md)、[doc/使徒教父著作-整合指南.md](doc/使徒教父著作-整合指南.md)：次經與使徒教父的信望愛 API 整合筆記

## 授權

[GPL-3.0-or-later](LICENSE)。內含的第三方函式庫與原 Python 版的聲明見發佈包內的 `THIRD_PARTY_NOTICES.md`。聖經與信望愛網站內容的著作權不在此授權範圍內，內容來源為[信望愛站](https://www.fhl.net/)。
