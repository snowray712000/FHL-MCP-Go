# M2-c: search_bible, search_bible_advanced, lookup_strongs, search_strongs_occurrences (se.php / sd.php)

- Date: 2026-10-16
- Plan ref: §13 M2, §9 (D4, D5, D6, D7, D12), tool-matrix D19/D20
- Status: done (offline + live client check); not committed
- Files touched: `internal/fhlapi/client.go` (+`SE`, `SD` and their types), `internal/bible/{search,strongs}.go`, `internal/bible/verse.go` (`API` gains `SE`, `SD`), `internal/tools/{tools_search.go,tools.go,m2c_test.go}`, test fakes, `internal/fhlapi/client_test.go`, `doc/m0/tool-matrix.md` (D19, D20), 6 upstream fixtures (`se_*`, `sd_*`), 5 goldens

## What
- 4 more real tools (11 of 29; 18 still "not implemented"). `lookup_strongs` was added because `search_strongs_occurrences` calls it first (Python does too); it is a contract tool anyway.
- Goldens byte-identical to Python M0 samples: `search_bible 愛 limit 3`, `search_bible_advanced 愛 John range limit 3`, `lookup_strongs 25 NT`, `lookup_strongs 99999 NT` (upstream "not found" = success record with `sn:""` and nulls, kept per D7), `search_strongs_occurrences 25 NT limit 3`.

## Why
Next M2 slice; all four go through se.php / sd.php and share the Strong's input parser.

## How / findings
- Upstream fixtures (6 live requests, 2 s apart, UA `FHL-MCP-Go-fixture-recorder/0.0.0.8`): `se_ai_limit3`, `se_ai_john_range_limit3`, `se_greek25_nt_limit3`, `se_empty_query` (`{"status":"Fail: Empty Query"}`), `sd_nt_25`, `sd_nt_99999` (`record_count:0` but one record with `sn:""`, `orig/dic_text/edic_text: null`).
- Parameter surface follows the frozen schemas: `search_bible` has NO `offset` (an `offset` argument is ignored, offset 0); `search_bible_advanced` has `offset`; range bounds are integers (numeric strings accepted, D6); `search_strongs_occurrences` limit default 50.
- Python behaviors kept: a missing or `0` range bound silently falls back to a whole-Bible search with the plain `search_bible` result shape (`scope:"all"`, no `range`); G/H prefix overrides the `testament` argument; leading zeros dropped; `testament` accepted case-insensitively; sd.php null/absent fields pass through (`rawOr`: absent -> `""`, null -> null); `ccnt` string/number -> int.
- D7 (approved) empty query -> `Invalid parameter 'query': must not be empty` before any upstream call. Assumption: whitespace-only counts as empty.
- **D19 (new, pending approval)**: in Python's range path of `search_bible_advanced` only `search_type:"keyword"` works (`greek_number`/`hebrew_number` are rejected, `greek`/`hebrew` pass the first check and fail in the endpoint) even though the tool description lists `keyword/greek_number/hebrew_number`. Go accepts the three documented values everywhere. Verified live: `search_bible_advanced 25 greek_number John(43..43)` returns John 3:16 with `<WG25>`.
- **D20 (new, finding, compat kept)**: `total_count` is upstream `record_count`, which equals the number of records returned when `limit` applies (fixtures: limit 3 -> `record_count` 3; live `H430 limit 2` -> `total_count` 2). So `total_count` is not the number of all matches; `showing` in `search_strongs_occurrences` always equals it. Copied as is; if a true total is wanted, upstream `count_only=1` may provide it (not tested).
- Upstream failures (`status` not success, HTTP, network) go through the existing `fhlapi.Error` -> `FHL API request failed - ...` mapping (D12 style, isError).
- Unverified: `gb=1` effect on se.php/sd.php (flag is sent; no simplified fixture recorded), non-`unv` versions, `offset` paging on the live API, whitespace-only query upstream behavior, Python's int() quirks on odd Strong's inputs (underscores, Unicode digits).
- Tests: `go vet` clean; `CGO_ENABLED=0 go test ./...` passes (request parameter mapping, validation without upstream calls, fallback/range cases, Strong's parser table, client query/parse tests).
- Live check: replaced my own background server (pid 29224, m2b build) with `%TEMP%\fhl-mcp-m2c.exe serve --port 8081` (left running); `search_strongs_occurrences H430 limit 2`, `search_bible_advanced 25 greek_number 43-43`, and empty `search_bible` query behaved as above.

## Open / next
- Pending approvals: D17, D18, D19 (D20 is informational).
- Next M2 slices: word analysis (`qp.php`: get_word_analysis, get_parsing_reference), commentary (`sc.php`, `ssc.php`, `st.php`), then `internal/cache` and retry/rate limit in `fhlapi`.
