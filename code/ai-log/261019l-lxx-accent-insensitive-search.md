# Accent-insensitive Greek keyword search in lxx (1.1.2.17)

- Date: 2026-10-19
- Plan ref: §9 (tool behaviour)
- Status: done
- Files touched: internal/bible/hebsearch.go, internal/bible/greeksearch.go, internal/bible/versions_test.go, internal/version/version.go, app_versions.md

## What
Greek keyword search in version lxx now uses the same folded matching as fhlwh (261019k) on `bible_lxx.json.gz`.

## Why
se.php matches lxx by exact form (FHLNUI z260927a); follow-up named in 261019k.

## How
`origKeyword` Greek branch accepts lxx. Unlike fhlwh/bhs, a failed file load falls back to se.php (it works for lxx), so `origKeyword` returns ok=false there.

## Decisions & assumptions
Behaviour change for existing lxx keyword users: more matches (accent/case/sigma insensitive, substring on folded text). Accepted by the user's request.

## Verification
go vet, go test ./... pass. Temporary live test (deleted): two spellings of "kai" both 18769 OT verses; "ierobo-am" 102. Not compared with the old se.php counts.

## Open issues / next
Retry/rate limit is the only FHLNUI item left.
