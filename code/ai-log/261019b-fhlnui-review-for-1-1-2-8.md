# FHLNUI review: what else can help the MCP tools (1.1.2.8 prep)

- Date: 2026-10-19
- Plan ref: §9 (tool behaviour)
- Status: done (investigation only, no code change)
- Files touched: this log, index.md

## What
Surveyed `D:\coding\信望愛\FHLNUI\FHLNUI` (`docs/api_reference.md`, `docs/z26*`) against the 29 MCP tools. Earlier rounds already took: citation parsing (261018a/b) and SN or/and/not/range (261019a). Candidates left, ranked:

1. **Footnotes** (`get_bible_footnote`, TCV + single id only). FHLNUI `z261010b`/`z261011a`: `rt.php` also serves `lcc` (呂振中), `csb`, `cnet`...; ids of lcc/csb restart per chapter, others are per book. Backend `rt-nui.php` (no `id` = whole chapter/book, `scope` field) exists on bkbible.fhl.net but production sync unconfirmed. 503 above ~6 concurrent requests on that host.
2. **Merged verses**: qsb.php returns `bible_text:"a"` for a verse merged into the previous one (西 2:20-21). Go does not detect it (grep found nothing) so tools output a literal "a".
3. **TWCB dictionaries**: `sbdag.php` (NT Greek) and `stwcbhdic.php` (OT Hebrew) return a JSON *string*; not used by Go (only `sd.php`/CBOL). `lookup_strongs` could add them.
4. **Version flags** from `uiabv.php` (`ntonly`/`otonly`/`strong`); FHLNUI keeps a 91-version snapshot. Check `list_bible_versions` (ab.php) exposes equivalent flags.
5. **LXX / fhlwh / bhs original text** with SN: derived offline from `bible_parsing.zip` (sqlite, ftp.fhl.net/FHL/COBS/data), see `z260928a/c`; `se.php` does not support fhlwh. Large, offline-data scope.
6. Request throttling lesson (concurrency ~4, retry 3) supports the existing TODO "retry + rate limit".

Already covered in Go: `same` (cognate) list of `sd.php`, citation relaxations, SN algebra, 500 cap paging.

## Why
User asked what else in FHLNUI helps the current MCP tools.

## How
Read FHLNUI `CLAUDE.md`, `api_reference.md`, rt/lxx/uiabv/節碼/同源字 docs; grepped Go sources for the endpoints.

## Decisions & assumptions
None. Ranking is a recommendation only.

## Verification
Static reading only; no live API call; `rt-nui.php` availability on production not tested.

## Open issues / next
Await user choice; likely 1.1.2.8 = merged-verse handling (item 2, small) or footnote versions (item 1).
