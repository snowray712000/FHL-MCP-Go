# M0 Prompts／Resources 盤點（Python 原專案）

- 盤點日期：2026-10-16
- 來源：`FHL-MCP-Server` 工作目錄（見 [source-manifest.md](source-manifest.md)）；Prompts 以實例化 `PromptManager` 取得，Resources 以靜態閱讀 `resources/handlers.py`、`server.py`。
- 機器可讀資料：[contract/prompts-inventory.json](contract/prompts-inventory.json)（19 項的參數與渲染結果）。

## 先決事實

1. **這兩類功能只存在於 stdio 的 `server.py`，HTTP 入口（29 工具）完全沒註冊**：HTTP 實測 `prompts/list`、`resources/list`、`resources/templates/list` 皆為空
   （見 [contract/README.md](contract/README.md)），`http_server.py` 的 server card 也寫 `"resources": False, "prompts": False`；
   但 `initialize` 的 capabilities 仍宣告 prompts／resources（FastMCP 預設）。
2. 因此 §13 M4 要遷移的是 **stdio 版行為**，沒有 HTTP 對照樣本可比；契約只能從原始碼與渲染結果建立。
3. stdio 與 HTTP 已分歧：`server.py` 的啟動日誌寫「Tools: 27、Prompts: 4」（過時字樣），實際 prompts 為 19；`server.py` 註冊的工具與 HTTP 的 29 項也不同（stdio 約 20 個 `name=` 字樣，未逐一比對）。

## Prompts：19 項（4 / 3 / 4 / 5 / 3，與計畫書 §4 相符）

只有一套實作：`prompts/templates.py` 是向後相容轉出，實際註冊在 `prompts/manager.py`；`PromptTemplate`（`prompts/base.py`）有 `name`、`description`、`arguments`、`render(**kwargs) -> str`。

| 類別 | 名稱 | 說明 | 參數（`*` 必填） | 缺必填參數時 |
| --- | --- | --- | --- | --- |
| 基礎 | `basic_help_guide` | 顯示 FHL Bible MCP Server 的完整使用指南，包含快速入門、工具說明和實用技巧 | `section` | — |
| 基礎 | `basic_uri_demo` | 展示和教導如何使用各種 Resource URI，包含互動式範例和最佳實踐 | `uri_type` | — |
| 基礎 | `basic_quick_lookup` | 快速查詢聖經經文，簡單直接 | `query`* | 拋 `ValueError` |
| 基礎 | `basic_tool_reference` | 顯示所有可用工具的詳細參考手冊，包含用法、參數和範例 | `tool_name`、`category` | — |
| 讀經 | `reading_daily` | 每日靈修讀經指引，包含默想和應用 | `passage` | — |
| 讀經 | `reading_chapter` | 閱讀和分析整章聖經，包含背景、結構、重點和應用 | `book`*、`chapter`*、`version` | 拋 `ValueError` |
| 讀經 | `reading_passage` | 閱讀和分析一段經文（可能跨越多個章節），包含背景、主題、解經和應用 | `book`*、`start_chapter`*、`start_verse`*、`end_chapter`*、`end_verse`*、`version` | 拋 `ValueError` |
| 研經 | `study_verse_deep` | 深入研讀一節經文，包含經文內容、原文分析、註釋等 | `book`*、`chapter`*、`verse`*、`version` | 拋 `ValueError` |
| 研經 | `study_topic_deep` | 研究聖經中的特定主題 | `topic`*、`version`、`max_verses` | 拋 `ValueError` |
| 研經 | `study_translation_compare` | 比較不同聖經譯本的翻譯 | `book`*、`chapter`*、`verse`*、`versions` | 拋 `ValueError` |
| 研經 | `study_word_original` | 深入研究希臘文或希伯來文單字 | `strongs_number`*、`testament`*、`max_occurrences` | 拋 `ValueError` |
| 特殊 | `special_sermon_prep` | 輔助講道準備，從釋經到應用的完整流程 | `passage`*、`sermon_type` | 拋 `ValueError` |
| 特殊 | `special_devotional` | 引導靈修默想，加深與神的關係 | `passage`、`theme` | — |
| 特殊 | `special_memory_verse` | 協助經文背誦，提供記憶技巧和應用提醒 | `verse`* | 拋 `ValueError` |
| 特殊 | `special_topical_chain` | 追蹤特定主題在聖經中的發展脈絡，建立主題經文鏈 | `topic`*、`testament` | 拋 `ValueError` |
| 特殊 | `special_bible_trivia` | 生成聖經知識問答題目，適合小組活動或教學 | `category`、`difficulty` | — |
| 進階 | `advanced_cross_reference` | 深入研究經文的交叉引用，建立經文網絡 | `reference`* | 拋 `ValueError` |
| 進階 | `advanced_parallel_gospels` | 比較四福音書的平行記載，分析異同與神學重點 | `event`*、`references` | 拋 `ValueError` |
| 進階 | `advanced_character_study` | 全面研究聖經人物，包含生平、性格、事蹟、屬靈教訓 | `character`、`focus`、`testament`、`version` | — |

共 45 個參數；參數 dict 只有 `name`／`description`／`required`，**沒有 default、型別、列舉值**（預設值寫在各 prompt 的 `render` 內，未曝光給客戶端）。

### 渲染行為（`server.py` 的 `get_prompt`）

- 回傳 `GetPromptResult`：`description` 固定為 `Rendered prompt: {name}`，`messages` 只有 1 則 `role: user`、`TextContent`，內容為 `render()` 的字串。
- 缺必填參數 → `validate_arguments` 拋 `ValueError("Missing required argument: …")`；名稱不存在 → `render_prompt` 回 `None`，再拋 `ValueError("Unknown prompt: …")`。
- 所有 `arguments` 值都以字串傳入；`chapter`、`max_verses` 等數字參數是否轉型，屬各 `render` 內部行為，尚未逐項核對。
- 19 項在「只給必填參數」時都能渲染，輸出 249–886 字元不等；完整文字在 `prompts-inventory.json`，可作為 M4 的 golden 樣本。

## Resources：7 個 URI（4 個 scheme，2 + 1 + 1 + 3）

由 `ResourceRouter`（`resources/handlers.py`）依 scheme 分派，內部直接呼叫對應的 tool 函式（非另走一條 API 路徑）。

| URI 範本 | 範例 | 實際呼叫 | 可選 query |
| --- | --- | --- | --- |
| `bible://verse/{version}/{book}/{chapter}/{verse}` | `bible://verse/unv/John/3/16` | `get_bible_verse` | `strong`、`simplified`（`true` 才為真） |
| `bible://chapter/{version}/{book}/{chapter}` | `bible://chapter/unv/Gen/1` | `get_bible_chapter` | `simplified` |
| `strongs://{testament}/{number}` | `strongs://nt/25` | `lookup_strongs` | `simplified` |
| `commentary://{book}/{chapter}/{verse}` | `commentary://John/3/16` | `get_commentary` | `commentary_id`、`simplified` |
| `info://versions` | 同左 | `list_bible_versions` | `simplified` |
| `info://books` | 同左 | `get_book_list` | `testament`、（`simplified` 讀取後未使用） |
| `info://commentaries` | 同左 | `list_commentaries` | `simplified` |

### `server.py` 的曝光方式與讀取行為

- `list_resources` 把**範例 URI**（如 `bible://verse/unv/John/3/16`）當作 `uri`、把範本字串當作 `name`，`mimeType` 一律 `application/json`；**沒有註冊 resource templates**。
  即客戶端只看得到 7 個固定範例，看不到參數化範本，與計畫書 §4「區分固定 Resources 與 Resource Templates」的要求有落差。
- `read_resource` 回傳 `json.dumps(result["content"], ensure_ascii=False, indent=2)`；`content` 是 tool 函式的回傳物件（結構化內容），
  不是 HTTP 工具那種「再包成字串」的格式，細節待取樣。
- scheme 不支援、`bible://` 類型不是 verse／chapter、`info://` 路徑不是 versions／books／commentaries、`strongs://` 約別不是 `OT`／`NT`（大小寫不拘）→ `ResourceError`（繼承 `FHLAPIError`）。
- 解析細節（與一般 URL 慣例不同，Go 版需決定是否相容）：`bible://verse/...` 的類型取自 URL host 位置；`strongs://nt/25` 的約別取自 host 並轉大寫；`commentary://John/3/16` 的書卷取自 host。

### 原版缺陷／風險（尚未決定是否保留）

- 路徑區段沒有 `unquote`（`grep` 0 筆）：書卷名若是百分比編碼的中文，不會被解碼。
- `chapter`、`verse`（commentary）、`commentary_id` 直接 `int()`，非數字會拋原始 `ValueError`，不是 `ResourceError`；`bible://verse` 的 `verse` 則保留字串（可含範圍如 `16-18`）。
- `info://books` 的 `simplified` 被解析卻未傳入；`list_resources` 的 `name`／`uri` 欄位語意顛倒（name 放範本）。
- 錯誤訊息混用中文與格式字串；未實測實際回傳。

## 結論與待使用者決定

- 計畫書「19 Prompts + 7 類 Resources」的數量與分類**經盤點屬實**，但「7 類」其實是 4 個 scheme、7 個 URI。
- 這兩類功能在 HTTP 路徑是 0，且無真實 HTTP 契約；是否列入首版仍是 §15 第 2 項待定，**未替使用者決定**。
- 若納入：Go SDK 應用 resource templates 正式曝光 4 個範本、3 個固定資源，並決定是否沿用原版範例式清單。

## 未涵蓋

- 未啟動 stdio 版實測 `prompts/get`、`resources/read` 的 wire 輸出；Prompts 渲染文字未逐篇審閱品質。
- 各 `render()` 對數字／空字串參數的處理未逐項核對；stdio 版 `server.py` 與 HTTP 的工具差異未比對。
