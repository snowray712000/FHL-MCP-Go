# M0 工具契約矩陣與差異清單（29 Tools）

- 取樣日期：2026-10-16
- 來源：Python HTTP server（`127.0.0.1:18081`，mcp 1.29.0），工作目錄狀態見 [source-manifest.md](source-manifest.md)
- 原始樣本：[contract/tool-samples.json](contract/tool-samples.json)（55 次呼叫；`structuredContent` 與 `content[0].text` 完全相同時，以一句說明取代以縮小檔案）
- 工具 schema：[contract/tools-list.json](contract/tools-list.json)；契約整體事實：[contract/README.md](contract/README.md)

## 取樣方式與上游流量

- 55 次 `tools/call`（29 個工具各至少 1 次，另有參數驗證、錯誤路徑、快取、繁簡、文章內容選項）。
- **實際送出的上游請求共 32 次**：單一程序、依序執行、有上游請求時間隔 ≥ 1.5 秒；User-Agent 為原版的 `FHL-Bible-MCP-Server/0.1.0`。
  其餘呼叫由本機靜態資料、快取或參數驗證回應，未打上游。
- 快取寫在 scratchpad（`.cache` 相對路徑，從 scratchpad 啟動），原專案的 `.cache` 未變動（最後修改時間早於本 session）。
- 上游 URL 由 `httpx` 日誌取得，因此下表的端點是**實測**而非讀碼推測。

## 上游端點（實測）

**全部使用 `https://bible.fhl.net/api/*.php`；文章用 `https://www.fhl.net/api/json.php`。本次取樣沒有任何請求使用 `/json/` 路徑**（計畫書 §3 原本需驗證的疑點，結論：實際行為以 `/api/` 為準；`api/client.py` 預設 `/json/` 的值在實際工具路徑中沒被用到）。查詢字串一律帶 `gb=`（預設 0）。

## 每個工具

「輸出」欄為成功時 JSON 頂層欄位（`content[0].text` 為 `indent=2`、`ensure_ascii=false` 的 JSON 字串，除非標註）。

| 類別 | 工具 | 上游端點 | 輸出 | 備註 |
| --- | --- | --- | --- | --- |
| 經文 | `get_bible_verse` | `qb.php`（`bid`,`chap`,`version`,`strong`,`sec`,`gb`） | `version`,`version_name`,`record_count`,`verses`,`navigation` | 不給 `verse` = 整章；`verse` 可為 `16-17` 等；`include_strong` 時文字內嵌 `<WG25>` 標記 |
| 經文 | `get_bible_chapter` | `qb.php`（無 `sec`） | 同上 | 與不給 `verse` 的 `get_bible_verse` 輸出相同 |
| 經文 | `query_verse_citation` | `qb.php` | 同上 | 引用需符合 `書卷 章:節`，**書卷與章之間必須有空白**（見 D10） |
| 搜尋 | `search_bible` | `se.php`（`orig=0`,`RANGE=0`,`limit`,`offset`…） | `total_count`,`query`,`search_type`,`scope`,`limit`,`offset`,`results` | 空字串查詢回錯誤 `'record'`（見 D7） |
| 搜尋 | `search_bible_advanced` | `se.php`（`RANGE=3`,`range_bid`,`range_eid`） | `total_count`,`query`,`search_type`,`range`,`limit`,`offset`,`results` | |
| 原文 | `get_word_analysis` | `qp.php`（`bid`,`chap`,`sec`） | `testament`,`book`,`chapter`,`verse`,`original_text`,`translation`,`word_count`,`words` | 含希臘文與變音符號 |
| 原文 | `get_parsing_reference` | `qp.php` | **非 JSON：Markdown 文字** | 以 `# 參考資料: 原文` 開頭（見 D15） |
| 原文 | `lookup_strongs` | `sd.php`（`N=0/1`,`k`） | `strongs_number`,`original_word`,`chinese_definition`,`english_definition`,`testament`,`related_words` | `strongs_number` 為補零字串 `"00025"`；字串內含 `\r\n` |
| 原文 | `search_strongs_occurrences` | `se.php`（`orig=1`,`RANGE=1`） | `strongs_info`,`occurrences` | |
| 註釋 | `get_commentary` | `sc.php`（`bid`,`chap`,`sec`） | `book`,`chapter`,`verse`,`commentary_count`,`commentaries` | |
| 註釋 | `list_commentaries` | `sc.php?validbook=1` | `total_count`,`commentaries` | 7 部 |
| 註釋 | `search_commentary` | `ssc.php`（`key`） | `keyword`,`total_count`,`results` | 常見詞（如「愛」）因上游 `record count > 500` 被吞成 0 筆（見 D12） |
| 主題 | `get_topic_study` | `st.php`（`N=4`,`count_only`,`keyword`） | `keyword`,`source`,`total_count`,`results`；`count_only` 時只有前三欄 | 「愛」完整結果約 20 KB |
| 版本 | `list_bible_versions` | `ab.php` | `total_count`,`versions` | 86 個版本，約 17 KB |
| 版本 | `search_available_versions` | `ab.php` 資料的本機篩選（本次由快取命中） | `total_count`,`filters`,`versions` | `filters` 含 `testament: null` |
| 書卷 | `get_book_list` | 無（靜態資料） | `total_count`,`testament`,`books` | 只測 `NT`，預設 `all` 未取樣 |
| 書卷 | `get_book_info` | 無（靜態資料） | `id`,`testament`,`names` | `Phm` 失敗、`Philemon`／`門` 成功（見 D3、D11） |
| 有聲 | `get_audio_bible` | `au.php`（`version`,`bid`,`chap`） | `version`,`version_name`,`book`,`book_eng`,`chapter`,`audio_files`,`navigation` | 只回 `media.fhl.net` 連結；`chapter` 是字串 `"3"` |
| 有聲 | `get_audio_chapter_with_text` | `au.php`，文字部分（本次由快取命中） | `audio`,`text` | `Phm` 失敗（見 D3） |
| 有聲 | `list_audio_versions` | 無（靜態資料） | `total_count`,`versions` | 19 個，`id` 16 缺號 |
| 次經 | `get_apocrypha_verse` | `qsub.php` | `status`,`query_type`,`book`,`book_id`,`chapter`,`verse`,`version`,`verse_count`,`verses` | 書卷 id 101–109；`verse` 未給時為 `null` |
| 次經 | `search_apocrypha` | `sesub.php`（`q`,`offset`,`limit`） | `status`,`query_type`,`keyword`,`total_count`,`returned_count`,`offset`,`results` | |
| 次經 | `list_apocrypha_books` | 無（靜態資料） | `status`,`query_type`,`id_range`,`book_count`,`books` | 9 本，含縮寫清單 |
| 教父 | `get_apostolic_fathers_verse` | `qaf.php` | 同次經 verse | 書卷 id 201–217 |
| 教父 | `search_apostolic_fathers` | `seaf.php` | 同次經 search | |
| 教父 | `list_apostolic_fathers_books` | 無（靜態資料） | 同次經 list | 8 本 |
| 註腳 | `get_bible_footnote` | `rt.php`（`bid`,`id`,`version=tcv`,`gb`） | `status`,`query_type`,`version`,`version_name`,`book_id`,`book_name`,`footnote_id`,`text` | 查無時 `status: "not_found"` + `message`（成功路徑） |
| 文章 | `search_fhl_articles` | `www.fhl.net/api/json.php`（`gb`,`title`…） | `status`,`query_type`,`total_count`,`returned_count`,`include_full_content`,`articles` | `limit` 不送上游；`articles[].id` 為字串 |
| 文章 | `list_fhl_article_columns` | 無（靜態資料） | `status`,`query_type`,`column_count`,`columns` | 12 欄 |

## 跨工具行為與差異清單

分類依計畫書 §12：相容／明確 bug fix／核准變更／未解決。**D2–D12 已於 2026-10-16 經使用者核准（D10 先照原樣；D12 調查後核准「回明確訊息，不當 0 筆」），並寫入計畫書 §9「已核准的原版差異」。D1 為先前已決定；D13–D16 為相容性描述，不需決策。**

| ID | Python 實測行為 | 建議的 Go 處理 | 分類／狀態 |
| --- | --- | --- | --- |
| D1 | 成功回傳同時有 `content[0].text` 與 `structuredContent.result`（同字串），宣告 `outputSchema {result:string}` | 同樣輸出 | 相容；**已決定**（plan §9） |
| D2 | 工具錯誤是 `isError:true` 的**結果**（不是 JSON-RPC 錯誤），文字前綴 `Error executing tool {name}: …`，無 structuredContent；**未知工具名也是這種形式**（`Unknown tool: x`）；參數驗證錯誤是 Pydantic 原文（含 `errors.pydantic.dev` 連結） | 保留 `isError:true` + 前綴；驗證訊息改為自訂簡潔文字 | 相容（結構）／變更（驗證文字）；**已核准**（2026-10-16）。M1 須確認 Go SDK 能讓未知工具回 `isError` 結果而非 JSON-RPC 錯誤，否則另案 |
| D3 | `get_book_info`、`get_audio_chapter_with_text` 遇到不認得的書卷，回 `InvalidParameterError.__init__() missing 1 required positional argument: 'value'`（例外建構錯誤，實測）；`get_bible_verse`／`get_bible_chapter` 則正確回 `Invalid parameter 'book': X - Book not found. Please check the spelling or use a valid book code.`。靜態稽核：`InvalidParameterError` 需要 `(parameter, value, reason)`，但單行單參數的呼叫點至少有 `tools/audio.py:72`、`commentary.py:39`／`141`／`201`、`info.py:117`、`strongs.py:141`，推測 `get_commentary`、`search_commentary`（空關鍵字）、`get_topic_study`（空關鍵字）、`lookup_strongs`／`search_strongs_occurrences`（未知書卷）也有同樣缺陷（未逐一實測，跨行呼叫點未稽核） | 全部工具統一用後者格式 | 明確 bug fix；**已核准** |
| D4 | `use_simplified` 在多數工具被接受但**沒有送出 `gb=1`**：`get_bible_verse` 實測 URL 為 `gb=0` 且輸出仍是繁體。只有 `get_bible_footnote`、`search_fhl_articles` 實測有送 `gb=1`（`endpoints.py` 另有一處 1334 行附近未驗證）。快取 key 也不含繁簡：`get_bible_chapter`／`search_bible` 帶 `use_simplified=true` 時由先前繁體結果的快取回應 | 實作 `gb=1` 並把繁簡納入快取 key | 明確 bug fix（行為變更）；**已核准** |
| D5 | `verse` 的 schema 為 `type: string, default: null`，但**明確傳 `null` 會被拒絕**（`Input should be a valid string`） | 接受 `null` 視同未提供（Claude／Codex 可能送 null） | 放寬（相容的超集）；**已核准** |
| D6 | 整數參數接受數字字串（`chapter:"3"` 成功），非數字（`"abc"`）驗證失敗 | 同樣接受數字字串 | 相容；**已核准** |
| D7 | 「查無資料」表示不一致：`get_bible_verse` 章數過大 → 成功、`record_count:0`、`verses:[]` 加 `navigation`；`lookup_strongs` 不存在 → 成功、`strongs_number:""`、其餘 `null`；`get_bible_footnote` → `status:"not_found"`；`search_bible` 空查詢 → 洩漏 `KeyError` 文字 `'record'` | 前三者保持原形狀；空查詢改為清楚的參數錯誤 | 相容（前三）／bug fix（空查詢）；**已核准** |
| D8 | 兩種輸出家族：舊工具（版本／搜尋／原文等）無 `status`；新工具（次經／教父／註腳／文章）有 `status`、`query_type` | 逐工具保持原形狀，不統一 | 相容；**已核准** |
| D9 | 型別不一致且須保留：音訊 `chapter` 為字串，經文 `chapter` 為整數；`strongs_number` 補零字串；文章 `id` 字串 | 逐欄位照樣輸出，以 golden 測試固定 | 相容；**已核准** |
| D10 | `query_verse_citation` 正規式 `^(.+?)\s+(\d+):(\d+(?:-\d+\|,\d+)*)$`：`約3:16-17` 失敗，`約 3:16-17` 成功 | 先照原樣；是否放寬為接受無空白，另議 | **已核准：先照原樣**（需要空白）；放寬與否日後再議 |
| D11 | 書卷別名表是資料：`Phm` 不被認得，`Philem`、`Philemon`、`門` 可以 | M2 從原始碼擷取完整別名表作為 fixture | 相容；**已核准**；M2 工作。**M2 查證（2026-10-16）**：Python `BOOK_ALIASES`（約 150 個簡體／縮寫別名）只被 `normalize_book_name`／`parse_reference`／`fuzzy_search` 使用，這三者**不在** 29 個 HTTP 工具的呼叫路徑上；實際解析只看 `BIBLE_BOOKS` 四種名稱＋數字碼，所以「保留中英別名行為」= 現有 `ResolveBookID`，不需另做別名表。若日後要支援簡體／縮寫別名，屬新功能，須另行核准 |
| D12 | `search_commentary` 以「愛」「耶穌」查詢回 0 筆。**原因已查明**：上游 `ssc.php` 對常見詞回 HTTP 200 的 `{"status":"Fail:record count > 500"}`，Python 找不到 `record` 就輸出 `total_count:0, results:[]`，把上游限制吞掉；冷僻詞「尼哥底母」回 `status:"success"`、`record_count:5`、`record[]`（欄位 `bchap,bid,book_name,bsec,chinesef,echap,engs,esec,tag,title`）。另：HTTP 版未暴露 `commentary_id` 參數 | 上游 `status` 非 `success` 時回明確訊息（`Fail:record count > 500` → 「結果超過 500 筆，請縮小關鍵字」），不當作 0 筆；以 `isError: true` 回報（我的實作假設），先適用 `search_commentary`，其他端點 M2 核對後再套用 | 明確 bug fix；**已核准**（2026-10-16） |
| D13 | 快取有效：相同 `get_bible_verse` 第二次呼叫 0 個上游請求（見 `cache_repeat_john316`） | 依計畫書 §9 的快取設計，行為對照 TTL | 相容；M2 |
| D14 | 文字 JSON 為 `indent=2`、`ensure_ascii=false`，字串內保留上游的 `\r\n` | 同樣的序列化設定 | 相容；建議 |
| D15 | `get_parsing_reference` 輸出為 Markdown（非 JSON），仍包在 `outputSchema {result:string}` 裡 | 同樣；以 James 1:1 fixture 驗證 | 相容；已有 fixture |
| D16 | 工具結果沒有分頁、沒有 `nextCursor`；`tools/list` 一次回 29 項 | 同樣 | 相容 |
| D17 | **新發現（M1，靜態稽核＋Go 實測）；已核准（2026-10-16，0.0.0.11）**：Python 的 `BIBLE_BOOKS` 中文簡寫 `多` 同時是提多書(56)與多俾亞傳(101)，索引字典後者覆蓋前者，而 `get_bible_verse` 先把書名轉成中文簡寫再轉回 bid：`Titus`／`提多書` → `多` → bid 101，推測會查到多俾亞傳（未實際對 Python 重現，只做離線 `get_book_id('多')=101` 驗證）。Go 版以書卷全名／英文名直接解析為 bid，`Titus`／`提多書`／`多` 皆為 56；實測 Go 版 `Titus 1:1` 回 `book_eng: Titus`。 | 屬明確 bug fix：提多書查提多書；`多` 單獨出現時解析為提多書(56)，多俾亞傳請用 `Tob`／`多俾亞傳`／`101` | 明確 bug fix；**已核准**；M2 其他工具照同一解析器 |
| D18 | **新發現（M2-a，靜態稽核＋golden 對照）；已核准修正（2026-10-16，0.0.0.12）**：`get_book_info` 的 `testament` 只有 `OT`／`NT` 兩種值，Python 以 `1<=id<=39` 判斷，**次經(101–112)與使徒教父(201–216)都被標成 `"NT"`**。另外有效但表中沒有的數字碼（如 `113`）會回 `names` 全為 `null` 且 `testament:"NT"`。`get_book_list` 不帶篩選時也包含這些書（共 86 筆） | M2-a 先**照原樣**輸出（byte 相容）；是否改為更正確的值（例如 `Apocrypha`／`Apostolic Fathers`）需核准，會改變輸出 | 明確 bug fix；**已核准**，0.0.0.12 起次經(101–115)回 `Apocrypha`、使徒教父(201–217)回 `Apostolic Fathers`，其餘不變 |
| D19 | **新發現（M2-c）；已核准（2026-10-16，0.0.0.13）**：`search_bible_advanced` 在指定書卷範圍時只有 `search_type:"keyword"` 能用：`greek_number`／`hebrew_number`（工具說明列出的值）被擋，`greek`／`hebrew` 通過第一關卻在端點層被擋。未指定範圍時（走一般搜尋）三種值都可用 | Go 在所有路徑都接受 `keyword`／`greek_number`／`hebrew_number`（實測上游 `RANGE=3` 搭配 `orig=1` 可用） | 明確 bug fix；**已核准** |
| D20 | **發現（M2-c；相容，照原樣）**：`search_bible`／`search_bible_advanced`／`search_strongs_occurrences` 的 `total_count` 取自上游 `record_count`，套用 `limit` 時等於實際回傳筆數，**不是**全部命中數（`limit 3` → 3；`H430 limit 2` → 2）；`showing` 因此恆等於 `total_count` | 先照原樣；若要真正總數，可評估上游 `count_only=1`（未測） | 相容；資訊 |
| D21 | **發現（M2-d；資訊）**：(a) `qp.php` 對不存在的節（如約 99:1）回**不合法的 JSON**（`"bid":,`），Go 以「上游回應不是合法 JSON」錯誤回報；(b) `get_parsing_reference` 每節一次上游請求（跨章再加每章一次 `qb.php` 查節數），依序進行，目前無上限、無快取，大範圍會產生大量請求；(c) 上游回 0 筆時 Python 會 `IndexError`，Go 回「no original-language data」；(d) 兩個工具只支援正典 1–66 卷，其他書卷回明確錯誤 | 快取與限流於 M2 後段處理；(b) 發版前需決定單次範圍上限 | 資訊；(b) 待決定 |
| D22 | **發現（M2-e；資訊，照原樣）**：(a) `get_commentary` 不給 `verse` 時，上游 `sc.php` 回 0 筆，輸出為 `"verse": null`、`commentary_count: 0`（沒有整章註釋；Go 不送 `sec`，Python 送空值，結果相同）；(b) 原版想輸出的 `navigation`（上一節／下一節）讀的是每筆記錄的 `prev`／`next`，但上游放在最上層，所以原版輸出從來沒有 `navigation`，Go 同樣不輸出；(c) `search_commentary` 結果的 `book` 欄保留上游補到 44 字元的尾端空白（與原版逐位元組相同）；(d) `search_commentary` 的 `total_count` 是回傳筆數；(e) 上游 `sc.php` 會依 `gb` 轉簡體（書名、標題、內文），`use_simplified` 依 D4 生效；(f) `ssc.php` 的 `Fail:record count > 500` 套用 D12，Go 以參數錯誤回報：`Invalid parameter 'keyword': 愛 - too many results (more than 500 records), use a more specific keyword (結果超過 500 筆，請縮小關鍵字)`（措辭為我的假設，待確認）；(g) HTTP 契約沒有 `commentary_id`／`topic_id` 參數，Go 也不暴露。 |

## 未涵蓋

- 每個工具只取樣 1–2 組參數；`search_type` 的其他值（如 `greek`、`hebrew`）、`scope`、`source`、`audio_version`、`text_version`、`testament` 篩選、`offset` 分頁等多數選項未逐一測試。
- 未測 `get_book_list` 預設 `all`、`get_topic_study` 其他 `source`、非 `unv` 版本、`get_bible_footnote` 其他版本。
- 未測上游逾時／重試／限流與非 JSON 回應；未量測 wire 層原始 JSON-RPC。
- D4 的 `endpoints.py:1334` 附近函式屬哪個工具、是否送 `gb=1` 未驗證。
- D3 除 `get_book_info`、`get_audio_chapter_with_text` 外，其餘受影響工具為靜態推測，未實測。
- D12 為直接對上游 `ssc.php` 發 3 次請求查證（`愛`、`耶穌`、`尼哥底母`），樣本見 [contract/upstream-ssc-samples.json](contract/upstream-ssc-samples.json)。
