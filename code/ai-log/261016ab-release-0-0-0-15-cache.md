# Release 0.0.0.15: on-disk response cache

- Date: 2026-10-16
- Plan ref: §9 "快取", §8 (data dir), §13 M2
- Status: done (not committed yet when written)
- Files touched: `internal/cache/{cache.go,cache_test.go}` (new), `internal/fhlapi/client.go` (+`Cache`, `endpointTTL`, `getJSON` split into `fetch` + cache layer, `isSuccess`), `internal/fhlapi/client_test.go`, `internal/app/{cachecfg.go,cachecfg_test.go,serve.go}`, `cmd/fhl-mcp/{main.go,main_test.go}`, plan §9, `doc/m0/tool-matrix.md` (D13 note), `internal/version/version.go`, `app_versions.md`, `.claude/CLAUDE.md`, `code/ai-log/index.md`

## What
- Cache for all 8 upstream endpoints: per-entry JSON files, atomic write, TTL, size cap (default 200 MB, oldest-first cleanup to 90%), in-process merging of concurrent identical requests, corrupt/expired -> miss. Only `status=="success"` + valid JSON bodies are stored.
- Config: `serve --data-dir` / `FHL_DATA_DIR`, `--cache-max-mb` / `FHL_CACHE_MAX_MB` (0 = off); default dir `os.UserCacheDir()/fhl-mcp`, cache under `cache/v1`. New command `fhl-mcp cache clear [--data-dir]`.

## Decisions & assumptions (mine, listed as TODO-able in the release notes)
- TTLs: qb/qp/sc/st/ab 7 d; se/ssc 1 d; sd permanent. Python cached only qb/se/qp/sd/ab (verses/word_analysis 7 d, search 1 d, sd+ab "permanent"); plan says commentary ~7 d, so sc/st/ssc are new; ab changed from permanent to 7 d because the version list can change.
- Key = full upstream URL incl. all params and `gb` (D4: simplified is its own entry). No stale-fallback.
- `cache clear` only deletes `*.json`/`.tmp-*` inside `v<N>` dirs under the cache dir; it does not coordinate with a running server (a missing file is a miss, the size counter self-corrects), which the plan's "與執行中的讀寫協調" is satisfied by in my reading - not verified with a live concurrent run.

## Verification
`go vet`, `CGO_ENABLED=0 go test ./...` pass. Live: rebuilt server with a temp `--data-dir`, `get_bible_verse` twice + `lookup_strongs` -> exactly 2 cache files (qb 7 d, sd permanent); `cache clear` removed 2. Not verified: actual upstream request counts on the live path, multi-process writers, macOS/Linux, `-race`, Windows rename-over-existing under contention.

## Open / next
Retry/backoff/rate limit in `fhlapi`, `doctor`, config file; remaining 12 tools; pending: D12 wording, protocol-version narrowing.
