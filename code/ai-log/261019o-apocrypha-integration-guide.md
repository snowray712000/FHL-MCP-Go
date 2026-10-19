# Apocrypha integration guide for another project

- Date: 2026-10-19
- Plan ref: none (user request)
- Status: done
- Files touched: `doc/次經-整合指南.md`, `code/ai-log/index.md`

## What
Wrote a zh-Hant design/implementation guide for the Apocrypha tools: qsub.php / sesub.php parameters and shapes, quirks, transformations, book table 101-115 (D24) with upstream names, tool output shapes, cache TTLs, test inventory and untested gaps. Modeled on `doc/使徒教父著作-整合指南.md` (261019n).

## Why
User asked to document how Apocrypha is fetched, what to watch out for, what conversions happen and what unit tests exist, for reuse in another project.

## How
Read `scripture.go`, `tools_scripture.go`, `client.go` (Sub*), `m2g_test.go`, `client_test.go`, fixtures/goldens, tool-matrix D24, ai-log 261016ad. Ran `go test ./...` (all pass). Probed the live upstream with ~20 read-only GETs (1.5 s apart) to fill gaps the earlier guide marked unverified.

New live facts (2026-10-19):
- `sesub.php` has the same cap as se.php: "Fail: record count > 500" when the returned count would exceed 500 (limit 0 / 501 / 100000 on a 4723-hit query fail; limit 500 pages fine). `search_apocrypha` passes limit through, no paging merge, so limit>500 surfaces as ErrUpstream.
- `bid=116` silently returns book 101 chapter 1 (no error); `chap=0` is treated as chapter 1; bad `sec` leaks SQL in `status`.
- Apocrypha hits have no intro rows (chap/sec 0), no HTML, no `【n】` marks; all numbers are ints (4723 hits scanned).
- `gb=1` is accepted by qsub.php and sesub.php and converts text and names to simplified. The Go client never sends it. The Apostolic guide says seaf/sesub have no `gb`; true only of this project's client, not of sesub upstream (seaf untested).
- Max chapter seen per book (lower bound) and per-book hit counts for "的" recorded in the guide.

## Decisions & assumptions
None new. Unverified items are labelled in the doc (gb=1 with simplified keyword, seaf gb, copyright terms, rate limits, official chapter counts). I did not edit the untracked Apostolic guide; its `gb` sentence is only cross-referenced from the new guide.

## Verification
`go test ./...` passes (tools tests `TestM2gGoldens`, `TestScripture*` re-run verbose). Live probes ran against the real FHL API; raw responses stayed in the scratchpad, not committed. Not verified: macOS/Linux, large-scale rate limiting.

## Open issues / next
- Consider paging merge for `search_apocrypha`/`search_apostolic_fathers` like `search_bible` (limit>500 currently fails).
- Consider optional simplified output via upstream `gb=1` for the 6 scripture tools.
- Correct the `gb` sentence in `doc/使徒教父著作-整合指南.md` §3 point 3 when that file is committed.
