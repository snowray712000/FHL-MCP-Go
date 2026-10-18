# lookup_strongs: TWCB dictionary option (1.1.2.10)

- Date: 2026-10-19
- Plan ref: §9 (tool behaviour; frozen contract changed on purpose, as in 261019d)
- Status: done
- Files touched: internal/fhlapi/client.go (+test), internal/bible/twcb.go (+test), internal/tools/tools_search.go, internal/tools/twcb_test.go, assets/tools-list.json, doc/m0/contract/tools-list.json, tests/fixtures/upstream/{sbdag_3303,stwcbhdic_128,sbdag_99999_500body}.json, internal/version/version.go, app_versions.md

## What
`lookup_strongs` gets optional `dictionary`: `cbol` (default, unchanged), `twcb` (new `TWCBEntry`), `both` (`StrongsEntry` + `twcb_definition`). Upstream `sbdag.php` (NT) / `stwcbhdic.php` (OT).

## Why
User picked the TWCB item from the FHLNUI review (261019b). Same "optional param on the existing tool" pattern as 261019d.

## How
- `Client.TWCB(nt, number, simplified)`; both endpoints work on `/api/` (no `/json/` switch needed, unlike rt-nui).
- Missing entry => upstream HTTP 500 with a truncated body; mapped to "no entry" for status 500 only.
- `cleanTWCB`: CRLF -> LF, `<div>` -> newline, other tags dropped, `&entities;` unescaped, `#ref|` -> `ref`, trailing blanks trimmed, 3+ newlines -> 2.
- Optional interface `TWCBAPI` keeps old fakes untouched.

## Decisions & assumptions
Param name/values, result shapes and the marker stripping are mine. `strongs_number` for TWCB is the 5-digit zero padded number (matches CBOL's `sn` style), not upstream's `sn` (which can be `3924a`). Letter-suffixed numbers still rejected by `ParseStrongsInput`.

## Verification
`go vet`, `go test ./...` pass (real-response fixtures; Greek checked without literals because of oxia/tonos). Temporary live tests (deleted) against bible.fhl.net: G3303 / H157 (gb=1) / G25 both / H128 twcb, and after the 500 mapping G99999 twcb -> "not found in the TWCB dictionary", G99999 both -> the usual empty CBOL entry (existing behaviour of the default path).

## Open issues / next
`search_strongs_occurrences` could show TWCB too; outputs can be tens of KB for common words (no truncation option).
