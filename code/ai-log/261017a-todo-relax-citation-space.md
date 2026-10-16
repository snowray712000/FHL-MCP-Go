# TODO: relax the space requirement in query_verse_citation

- Date: 2026-10-17
- Plan ref: §9 D10
- Status: done (recorded only; no code change)
- Files touched: `app_versions.md`, `code/ai-log/index.md`

## What
Added a TODO in `app_versions.md` to later relax `query_verse_citation`'s "space required after the book name" rule. Nothing implemented.

## Why
User tested the server from Codex successfully; the AI client first sent `約3:16`, got the D10 error, and retried with `約 3:16`. User decided to defer the improvement ("add to TODO, study later").

## How
Checked first: Go regex/behavior matches Python (`^(.+?)\s+(\d+):…`), D10 in `doc/m0/tool-matrix.md` is "approved: keep as-is, revisit later". Not a regression.

## Decisions & assumptions
User decided to defer. Design notes for later: switch to `\s*` while splitting on the last `digits:digits` group, keep names with leading digits (`1John 3:16`) working, add tests; update D10, tool description examples, and the error text.

## Verification
None needed (docs only). The Codex session itself was reported by the user, not reproduced here; whether `setup`/`disconnect`/`doctor` were exercised against Codex is unknown, so the existing Codex TODO was left in place.

## Open issues / next
Implement the relaxation in a future version once the design is studied.
