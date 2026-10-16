# FHL-MCP-Go 重寫計畫書

- 日期：2026-10-16（Asia/Taipei）
- 版本：討論版 v0.1
- 原專案：`D:\coding\信望愛\FHL-MCP-Server`
- 建議新專案：`D:\coding\信望愛\FHL-MCP-Go`（尚未建立）
- 本次交付：規劃文件；尚未實作 Go 程式，未改動原專案或客戶端設定。

## 1. 目標與決策狀態

以 Go 重寫信望愛聖經 MCP 伺服器，提供 Windows 10/11、Apple Silicon Mac 及 Ubuntu 使用者可直接下載執行的版本。一般使用者不需安裝 Docker、Go、Python、Node.js 或編譯工具。

### 已確認需求

1. 採獨立 HTTP 服務，不採 stdio，不以 stdio 代理作為預設連線方式。
2. 主要連接 Claude Code 與 OpenAI Codex 的本機執行環境。
3. 以同一台電腦使用為主；不以區網分享、公網代管為首版目標。
4. 免安裝、按需手動啟動；不預設登錄開機或登入自動啟動，也不安裝常駐服務。
5. 命令列是跨平台核心，無桌面 Ubuntu 必須可完整操作。
6. Windows 與 Mac 儘量提供雙擊入口。
7. 使用者可自行將啟動入口加入系統啟動設定，文件提供範例。
8. 同一使用者與同一執行環境設定一次 MCP，供多個專案共用。

### 建議採用、仍可調整

- 專案名稱 `FHL-MCP-Go`，不另加 Server；產品顯示名「信望愛聖經 MCP 伺服器」。
- 主執行檔 `fhl-mcp`／`fhl-mcp.exe`。
- MCP 網址 `http://127.0.0.1:8081/mcp`。
- 第一里程碑完成現有 HTTP 的 29 個工具；完整重寫目標再包含 19 個 Prompts 與 7 個 Resource URI（4 個 scheme）。
- **已決定（2026-10-16，使用者確認）：首版只交付 29 個 Tools；Prompts 與 Resources 移到下一版本（M4）。** 首版不註冊、不宣告這兩類 capability，文件與發佈說明不得宣稱已支援。
- **首版完成定義**：M0–M3、M5、M6 全部達成（29 工具契約與差異已結案、本機 HTTP 服務、`setup`／`doctor`／`status`／`stop`／`disconnect`、三平台封裝與原生驗收）；M4 不影響首版驗收。
- 首版平台基準（**已定案 2026-10-16**，見 §15）：Windows x64、macOS 13+ arm64、Ubuntu 22.04/24.04 x64（macOS 13 為 Go 1.27 工具鏈的下限，2026-10-16 使用者核准，見 §15）。

## 2. 產品使用方式

### 第一次使用

1. 下載對應平台壓縮檔並解壓縮。
2. 執行 `fhl-mcp setup`，或雙擊「首次設定」入口。
3. 選擇連接 Claude Code、Codex，或兩者。
4. 工具登錄使用者層級 HTTP MCP，檢查結果並顯示摘要。
5. 執行 `fhl-mcp serve` 或雙擊「啟動」入口。
6. 在客戶端重新載入 MCP 或開啟新工作階段；依實際版本給出指引。

`setup` 僅進行連接設定，不安裝系統服務、不新增自動啟動、不要求系統管理員權限。服務本身不需要客戶端存在才能啟動。

### 平日使用

- 需要查經時啟動程式，服務維持執行，Claude Code 與 Codex 可同時連線。
- 使用結束後按 Ctrl+C 或執行 `fhl-mcp stop`。
- 重新開機後服務不自動執行，已登錄的 MCP 連接設定仍保留。
- 服務未啟動時，客戶端可能顯示未連線；客戶端是否重試及如何重連必須實測，不能保證啟動後所有既有對話立即恢復。

### 各平台入口

| 平台 | 一般啟動 | 雙擊便利入口 | 停止方式 |
| --- | --- | --- | --- |
| Windows | `.\fhl-mcp.exe serve` | `啟動-FHL-MCP.cmd`；`首次設定.cmd` | Ctrl+C／`stop` |
| Apple Silicon Mac | `./fhl-mcp serve` | `Start-FHL-MCP.command`；`Setup.command` | Ctrl+C／`stop` |
| Ubuntu | `./fhl-mcp serve` | 不要求桌面環境 | Ctrl+C／`stop` |

雙擊入口明確保留終端視窗，顯示網址與停止方法；首版不另做原生 GUI、系統匣或網頁管理介面。腳本正確處理中文、空白路徑及自身所在目錄。Mac 腳本與二進位檔需保留執行權限；下載隔離與 Gatekeeper 行為在簽章、公證與實機驗證階段處理。

首版不內建背景 daemon。自行加入系統啟動的進階使用者，應使用各平台服務管理機制，不在登入腳本中直接執行會阻塞的前景指令；相關範例清楚標示為選用，並提供移除步驟。

## 3. 原專案調查基準

2026-10-16 以本機工作目錄靜態檢視；Git HEAD 為 `03e9cb0`，有多個已修改及未追蹤檔案。HEAD 不能代表此次規劃的完整來源。尚未執行原版測試或即時 API 驗證，README 的歷史測試數、覆蓋率與成功宣告不視為本次驗證成果。

| 原始位置 | 已觀察內容 | 遷移意義 |
| --- | --- | --- |
| `pyproject.toml` | Python；MCP、httpx、Pydantic、Starlette、Uvicorn、Smithery 等依賴 | 使用 Go SDK、標準 HTTP、明確資料型別取代 |
| `Dockerfile`、`main.py` | Python HTTP 入口，port 8081，Docker 安裝 gcc | 不因存在 gcc 就推定必須保留 C 依賴；逐項稽核實際需求 |
| `http_server.py` | 29 個 `@mcp.tool()`；Streamable HTTP；根路徑轉向 `/mcp` 的相容處理 | 以實際 MCP 列舉與呼叫結果建立契約 |
| `server.py` | 另一套 stdio 註冊，另有 Resources／Prompts | 抽出功能，不搬運 stdio 啟動邏輯 |
| `http_tools.py` | HTTP 包裝與部分直接資料處理 | 與 tools、API 層合併盤點，避免漏掉包裝層行為 |
| `api/client.py`、`api/endpoints.py` | 上游 HTTP、重試、快取、查詢與解析 | 遷移為單一可注入的 API client |
| `config.py` | 設定檔、環境變數、預設值 | 統一為清楚的 Go 設定優先序 |
| `utils/cache.py` | JSON 檔案快取、多類 TTL | 首版無須引進資料庫 |
| `prompts/manager.py` | 靜態註冊 19 個 Prompts | 後續確認 schema 與實際渲染輸出 |
| `resources/handlers.py` | 經文、章、Strong's、註釋、資訊等 7 類 URI | 以讀取／列舉／模板契約重新整理 |
| `tools/parsing_reference.py`、`utils/parsing_format.py` | 本機未追蹤的原文分析功能 | 必須納入來源快照 |
| `tests/fixtures/parsing_reference_james_1_1.md` | 現有原文分析 fixture | 納入回歸測試基準 |

現有 HTTP 對外監聽 `0.0.0.0`、開放 CORS，並為雲端部署關閉 DNS rebinding 防護。新版本定位為本機工具，應重新設計 Host／Origin 防護，不直接沿用雲端設定。

設定中出現 `/api/`，基礎 API client 預設出現 `/json/`；文章另呼叫 `https://www.fhl.net/api/json.php`。需驗證各功能真正使用的端點，不做全域字串替換。

### 來源保存程序

1. 產生來源檔案清單、SHA-256 與 Git 狀態摘要，記錄盤點日期。
2. 包含已修改與未追蹤的相關原始碼、文件、fixtures；排除虛擬環境、快取及秘密資料。
3. 在新專案保存 manifest 與契約樣本，原 Python 專案保留作為比對參考。
4. 遷移過程原專案若繼續修改，另列差異與納入版本，避免基準漂移。

## 4. 功能遷移清單

以下名稱來自 HTTP 入口靜態註冊，共 29 項。實作前以 `tools/list` 匯出最終 schema，凍結參數型別、required、default、description 與結果格式。

| 類別 | 必須盤點與遷移的工具 |
| --- | --- |
| 經文（3） | `get_bible_verse`、`get_bible_chapter`、`query_verse_citation` |
| 搜尋（2） | `search_bible`、`search_bible_advanced` |
| 原文（4） | `get_word_analysis`、`get_parsing_reference`、`lookup_strongs`、`search_strongs_occurrences` |
| 註釋與主題（4） | `get_commentary`、`list_commentaries`、`search_commentary`、`get_topic_study` |
| 版本與書卷（4） | `list_bible_versions`、`search_available_versions`、`get_book_list`、`get_book_info` |
| 有聲聖經（3） | `get_audio_bible`、`get_audio_chapter_with_text`、`list_audio_versions` |
| 次經（3） | `get_apocrypha_verse`、`search_apocrypha`、`list_apocrypha_books` |
| 使徒教父（3） | `get_apostolic_fathers_verse`、`search_apostolic_fathers`、`list_apostolic_fathers_books` |
| 註腳（1） | `get_bible_footnote` |
| 文章（2） | `search_fhl_articles`、`list_fhl_article_columns` |

保留中文／英文書卷別名、章節範圍、多節引用、譯本代碼、繁簡、Strong's 編號、次經與使徒教父特殊書卷編號、搜尋限制、註腳與文章內容選項。特別測試希臘文、希伯來文、變音符號與原文分析 Markdown 排版。

### Resources 與 Prompts（下一版本範圍，首版不交付）

M0 盤點結果見 `doc/m0/prompts-resources-inventory.md`：這兩類只存在於原版 stdio `server.py`，HTTP 入口實測為 0 項，沒有 HTTP 契約樣本；下列為下一版本的建議範圍。

- Resources：`bible://verse/{version}/{book}/{chapter}/{verse}`、`bible://chapter/{version}/{book}/{chapter}`、`strongs://{testament}/{number}`、`commentary://{book}/{chapter}/{verse}`、`info://versions`、`info://books`、`info://commentaries`。
- Prompts：基礎 4、讀經 3、研經 4、特殊 5、進階 3，共 19 項；盤點名稱、參數、必填性與渲染結果。
- 透過同一 HTTP MCP 入口提供，不新增 stdio；各客戶端是否展示 Resources／Prompts 分別驗收，不以 server 宣告支援等同客戶端完整展示。
- 適當區分固定 Resources 與 Resource Templates，避免僅列出示例 URI 冒充完整參數化介面。
- 已同意首版只交付 Tools：上述兩類明確移入下一版本（M4），完成定義見 §1。為日後擴充預留：`internal/mcpserver` 的註冊來源設計成可再加入 resources／prompts，但首版不建立空殼套件、不宣告空 capability（原版 HTTP 曾宣告 prompts／resources 卻為空列表）。

## 5. Go 架構

採官方 `github.com/modelcontextprotocol/go-sdk/mcp`，配合標準庫 `net/http`、`encoding/json`、`context`、`log/slog`、`embed`。第一階段鎖定可用的 Go 與 SDK 確切版本，提交 `go.mod`、`go.sum`，記錄支援的 MCP 協定與客戶端版本。

```text
Claude Code / Codex
        │ Streamable HTTP，127.0.0.1
        ▼
HTTP 路由、Host/Origin 驗證、生命週期
        ▼
官方 MCP SDK：Tools（首版）／Resources、Prompts（M4，下一版本）
        ▼
聖經領域服務、書卷正規化、資料格式化
        ▼
FHL API client ── 檔案快取
        ▼
信望愛 API
```

建議目錄：

```text
FHL-MCP-Go/
  cmd/fhl-mcp/             # CLI、程序入口
  internal/app/           # 啟動、停止、程序狀態
  internal/config/        # 設定載入、驗證
  internal/transport/     # HTTP MCP 路由與防護
  internal/mcpserver/     # 一份功能註冊來源
  internal/tools/         # 29 個工具 adapter
  internal/resources/     # URI 與模板（M4，下一版本才建立）
  internal/prompts/       # 提示模板與渲染（M4，下一版本才建立）
  internal/bible/         # 書卷、章節、Strong's 等領域處理
  internal/fhlapi/        # 上游請求、回應型別、錯誤與重試
  internal/cache/         # 有上限的檔案快取
  internal/integration/   # Claude Code / Codex 登錄與檢查
  assets/                 # 可 embed 的模板與靜態資料
  tests/fixtures/         # 經審核的上游與輸出樣本
  scripts/                # 建置、打包、雙擊入口
  docs/                   # 快速開始、遷移、疑難排解
  go.mod
  go.sum
  LICENSE
```

不保留兩套功能註冊，不把每個 Python 函式逐行翻譯。領域邏輯與 SDK 分離，上游 client 可用本機 mock 替換。首版目標 `CGO_ENABLED=0`，避免額外 C runtime 部署需求；若依賴引入 cgo，必須先評估並記錄原因。

## 6. HTTP 與程序生命週期

- 正式 MCP 入口 `/mcp`，預設 `127.0.0.1:8081`。
- `/health` 僅回報本機服務健康與版本，不因上游短暫故障就判定程序死亡；上游連通性由 `doctor` 另行檢查。
- `/` 可提供簡單文字指引；原版根路徑 MCP 請求重寫如需保留，作為有測試的相容選項。
- 由 SDK 處理 JSON-RPC、初始化、版本協商及 HTTP 傳輸語意，不自行手刻一般 REST 取代 MCP。
- 依鎖定 SDK 與目標協定決定 session 模式，驗證多客戶端與重連；不直接假設原版 session 實作可沿用。
- 設定 body 大小、header／idle timeout、上游 timeout 與並行上限；讀寫 timeout 需考慮 MCP 串流，不能讓固定短 timeout 截斷合法連線。
- 預設驗證 Host、Origin，不開放任意瀏覽器來源。CORS 不當作身分驗證。
- 不提供無驗證的 HTTP 停機／改設定端點。`stop` 使用限本機的控制通道，或帶高熵控制 token 的 loopback 介面，token 儲存於使用者受限檔案。
- `status` 檢查程序身分、產品版本及設定埠；不能只以某個 port 能回應就認定是本程式。
- 重複啟動顯示「已執行」與既有網址。port 被其他程式占用則清楚失敗，不默默改成新 port 使客戶端設定失效。
- Ctrl+C 觸發 graceful shutdown，取消上游工作，設定最長結束等待，釋放 listener 與程序狀態。
- `stop` 不單靠 PID 殺程序，避免 PID 重用誤停其他軟體；處理 stale lock 與非正常結束。
- 區網／公網、TLS、共享身分認證不列首版功能；日後增加必須另做完整設計。

## 7. CLI 規格與客戶端連接

下列指令為計畫中的產品介面，尚未實作：

| 指令 | 行為 |
| --- | --- |
| 無子指令／`serve` | 前景啟動服務，顯示網址、版本、停止方法 |
| `setup` | 選擇客戶端，登錄使用者層級 HTTP MCP |
| `setup --client claude`／`--client codex` | 僅設定指定客戶端 |
| `status` | 顯示執行狀態與網址，可選 `--json` |
| `stop` | 經身分驗證停止本工具 |
| `doctor` | 本機設定、服務、上游、客戶端登錄分項診斷 |
| `disconnect --client ...` | 僅移除本工具的連接設定，保留其他 MCP |
| `cache clear` | 清除本工具快取，與執行中的讀寫協調 |
| `version`、`help` | 版本、支援平台、參數與使用說明 |

`serve` 支援 `--host`（首版本機限制）、`--port`、`--config`、`--data-dir`、`--log-level`。固定埠與 client 設定同步；改埠後 `doctor` 能識別不一致並提供重新 setup 指引。

### 官方登錄方式

Claude Code 使用 user scope：

```sh
claude mcp add --transport http --scope user fhl http://127.0.0.1:8081/mcp
```

Codex 使用使用者設定：

```sh
codex mcp add fhl --url http://127.0.0.1:8081/mcp
```

Claude Code 的 user scope 儲存於 `~/.claude.json`，可跨專案使用；Codex 預設使用 `~/.codex/config.toml`，也有專案層設定。setup 優先使用實際可用的官方 CLI，檢查版本與 `--help`，不硬編碼所有版本一定相容。

### setup 的可靠性

1. 偵測 CLI 是否存在，分開回報未安裝、找不到 PATH、設定錯誤、無法連線。
2. 沒有 CLI 時顯示精確的手動連接方式；不因客戶端不存在就擅自安裝。
3. 已有相同名稱與網址時視為成功；同名不同網址時顯示衝突，不能直接覆寫。
4. 每個客戶端獨立處理、獨立回報，部分成功後可安全重跑。
5. 不改寫整份設定、不刪除其他 MCP；如必須直接編輯，先備份、結構化合併、原子寫入並可復原。
6. 尊重 `CODEX_HOME` 等客戶端自訂位置；Windows、WSL、SSH 遠端、容器環境分別辨識。
7. 不將登錄成功等同連線成功；另跑 initialize、tools/list 及代表性工具 smoke test。
8. 客戶端既有批准／信任機制仍依該產品規則處理，不透過 setup 關閉。

「一次設定」的範圍是同一帳號與同一執行環境；不能承諾 Windows 設定自動適用 WSL 或遠端主機。Codex 在此指本機 CLI／相關本機客戶端，不承諾雲端執行環境或一般 ChatGPT 網頁能直接連線使用者的 localhost。

## 8. 設定、檔案與升級

採單一 UTF-8 JSON 設定檔，優先序：CLI > `FHL_` 環境變數 > 指定／預設設定檔 > 內建預設值。錯誤型別、非法 port、無效 URL 應顯示可理解的訊息，不悄悄忽略。

免安裝是不用安裝程式與執行環境，不等同完全不寫入使用者資料。

- 預設使用 Go `os.UserConfigDir`／`os.UserCacheDir` 對應平台位置，不依賴啟動時的工作目錄。
- 不要求寫入執行檔所在目錄，避免唯讀、下載、隨身碟權限問題。
- 需要所有檔案隨資料夾攜帶時，可用 `--data-dir ./data`；setup 仍會另外影響客戶端的使用者設定，文件必須說清楚。
- 舊 Python 快取不直接共用，避免 schema 汙染；新快取加格式版本。
- 升級為停止服務、替換程式、再啟動；設定與快取目錄保留。客戶端連接網址不變則無須每次重新登錄。
- 退出使用時先 disconnect，再刪除程式與選擇性清理資料；使用者若自行設定自動啟動，也自行移除相應入口。

## 9. 上游、快取與輸出相容

### 上游請求

- 使用共用 `http.Client` 與連線池，URL query 正確編碼，支援請求取消。
- 依端點建立明確型別與驗證；處理正常 JSON、錯誤 JSON、非 JSON、空資料與異常狀態碼。
- timeout、重試次數、指數退避與總時間上限可控；只重試合適的讀取請求與暫時性錯誤，尊重限流與 `Retry-After`。
- 不以大量即時測試轟炸上游；預設 CI 使用 fixture/mock，真實 API 驗證採低頻 opt-in。
- HTML／Markdown 解析依現有 fixtures 與實際回應驗證，不能假設所有資料皆為純 JSON。
- 音訊功能先保留連結與相關文字語意，不下載或重新分發音檔。

### 快取

- JSON 檔案快取、原子寫入、跨請求同步與同查詢合併，防止同時請求放大流量。
- cache key 包含端點、所有影響結果的參數、繁簡與資料格式版本。
- 原版經文、原文分析、註釋約 7 天；搜尋約 1 天；Strong's 與部分目錄永久。先建立行為對照，再決定永久類別是否調整為有限 TTL。
- 設定磁碟容量上限與逐步清理；不在啟動時遞迴掃描全部快取，也不空閒輪詢全部檔案。
- 損壞快取視為 miss，記錄診斷後重取；失效資料是否容許 fallback 必須有明確標示。
- 快取只限本機使用，不隨 release 打包上游內容；保留原專案的資料來源與版權說明。

### 輸出與錯誤

- 保留工具名稱、參數語意與可觀察結果；空值、空陣列、缺欄位和字串／數字型別分別驗證。
- HTTP 原版以 MCP text content 回傳 JSON 字串（`indent=2`、`ensure_ascii=false`），M0 實測（`doc/m0/contract/`）另確認：每個工具宣告相同的 `outputSchema`（`{"result": string}`，required），成功時 `structuredContent.result` 與 `content[0].text` 是同一個字串。
- **已決定（2026-10-16，使用者確認）**：Go 版同時輸出 text content 與 `structuredContent: {"result": "<同一字串>"}`，並對 29 個工具宣告相同 `outputSchema`；不可只改成 structuredContent，也不可省略任一者。需測試 Claude Code 與 Codex 兩客戶端皆能正常取用。
- 原版錯誤路徑為 `isError: true` 加純文字訊息，且有已知缺陷（如 `get_book_info` 查無書卷時為 Python 例外建構錯誤）；Go 版錯誤格式於 M0 差異表另定，歸為「明確 bug fix」。
- 調整舊 bug 必須記錄預期新行為與測試，避免把所有差異當成可接受。
- 區分 JSON-RPC 協定錯誤、工具執行失敗、查無資料與上游故障；使用 SDK 合適的錯誤表示。
- 有動態欄位的差異測試先正規化時間戳與非決定性資訊，不比較排版空白即可判定所有語意相同。

### 已核准的原版差異（M0，2026-10-16，使用者確認）

依據 M0 實測的差異清單（`doc/m0/tool-matrix.md` D1–D16，樣本在 `doc/m0/contract/`），以下為 Go 版的規格；實作與測試以此為準。

| ID | Go 版行為 |
| --- | --- |
| D2 | 工具錯誤維持 `isError: true` 的結果，文字前綴 `Error executing tool {name}: `，無 `structuredContent`；未知工具名稱也維持此形式。參數驗證訊息改為自訂、簡潔、不含 Pydantic 原文。**M1 須確認 Go SDK 能讓未知工具回 `isError` 結果（而非 JSON-RPC 錯誤），否則另案決定。** |
| D3 | 所有工具遇到不認得的書卷，統一回 `Invalid parameter 'book': X - Book not found. Please check the spelling or use a valid book code.`；不重現原版 `InvalidParameterError` 建構錯誤。 |
| D4 | `use_simplified` 在所有接受它的工具都真正送出 `gb=1`，且繁簡納入快取 key（原版多數工具忽略此參數）。 |
| D5 | 可選參數明確傳 `null` 視同未提供（原版會拒絕）。 |
| D6 | 整數參數接受數字字串（如 `"3"`），非數字才報錯。 |
| D7 | 「查無資料」保持原形狀：章數過大 → 成功、`record_count:0`；`lookup_strongs` 不存在 → 成功、`strongs_number:""` 其餘 `null`；註腳查無 → `status:"not_found"`。`search_bible` 空查詢改為清楚的參數錯誤（原版洩漏 `'record'`）。 |
| D8 | 新舊兩種輸出家族（有／無 `status`、`query_type`）逐工具保持原形狀，不統一。 |
| D9 | 型別差異逐欄位保留（音訊 `chapter` 字串、`strongs_number` 補零字串、文章 `id` 字串），以 golden 測試固定。 |
| D10 | `query_verse_citation` 先照原樣：書卷與章之間須有空白；是否放寬日後再議。 |
| D11 | 書卷別名表於 M2 從原始碼擷取為 fixture，保留中英別名行為。 |
| D12 | 已調查：上游 `ssc.php` 對常見詞回 `status:"Fail:record count > 500"`，原版吞成 0 筆。**已核准（2026-10-16）：上游 `status` 非 `success` 時回明確訊息，不當 0 筆**；`Fail:record count > 500` 對應「結果超過 500 筆，請縮小關鍵字」。實作假設（未經使用者逐字確認）：以 D2 的 `isError: true` 錯誤結果回報；此規則先確定適用 `search_commentary`，其他端點的 `status` 慣例（尤其與 D7 保留的「查無資料」形狀的界線）於 M2 逐端點核對後再套用。 |
| D17 | 已核准（2026-10-16，0.0.0.11）：Python 的中文簡寫 `多` 同時是提多書 (56) 與多俾亞傳 (101)，原版 `Titus`／`提多書` 因而推測會查到多俾亞傳。Go 以書卷全名／英文名直接解析：`Titus`／`提多書`／單獨的 `多` 皆為提多書 (56)；多俾亞傳請用 `Tob`／`多俾亞傳`／`101`。所有以書卷名稱為參數的工具共用同一解析器。 |
| D18 | 已核准（2026-10-16，0.0.0.12）：`get_book_info` 的 `testament`：1–39 為 `OT`、101–115（次經，含表中沒有的有效碼如 113）為 `Apocrypha`、201–217（使徒教父）為 `Apostolic Fathers`，其餘為 `NT`。原版把 1–39 以外全標成 `NT`。`get_book_list` 不變。 |

D1（text + `structuredContent`）已於上方決定；D13–D16 為相容性描述，不需另行決策。D3 的靜態稽核另指出 `get_commentary`、`search_commentary`、`get_topic_study`、`lookup_strongs`、`search_strongs_occurrences` 可能有同樣缺陷（未實測），納入 D3 的測試範圍。

## 10. 效能與資源驗收

不在量測前承諾「不耗資源」。關閉後不應遺留程序、子程序或背景排程；執行但無查詢時不主動抓取聖經、不建立健康輪詢、不預先下載資料。

建議初期量測目標（尚非保證，M1 後校正）：

| 項目 | 暫定目標／方法 |
| --- | --- |
| 暖啟動至 `/health` 就緒 | 參考機 2 秒內；另記錄冷啟動與首次防毒掃描 |
| 空閒 CPU | 就緒後觀察 5 分鐘，平均接近 0%，初始預算 < 0.5% 單核心 |
| 空閒記憶體 | 初始 RSS／working set 預算 < 100 MiB；分別記錄平台與定義 |
| 持續使用 | 100 次代表性查詢後觀察回收與檔案數，檢查成長是否有上限 |
| 停止 | 一般情況 5 秒內正常結束，沒有遺留 listener 或子程序 |
| 登入／開機影響 | 預設不登錄自動啟動；重開機後確認無 FHL 程序 |

記錄 OS、CPU、RAM、工具鏈、測試條件、空閒與查詢負載。上游延遲與本機處理時間分開，不把網路快慢當成 Go 效能。

## 11. 編譯、封裝與發佈

| 發佈目標 | GOOS/GOARCH | 建議交付檔 | 驗收基準 |
| --- | --- | --- | --- |
| Windows 10/11 x64 | windows/amd64 | `fhl-mcp-go_VERSION_windows_amd64.zip` | Win 10 與 Win 11 各自驗證 |
| Apple Silicon Mac | darwin/arm64 | `fhl-mcp-go_VERSION_darwin_arm64.tar.gz` | macOS 13+（Go 1.27 下限） |
| Ubuntu x64 | linux/amd64 | `fhl-mcp-go_VERSION_linux_amd64.tar.gz` | Ubuntu 22.04 與 24.04 |

Mac 使用 tar.gz 以保留執行權限；可另產 ZIP，但需測試解壓縮工具保留 mode。最終封裝形式可在 beta 使用者試用後調整。首版不含 Windows ARM64、Ubuntu ARM64、Intel Mac、Windows 32 位元；之後有需求可擴充。

發佈規格：

1. 以純 Go 建置、內嵌必要模板與固定資料；不將開發環境依賴帶給使用者。
2. 固定 Go 與 SDK 版本，CI 建置三平台，寫入 version／commit／build date。
3. 編譯成功不等於平台相容，各 OS 原生環境執行 smoke test 與客戶端測試。
4. release 含二進位檔、雙擊入口、快速開始、LICENSE、必要第三方聲明、SHA-256 清單。
5. 檢查 release 無 API 憑證、機器個資、本機路徑、快取內容與私人設定。
6. Windows 簽章、Mac 二進位檔及腳本封裝的簽章／公證／Gatekeeper 檢驗列為公開發佈工作；所需憑證與帳號是交付依賴，未具備時標示限制，不指導使用者停用系統防護。**首版 beta 不簽章、不公證（已定案，見 §15）**，本項檢驗於正式公開版執行。
7. 首版手動更新，不新增自動更新常駐程序。
8. 開發測試可使用 Docker，但正式使用流程不得依賴 Docker。

## 12. 測試與完成定義

### 契約與功能

- 匯出原版 Tools schema；29 項名稱、參數與代表性輸出都有對照表及測試。
- 書卷別名、引用範圍、版本、繁簡、Strong's、特殊書卷、註腳與文章選項有案例。
- 原文分析使用現有 James 1:1 fixture，另補 HTML 變形、缺資料與 Unicode 案例。
- 首版不含 Prompts／Resources；測試需確認 `initialize` 未宣告這兩類 capability。M4 納入時，每項再補列舉、讀取／渲染、錯誤參數測試。
- 所有差異分為相容、明確 bug fix、核准變更或未解決，不能以測試總數代替盤點。

### MCP 與 HTTP

- initialize、tools/list、tools/call、適用的 capabilities，以及協定協商可正常運作。
- Claude Code 與 Codex 同時連線，不互相污染參數、繁簡設定與 session。
- 驗證 Host／Origin、無效請求、斷線取消、重啟重連、port 衝突、重複啟動與優雅停止。
- 控制介面無 token、錯誤 token、其他使用者與錯誤程序不能停機。

### 使用流程與平台

- 三平台乾淨環境不安裝 Docker、Go、Python、Node.js 即能啟動。
- Windows 10/11、Mac arm64、Ubuntu 22.04/24.04 有明確 pass／fail 記錄。
- 中文／空白／唯讀路徑、非系統管理員帳號與不同工作目錄可用。
- 在專案 A setup 後於專案 B 連線成功，沒有新增每專案設定檔。
- 客戶端未安裝、同名衝突、部分登錄成功、重跑 setup 及 disconnect 都有驗證。
- 關機再啟動作業系統後不自行執行，MCP 登錄仍存在；手動啟動後可依文件重連。
- 發佈包執行權限、Mac Gatekeeper 與 Windows 下載檔案執行體驗有實測紀錄。

開發檢查以 `go test ./...`、`go vet ./...` 為基礎；並行共享資料在適用原生環境跑 race 檢查。真實上游整合測試獨立啟用，避免測試依賴上游可用性。公開版本需完成弱點檢查及結果審查。

## 13. 開發階段與交付物

| 階段 | 工作 | 退出條件／交付 |
| --- | --- | --- |
| M0 來源與契約 | 保存工作目錄 manifest、匯出工具 schema、盤點 fixtures、確認完整範圍 | 可追溯遷移矩陣與差異清單 |
| M1 最小垂直切片 | Go SDK、HTTP、經文查詢、serve、雙客戶端手動連線 | 三平台編譯且原生 smoke，兩客戶端成功呼叫同一服務 |
| M2 核心功能 | 經文、搜尋、原文、書卷、版本、註釋與快取 | fixture 比對與錯誤案例通過 |
| M3 完整工具 | 次經、教父、註腳、文章、音訊，完成 29 工具 | schema 與語意差異已結案 |
| M4 擴充協定功能（**延後至下一版本，不屬首版**） | 遷移 19 Prompts、7 個 Resource URI（含 Resource Templates） | HTTP 與實際客戶端能力驗收；首版 M5、M6 不等待 M4 |
| M5 使用體驗 | setup、doctor、status、stop、disconnect、雙擊入口、說明 | 跨專案只設定一次，關機不自啟，停止無遺留 |
| M6 發佈驗收 | 三平台完整封裝、資源量測、簽章／公證、beta 使用者試用 | 乾淨機器驗收報告與可下載套件 |

原則上每階段可獨立展示，不等最後才測客戶端或跨平台。工期在 M0 後按實際 schema、解析複雜度、可取得測試機與憑證重新估算；目前不提供缺乏依據的日期承諾。

## 14. 風險與處置

| 風險 | 處置 |
| --- | --- |
| 原始碼與 README 數量、版本資訊不一致 | 以來源 manifest、MCP 列舉與測試固定基準 |
| 本機未提交新功能被遺漏 | 明確包含未追蹤解析器、fixture 與文件 |
| Go／MCP SDK 協定演進 | 鎖版，記錄版本矩陣，實測客戶端，升級另跑相容測試 |
| 上游回應不穩定或格式變動 | fixtures、型別驗證、低頻 smoke、可診斷錯誤 |
| 輸出看似相同但語意變動 | schema 與語意差異比對，特別處理 null、字串與數字 |
| Windows／WSL／遠端 localhost 混淆 | doctor 顯示目前環境，文件分開說明，不自動暴露區網 |
| 使用者忘記啟動服務 | 雙擊入口與簡單診斷；文件清楚說明按需啟動，不暗中常駐 |
| 搬移執行檔使自訂自動啟動失效 | 選用自啟文件要求固定路徑，升級與移除步驟成對提供 |
| Mac／Windows 下載執行阻擋 | 正式發佈簽章與實機驗證，beta 明示狀態 |
| 停止控制誤殺其他程序 | 驗證控制 token 與產品身分，不只依 PID |
| 快取內容、版權或磁碟成長 | 僅本機快取、有上限、可清除、保留資料來源聲明，不隨包發佈 |

## 15. 待定項目與建議預設

以下不阻擋本機按需啟動架構，但實作前應在 M0 定案：

1. ~~專案命名~~ **已定案（2026-10-16）**：`FHL-MCP-Go`（使用者確認採用）。
2. ~~完整功能範圍~~ **已定案（2026-10-16）**：首版只交付 29 Tools；Prompts、Resources 移到下一版本（M4）。
3. ~~Ubuntu 最低版與 CPU~~ **已定案（2026-10-16）**：Ubuntu 22.04／24.04 x64；ARM64 列後續需求。
4. ~~macOS 最低版~~ **已定案（2026-10-16，修訂）**：macOS 13+ arm64，以 Go 1.27.x 建置。原訂 12+，但 Go 1.26 是最後支援 macOS 12 的版本、Go 1.27 起需 macOS 13（Ventura）以上；使用者選擇升工具鏈、放棄 macOS 12。此限制須寫入 README 與發佈說明；實機驗證（macOS 13+ arm64）仍待 M1 / M5 執行
5. ~~正式發佈憑證~~ **已定案（2026-10-16）**：首版為 beta，**不簽章、不公證**；文件標示 Windows SmartScreen 與 macOS Gatekeeper 可能攔阻，不指導停用系統防護。正式公開版再補 Windows 簽章與 Apple 公證（憑證與帳號由發佈者提供）。
6. ~~Smithery／舊根路徑相容~~ **已定案（2026-10-16）**：不列首版；只提供 `/mcp`，不做 Smithery well-known 端點與每請求設定中介層，也不做根路徑 `/`→`/mcp` 轉向。若有既有使用者再另案列相容工作。

§15 全部定案，M0 退出條件中的「確認完整範圍」已滿足。

## 16. 參考資料

以下官方資料於規劃期間查閱；實作啟動時應再次核對版本、指令與平台限制。

- [MCP 官方 Go SDK](https://github.com/modelcontextprotocol/go-sdk)：官方實作、支援版本與功能文件。
- [MCP Streamable HTTP 傳輸規格（2025-11-25 基準）](https://modelcontextprotocol.io/specification/2025-11-25/basic/transports)：傳輸與本機 Host／Origin 防護基準；實作不限定只支援此歷史版本，依鎖定 SDK 與客戶端協商。
- [Go 最低系統要求](https://go.dev/wiki/MinimumRequirements)：工具鏈與 OS 相容性。
- [Go FAQ：連結與執行依賴](https://go.dev/doc/faq#Why_no_dependencies)：純 Go 與 cgo 的差異。
- [Claude Code MCP 文件](https://code.claude.com/docs/en/mcp#user-scope)：user scope 可跨專案使用。
- [Codex MCP 官方文件](https://learn.chatgpt.com/docs/extend/mcp?surface=cli)：HTTP MCP 與使用者設定。

## 17. 執行檢查清單

- [ ] 鎖定需求、名稱、範圍與平台基準。
- [ ] 保存原始工作目錄 manifest 與未提交功能。
- [ ] 建立新專案、Go／SDK 鎖版與 CI。
- [ ] 先完成經文查詢到兩客戶端的 HTTP 垂直切片。
- [ ] 完成 29 Tools 契約與差異測試。
- [ ] ~~完成 Resources／Prompts~~ 移至下一版本（M4），首版不含。
- [ ] 完成檔案快取、設定優先序及可讀錯誤。
- [ ] 完成 setup、doctor、status、stop、disconnect。
- [ ] 完成 Windows／Mac 雙擊入口與 Ubuntu 無桌面說明。
- [ ] 完成原生平台與跨專案連線測試。
- [ ] 量測空閒、負載、啟動與停止資源。
- [ ] 確認預設不自啟、不常駐、不遺留程序。
- [ ] 完成簽章／公證或標記尚未具備的發佈限制。
- [ ] 交付 release 套件、校驗碼、快速開始及驗收報告。
