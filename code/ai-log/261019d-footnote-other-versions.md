# get_bible_footnote: other versions and whole chapter/book (1.1.2.9)

- Date: 2026-10-19
- Plan ref: §9 (tool behaviour; frozen contract changed on purpose)
- Status: done
- Files touched: internal/fhlapi/client.go (+test), internal/bible/footnote_multi.go, internal/tools/tools_footnote.go, internal/tools/footnote_multi_test.go, internal/tools/m2i_test.go, internal/bible/verse_test.go, assets/tools-list.json, doc/m0/contract/tools-list.json, tests/fixtures/upstream/rt*.json, internal/version/version.go, app_versions.md

## What
`get_bible_footnote` gains optional `version` (default tcv) and `chapter`; `footnote_id` becomes optional (omitted = list all footnotes of the chapter for lcc/csb, of the book for others). Classic call (TCV + id) is untouched.

## Why
User picked item 1 of the FHLNUI review (ai-log 261019b) and chose "optional params on the existing tool" (AskUserQuestion, 2026-10-19). FHLNUI z261010b/z261011a.

## How
- New `Client.RTMulti`: id>=0 -> `rt.php`, else `rt-nui.php`, both on `https://bible.fhl.net/json/` (`jsonBase()`); `/api/` has no rt-nui.php and ignores `engs` on rt.php (live-measured).
- Upstream needs `engs` (not `bid`) with its own abbreviations: leading digit + space (`1 Sam`) and 9 exceptions (`rtEngsExceptions`), verified for all 66 books with cnet.
- lcc/csb restart ids per chapter -> chapter required; other versions ignore chapter (not sent), scope is book.
- `FootnoteMultiAPI` optional interface so ~10 existing fake APIs stay unchanged.
- Version code whitelisted `[A-Za-z0-9]{1,16}` before sending (rt.php echoes SQL on fail); upstream `fail` -> fixed-message ErrUpstream.

## Decisions & assumptions
Result shapes (`bible_footnotes`, `scope`, `record_count`, `footnotes[]`) are mine. Tool description text rewritten. Only canon books 1-66. Whole-book lists can be big (Ps cnet 3572); no cap added.

## Verification
`go vet`, `go test ./...` pass (new tests with real-response fixtures). Temporary live test (deleted) against bible.fhl.net: lcc Hab 2 single/list, cnet 1Kings single, cnet Philem book, ncv -> error, lcc ch 99 -> empty. csb and tcv-list not run live beyond curl. No Python comparison (Python has no such feature).

## Open issues / next
Consider a cap or paging for huge book lists; list_bible_versions could mark which versions have footnotes; 'chapter' for tcv is ignored.
