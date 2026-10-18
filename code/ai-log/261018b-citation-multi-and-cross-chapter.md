# query_verse_citation: multi-segment and cross-chapter (class B)

- Date: 2026-10-18
- Plan ref: §9 D10
- Status: done
- Files touched: internal/bible/citation.go (rewritten parser), internal/tools/m2a_citation_b_test.go (new), plan D10 row, app_versions.md TODO line, this log
- Supersedes the parser design of 261018a (same file; class A forms still accepted)

## What
`query_verse_citation` accepts `;`-separated segments (`約 3:16; 羅 8:28`, full-width `；`), segments without a book reuse the previous book, cross-chapter ranges (`3:16-4:2`, `3:16-6:2`, `3:16-4:end`), and chapter chains inside a segment (`3:16,5:1-2`). The result is one object: verses concatenated in citation order, `record_count` summed, `navigation` omitted, `version`/`version_name` from the first request.

## Why
Second half of the FHLNUI review (user asked for class B after class A): FHLNUI's `splitReference`/`GetAddresses` handle these shapes (`type a` cross-chapter, `;` book inheritance, `c:v,v` comma lists).

## How
- Parser is now segment-based: `ParseCitationQueries` -> `[]CitationQuery{Book,Chapter,Spec}`; `ParseCitation` kept as a thin wrapper that requires exactly one query (old tests/API unchanged).
- Book is peeled by trying the shortest prefix ending before a digit that resolves (`1 John 4:8` works; bare digits never resolve as numeric book codes). Unresolvable text before the first digit is passed through so `GetVerse` still reports "Book not found".
- Cross-chapter `c1:v-c2:w` expands to `c1:v-end`, whole middle chapters, `c2:1-w`; no verse-count table is needed because `end` is answered by fetching the whole chapter and filtering (from 261018a).
- Adjacent items of the same book+chapter are merged (`16`,`18` -> `16,18`) so plain citations remain one upstream request and existing goldens hold.
- Cap `maxCitationQueries = 40` expanded requests per citation (like D21's guard); over the cap = invalid citation.

## Decisions & assumptions
- Bare number with no chapter in scope = whole chapter, later bare numbers = verses of it (`約 3,5` = ch 3 + verse 5), copied from FHLNUI. `約3:16;18` makes chapter 18 (new segment), not verse 18.
- Chapter ranges (`太 5-7`), reversed cross-chapter (`4:2-3:16`) are rejected.
- Output shape for multi-request results (no navigation, first request's version fields) is my choice; no Python equivalent exists.
- Still not done: Simplified Chinese book names (ResolveBookID exact, D11); `get_bible_verse` itself unchanged (only the citation tool parses these forms).

## Verification
`go vet ./...`, `go test ./internal/...` pass: `TestCitationQueriesExpansion` (expansion table incl. rejections and the cap) and `TestCitationMultiQueryResult` (fake qb: request sequence and combined result). Class A tests in m2a_test.go still pass unchanged. Not run against the live FHL API; no version bump, no commit.

## Open issues / next
Decide release (1.1.2.3 dev build vs roll into 1.1.3). Consider documenting the new forms in the tool description if the frozen contract allows (contract text is frozen from Python; not touched).
