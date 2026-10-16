# M2-j / release 0.0.0.20: search_fhl_articles, list_fhl_article_columns (www.fhl.net/api/json.php) + D25 + raw-JSON unescape fix

- Date: 2026-10-16
- Plan ref: §13 M2, §9 D25, tool-matrix D25
- Status: done (offline + live client check). All 29 contract tools implemented.
- Files touched: `internal/fhlapi/client.go` (+`Articles`, `ArticlesBaseURL`, `getJSONAt`), `internal/bible/{articles.go,raw.go,raw_test.go,audio.go,scripture.go,footnote.go,verse.go}`, `internal/tools/{tools_articles.go,tools.go,m2j_test.go,tools_test.go}`, `internal/app/serve_test.go`, `internal/fhlapi/client_test.go`, stub `Articles` in 9 fakes, fixtures `articles_*`, 6 goldens, `tests/oracle/articles_oracle.py`, plan §9 + tool-matrix (D25), `internal/version/version.go`, `app_versions.md`, `.claude/CLAUDE.md`

## What
Last 2 tools. Output = Python `http_search_articles` / `http_list_article_columns`, except D25(a): article text read from `txt`.

## How / verification
- Oracle runs Python's own `FHLAPIEndpoints.search_articles` (AST-extracted, `httpx` replaced by replay) + `http_search_articles` with the single `content`->`txt` patch (asserted exactly 2 sites). Columns golden = live M0 sample (byte-identical). Titles/authors/abstracts of the oracle output match the live M0 sample.
- json.php `status` is numeric; only `status==1` is cached (new `getJSONAt` with a per-call `ok`). Articles base URL is separate from `BaseURL`.
- Found while live-testing: passthrough `json.RawMessage` kept upstream `\uXXXX` escapes (Python prints characters). Earlier goldens did not catch it because the fixtures I had rewritten contained real characters. Fixed with `norm` (decode with UseNumber, re-encode without HTML escaping) in `jsonOrNull`, `rawOrDefault`, audio direct fields, footnote id. Article fixtures restored to upstream-style escapes (trimmed to 3 records, 118 KB -> 9 KB).
- Tests: `TestEveryContractToolIsImplemented` replaces the not-implemented probes; the end-to-end "not implemented" case in `serve_test.go` was removed (no unimplemented tool left) - the message formatting in `mcpserver/server.go` now has no end-to-end test.

## Decisions & assumptions
- User chose "修正，讀 txt". Limit validation 1..200 is mine (documented range). Column list kept as Python's (12), flagged for decision (D25b).
- Preview counts Unicode code points like Python (`[]rune`).

## Open / next
All 29 tools done. Remaining (see TODOs): retry/backoff/rate limit, `doctor`, config file, Codex + two-client + reconnect verification, macOS/Linux runtime, pending decisions (D12 wording, protocol version, cache TTLs, D24 names, D25b columns), `get_book_list/get_book_info/ResolveBookID` 101+/201+ numbering (D24 follow-up), M3+ (setup/stop/doctor, packaging).
