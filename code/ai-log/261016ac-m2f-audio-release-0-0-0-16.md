# M2-f / release 0.0.0.16: get_audio_bible, get_audio_chapter_with_text, list_audio_versions (au.php)

- Date: 2026-10-16
- Plan ref: §13 M2, §9 (D3, D4, D5, D6), tool-matrix D23
- Status: done (offline + live client check); not committed when written
- Files touched: `internal/fhlapi/client.go` (+`AU`, `AUResponse`, `au.php` TTL), `internal/bible/audio.go` (new), `internal/bible/verse.go` (`AudioAPI` embedded in `API`), `internal/bible/commentary.go` (interface decl), `internal/tools/{tools_audio.go,tools.go,m2f_test.go}` + stub `AU` in 5 fakes, `internal/fhlapi/client_test.go`, 9 upstream fixtures (`au_*.json`), 10 goldens, `tests/oracle/audio_*`, `doc/m0/tool-matrix.md` (D23), `internal/version/version.go`, `app_versions.md`, `.claude/CLAUDE.md`

## What
3 more real tools (20 of 29; 9 left: apocrypha x3, apostolic fathers x3, `get_bible_footnote`, `search_fhl_articles`, `list_fhl_article_columns`). Output shape = Python (`version`, `version_name`, `book`, `book_eng`, `chapter` (string), `audio_files{mp3,ogg}`, optional `navigation`); `get_audio_chapter_with_text` = `{audio, text}` where `text` is the `get_bible_chapter` result.

## How / verification
- 9 live `au.php` requests (2 s apart, UA `FHL-MCP-Go-fixture-recorder/0.0.0.16`). Goldens: `list_audio_versions`, `get_audio_bible_john3`, `get_audio_chapter_with_text_phm1` = live Python M0 samples (byte-identical); 7 more `get_audio_bible_*` from running Python's own `tools/audio.py` against a replay stub (oracle first checked equal to the samples).
- Fields stay raw JSON (upstream sends numbers as strings); Python truthiness (`truthy(rawOr(...))`) decides whether mp3/ogg/prev/next appear.
- `go vet`, `CGO_ENABLED=0 go test ./...` pass; live via `fhl` MCP: Romans 8 taiwanese, Philemon 1 with text + `use_simplified`, version list, invalid `audio_version`.

## Decisions & assumptions
- D4 applied to the audio request of `get_audio_chapter_with_text` (au.php honours `gb` for book names); `get_audio_bible` has no `use_simplified` in the contract -> always `gb=0`.
- `audio_version` validated before the book (Python order); message `Invalid parameter 'audio_version': x - Must be one of: ...` (my wording, same style as other D2/D3 messages).
- `au.php` is cached 7 days (assumption; Python did not cache it).
- No chapter existence check (matches Python; D23).

## Release
Bumped to 0.0.0.16 (user: "接著下一個工具" as 0.0.0.16). Not yet committed at the time of writing.

## Open / next
Remaining 9 tools; retry/rate limit; pending: D12 wording, protocol-version narrowing, cache TTL confirmation.
