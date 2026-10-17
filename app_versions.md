# app_versions

版本更新紀錄（對應 FHLNUI 的 `app_versions.json`）。最新的在最上面；`TODOs` 區塊永遠放第一。

- last: **1.1.2.2**
- 版次格式：`主.次.修.小`（例 `1.1.2.1`），第三碼的奇偶有意義（使用者定，2026-10-17）：
  - 第三碼奇數（如 `1.1.1`、`1.1.3`、`1.2.1`）是正式發佈版，三段式；只有正式版會做發佈包、發到 GitHub Release 與 npm。
  - 第三碼偶數（`1.1.2.x`）是小變更的開發版，給研發用，不做正式發佈；每次小變更升最後一碼。
  - 累積到一個段落，整理 `1.1.2.*` 為 `1.1.3`；有重要功能時整理為 `1.2.1`（同樣整理 `1.1.2.*`）。
  - npm 只接受三段式版本，所以開發版不會發到 npm。
  - 舊的 `0.0.0.x` 已在 `1.1.1` 整理完畢。
- 發版時：先更新本檔，再 commit，標題為 `發版 {版次}：{摘要}`（見 `.claude/CLAUDE.md` 的 commit 慣例）。
- 程式內版本字串在 `internal/version/version.go`，須與 `last` 一致 (有測試檢查)；`fhl-mcp version` 會顯示。

## 寫法（給使用者看 vs. 給工程師看）

- 給使用者看的項目：`🆕 新功能`、`✨ 優化`、`🐞 Bug`、`ℹ️ 說明`。用白話寫，非工程師也能懂（例：「Claude 與 Codex 都能支援了」）。
  `新功能`、`優化` 寫得短；`說明` 是比較詳細的描述，同樣白話。
- 給工程師看的細節一律標 `📝工程師筆記:`，放在該版本的最尾部。若有次分類，寫成 `📝工程師筆記-✨新增`、`📝工程師筆記-🐞發現`、`📝工程師筆記-ℹ️說明` 等。
- TODOs 區塊也用白話寫；技術細節放在版本內的工程師筆記。

## TODOs

- ✨ 優化: npm 啟動器出錯時顯示真正的原因（現在所有錯誤都說「找不到執行檔套件」），並提示「先解除安裝殘留的平台套件再重裝」（需要新的正式版才能發到 npm）
- ✨ 優化: 快速開始補上 Ubuntu 用 apt 裝的 npm 遇到 `EACCES` 權限錯誤時，改用自己的全域資料夾的做法
- ✨ 優化: 「用引用查經文」放寬書卷名後面一定要有空格的限制（現在 `約 3:16` 可以、`約3:16` 會被拒絕，這是原 Python 版就有的行為，D10 先照原樣；Codex 實測時 AI 第一次就用了 `約3:16`，要多重試一次才成功）。怎麼放寬、書名本身含數字的寫法（如 `1John3:16`）怎麼處理，之後再研究
- ✨ 優化: 上游暫時失敗時自動重試，並控制對信望愛網站送出請求的速度（目前有快取與合併相同請求，但還沒有重試與限速）
- ✨ 優化: 在 Mac 上確認兩個 `.command` 雙擊入口（`Setup.command`、`Start-FHL-MCP.command`）；在 Ubuntu 上確認 `disconnect` 真的移除登錄（上次測試時 Ubuntu 還沒登錄，只驗到「沒有登錄就略過」），以及 Ubuntu 上 Codex 的連接與移除；也還沒分別確認 macOS 13 與更新版本
- ✨ 優化: 支援設定檔與 `FHL_` 環境變數來指定埠等選項（計畫書 §8；目前埠只能用 `--port` 指定）
- ✨ 優化: 兩個 AI 工具同時連線、伺服器重開後自動重連的測試
- ✨ 優化: 發佈包在 Windows 11 上的驗收（目前只在 Windows 10 驗過 Windows 發佈包）
- ✨ 優化: 自動建置（CI）與公開發佈前的弱點檢查（目前專案還沒有遠端倉庫，所以還沒設 CI；發佈包由本機一行指令產生）
- ✨ 優化: 邀請 beta 使用者在乾淨的電腦上試用，並整理驗收報告
- ⏳ 待決定: 「搜尋註釋」遇到太常見的字（如「愛」）時，上游會拒絕並說結果超過 500 筆；原 Python 版把它當成 0 筆，Go 版已核准改為明確回報，錯誤訊息的措辭請確認 (D12／D22)
- ⏳ 待決定: 次經（以及接著的使徒教父）書卷的中文全名與英文名，有一部分是我依信望愛網站回傳的簡稱與英文縮寫整理的（例如「以斯拉上」「三童歌」「蘇撒拿傳」），請確認用字 (D24)
- ⏳ 待決定: 「信望愛文章專欄清單」是原 Python 版自己維護的 12 個專欄，實測除了 `women3`（麻辣姊妹）之外，其餘代碼用來搜尋都查不到文章；網站實際有 `lk`（家庭之歌）、`perseverance`、`reading`、`theo`、`thought`、`writer2` 等。要不要改成網站實際的專欄代碼？(D25)
- ⏳ 待決定: 是否讓本程式只講跟原 Python 版相同的 MCP 協定版本 (2025-11-25)，不啟用更新的版本

## 1.1.2.2

- ✨ 優化: 快速開始的「取得與驗證」分成「從 github 下載」與「使用 npm 安裝」兩段，npm 安裝只要 `sudo npm i -g @fhlnet/fhl-mcp`，裝好後用 `fhl-mcp version` 確認，再用 `fhl-mcp setup` 連接 Claude Code 或 Codex。
- ℹ️ 說明: 使用者在 Ubuntu 與 Apple 晶片 Mac（M2）上實測 `status`、`doctor`、`doctor --offline`、`stop`（含重複停止）、`disconnect`（含重複移除），結果都符合預期：服務沒在跑時 `doctor` 只提醒不報錯，重複 `stop`／`disconnect` 只回報「沒有」而不會出錯。Mac 上 Claude Code 與 Codex 的登錄都確實被移除，且沒有動到其他項目。Ubuntu 測試時尚未登錄，所以只驗到「沒有登錄就略過」；Mac 的雙擊入口、Ubuntu 的 Codex 與兩個系統的細部版本仍待確認（見上方待辦）。這一版沒有功能變化。
- 📝工程師筆記-ℹ️說明: 實測使用 npm 安裝的 `fhl-mcp`（版本 1.1.1，linux/amd64 與 darwin/arm64，go1.27.2），非發佈包解壓縮版。`doctor` 九項檢查在服務執行中、客戶端皆已登錄時為 9 項正常；`--offline` 時上游節顯示「已指定 --offline，略過」計為需留意。Ubuntu 沒裝 Codex，`doctor`／`disconnect` 提示找不到 codex 並略過，行為與設計一致。快速開始原稿把 `fhl-mcp setup` 誤寫成 `fhl-map setup`，已更正。詳見 ai-log 261017g。

## 1.1.2.1

- ℹ️ 說明: 版次規則定案：第三碼奇數是正式版、偶數是研發用的小變更版（細節見上方「版次格式」）。這一版沒有功能變化。
- ℹ️ 說明: 在 Ubuntu 用 npm 安裝時，如果曾經在套件還沒完整上架時裝過，系統可能殘留空的舊佔位版，之後執行 `fhl-mcp` 會說「找不到執行檔套件」。處理方式：`sudo npm uninstall -g @fhlnet/fhl-mcp @fhlnet/fhl-mcp-linux-x64`，刪除 `/usr/local/lib/node_modules/@fhlnet`，`sudo npm cache clean --force`，再 `sudo npm i -g @fhlnet/fhl-mcp`（不要單獨安裝平台套件）。已在 Ubuntu 實測重裝成功。
- 📝工程師筆記-✨新增: `tools/npmpack` 現在會拒絕不是三段式 semver 的版本（例如 `1.1.2.1`），避免產生 npm 不接受的套件。
- 📝工程師筆記-ℹ️說明: npm 上四個套件的 `0.0.0-stage` 佔位版已由使用者標為棄用。發佈順序教訓：1.1.1 的主包比 linux-x64 平台包早約 4 分鐘上架，使用者在空檔安裝而殘留佔位版；之後務必三個平台包都上架後才發主包。詳見 ai-log 261017e。

## 1.1.1

- 🆕 新功能: 正式版。把 0.0.0.1 ~ 0.0.0.23 整理成第一個正式發佈：29 個查經工具（經文、搜尋、原文字彙、註釋、語音、次經、使徒教父、註腳、信望愛文章）全部可用，可在 Windows、Apple 晶片 Mac、Ubuntu 上一行指令連接 Claude Code 與 Codex。
- ℹ️ 說明: 這個版本的程式內容與 0.0.0.23 相同，只是版號改為正式版。使用者回報 Windows、Mac、Ubuntu 上 Claude Code 與 Codex 都已連接成功（Mac 與 Ubuntu 依回報，未逐項測試）。尚未有數位簽章與 Apple 公證，第一次執行時 Windows 或 macOS 可能跳出警告。授權為 GPL-3.0-or-later。
- 📝工程師筆記-ℹ️說明: 版次 `1.1.1` 為三段式，與 `主.次.修.小` 慣例不同，是使用者指定；`tools/release` 的版本常數正規表示式原本只認四段式，已放寬為三或四段；`internal/version` 測試只比對字串。發佈包檔名為 `fhl-mcp-go_1.1.1_<os>_<arch>`。TODOs 已移除 Codex 與 Mac/Ubuntu 的實測項目（依使用者回報完成），其餘縮小範圍。

## 0.0.0.23

- 🆕 新功能: Ubuntu 的發佈包現在附 `setup.sh`（首次設定）與 `start.sh`（啟動服務），在終端機執行 `./setup.sh`、`./start.sh` 即可，不必自己打 `fhl-mcp` 完整指令。
- ✨ 優化: `setup`（以及 `doctor`）找不到 Claude Code 或 Codex 時，現在會多告訴你怎麼辦：先用 `npm i -g @anthropic-ai/claude-code`（Claude Code）或 `npm i -g @openai/codex`（Codex）安裝，裝好後開一個新的終端機視窗再執行一次。已在 Windows 與 Mac 上驗證照這個做法就能成功連接。
- ℹ️ 說明: `doc/快速開始.md` 新增 Ubuntu 的解壓縮與執行步驟，以及找不到 CLI 時的處理。
- 📝工程師筆記-✨新增: `scripts/setup.sh`、`scripts/start.sh`（0755，LF，`exec ./fhl-mcp setup|serve "$@"`）；`tools/release` 的 linux 目標納入並加進稽核白名單。
- 📝工程師筆記-✨新增: `clients.Client.InstallHint()`，在 `setup` 印出手動指令的同一處（CLIMissing 且為 setup 時）與 `doctor` 的客戶端登錄節印出；只印文字，不會自動安裝。Mac 實測回報：npm 全域安裝後 `setup` 成功。未做：PATH 以外的常見位置備援搜尋（`~/.local/bin`、`/opt/homebrew/bin` 等），有需要再補。

## 0.0.0.22

- 🆕 新功能: 有可以下載的發佈包了：Windows（zip）、Apple 晶片 Mac（tar.gz）、Ubuntu（tar.gz）各一個，裡面有程式、雙擊用的入口檔、快速開始、授權條款與第三方聲明；另附 `SHA256SUMS.txt` 讓你核對檔案沒被改過。解壓縮到哪裡都能用（路徑有中文或空白也可以）。
- ✨ 優化: `fhl-mcp version` 現在會顯示這份程式是從哪個版本的原始碼、哪一天建出來的，回報問題時比較好對照。
- ℹ️ 說明: 本程式的授權定為 GPL-3.0-or-later（著作權人：財團法人信望愛資訊文化藝術基金會）。授權只涵蓋程式碼，不包含聖經與網站內容的著作權（有些譯本只授權給信望愛網站）。原 Python 版是 MIT 授權，聲明已附在 `THIRD_PARTY_NOTICES.md`。測試版沒有數位簽章與 Apple 公證，第一次執行時 Windows 或 macOS 可能跳出警告。目前只在 Windows 10 上實測過發佈包：解壓到有中文與空白的路徑、啟動約 0.6 秒、閒置時幾乎不耗 CPU、停止約 0.1 秒、停止後沒有遺留程序或連線；Mac 與 Ubuntu 的發佈包還沒在實機試過（見上方待辦）。
- 📝工程師筆記-✨新增: M6：`tools/release`（`go run ./tools/release [-out dist] [-targets windows/amd64,...] [-allow-dirty]`，純標準函式庫）。對每個目標以 `CGO_ENABLED=0 -trimpath -buildvcs=false -ldflags "-s -w -X …/internal/version.Commit=<短 hash，未提交時加 -dirty> -X …Date=<commit 日期>"` 建置；封存檔檔名 `fhl-mcp-go_<版本>_<os>_<arch>.zip|.tar.gz`（Windows zip，其餘 tar.gz），內含單一頂層資料夾；內容固定為執行檔、`LICENSE`、`THIRD_PARTY_NOTICES.md`、`快速開始.md`，以及 Windows 的兩個 `.cmd`（轉 CRLF）、Mac 的兩個 `.command`（0755）；Ubuntu 沒有雙擊入口。檔案排序、mtime＝commit 時間、uid/gid 0，所以同一 commit 重跑結果逐位元組相同（實測兩次 `SHA256SUMS.txt` 一致）。版本常數與 `app_versions.md` 的 `last` 不一致、或工作樹有未提交變更（除非 `-allow-dirty`）會拒絕。建置後若目標等於本機平台，會實跑 `version` 確認含版本、commit、日期。`THIRD_PARTY_NOTICES.md` 由 `release/NOTICE-head.md` 加上實際連結進該目標執行檔的每個模組的 LICENSE（由 `go list -deps` 與模組快取取得，不會過期）。
- 📝工程師筆記-✨新增: 發佈稽核 (`audit`)：檔案必須在白名單內（不可出現快取、狀態檔或其他檔案），且每個檔案（含執行檔）都不能含建置機器的 repo 路徑、使用者家目錄或 TEMP 路徑（正反斜線各查）。`tools/release/main_test.go` 測稽核、tar/zip 權限位元與可重現性。`internal/version` 新增 `Commit`／`Date`（ldflags 注入）與 `Describe()`；`version` 常數仍是 `Version`，由測試對照 `app_versions.md`。`LICENSE`（GNU 官方 gpl-3.0.txt，35149 bytes）與 `release/**` 在 `.gitattributes` 固定 LF。
- 📝工程師筆記-ℹ️說明: 授權決策（使用者 2026-10-16）：GPL-3.0-or-later、著作權人為財團法人信望愛資訊文化藝術基金會。相容性核對：連結的函式庫為 go-sdk（Apache-2.0，部分早期貢獻為 MIT）、jsonschema-go／segmentio/asm／segmentio/encoding（MIT）、yosida95/uritemplate（BSD-3）、golang.org/x/*（BSD-3），皆與 GPLv3 相容；原 Python 版 MIT 的聲明（含其「聖經內容不在授權內」的提醒）放進 `THIRD_PARTY_NOTICES.md`。原始碼檔頭未加 GPL 標頭（GPL 不強制，需要時另案）。
- 📝工程師筆記-ℹ️說明: Windows 10 實測（`go run ./tools/release` 產生的 zip 解壓到 `%TEMP%` 下名為「fhl 測試 路徑」的資料夾，`serve --port 18082 --data-dir` 在該路徑下）：`/health` 就緒 0.6 秒；啟動後工作集 12.5 MB（私有位元組 49 MB）；閒置 60 秒 CPU 0.000 秒；經 MCP（initialize + 100 次 `get_bible_verse`，50 筆不同經文各一次、其餘重複命中快取，間隔 0.1 秒）全部成功，最慢一次 625 毫秒（上游），快取命中約 20 毫秒；之後工作集 20 MB、私有位元組 52 MB、控制代碼 329、執行緒 11，快取寫入 50 個檔案；`stop` 0.12 秒結束，無遺留 listener、程序或 run 狀態檔。未驗證：計畫書 §10 要求的閒置 5 分鐘（只量 60 秒）、冷啟動與防毒首次掃描、Mac／Linux 的 RSS、Windows 11、解壓縮目錄為唯讀的情況。

## 0.0.0.21

- 🆕 新功能: `fhl-mcp setup`：一行指令（或雙擊「首次設定」）就能把本服務登錄到 Claude Code 與 Codex，做一次之後所有專案都能用。可以選只連其中一個；登錄完如果服務正好在跑，還會順便做一次連線測試。
- 🆕 新功能: `fhl-mcp status` 看服務有沒有在跑，`fhl-mcp stop` 乾淨地停止它（不用再找視窗按 Ctrl+C）。
- 🆕 新功能: `fhl-mcp doctor` 一次檢查環境、設定、服務能不能連、信望愛網站通不通、Claude Code／Codex 登錄對不對，哪裡有問題會告訴你怎麼處理。
- 🆕 新功能: `fhl-mcp disconnect` 只移除本服務在 Claude Code／Codex 的登錄，不動你其他的 MCP；不是本服務的同名項目不會被刪。
- 🆕 新功能: 附上雙擊用的入口檔（Windows：`啟動-FHL-MCP.cmd`、`首次設定.cmd`；Mac：`Start-FHL-MCP.command`、`Setup.command`），放在 `scripts/`，之後打包時會跟程式放在同一個資料夾。
- ✨ 優化: 已經有一份在跑的時候再執行 `serve`，會直接告訴你「已在執行」和網址，不會報埠被占用的錯。埠真的被別的程式占住時還是明確失敗，不會偷偷換埠。
- ℹ️ 說明: `setup` 不會自動啟動服務、不會設定開機自動執行，也不需要管理員權限。已經登錄相同網址就當作成功，重跑是安全的；同名但網址不同時只會提醒，不會蓋掉。`stop` 一定要驗證過是本服務才會停（不是只看哪個程式占了埠或 PID）；舊版（0.0.0.20 以前）沒有這個功能，所以要先到原本的視窗按 Ctrl+C 關掉，再用新版重新啟動，之後才能用 `stop`。如果你用 `--data-dir` 啟動，`status`／`stop`／`doctor` 也要加上同樣的 `--data-dir`；換過 `--port` 同理。這一版沒有改動任何查經功能。
- 📝工程師筆記-✨新增: M5：執行狀態 `internal/runstate`（`<資料目錄>/run/state-<port>.json`，內含 instance id、PID、網址、版本、32 bytes 隨機控制 token；暫存檔＋rename 原子寫入、0600；`serve` 啟動後寫入、結束時只刪自己 instance 的檔）。控制通道：同一個 listener 上的 `GET /_control/info`、`POST /_control/stop`（`internal/transport/control.go`），需 `Authorization: Bearer <token>`（`subtle.ConstantTimeCompare`）、Host 必須是 loopback（防 DNS rebinding）、外包 `http.CrossOriginProtection`（瀏覽器跨來源 403）；沒帶 token／錯 token／錯 scheme 一律 401，不存在無驗證的停機端點。`/health` 維持不依賴上游，只回 name／version。`transport.NewHandler` 加 `...Option`（`WithControl`），既有呼叫不變。
- 📝工程師筆記-✨新增: `internal/app/lifecycle.go`：`Probe` 先打 `/health`（必須 `name == FHL-MCP-Go`，否則視為 `port_in_use`；連線被拒才算 `stopped`；逾時／非 HTTP 也算占用）、再讀狀態檔，最後用 token 打 `/_control/info` 比對 instance id 才判為 `running`；沒有或對不上則 `running_unverified`（`stop` 拒絕動作）。順序是刻意的：`Serve` 先寫狀態檔才開始回應，所以能回應就一定有狀態檔（先讀檔會有競態）。`Stop` 只送 `/_control/stop`，再輪詢到 `/health` 不再回應（最多 10 秒），絕不用 PID 送訊號。過期狀態檔（程序已死）在 `serve`／`stop` 時清除。`ResolveDataDir` 從 `ResolveCacheDir` 抽出（資料目錄＝`--data-dir`＞`FHL_DATA_DIR`＞使用者快取目錄下的 `fhl-mcp`）。`Config.Run` 為 nil 時不寫狀態檔、不開控制端點（既有測試與 `app.Serve` 呼叫不變）。`app.Smoke`：真實 MCP session（initialize → tools/list → 呼叫靜態工具 `get_book_list`，不碰上游）。
- 📝工程師筆記-✨新增: `internal/clients`：經官方 CLI 登錄，不直接改任何客戶端設定檔。Claude：`mcp get fhl`（解析 `Scope:`／`URL:`）、`mcp add --transport http --scope user fhl URL`、`mcp remove fhl --scope user`（依 get 回報的 scope）；Codex：`mcp get fhl --json`（遞迴找 `url`，退而求其次抓文字中的 URL）、`mcp add fhl --url URL`、`mcp remove fhl`。`get` 非 0 且輸出符合 `no mcp server|not found|…` 才視為「未登錄」，其他錯誤一律回報失敗（不當成未登錄）。`add` 後再 `get` 讀回確認，不信任 exit code。`disconnect` 只移除網址為 loopback＋`/mcp` 的項目，否則 `refused`。每個客戶端獨立回報；`setup` 在 `--client all`／互動選擇時，未安裝的客戶端不算失敗（只要至少一個成功），明確指定未安裝的客戶端則 exit 1。沒指定 `--client` 且 stdin 是終端機時詢問 1/2/3，非終端機＝all。CLI 透過 `Runner` 介面注入，測試用有狀態的 fake。
- 📝工程師筆記-✨新增: CLI：`status`（`--json`；exit 0＝執行中含未驗證，1＝未執行／埠被占）、`stop`（未執行 exit 0；未驗證／埠被占 exit 1）、`setup`、`disconnect`、`doctor`（`--offline` 略過上游；分「環境／設定／服務／上游／客戶端登錄」五節，有 ✘ 才 exit 1，未安裝客戶端與未啟動服務只是 ⚠；偵測 WSL 並提醒 localhost 不一定相通）。共用旗標 `--host --port --data-dir`。`scripts/` 的 `.cmd` 第一行後必須先 `chcp 65001` 再放中文（cmd 先以系統碼頁剖析，中文 rem 會吃掉換行而執行出錯；已在 Windows 實測），`.gitattributes` 設 `*.cmd` CRLF、`*.command`／`*.sh` LF，兩個 `.command` 以 `git add --chmod=+x` 設為 100755。
- 📝工程師筆記-ℹ️說明: 驗證：`go vet`、`CGO_ENABLED=0 go test ./...` 通過（新增 runstate、control endpoint 授權／Host／跨來源、Probe 各狀態與過期檔、`Stop` 拒絕無對應狀態或外來程式、clients 的 setup／disconnect 各狀態、`status`→`serve` 重複啟動→`stop` 全循環、`setup`／`disconnect`／`doctor` 的退出碼與輸出）。三平台（windows/amd64、darwin/arm64、linux/amd64）`CGO_ENABLED=0` 皆可編譯。真實實測（Windows）：獨立埠 18081 的 `serve`→`status`→第二次 `serve`→`doctor`→`stop`→`status`；對這台電腦實際的 Claude Code CLI 跑 `setup --client claude`（已登錄 8081，回「相同網址不需變更」並通過連線測試）；對仍在跑的舊版（0.0.0.20，無控制通道）8081 服務跑 `stop`，確認拒絕動作；`doctor` 連真實上游通過；兩個 `.cmd` 入口經 `cmd /c` 實跑。未驗證：Codex CLI（本機沒有）；`.command` 腳本與 Mac／Ubuntu 上的任何指令；Windows 上狀態檔實際的存取權限（沿用 `%LocalAppData%` 預設只有使用者／系統／管理員可讀的 ACL，沒有另外設 ACL）；`--data-dir` 指到共用或同步資料夾時 token 會跟著被同步；`disconnect` 真的移除（只在 fake 與單元測試跑過，沒對使用者實際的 `fhl` 登錄動手）；`status`／`stop` 對 IPv6 `--host ::1` 的行為；兩個 AI 工具同時連線、重開後重連（仍在 TODOs）。

## 0.0.0.20

- 🆕 新功能: 可以搜尋信望愛網站的文章（依標題、作者、內文、摘要、專欄、日期，可選簡體），預設回摘要與約 200 字的內容預覽，也可要求回完整內容；也可以列出文章專欄。
- 🐞 Bug: 原 Python 版讀錯欄位，所以文章的「內容預覽」永遠沒有、要求完整內容也只會得到空白。這一版修正了，預覽與完整內容都會正確出現（你選的處理方式）。
- ✨ 優化: 搜尋文章時完全沒給條件，或一次要的篇數不在 1～200 之間，會直接說明哪裡不對。
- ℹ️ 說明: 到這一版，29 個功能全部真正能用，不再有「尚未實作」的項目。這一版也修正了一個小問題：部分功能（文章、有聲聖經、次經、註腳）回傳的中文可能顯示成 `\u662f` 這樣的跳脫碼，現在都會顯示成正常的中文。更新後需要把舊的 `fhl-mcp serve` 關掉再重新啟動。文章專欄清單與網站實際代碼不一致的問題見上方待決定。
- 📝工程師筆記-✨新增: M2-j：`search_fhl_articles`、`list_fhl_article_columns`。`fhlapi.Articles`（`www.fhl.net/api/json.php`，新增 `Client.ArticlesBaseURL`；`getJSON` 抽出 `getJSONAt(base, endpoint, …, ok, …)` 讓不同伺服器與不同「成功」判斷可共用快取；json.php 的 `status` 是數字，`status==1` 才快取，`endpointTTL` 1 天）、`internal/bible/articles.go`（`ArticleAPI`、`SearchArticles`、靜態 12 欄 `ListArticleColumns`）。搜尋條件空字串視為未提供 (Python 真值判斷)；`limit` 在取得結果後才截斷 (`record_count` 改為 `limit`)，驗證 1–200；網站 `status:0` 的 `no data`／`data too much`／其他訊息依 Python 原樣當成一般結果回傳。
- 📝工程師筆記-✨新增: `internal/bible/raw.go` 的 `norm`：以 raw JSON 原樣傳遞的上游值，其 `\uXXXX` 跳脫會原封不動輸出 (Python 會解碼再以 ensure_ascii=False 輸出)；套用於 `jsonOrNull`、`rawOrDefault`、有聲聖經直接傳遞的欄位、註腳 id。既有 golden 因 fixture 當初含實際字元而未暴露，本版把文章 fixture 還原成上游真實的 `\u` 跳脫並加 `TestNormUnescapesAndKeepsNumbers`。
- 📝工程師筆記-🐞發現: D25：(a) 原版 `http_search_articles` 讀 `article.get("content")`，上游實際欄位為 `txt`，所以 `content_preview` 永遠不出現、`include_content=true` 只回 `""`；已核准修正（讀 `txt`）。(b) 原版專欄清單 12 欄多數代碼上游回 `no data`，待決定。(c) 不給任何條件原版丟 D3 型例外，Go 回 `Invalid parameter 'search_params': …`；`limit` 原版不檢查（0 回空、負數切片怪異）。(d) 上游單次最多回 10 筆。詳見 `doc/m0/tool-matrix.md`。
- 📝工程師筆記-ℹ️說明: 驗證：`go vet`、`CGO_ENABLED=0 go test ./...` 通過。5 份 golden 由 Python 自己的 `search_articles`（AST 取出、`httpx` 換成重播）與 `http_search_articles`（僅 `content`→`txt` 兩處）在錄下的 fixture 上產生 (`tests/oracle/articles_oracle.py`)，`list_fhl_article_columns` 為 M0 實測樣本且逐位元組相同；oracle 產生的標題／作者／摘要與 M0 實測樣本一致。文章 fixture 為節錄前 3 筆 (原 10 筆約 118 KB)。經 MCP 實測：`women3` 專欄 limit 1、不給條件的錯誤。「尚未實作」的探測改為 `TestEveryContractToolIsImplemented` (29 個契約工具皆有 handler，並刪除一個 handler 驗證 `ErrNotImplemented` 路徑)；`serve_test` 裡走完整伺服器的「尚未實作」案例移除，該訊息格式 (`mcpserver/server.go`) 目前沒有端對端測試。未驗證：`data too much` 實際回應 (上游從未出現，只以合成資料測)、`gb=1` 文章內容是否完整轉簡體、長文章 (完整內容) 在客戶端的顯示。

## 0.0.0.19

- 🆕 新功能: 可以查現代中文譯本 (TCV) 的經文註腳（翻譯的其他說法、古卷差異等），可選簡體。
- ℹ️ 說明: 只有現代中文譯本有註腳；每卷書有自己的註腳編號，從 1 開始。查不到的編號（或不存在的書卷）會回「找不到書卷 N 的註腳 #M」，這是正常的回答而不是錯誤，與原 Python 版相同。目前 29 個功能裡已有 27 個真正能用，其餘 2 個（信望愛文章）的名稱還在清單上，但呼叫時會明確回覆「尚未實作」。更新後需要把舊的 `fhl-mcp serve` 關掉再重新啟動。
- 📝工程師筆記-✨新增: M2-i：`get_bible_footnote`。`fhlapi.RT` (`rt.php`，固定 `version=tcv`，送 `gb`，`endpointTTL` 7 天；`record.id` 保留 raw JSON)、`internal/bible/footnote.go` (`FootnoteAPI`、`GetFootnote` 回 `*Footnote` 或 `*FootnoteNotFound`；`record_count<=0` → `status:"not_found"`，訊息 `找不到書卷 %d 的註腳 #%d`；不檢查書卷範圍，同 Python)。`book_name` 取上游頂層 `engs`。
- 📝工程師筆記-ℹ️說明: 驗證：`go vet`、`CGO_ENABLED=0 go test ./...` 通過。3 份 golden (創 1 #1、簡體、書卷 999／#999 查無) 都是 M0 實際樣本，逐位元組相同；測試涵蓋參數數字字串 (D6)、缺參數、`gb` 傳遞。經 MCP 對真實上游實測：約 43 #10、書卷 999 查無。「尚未實作」的探測工具改為 `search_fhl_articles`。未驗證：`record_count>1`（上游實測最多 1 筆）、其他書卷的註腳編號範圍。

## 0.0.0.18

- 🆕 新功能: 可以查使徒教父文獻（革利免書信、伊格那丟書信、坡旅甲書信、十二使徒遺訓、巴拿巴書信、黑馬牧人書、帕皮亞殘篇等）的經文，可以在其中搜尋關鍵字，也可以列出所有書卷與可用簡稱。
- 🐞 Bug: 和次經一樣，原 Python 版使徒教父的書卷編號與信望愛網站實際的不同（例如「黑馬牧人書」會查到伊格那丟致他拉勒人書，「十二使徒遺訓」查不到），搜尋結果的書名也會標錯。這一版改用網站實際的編號，書名與內容一致。
- ✨ 優化: 工具說明裡寫可以用的簡稱（`革二`、`黑`、`訓`、`帕` 等）原版有些查不到，現在都能用；`伊`（伊格那丟共有七封書信）和 `巴` 的歧義也處理了（`巴` 是巴拿巴書信，`伊` 請寫明是哪一封）。
- ℹ️ 說明: 「使徒教父書卷清單」現在列出 17 卷（原版 8 卷），編號 201～217 是網站實際編號。目前 29 個功能裡已有 26 個真正能用，其餘 3 個（註腳、信望愛文章）的名稱還在清單上，但呼叫時會明確回覆「尚未實作」。更新後需要把舊的 `fhl-mcp serve` 關掉再重新啟動。
- 📝工程師筆記-✨新增: M2-h：`get_apostolic_fathers_verse`、`search_apostolic_fathers`、`list_apostolic_fathers_books`，與次經共用 0.0.0.17 的 `internal/bible/scripture.go`（`ApostolicFathers` 集合、`fhlapi.QAF`／`SEAF`），本版只在 `tools_scripture.go` 註冊工具並補資料與測試：5 份 fixture、6 份 golden（同樣由 Python 自己的 `http_*` 函式在換成上游編號表後產生，`scripture_oracle.py` 重跑結果與 golden 逐位元組相同）。版本欄位預設值沿用原版 HTTP 工具的 `af`（上游實際回 `afhuang`）。
- 📝工程師筆記-ℹ️說明: 驗證：`go vet`、`CGO_ENABLED=0 go test ./...` 通過；`TestScriptureBookNumbers` 驗證各簡稱送出的 `bid`（`巴`→213、`坡`→210、`訓`→212 等）。經 MCP 對真實上游實測：十二使徒遺訓 1:1（bid 212）、搜尋「教會」（命中帕皮亞殘篇，書名正確）、`伊` 被拒。D24 內容見 `doc/m0/tool-matrix.md`。未驗證：版本欄位 `af` 預設值（上游一律有回）、極長內容（帕皮亞殘篇單筆約 3 KB）的客戶端顯示。

## 0.0.0.17

- 🆕 新功能: 可以查次經的經文（指定書卷、章、節，節可寫 `1`、`1-5`、`1,3,5`），可以在次經裡搜尋關鍵字，也可以列出次經的所有書卷與可用簡稱。
- 🐞 Bug: 原 Python 版的次經書卷編號和信望愛網站實際的編號不一樣：查「多俾亞傳」其實會拿到瑪加伯上的經文，卻還標成多俾亞傳；搜尋結果的書名也會標錯。這一版改以信望愛網站實際的編號為準，查到的經文與標示的書名一致。
- ✨ 優化: 工具說明裡寫可以用、但原版其實查不到的簡稱（例如 `1Mac`、`2Mac`、`便西拉智訓`）現在都能用；`但補`（但以理補篇其實包含三卷）會告訴你分別是「三童歌」「蘇撒拿傳」「彼勒與大龍」。節的格式寫錯、或搜尋關鍵字是空的時候，會直接說明哪裡不對，不會再把網站內部的錯誤訊息丟給你。
- ℹ️ 說明: 這是你決定的處理方式（只修次經與使徒教父這 6 個功能，「書卷清單」「書卷資訊」裡 101 號以上的書卷仍照原版，之後再決定）。所以「次經書卷清單」現在列出 15 卷（原版 9 卷），書卷編號是網站實際的 101～115，與原版不同；書卷的中文全名有些是依網站的簡稱整理的，詳見上方待決定。目前 29 個功能裡已有 23 個真正能用，其餘 6 個（使徒教父、註腳、信望愛文章）的名稱還在清單上，但呼叫時會明確回覆「尚未實作」。更新後需要把舊的 `fhl-mcp serve` 關掉再重新啟動。
- 📝工程師筆記-✨新增: M2-g：`get_apocrypha_verse`、`search_apocrypha`、`list_apocrypha_books`。`fhlapi`：`QSub`／`QAF`／`SESub`／`SEAF` (`qsub`／`qaf`／`sesub`／`seaf.php`，數字欄位保留 raw JSON，`endpointTTL`：qsub／qaf 7 天、sesub／seaf 1 天)，使徒教父那兩組已先放好、這版尚未註冊工具。`internal/bible/scripture.go`：`ScriptureAPI` 介面、`ScriptureSet`、書卷表由內嵌的 `scripture_books.json` 載入 (上游編號、別名依輸出順序列：中文簡稱、中文全名、英文縮寫、英文名)；`GetScriptureText`／`SearchScripture`／`ListScriptureBooks`。書名解析：數字碼 (101–115)、中文完全比對、英文不分大小寫；不使用全域 `ResolveBookID` (其 101 以上的表沿用原版編號)。`verse` 以 `^\d+(-\d+)?(,\d+(-\d+)?)*$` 驗證後才送出 (去頭尾空白)，空查詢 (`query`) 為參數錯誤；`limit` 預設 10、`offset` 預設 0 (與原版 HTTP 工具相同，`limit=0` 上游視為不限)。
- 📝工程師筆記-🐞發現: D24 (已核准處理方式)：實測 `qsub.php` 的 `bid` 101=馬一、102=馬二、103=多比、104=滴、105=便西拉、106=所羅門智訓、107=拉上、108=拉下、109=巴錄、110=耶信、111=禱、112=三童歌、113=蘇撒拿、114=勒、115=斯補；原版表 (101=多俾亞傳…) 與此不符，等於送 `bid=101` 卻標成多俾亞傳。`qaf.php` 201–217 同樣與原版表不同 (原版 205=黑馬牧人書，實際 205=伊格那丟致他拉勒人書；十二使徒遺訓實際為 212)。另：`sec` 寫錯時上游回的 `status` 含完整 SQL 字串；`record_count` 在有 `limit` 時為本頁筆數。詳見 `doc/m0/tool-matrix.md` D24。
- 📝工程師筆記-ℹ️說明: 驗證：`go vet`、`CGO_ENABLED=0 go test ./...` 通過。8 份 golden 由 Python 自己的 `tools/http_tools.py` 函式 (以 AST 原樣取出) 在錄下的 fixture 上執行產生，僅書卷表換成 `scripture_books.json` (`tests/oracle/scripture_oracle.py`)；此 oracle 的清單格式化程式先用「原版自己的表」重現 M0 實測樣本 `list_apocrypha_books` 與 `list_apostolic_fathers_books`，逐位元組相同。經 MCP 對真實上游實測：多 1:1-2、搜尋「智慧」(書名正確標為瑪加伯下)、`John` 被拒。未驗證：`gb=1`／簡體 (這幾個工具契約沒有 `use_simplified`)、大量結果、書卷全名用字 (見待決定)。

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
