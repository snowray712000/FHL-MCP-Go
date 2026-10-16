# M2-e: get_commentary, list_commentaries, search_commentary, get_topic_study (sc / ssc / st.php)

- Date: 2026-10-16
- Plan ref: §13 M2, §9 (D3, D4, D5, D12), tool-matrix D22
- Status: done (offline + live client check); not committed
- Files touched: `internal/fhlapi/client.go` (+`SCList`/`SC`/`SSC`/`ST`, `ErrTooManyRecords`), `internal/bible/commentary.go` (new; `CommentaryAPI` embedded in `API`), `internal/tools/tools_commentary.go`, `internal/tools/tools.go`, tests (`m2e_test.go`, `client_test.go`, stub methods in the other fakes), 11 upstream fixtures, 11 goldens, `tests/oracle/commentary_*`, `doc/m0/tool-matrix.md` (D22), `app_versions.md` TODOs, `.claude/CLAUDE.md`

## What
- 4 more real tools (17 of 29; 12 still "not implemented"). Output shapes are the Python ones (`commentaries[]`, `results[]`, `count_only` shape).
- `use_simplified` is honoured (D4): `sc.php` converts titles/names/body with `gb=1` (verified live), `ssc`/`st` get `gb` too.

## Why
Next M2 slice requested by the user ("註釋類（sc.php、ssc.php、st.php）").

## How / verification
- 14 live upstream requests (2 s apart, UA `FHL-MCP-Go-fixture-recorder/0.0.0.9`) -> `tests/fixtures/upstream/{sc_*,ssc_*,st_*}.json`.
- Goldens: `list_commentaries`, `get_commentary_john316`, `get_topic_study_love{,_count}` equal the **live Python M0 samples** (byte-identical). The other 7 (empty chapter, no verse, search hit/none, topic torrey_en/none/none_count) come from running the **Python reference's own `tools/commentary.py`** against a replay stub of `FHLAPIEndpoints` (oracle first checked equal to the 4 live samples). Scripts: `tests/oracle/commentary_oracle.py` + `commentary_stub_endpoints.py` (assembly instructions at the top).
- `go vet` clean; `CGO_ENABLED=0 go test ./...` passes (goldens, request mapping incl. source->N, validation w/o upstream calls, D12 error, client query/parse/status tests).
- Live via the `fhl` MCP (restarted dev server = `%TEMP%\fhl-mcp-m2e.exe`, port 8081, /health 0.0.0.9): list_commentaries, get_commentary 約翰福音 3:16 with `use_simplified` (simplified body), get_topic_study count_only, search_commentary 愛 -> D12 error.

## Decisions & assumptions
- **D12 applied** to `ssc.php` (`Fail:record count > 500` -> `ErrTooManyRecords`): reported as `Invalid parameter 'keyword': 愛 - too many results (more than 500 records), use a more specific keyword (結果超過 500 筆，請縮小關鍵字)`. Wording and the "invalid parameter" framing are my assumption (listed as a TODO 待決定).
- Python quirks kept (see D22): `verse` omitted -> upstream 0 records, `verse: null`; `navigation` never emitted (prev/next are top-level upstream, Python reads per-record); `book` of search hits keeps 44-char space padding; `search_commentary.total_count` = number of returned hits; `commentary_id`/`topic_id` not exposed (not in the HTTP contract).
- Unknown book -> D3 unified message; empty keyword -> `Invalid parameter 'keyword': must not be empty`; bad `source` -> `Must be one of: all, torrey_en, naves_en, torrey_zh, naves_zh`.
- Not verified: `st.php` other sources live (only torrey_en and `all`), `gb=1` for `ssc`/`st`, more than 500-record behaviour of `st.php`; U+2028/2029 in text would be escaped by Go's encoder (Python leaves raw), none seen in fixtures.

## Open / next
- Remaining 12 tools: audio x3 (`get_audio_bible`, `get_audio_chapter_with_text`, `list_audio_versions`), apocrypha x3, apostolic fathers x3, `get_bible_footnote`, `search_fhl_articles`, `list_fhl_article_columns`; plus `internal/cache` + retry/rate limit (still none), pending approvals D17/D18/D19/D21(b).
