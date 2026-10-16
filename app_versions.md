# app_versions

版本更新紀錄（對應 FHLNUI 的 `app_versions.json`）。最新的在最上面；`TODOs` 區塊永遠放第一。

- last: **0.0.0.16**
- 版次格式：`主.次.修.小`（例 `0.0.0.1`）。小改動升最後一碼；整理一批小版本時，另立一個較高版次並註明整理範圍。
- 發版時：先更新本檔，再 commit，標題為 `發版 {版次}：{摘要}`（見 `.claude/CLAUDE.md` 的 commit 慣例）。
- 程式內版本字串在 `internal/version/version.go`，須與 `last` 一致 (有測試檢查)；`fhl-mcp version` 會顯示。

## 寫法（給使用者看 vs. 給工程師看）

- 給使用者看的項目：`🆕 新功能`、`✨ 優化`、`🐞 Bug`、`ℹ️ 說明`。用白話寫，非工程師也能懂（例：「Claude 與 Codex 都能支援了」）。
  `新功能`、`優化` 寫得短；`說明` 是比較詳細的描述，同樣白話。
- 給工程師看的細節一律標 `📝工程師筆記:`，放在該版本的最尾部。若有次分類，寫成 `📝工程師筆記-✨新增`、`📝工程師筆記-🐞發現`、`📝工程師筆記-ℹ️說明` 等。
- TODOs 區塊也用白話寫；技術細節放在版本內的工程師筆記。

## TODOs

- ✨ 優化: 上游暫時失敗時自動重試，並控制對信望愛網站送出請求的速度（目前有快取與合併相同請求，但還沒有重試與限速）
- 🆕 新功能: 補齊其餘 9 個查詢功能（目前「查經文」「查整章」「查引用」「書卷清單」「書卷資訊」「版本清單」「版本篩選」「搜尋經文」「進階搜尋」「查原文字典」「查原文編號出處」「逐字原文分析」「結構分析參考資料」「查註釋」「註釋書清單」「搜尋註釋」「主題查經」「有聲聖經連結」「有聲聖經加經文」「有聲聖經版本清單」可用，其他功能已列在清單上，但呼叫會回「尚未實作」）
- ✨ 優化: 在 Codex 實際連線驗證（這台電腦只找到 Codex 桌面版，沒有命令列工具，尚未測）
- ✨ 優化: 兩個 AI 工具同時連線、伺服器重開後自動重連的測試
- ✨ 優化: 確認之後的版本會在 macOS 13 以上的 Apple 晶片 Mac 與 Linux 上實際跑得起來（目前只有在 Windows 上實測，Mac／Linux 只確認能編譯出程式）
- ⏳ 待決定: 「搜尋註釋」遇到太常見的字（如「愛」）時，上游會拒絕並說結果超過 500 筆；原 Python 版把它當成 0 筆，Go 版已核准改為明確回報，錯誤訊息的措辭請確認 (D12／D22)
- ⏳ 待決定: 是否讓本程式只講跟原 Python 版相同的 MCP 協定版本 (2025-11-25)，不啟用更新的版本

## 0.0.0.16

- 🆕 新功能: 可以查某一章的有聲聖經（mp3／ogg 連結與上一章、下一章），支援和合本、台語、客家話、廣東話、希伯來文、希臘文等 19 種錄音版本。
- 🆕 新功能: 可以同時取得整章的有聲聖經連結與經文（也可選簡體）。
- 🆕 新功能: 可以列出所有有聲聖經版本。
- ℹ️ 說明: 目前 29 個功能裡已有 20 個真正能用，其餘 9 個（次經、使徒教父、註腳、信望愛文章）的名稱還在清單上，但呼叫時會明確回覆「尚未實作」。有聲聖經只提供信望愛網站的連結，不會下載或轉發音檔；查不存在的章（例如約翰福音 99 章）時，網站仍會給出連結，但連結可能打不開（與原 Python 版相同）。更新後需要把舊的 `fhl-mcp serve` 關掉再重新啟動。
- 📝工程師筆記-✨新增: M2-f：`get_audio_bible`、`get_audio_chapter_with_text`、`list_audio_versions`；`fhlapi.AU` (`au.php`，欄位保留 raw JSON 因上游數字為字串，`endpointTTL` 7 天)、`internal/bible/audio.go` (`AudioAPI` 介面、靜態版本表、`GetAudio`／`GetAudioWithText`；導覽依 Python 的真值判斷 `pbid`&`pchap`／`nbid`&`nchap`，`audio_files` 只放有值的連結)。`get_audio_chapter_with_text` 的經文部分重用 `GetChapter`。
- 📝工程師筆記-ℹ️說明: 驗證：`go vet`、`CGO_ENABLED=0 go test ./...` 通過。新增 10 份 golden：3 份 (`list_audio_versions`、`get_audio_bible_john3`、`get_audio_chapter_with_text_phm1`) 與 M0 實際樣本逐位元組相同，其餘 7 份 (創 1、啟 22、台語、希伯來文、約 99、次經 101) 由 Python 自己的 `tools/audio.py` 在錄下的 `au.php` fixture 上執行產生 (`tests/oracle/audio_*`，先與實測樣本核對相符)。Claude Code 經 MCP 實測三個工具與無效版本錯誤皆正常。未驗證：簡體的經文以外部分只做 `au.php gb=1` 實測 (書名轉簡體)、其他錄音版本的實際連結是否可播。
- 📝工程師筆記-🐞發現: D23 (`au.php` 對不存在的章仍回 success；創 1／啟 22／次經 101 的導覽怪異值照上游原樣；字串型數字不轉型；`use_simplified` 在 with_text 依 D4 也送到音訊請求；無效 `audio_version` 與 D3 的處理)。詳見 `doc/m0/tool-matrix.md`。

## 0.0.0.15

- ✨ 優化: 查過的經文、搜尋、字典、註釋等結果會暫存在這台電腦上，同樣的查詢再問一次幾乎立刻回答，也不會再打擾信望愛網站；多個查詢同時問同一件事時，只會向網站問一次。
- 🆕 新功能: 新增 `fhl-mcp cache clear` 指令，可以一鍵清除暫存；也可以用 `--data-dir` 指定資料放哪裡、`--cache-max-mb` 限制暫存最多占多少空間（預設 200 MB，設 0 就不暫存）。
- ℹ️ 說明: 暫存只放在你的電腦，不會隨程式發佈。經文、原文分析、註釋、主題、版本清單約保留 7 天，搜尋約 1 天，原文字典（Strong's）不過期；空間超過上限時會自動刪掉最久沒用的。查詢失敗或網站回覆異常時不會被暫存，下次會重新詢問。保留天數是我的暫定值，想調整請告訴我。更新後需要把舊的 `fhl-mcp serve` 關掉再重新啟動。
- 📝工程師筆記-✨新增: `internal/cache`：每筆一個檔案 (`<資料目錄>/cache/v1/<sha256(key)>.json`，內含 `v`、`key`、`cached_at`、`expires_at`、`body`)、原子寫入、損壞／過期／key 不符視為 miss 並刪除、容量只在第一次寫入時掃描一次目錄、超過上限刪最久未使用至 90% (讀取時 best-effort 更新 mtime)、`Do` 以 key 合併並行請求 (leader 的 context 被取消時 follower 自行重試)。`fhlapi.Client.Cache` 為 nil 時不快取；`getJSON` 拆成 `fetch` (HTTP) 與快取層，key = 完整上游網址 (含 `gb` 與所有參數)，只存合法 JSON 且 `status=="success"` 的回應。`endpointTTL`：qb／qp／sc／st／ab 7 天、se／ssc 1 天、sd 永久。CLI：`--data-dir`／`FHL_DATA_DIR`、`--cache-max-mb`／`FHL_CACHE_MAX_MB`、`cache clear`。
- 📝工程師筆記-ℹ️說明: 驗證：`go vet`、`CGO_ENABLED=0 go test ./...` 通過 (快取單元測試含 TTL／損壞／原子寫入／容量／並行合併／leader 取消／Clear；客戶端測試含重複請求、各參數獨立、失敗不快取、並行合併；CLI 與設定解析測試)。以 MCP 實測：同一查詢第二次不產生新檔，兩筆請求各留一個檔案，`cache clear` 清除 2 個。未驗證：真實上游的重複請求次數 (沒有計數器可見)、多個程序同時寫同一目錄、macOS／Linux、`-race`。尚未做：重試與限速、`doctor`、設定檔。TTL 與 200 MB 上限為暫定，待你確認。

## 0.0.0.14

- ✨ 優化: 「結構分析參考資料」現在單次最多查 30 節，超過時會直接告訴你「請分成較小的範圍」，不會再默默對信望愛網站連續送出一大堆請求。
- ℹ️ 說明: 30 節是你決定的上限（每一節都要向信望愛網站查一次，目前還沒有快取）。節數用逗號、範圍或跨章寫法都一樣計算；超過就整個拒絕，不會只回一部分。
- 📝工程師筆記-ℹ️說明: D21(b) 決定並寫入計畫書 §9 與 `doc/m0/tool-matrix.md`。`internal/bible/words.go`：`MaxParsingVerses = 30`；`expandVerseString`／`seq` 在展開前以算術判斷，極大範圍 (如 `1-9223372036854775807`) 不會配置記憶體；跨章範圍橫跨 30 章以上不查章節長度，逐章累計超過即停；錯誤為 `Invalid parameter 'verse': …`，且發生在任何 `qp.php` 請求之前。測試：`TestParsingReferenceVerseCap` (邊界 30 通過、31 拒絕、溢位／極大值)、`TestParsingReferenceCap` (工具層，確認未呼叫 `qp.php`)；既有 golden 不受影響。

## 0.0.0.13

- 🐞 Bug: 「進階搜尋」指定書卷範圍時，現在也可以用希臘文／希伯來文原文編號搜尋（例如只在新約找 G25），不再只限關鍵字。
- ℹ️ 說明: 這項修正是你核准的（原 Python 版說明裡寫可以用，實際在指定範圍時卻被擋掉）。沒有指定範圍的搜尋本來就可以，行為不變。
- 📝工程師筆記-ℹ️說明: D19 核准並寫入計畫書 §9 與 `doc/m0/tool-matrix.md`；行為早已在 M2-c 實作 (`SearchAdvanced` 三種 `search_type` 皆可，上游 `RANGE=3` 搭配 `orig=1`／`2` 實測可用)。本版補上 `hebrew_number` 範圍搜尋的請求對應測試 (`orig=2`、`RANGE=3`、`range_bid=1`、`range_eid=39`)。

## 0.0.0.12

- 🐞 Bug: 「書卷資訊」查次經（如多俾亞傳）時，不再被標成「新約」，改標「次經」(Apocrypha)；查使徒教父（如革利免前書）改標「使徒教父」(Apostolic Fathers)。聖經 66 卷的標示不變。
- ℹ️ 說明: 這項修正是你核准的（原 Python 版把舊約以外的書卷全標成新約）。這會讓這個功能的輸出和原 Python 版不同，只影響次經與使徒教父。「書卷清單」沒有變動。
- 📝工程師筆記-ℹ️說明: D18 核准並寫入計畫書 §9 與 `doc/m0/tool-matrix.md`；`internal/bible/bookinfo.go` 新增 `testamentOf` (1–39 `OT`、101–115 `Apocrypha`、201–217 `Apostolic Fathers`、其餘 `NT`；範圍同 `validNumericID`，表中沒有的有效碼如 113、115 也依範圍標示)。測試：`TestBookInfoTestamentLabels`；`TestBookInfoEdgeCases` 的 115 期望值由 `NT` 改為 `Apocrypha`。66 卷的 golden 不受影響。

## 0.0.0.11

- 🐞 Bug: 查「提多書」時不會再誤查成「多俾亞傳」；不論寫「提多書」「Titus」或簡寫「多」，都是提多書。要查多俾亞傳請寫「多俾亞傳」或「Tob」。
- ℹ️ 說明: 這項修正是你核准的（原 Python 版的書卷簡寫「多」同時對應兩本書，結果會查到多俾亞傳）；所有以書卷名稱查詢的功能都套用同一套規則。這一版沒有其他功能變動。
- 📝工程師筆記-ℹ️說明: D17 核准並寫入計畫書 §9、`doc/m0/tool-matrix.md`；行為早已在 M1 實作 (`ResolveBookID` 以全名／英文名直接解析)，本版新增 `TestTitusIsNotTobit` (`Titus`／`提多書`／`多`／`56` → bid 56，`Tob`／`多俾亞傳`／`101` → bid 101) 當回歸測試。仍待決定：D18、D19、D21、D12 措辭、協定版本。

## 0.0.0.10

- 🆕 新功能: 可以查某一節的註釋（也可選簡體），並列出有哪些註釋書。
- 🆕 新功能: 可以在註釋裡搜尋關鍵字，看哪些書卷章節的註釋提到它。
- 🆕 新功能: 可以查主題查經（Torrey、Nave 的主題資料，中英文皆可），也能只查有幾筆。
- ✨ 優化: 在註釋裡搜尋太常見的字（例如「愛」）時，會明確告訴你「結果超過 500 筆，請縮小關鍵字」，不再當成沒有結果。
- ℹ️ 說明: 目前 29 個功能裡已有 17 個真正能用，其餘 12 個（有聲聖經、次經、使徒教父、註腳、信望愛文章）的名稱還在清單上，但呼叫時會明確回覆「尚未實作」。
- ℹ️ 說明: 查註釋時如果沒給節數，信望愛網站不會回傳整章註釋，所以結果是空的（與原 Python 版相同）。這一版仍沒有快取；更新後需要把舊的 `fhl-mcp serve` 關掉再重新啟動。
- 📝工程師筆記-✨新增: M2-e：`get_commentary`、`list_commentaries`、`search_commentary`、`get_topic_study`，經 `fhlapi.SC`／`SCList`／`SSC`／`ST` 查 `sc.php`／`ssc.php`／`st.php`；`use_simplified` 依 D4 生效 (實測 `sc.php` 會轉簡體)；HTTP 契約無 `commentary_id`／`topic_id`，不暴露。
- 📝工程師筆記-ℹ️說明: 驗證：`go vet`、`CGO_ENABLED=0 go test ./...` 通過。新增 11 份 golden：4 份 (`list_commentaries`、`get_commentary_john316`、`get_topic_study_love`、`get_topic_study_love_count`) 與 M0 實際樣本逐位元組相同，其餘 7 份由 Python 自己的 `tools/commentary.py` 在錄下的 fixture 上執行產生 (對照器 `tests/oracle/commentary_*`，先與 4 份實測樣本核對相符)。Claude Code 經 MCP 實測 4 個工具皆正常。
- 📝工程師筆記-🐞發現: D12 套用於 `ssc.php` (`Fail:record count > 500` 回參數錯誤，措辭為我的假設，待確認)；D22 (無 `verse` 時回 0 筆且 `verse:null`；Python 想輸出的 `navigation` 因 prev/next 在上游最上層而從不出現，Go 同樣不輸出；搜尋結果 `book` 欄保留上游補的 44 字元空白；`total_count` 為回傳筆數)。詳見 `doc/m0/tool-matrix.md`。

## 0.0.0.9

- 🆕 新功能: 可以查整章經文，也可以直接用「約 3:16-17」這種寫法查經文。
- 🆕 新功能: 可以列出所有書卷、查單一書卷的資料；可以列出所有聖經版本，並依「有沒有原文編號」「舊約或新約」篩選。
- 🆕 新功能: 可以搜尋經文（用關鍵字，或用希臘文／希伯來文原文編號），也能限定在某幾卷書內搜尋。
- 🆕 新功能: 可以查原文字典（例如 G25、H430），也能查某個原文編號在聖經裡出現在哪些經文。
- 🆕 新功能: 可以逐字查某一節的原文分析，以及產生給結構分析用的原文參考資料（含單節、多節、跨章）。
- ✨ 優化: 搜尋時如果沒輸入關鍵字，會明確告訴你「不能是空的」，不再冒出看不懂的錯誤字樣。
- ℹ️ 說明: 目前 29 個功能裡已有 13 個真正能用，其餘 16 個（註釋、主題查經、有聲聖經、次經、使徒教父、註腳、信望愛文章）的名稱還在清單上，但呼叫時會明確回覆「尚未實作」。
- ℹ️ 說明: 這一版還沒有快取，每次查詢都會直接問信望愛網站；「結構分析參考資料」查很多節時會連續送出很多請求，請先用小範圍。更新後需要把舊的 `fhl-mcp serve` 關掉再重新啟動，才會用到新功能。
- ℹ️ 說明: 有三件事等你決定要不要採用（見上方 TODOs）：提多書的查詢修正、次經書卷的約別標示、進階搜尋指定範圍時也允許用原文編號。
- 📝工程師筆記-✨新增: M2-a (81070b2)：`get_bible_chapter`、`query_verse_citation` (沿用 Python 正規式，書名與章之間需空白，D10)、`get_book_list`、`get_book_info`；D11 查明 Python `BOOK_ALIASES` 不在 29 個 HTTP 工具的呼叫路徑上，故不另做別名表。
- 📝工程師筆記-✨新增: M2-b (dea2b65)：`list_bible_versions`、`search_available_versions` (`fhlapi.AB`)；實測上游 `ab.php` 不理會 `gb`，簡體選項不影響版本名稱。
- 📝工程師筆記-✨新增: M2-c (a5478a7)：`search_bible`、`search_bible_advanced`、`lookup_strongs`、`search_strongs_occurrences` (`fhlapi.SE`／`SD`)；空查詢回參數錯誤 (D7)；`search_bible` 契約無 `offset` 參數，收到也忽略；範圍搜尋缺一端或為 0 時照 Python 退回全聖經搜尋。
- 📝工程師筆記-✨新增: M2-d：`get_word_analysis`、`get_parsing_reference` (`fhlapi.QP`)；`internal/bible/parsingfmt.go` 為 Python `utils/parsing_format.py` 的移植 (token id、`+w+u+` 異文、希伯來文 RLI／PDI／LRM、舊約行序反轉)；圖例字串由 Python 原始碼產生 (`legend.go`)；只支援正典 1–66 卷；跨章範圍以 `qb.php` 取章節數。
- 📝工程師筆記-ℹ️說明: 驗證：`go vet`、`CGO_ENABLED=0 go test ./...` 通過。golden 共 22 份輸出與 Python 一致：其中 18 份與 M0 實際樣本逐位元組相同，其餘 4 份為對照器產生；`get_word_analysis` (創 1:1)、`get_parsing_reference` (雅 1:1-2、創 1:1、得 1:22–2:1) 由 Python 自己的 `parsing_format.py` 在錄下的 fixture 上執行產生 (對照器 `tests/oracle/`，先與兩份實測樣本核對相符)，舊約與跨章路徑因此是對 Python 程式碼而非實際 Python 伺服器驗證。Claude Code 經 MCP 實測各批工具皆成功，且伺服器重啟後自動重連。
- 📝工程師筆記-🐞發現: D17 (`Titus`／提多書 → `多` → 多俾亞傳，待核准)；D18 (`get_book_info` 對次經與使徒教父的 `testament` 都回 `NT`，目前照原樣，待核准)；D19 (範圍搜尋只接受 `keyword`，Go 改為三種型別皆可，待核准)；D20 (`total_count` 等於回傳筆數而非總命中數，照原樣)；D21 (`qp.php` 對不存在的節回不合法 JSON；`get_parsing_reference` 逐節請求無上限，待決定)。詳見 `doc/m0/tool-matrix.md`。
- 📝工程師筆記-ℹ️說明: 尚未做：`internal/cache`、重試與限流 (D13)；其餘 16 個工具 (`sc.php`／`ssc.php`／`st.php` 註釋與主題、`au.php`、`qsub.php`／`sesub.php`、`qaf.php`／`seaf.php`、`rt.php`、文章 `json.php` 與靜態清單類)；Codex 驗證；`-race`；macOS／Linux 執行；雙客戶端並行。

## 0.0.0.8

- 🆕 新功能: 可以啟動本程式並在 Claude Code 查經文（例如「約 3:16」），結果與原本的 Python 版完全一樣。
- ℹ️ 說明: 目前只有「查經文」這一個功能真正能用；其餘 28 個功能的名稱與說明已經列出來，但呼叫時會明確回覆「尚未實作」。啟動方式是在命令列執行 `fhl-mcp serve`，Claude Code 要先註冊一次（`claude mcp add --transport http --scope user fhl http://127.0.0.1:8081/mcp`）。
- ℹ️ 說明: 只接受本機連線，並擋掉其他網站從瀏覽器發來的請求。Codex 還沒驗證，Mac 與 Linux 版本尚未實機測試。
- 📝工程師筆記-✨新增: M1 垂直切片：`cmd/fhl-mcp` (`serve`／`version`／`help`；`--host` 僅限 loopback、`--port` 預設 8081 被占用即失敗、`--log-level`、Ctrl+C 優雅關閉)；`internal/{app,transport,mcpserver,tools,bible,fhlapi,version}` 與 `assets/` (內嵌凍結的 `tools-list.json`)。
- 📝工程師筆記-✨新增: `internal/mcpserver` 以低階 `AddTool` 逐字註冊契約 29 個工具；成功回 text + `structuredContent{result}`，錯誤回 `isError` 且文字前綴 `Error executing tool X: `；未知工具以 receiving middleware 轉成同形式 (D2)；capabilities 只有 tools (`listChanged:false`)。serverInfo 名稱 `FHL-MCP-Go`、版本為產品版本 (Python 為 `FHL Bible MCP Server`／`1.29.0`)，為假設。
- 📝工程師筆記-✨新增: `get_bible_verse` 真實實作：參數 null 視同未給 (D5)、整數接受數字字串 (D6)、錯誤訊息自訂簡短 (D2)、不認得的書卷訊息同 D3、`use_simplified` 真正送 `gb=1` (D4)；輸出以 `MarshalIndent2` 對齊 Python `json.dumps(indent=2, ensure_ascii=False)`，3 個 Python 樣本逐位元組相同。
- 📝工程師筆記-✨新增: 防護：SDK 預設擋非 localhost 的 `Host`，`http.NewCrossOriginProtection` 擋非同源 `Origin`；`/health` 不連上游；根路徑與其他路徑回 404 (不做根路徑相容)。
- 📝工程師筆記-✨新增: 測試 (離線，fixtures 為 4 份實際抓取的 `qb.php` 回應與 3 份 Python golden)：書卷解析、上游 client (httptest)、經文服務、參數驗證、經真實 loopback 監聽與官方 SDK client 的端到端 (initialize／tools/list 與契約逐項相等／查經文／錯誤形式／Host 與 Origin／port 占用／優雅關閉)、版本字串與本檔 `last` 一致。`CGO_ENABLED=0` 已交叉編譯 windows/amd64、darwin/arm64、linux/amd64。
- 📝工程師筆記-🐞發現: D17：Python `BIBLE_BOOKS` 中文簡寫「多」同時是提多書 (56) 與多俾亞傳 (101)，後者覆蓋前者；`get_bible_verse` 先轉中文簡寫再轉回 bid，`Titus`／`提多書` 推測會查到 101 (靜態推斷與離線 `get_book_id('多')=101`，未對 Python 實際重現)。Go 版皆解析為 56，實測 `Titus 1:1` 回 `book_eng: Titus`。待核准，記於 `doc/m0/tool-matrix.md`。
- 📝工程師筆記-ℹ️說明: 驗證：`go vet` 與 `go test` 通過；Windows 版實際啟動並由 Claude Code 新對話呼叫 `get_bible_verse` 成功 (使用者驗證)。未驗證：Codex、`-race` (此環境無 cgo)、macOS／Linux 執行、同時雙客戶端、重啟後重連、`stateless` 與 stateful 模式取捨。
- 📝工程師筆記-ℹ️說明: 假設事項：模組路徑 `github.com/snowray712000/FHL-MCP-Go`；SDK 預設 stateful 模式；未收窄 `SupportedProtocolVersions`。詳見 `code/ai-log/261016p`。

## 0.0.0.7

- ℹ️ 說明: 專案仍在打底階段，還沒有可以執行的程式。這一版先把開發用的工具和套件版本定下來，也實際試過「呼叫不存在的工具會怎麼回應」「擋掉不是自己電腦發出的連線」這兩件事，確認做法可行。
- ℹ️ 說明: Mac 的最低版本由 macOS 12 改為 macOS 13（Apple 晶片）。原因是新版 Go 工具鏈不再支援 macOS 12；這是你核准的取捨。
- 📝工程師筆記-✨新增: 鎖定 `go 1.27.2` 與 `github.com/modelcontextprotocol/go-sdk v1.8.0` (`go.mod` / `go.sum`)；模組路徑 `github.com/snowray712000/FHL-MCP-Go` 為假設 (計畫書未規定、repo 尚無 remote)，日後可改。
- 📝工程師筆記-✨新增: 計畫書 §1 / §15 / 發佈表與 `.claude/CLAUDE.md` 改為 macOS 13+ arm64、以 Go 1.27.x 建置 (Go 1.26 是最後支援 macOS 12 的版本；依據為網路搜尋結果，來源是 golang/go 的第三方鏡像與 Go wiki，需對照 go.dev/doc/go1.27 確認；macOS 13+ 真機尚未驗證)。
- 📝工程師筆記-✨新增: 試驗 (a) D2：SDK 對未知工具預設回 JSON-RPC `-32602` 協定錯誤；以 `AddReceivingMiddleware` 攔截 `tools/call` 可改回 `isError:true` 結果 (`Error executing tool X: Unknown tool: X`)，與原版相容。已實測。
- 📝工程師筆記-✨新增: 試驗 (b) D1：以低階 `Server.AddTool` 直接載入契約 `tools-list.json`，回傳 `Content` 文字 + `StructuredContent{result}`；`tools/list` 與 Python 契約 29 個工具逐項 JSON 相等。低階 `AddTool` 不驗證輸入，錯誤格式與型別放寬需自行實作，不用泛型 `AddTool[In,Out]`。
- 📝工程師筆記-✨新增: 試驗 (c) Host／Origin：SDK 預設擋非 localhost 的 `Host` (403)，但不擋 `Origin`；以 `http.NewCrossOriginProtection().Handler(h)` 包住後，惡意 Origin／其他 port 的 localhost Origin／`Sec-Fetch-Site: cross-site` 皆 403，無 Origin 的 CLI 客戶端與同源皆 200。試驗程式放 scratchpad，不在 repo。
- 📝工程師筆記-ℹ️說明: 尚未驗證：真實 Claude Code／Codex 連線、macOS 實機、`-race`、跨平台編譯；`initialize` 回應 (capabilities 含 `logging`) 與 `initialize.json` 的差異待做正式 server 時比對。詳見 `code/ai-log/261016m`、`261016n`。

## 0.0.0.6

- 📝工程師筆記-✨新增: M0 收尾：計畫書 §15 全部定案 — 專案名稱 `FHL-MCP-Go`；首版只交付 29 Tools；平台 Ubuntu 22.04/24.04 x64 + macOS 12+ arm64 (最終 macOS 最低版 M1 鎖定) + Windows x64；首版 beta 不簽章、不公證；不做 Smithery 與根路徑 `/` 相容 (只提供 `/mcp`)
- 📝工程師筆記-✨新增: `.claude/CLAUDE.md` 同步 §15 定案與 D2–D12 核准；`code/ai-log/261016l` 記錄決策與 M1 試驗清單
- 📝工程師筆記-ℹ️說明: M0 退出條件全部滿足 (來源 manifest、契約基準、Prompts／Resources 盤點、工具矩陣與差異清單、範圍與平台定案)；仍無 Go 程式碼，下一步為 M1

## 0.0.0.5

- 📝工程師筆記-✨新增: `doc/m0/tool-matrix.md`：29 個工具契約矩陣 (實測上游端點、輸出欄位、備註) 與原版差異清單 D1–D16
- 📝工程師筆記-✨新增: `doc/m0/contract/tool-samples.json` (55 次呼叫、32 次上游請求的原始樣本)、`contract/upstream-ssc-samples.json` (D12 上游回應)
- 📝工程師筆記-✨新增: 計畫書 §9「已核准的原版差異」：D2–D12 經使用者核准 (錯誤格式維持 `isError`、不認得書卷統一訊息、`use_simplified` 真正生效並納入快取 key、`null`／數字字串放寬、查無資料形狀保留、上游 `status` 非 success 回明確訊息)；D10 先照原樣
- 📝工程師筆記-🐞發現: Python 原版 `use_simplified` 在多數工具被忽略；`InvalidParameterError` 建構錯誤不只 `get_book_info`；`search_commentary` 常見詞被上游 `record count > 500` 吞成 0 筆
- 📝工程師筆記-ℹ️說明: 實測上游一律為 `bible.fhl.net/api/*.php` (文章為 `www.fhl.net/api/json.php`)，未使用 `/json/`；M1 須確認 Go SDK 能讓未知工具回 `isError` 結果；仍無 Go 程式碼

## 0.0.0.4

- 📝工程師筆記-✨新增: `doc/m0/prompts-resources-inventory.md`：盤點 19 個 Prompts (4/3/4/5/3、45 個參數) 與 7 個 Resource URI (4 個 scheme)，並記錄 stdio `server.py` 的曝光方式與原版缺陷 (路徑未 `unquote`、`int()` 例外、`info://books` 忽略 `simplified`)
- 📝工程師筆記-✨新增: `doc/m0/contract/prompts-inventory.json`：19 個 Prompts 的參數與渲染結果樣本 (M4 可作 golden)
- 📝工程師筆記-✨新增: 決定「首版只交付 29 個 Tools，Prompts 與 Resources 移到下一版本 (M4)」；計畫書 §1 加入首版完成定義 (M0–M3、M5、M6)，§4、§5、§12、§13、§15、§17 與 `.claude/CLAUDE.md` 同步
- 📝工程師筆記-ℹ️說明: Prompts／Resources 只存在於 stdio 的 `server.py`，HTTP 入口為 0 項；其餘 27 個工具尚未取樣輸出；仍無 Go 程式碼

## 0.0.0.3

- 📝工程師筆記-✨新增: M0 契約基準 `doc/m0/contract/`：Python HTTP server 實測匯出 `initialize`、`tools/list` (29 個工具含 inputSchema / outputSchema)、空的 prompts / resources 列表、3 筆 `tools/call` 樣本，附 `README.md` 整理事實
- 📝工程師筆記-✨新增: 決定輸出格式：Go 版同時輸出 text content 與 `structuredContent{result}`，並宣告相同 `outputSchema` (計畫書 §9、`.claude/CLAUDE.md` 已同步)
- 📝工程師筆記-🐞發現: Python 原版 `get_book_info` 查無書卷時回 `InvalidParameterError.__init__() missing ... 'value'` (例外建構錯誤)，Go 版歸為明確 bug fix，錯誤格式待定
- 📝工程師筆記-ℹ️說明: HTTP 入口的 Prompts / Resources 皆為 0 項 (19 Prompts、7 類 Resources 僅存在於 stdio 的 `server.py`)；其餘 27 個工具尚未取樣輸出；仍無 Go 程式碼

## 0.0.0.2

- 📝工程師筆記-✨新增: M0 來源 manifest `doc/m0/source-manifest.tsv`：Python 原專案 254 個檔案的 SHA-256 與 Git 狀態 (237 未動 / 7 已修改 / 10 未追蹤；HEAD `03e9cb0`)
- 📝工程師筆記-✨新增: `doc/m0/source-manifest.md`：manifest 範圍、未追蹤檔案說明、已核對事實 (HTTP 入口 29 個工具名稱與計畫書一致)
- 📝工程師筆記-ℹ️說明: 仍無 Go 程式碼；尚未匯出 `tools/list` schema，也尚未盤點 Prompts／Resources

## 0.0.0.1

- 📝工程師筆記-✨新增: 重寫計畫書 `doc/FHL-MCP-Go-重寫計畫書.md` (討論版 v0.1)
- 📝工程師筆記-✨新增: `.claude/CLAUDE.md`：專案現況、ai-log 流程、架構限制、git commit 慣例 (沿用 FHLNUI)
- 📝工程師筆記-✨新增: `code/ai-log/`：AI 工作紀錄與 `index.md` 索引
- 📝工程師筆記-✨新增: 本檔 `app_versions.md`
- 📝工程師筆記-ℹ️說明: 尚無任何 Go 程式碼，此版僅為規劃與專案骨架
