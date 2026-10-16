# Decision: approve differences D2-D12; D12 investigated

- Date: 2026-10-16
- Plan ref: §9, §12, §13 M0
- Status: done (D12 behavior approved later the same day, see Decisions)
- Files touched: `doc/FHL-MCP-Go-重寫計畫書.md` (§9 new subsection), `doc/m0/tool-matrix.md` (D2-D12 status/columns), `doc/m0/contract/upstream-ssc-samples.json`, `app_versions.md` (TODOs)

## What
Recorded user's approval of D2-D12 from [261016j](261016j-m0-tool-sampling-and-diff-list.md) ("D10 先照原樣，D12 先調查") in plan §9
"已核准的原版差異" and in the matrix status column. Investigated D12 and extended D3 with a static audit.

## Why
User decision in chat; plan must be the source of truth for Go behavior.

## How
- D12: read `tools/commentary.py` + `endpoints.py`; sent 3 direct GETs to `https://bible.fhl.net/api/ssc.php` (`愛`, `耶穌`, `尼哥底母`, >=2 s apart, same UA).
  Common words return HTTP 200 `{"status":"Fail:record count > 500"}`; Python finds no `record` key and returns `total_count:0`. A rare word returns
  `status:"success"`, `record_count`, `record[]`. Saved under `contract/upstream-ssc-samples.json`. Also noted HTTP wrapper omits `commentary_id`.
- D3 audit: `InvalidParameterError(parameter, value, reason="")` called with one arg at `tools/audio.py:72`, `commentary.py:39/141/201`, `info.py:117`, `strongs.py:141`
  -> same ctor bug likely in get_commentary, search_commentary, get_topic_study, lookup_strongs, search_strongs_occurrences (runtime-verified only for get_book_info and get_audio_chapter_with_text).
- Matrix updated by script (rows rebuilt, escaped pipe in D10 kept).

## Decisions & assumptions
- Decided by user: D2-D12 as proposed; D10 keep as-is; D12 investigate first.
- D12 follow-up, approved by user: when upstream `status` != `success`, return an explicit message ("結果超過 500 筆，請縮小關鍵字" for `Fail:record count > 500`), not 0 results. Plan §9 and matrix updated.
- My assumptions (flagged in plan, not user-confirmed): report it as an `isError:true` result (per D2); apply to `search_commentary` first; other endpoints' `status` conventions (vs the D7 "no data" shapes) get checked per endpoint in M2.
- Open technical risk written into plan: D2 requires Go SDK to return unknown-tool as an `isError` result; verify in M1.

## Verification
Re-checked D-row cell counts in the matrix (all 16 rows well formed; D10 pipe escape intact). Upstream responses saved and parsed.
NOT verified: other InvalidParameterError sites at runtime, multi-line call sites, Go SDK behavior for D2.

## Open issues / next
- Ask user: remaining §15 items (Ubuntu/macOS minimums, signing, Smithery compat).
- Then M0 exit check -> M1 (pin Go + go-sdk versions; verify D2 unknown-tool behavior early).
