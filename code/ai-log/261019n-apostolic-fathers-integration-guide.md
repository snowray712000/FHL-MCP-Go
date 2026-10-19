# Apostolic Fathers integration guide for another project

- Date: 2026-10-19
- Plan ref: none (user request)
- Status: done
- Files touched: `doc/使徒教父著作-整合指南.md`, `code/ai-log/index.md`

## What
Wrote a zh-Hant design/implementation guide: upstream qaf.php / seaf.php parameters and response shapes, quirks, transformations, book table 201-217 (D24), tool output shapes, cache TTLs, test inventory and untested gaps.

## Why
User wants to reuse the Apostolic Fathers access in another project.

## How
Read `scripture.go`, `tools_scripture.go`, `client.go` (Sub*), `scripture_books.json`, `m2g_test.go`, fixtures/goldens, tool-matrix D24, ai-logs 261016ad/ae. Ran `go test ./...` (all pass).

## Decisions & assumptions
None new. Items not verified are labelled in the doc (seaf limit cap, chapter counts, copyright terms).

## Verification
`go test ./...` passes. No live upstream call this time; response examples come from recorded fixtures.

## Open issues / next
Intro hits (chap 0 / sec 0, HTML, several KB) are passed through unchanged; consider a policy if token size matters.
