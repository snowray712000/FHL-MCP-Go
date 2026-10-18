# Search keyword script follows simplified/traditional mode (1.1.2.5)

- Date: 2026-10-18
- Plan ref: §9 D4
- Status: done
- Files touched: internal/hanzi/{hanzi.go,tables.go,hanzi_test.go}, internal/fhlapi/{client.go,client_test.go}, release/NOTICE-head.md, internal/version/version.go, app_versions.md

## What
Search keywords are converted to the script matching `gb` before the upstream call (se.php `q`, ssc.php `key`, st.php `keyword`, json.php `title/author/txt/abst`); sesub.php/seaf.php (no `gb`) always get traditional.

## Why
TODO from 261018d: upstream picks its table by `gb` and does not convert the keyword (`爱` hits only with gb=1).

## How
New `internal/hanzi` with char-by-char maps generated from zhconv 1.4.3 `zhcdict.json` single-char entries (no Go dependency; data is GPL-2.0-or-later, notice added). Conversion sits in `fhlapi.Client` where params are set, i.e. before the cache key. Phrase-level ambiguity (发 -> 發/髮) is not handled: dictionary default wins.

## Decisions & assumptions
Embedded table instead of a dependency (user did not choose; embedded keeps CGO-free/no new module). Assumed sesub/seaf data is traditional-only (not probed live).

## Verification
`go vet`, `CGO_ENABLED=0 go test ./...` pass; new tests for SE keyword and hanzi. Not verified: live upstream after the change, sesub/seaf with simplified input, a real MCP client.

## Open issues / next
Probe sesub/seaf live; phrase-aware conversion if rare characters matter.
