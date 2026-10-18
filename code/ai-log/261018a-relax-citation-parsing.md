# Relax query_verse_citation parsing (class A)

- Date: 2026-10-18
- Plan ref: §9 D10
- Status: done
- Files touched: internal/bible/citation.go, internal/tools/m2a_test.go, doc/FHL-MCP-Go-重寫計畫書.md (D10 row), app_versions.md (TODO line), this log

## What
`ParseCitation`/`QueryCitation` now accept: no space after book (`約3:16`, `1John 4:8`), full-width digits and punctuation (`３：１６－１８，２０`, `～`, `、`, `路1．23`), `#...|` wrapper, trailing `. 。 ; ；`, `n-end`/`n-e` (case-insensitive), whole chapter (`約 3`), one-chapter books (bare `猶` / `猶大書` = chapter 1; `猶 5`, `猶 5-7`, `猶 23-end` = verses of chapter 1), book names with stray spaces/trailing dot (`Gen. 1`).

## Why
User asked (after reviewing FHLNUI reference parsing: `splitReference`, `ParsingReferenceDescription`, `GetAddresses`, `SmartDescriptEndParsing`) to relax single-citation parsing; real Codex sessions failed first try on `約3:16`. D10 had kept the Python strict regex.

## How
- Normalise (full-width map, strip `#`/`|`/trailing punct) then two regexes: colon form `book chap : spec`, no-colon form `book spec`. Lazy book capture + `resolveBookLoose` (tries as-is, trailing `.` stripped, spaces collapsed, spaces removed) so books starting with digits (`1 John`) work.
- Pure digits (`31`) are rejected so a numeric book code is not read as a bare one-chapter book.
- `end`: no verse-count table exists in Go, so `QueryCitation` fetches the whole chapter and filters locally; `record_count` recomputed, `navigation` dropped (it would describe the whole chapter). Non-end specs pass through unchanged, so existing goldens are untouched.
- `parseVerseSpec` swaps reversed ranges only for the local `end` filter path; for normal requests the spec is forwarded as typed (upstream behaviour unverified).

## Decisions & assumptions
- `猶 5` = verse 5 of chapter 1 (Bible-citation convention), unlike FHLNUI which treats a bare number as a chapter and rejects it for one-chapter books. Assumption; user can reverse.
- Chapter ranges (`約 3-4`) rejected with the format error.
- Not done (class B / later): multiple citations with `;`, cross-chapter `3:16-4:2`, chain `1:1,3:5`, Simplified Chinese book names (ResolveBookID is exact, D11), `ParseCitation` returns the resolved spelling (`Gen.` -> `Gen`).

## Verification
`go vet ./...`, `go test ./...` all pass (new cases in `TestCitationParsing`, new `TestCitationRelaxedQueries`: end filter uses the Philemon chapter fixture; full-width citation reaches upstream as `43/3/16-17`). Not verified against the live FHL API (e.g. reversed ranges, `end` on real data). No version bump / commit made.

## Open issues / next
Class B (multi-citation, cross-chapter). Decide whether to release as 1.1.2.3. Plan §9 D10 row updated.
