# M2-g / release 0.0.0.17: get_apocrypha_verse, search_apocrypha, list_apocrypha_books (qsub.php / sesub.php) + D24

- Date: 2026-10-16
- Plan ref: §13 M2, §9 D24 (new), tool-matrix D24
- Status: done (offline + live client check)
- Files touched: `internal/fhlapi/client.go` (+`QSub/QAF/SESub/SEAF`, `Sub*` response types, TTLs), `internal/bible/scripture.go` + `scripture_books.json` (new), `internal/bible/verse.go` (`ScriptureAPI` embedded in `API`), `internal/tools/{tools_scripture.go,tools.go,m2g_test.go}` + stub methods in 7 fakes, `internal/fhlapi/client_test.go`, 4 qsub + 3 sesub fixtures, 8 goldens, `tests/oracle/scripture_oracle.py`, plan §9 + `doc/m0/tool-matrix.md` (D24), `internal/version/version.go`, `app_versions.md`, `.claude/CLAUDE.md`

## What
3 more real tools (23 of 29). The Apostolic Fathers code (`QAF/SEAF`, `ApostolicFathers` set, table) was written together with this but its tools are registered only in 0.0.0.18 (user numbered them separately).

## Why
User asked "0.0.0.17 作次經 .18 作使徒教父 .19 get_bible_footnote .20 望愛文章 2 個", commit each when done.

## How
- Live probing (2 s apart not needed, ~40 GETs) showed the Python book table does not match upstream numbering (D24): Python sent `bid=101` for Tobit and got 1 Maccabees text labelled Tobit. Upstream docs (https://bible.fhl.net/json/) confirm bids 101–115 and that `chineses` is the real selector.
- User chose "只修這 6 個工具": tools use upstream numbering/names; global book table (`get_book_list/get_book_info/ResolveBookID` 101+) untouched.
- Book table = `internal/bible/scripture_books.json` (embedded): id, name_zh, name_en, ordered aliases. Resolution: digits (in range), exact Chinese, ASCII case-insensitive. Aliases include Python's documented names where they map unambiguously, plus upstream `chineses`/`engs`. Ambiguous: `但補` (3 books), `伊` (7 letters) -> error.
- Output formatting = Python `http_*` functions (field order, `book_id` "未知" fallback, defaults). `book` is always the resolved book's `name_zh` (Python echoed its own table by exact key or the raw input).
- Added validation (divergence, improvement): `verse` regexp (upstream leaks SQL in `status` otherwise), empty `query` (upstream `Fail: Empty Query`). Error for non-member book: `Invalid parameter 'book': x - Not a book of this collection (101-115). Use the matching list_*_books tool ...` (my wording).
- Oracle: AST-extracted `http_*` functions run on fixtures with the table swapped for the JSON; formatter first validated by reproducing the two live M0 list samples using Python's own tables (validation script in scratchpad, not committed).

## Decisions & assumptions
- Full Chinese names for books Python lacked (以斯拉上/下, 瑪拿西禱詞, 三童歌, 蘇撒拿傳, 彼勒與大龍, 以斯帖補篇) and English names are my assumptions from upstream `chineses`/`engs`; Python names kept for books it had (e.g. 多俾亞傳, 瑪加伯上). Logged as ⏳ in TODOs. Polycarp aliases (`坡`,`Pol`,`Polycarp`) map to 210 (to the Philippians) - assumption.
- Same defaults as Python HTTP tools: `limit` 10, `offset` 0. TTLs: qsub/qaf 7 d, sesub/seaf 1 d (assumption, Python cached verses 7 d / search 1 d).

## Verification
`go vet`, `CGO_ENABLED=0 go test ./...` pass. Live via `fhl` MCP: Tobit 1:1-2 (bid 103), search 智慧 (hits labelled 瑪加伯下), `John` rejected. Not verified: `gb`, big results, Codex, macOS/Linux.

## Open / next
0.0.0.18 registers the apostolic tools (goldens/fixtures already recorded in scratchpad `full/`); then footnote (rt.php) and articles. `get_book_list`/`get_book_info` 101+ numbering still Python's (follow-up).
