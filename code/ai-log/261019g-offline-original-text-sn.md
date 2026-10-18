# Offline original-text files for Strong's numbers in lxx / bhs / fhlwh (1.1.2.12)

- Date: 2026-10-19
- Plan ref: §9 (tool behaviour)
- Status: partial (verse lookup done; SN search over these files not done)
- Files touched: internal/fhlapi/client.go, internal/bible/origtext.go, internal/bible/verse.go, tests, internal/version/version.go, app_versions.md

## What
`get_bible_verse` (and `query_verse_citation`) with `include_strong=true` and version lxx / bhs / fhlwh now return text with `<WG..>` / `<WH..>` numbers from FHL's pre-built files; otherwise unchanged.

## Why
FHLNUI review item 5. qb.php gives no numbers for these versions (strong=0 and 1 identical, measured). FHLNUI builds the files offline from bible_parsing.zip and publishes them.

## How
User said the files should be on bible.fhl.net/NUI/; names unknown. Found `https://bible.fhl.net/NUI/index/bible_{lxx,bhs,fhlwh,bhs_code}.json.gz` (HTTP 200, 0.6-1.7 MB gz, ~9 MB json for lxx). `Client.OrigData` downloads via the file cache (7 d), checks gzip magic, decompresses (64 MB cap). `bible.getVerseOrig` parses once per process (mutex map in Service), filters by chapter/verse spec, builds navigation from file order. Optional interface `OrigDataAPI` so fakes need no change. Any failure (no interface, download/parse error) falls back to qb.php exactly as before. Rejected: shipping the data in the binary (size, staleness); generating from bible_parsing.zip ourselves (sqlite + alignment code, FHLNUI already does it).

## Decisions & assumptions
Assumes the NUI/index path is stable (it is a UI asset, not an API; may move with a UI rebuild - FHLNUI vite adds ?v= only on its own requests). Silent fallback means a user asking for numbers may get none if the download fails. bhs_code (Hebrew search) not used. Contract JSON untouched.

## Verification
go vet / go test ./... pass. Temporary live test (deleted) returned Gen 1:1 lxx and bhs and John 1:1 fhlwh with numbers and sane navigation. Not verified: cache path with the real file cache against the real file; memory use of parsed lxx (~22k rows, fine).

## Open issues / next
SN search in lxx/bhs offline (needs an index); bhs_code for Hebrew keyword search; TWCB in search_strongs_occurrences; retry/rate limit.
