# M2-a: get_bible_chapter, query_verse_citation, get_book_list, get_book_info

- Date: 2026-10-16
- Plan ref: §13 M2, §9 (D3, D5, D6, D10, D11), tool-matrix D17/D18
- Status: done (offline); not released, not committed
- Files touched: `internal/bible/{bookinfo,citation}.go`, `internal/tools/{tools.go,m2a_test.go,tools_test.go}`, `internal/app/serve_test.go`, `tests/fixtures/golden/` (6 new), `tests/fixtures/upstream/{qb_philem_1,qb_john_3_16_17}.json`, `doc/m0/tool-matrix.md` (D11 note, D18)

## What
- 4 more real tools (5 of 29 implemented; 24 still return the "not implemented" isError). `get_bible_chapter` = `GetVerse` without verse; `query_verse_citation` = regex parse then `GetVerse`; `get_book_list`/`get_book_info` are static (no upstream).
- Outputs byte-identical to the Python M0 samples for: book list NT, book info 約/門/Philemon, chapter 門 1 (Philem, 25 verses), citation `約 3:16-17` (6 goldens, test `TestM2aGoldens`).

## Why
First M2 slice: tools that reuse the existing qb.php path or need no upstream, so they close quickly with exact goldens.

## How / findings
- D11 check: Python `BOOK_ALIASES` is only used by `normalize_book_name`/`parse_reference`/`fuzzy_search`, none of which the 29 HTTP tools call. Real resolution = the 4 name columns + numeric code = existing `ResolveBookID`. So no alias table was built; alias/simplified-name support would be a new feature (needs approval). Recorded in tool-matrix D11.
- D18 (new, pending): Python reports `testament:"NT"` for every id outside 1-39 (Apocrypha 101-112, Fathers 201-216, and valid codes without table row like 113 -> null names). Kept byte-compatible; changing it needs approval.
- `get_book_list`: testament filter is case-insensitive, unknown value returns all 86 entries and echoes the input; null/empty -> "all".
- Citation: same regex as Python (`^(.+?)\s+(\d+):(\d+(?:-\d+|,\d+)*)$`, space required per approved D10); Go `\d` is ASCII-only (Python also accepts full-width digits) - negligible, not special-cased. Error message is custom English (D2): `Invalid parameter 'citation': <input> - Invalid citation format. Expected ...`; bad citation never reaches upstream.
- Unknown book in `get_book_info` uses the unified D3 message.
- Fixtures: 2 live `qb.php` requests (Philem 1; John 3:16-17 no strong), UA `FHL-MCP-Go-fixture-recorder/0.0.0.8`, 2 s apart. Fixture files are not line-ending converted (`.gitattributes`).
- Tests: `go vet` clean, `CGO_ENABLED=0 go test ./...` passes. e2e "not implemented" case now uses `list_bible_versions`.

## Open / next
- Next M2 slices: `list_bible_versions`/`search_available_versions` (ab.php), search (`se.php`), word analysis/strongs (`qp.php`/`sd.php`), commentary, then cache (`internal/cache`) and retry/rate limits in `fhlapi`.
- Pending approvals: D17, D18.
- Live check done: stopped my old 0.0.0.7 background server (pid 30240, started by me earlier), started the new build (`%TEMP%hl-mcp-m2a.exe serve --port 8081`, health reports 0.0.0.8), then called `get_book_info 提多書` (id 56, Titus) and `query_verse_citation 約 3:16-17` through this session's `fhl` MCP: both OK. This session had connected to the previous server instance, so the call also shows Claude Code recovers after a server restart (single client; the two-client and Codex cases are still open). The new server is left running in the background.
