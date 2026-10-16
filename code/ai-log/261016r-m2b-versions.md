# M2-b: list_bible_versions, search_available_versions (ab.php)

- Date: 2026-10-16
- Plan ref: §13 M2, §9 (D4, D5), tool-matrix rows 版本
- Status: done (offline + live client check); not committed
- Files touched: `internal/fhlapi/client.go` (+`AB`, `ABRecord/ABResponse`), `internal/bible/{verse.go,versions.go}` (`VerseAPI` -> `API` with `QB`+`AB`), `internal/tools/{tools.go,args.go}` (+`optStringPtr/optBoolPtr`), tests, `tests/fixtures/upstream/ab.json`, goldens `list_bible_versions.json`, `search_available_versions_strongs.json`

## What
- 2 more real tools (7 of 29; 22 still "not implemented"). `ListVersions` maps ab.php records exactly like Python (testament nt_only wins over ot_only; proc 1-4 -> greek/hebrew/roman/openhan; missing `candownload`/`version` -> false/""). `SearchVersions` filters locally: has_strongs exact match; testament OT/NT (case-insensitive) keeps "both" + that side only; any other value ("both", "", "xx") = no filter; filters echoed as given (null when absent/null, D5).
- Both goldens (86 versions; 3 with Strong's) byte-identical to Python samples; also verified end to end through the SDK client (`TestListBibleVersionsEndToEnd`) and live through the `fhl` MCP in Claude Code.

## Why
Next M2 slice; ab.php is a single no-argument endpoint, so it also exercises the AB path of the upstream client cheaply.

## How / findings
- Upstream: 1 live request recorded as `ab.json` (9601 bytes, 86 records: 52 both / 32 nt_only / 2 ot_only; fields book,cname,proc,strong,ntonly,otonly only). Plus 1 live request `ab.php?gb=1`: byte-identical to `gb=0`, so **upstream ignores `gb` for ab.php** — `use_simplified` has no visible effect on version names (still traditional). Go still sends `gb` (D4, harmless, matches Python's always-send-gb). Not a Go bug; do not "fix" by converting names locally without approval.
- Upstream data quirk kept as is: version `genbklcl` has an empty `name`.
- `Service.API` is now the `API` interface (QB+AB); test fakes got an `AB` stub; the e2e fake upstream routes `/ab.php` to `ab.json`. e2e "not implemented" case now uses `list_commentaries`.
- No cache yet: every call hits ab.php (Python cached it permanently). Cache comes with `internal/cache` (D13), so repeated list calls are 1 upstream request each for now.
- Live check: replaced my own background server (pid 18988, m2a build) with the new build (`%TEMP%\fhl-mcp-m2b.exe serve --port 8081`, left running); `search_available_versions {testament:"NT", has_strongs:true}` -> 3 versions; `list_bible_versions {use_simplified:true}` -> 86. NOTE: do not print the 17 KB list into the transcript again.

## Open / next
- Pending approvals: D17, D18. Next M2 slices: search (`se.php`: search_bible, search_bible_advanced, search_strongs_occurrences), word analysis (`qp.php`, `sd.php`), commentary (`sc.php`, `ssc.php`, `st.php`), then `internal/cache` + retry/rate limit.
