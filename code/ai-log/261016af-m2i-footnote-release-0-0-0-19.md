# M2-i / release 0.0.0.19: get_bible_footnote (rt.php)

- Date: 2026-10-16
- Plan ref: §13 M2, §9 D4 (gb honoured), D7 (not_found is a success answer)
- Status: done (offline + live client check)
- Files touched: `internal/fhlapi/client.go` (+`RT`, `RTResponse`, TTL), `internal/bible/footnote.go` (new), `internal/bible/verse.go` (`FootnoteAPI` embedded), `internal/tools/{tools_footnote.go,tools.go,m2i_test.go,tools_test.go}`, `internal/app/serve_test.go` (unimplemented probe -> `search_fhl_articles`), stub `RT` in 8 fakes, `internal/fhlapi/client_test.go`, 3 rt fixtures, 3 goldens, `internal/version/version.go`, `app_versions.md`, `.claude/CLAUDE.md`

## What
1 more real tool (27 of 29; left: `search_fhl_articles`, `list_fhl_article_columns`).

## How / verification
- Python `http_get_bible_footnote`: `status success` + `record_count>0` -> object with `version:"TCV"`, `version_name`, `book_id` (as requested), `book_name` (= upstream `engs`), `footnote_id` (record id), `text`; else `status:"not_found"`+`message`.
- All 3 goldens are the live M0 samples (byte-identical). Fixtures recorded from rt.php (gen 1 #1 gb0/gb1, 999/999). Upstream returns success with `record_count:0` and `engs:""` for unknown book/id (also for bid 0 and negative ids).
- `go vet`, `CGO_ENABLED=0 go test ./...` pass; live via `fhl` MCP: John #10, 999/999.

## Decisions & assumptions
- No book-range validation (matches Python; sample `err_footnote_not_found`). `rt.php` cached 7 days (Python: verses strategy 7 d). Official docs list `engs` as the selector, but `bid` (used by Python) works.

## Open / next
0.0.0.20 articles (www.fhl.net/api/json.php). After that no tool is "not implemented": the probe tests in `tools_test.go`/`serve_test.go` need another way to test the not-implemented path.
