# M2-d: get_word_analysis, get_parsing_reference (qp.php)

- Date: 2026-10-16
- Plan ref: §13 M2, §9 (D2, D3, D4, D6), tool-matrix D15/D21
- Status: done (offline + live client check); not committed
- Files touched: `internal/fhlapi/client.go` (+`QP`), `internal/bible/{parsingfmt,words,legend}.go` (+`parsingfmt_test.go`), `internal/bible/verse.go` (`API` gains `QP`), `internal/tools/{tools_words.go,tools.go,m2d_test.go}` and test fakes, `internal/fhlapi/client_test.go`, `tests/fixtures/upstream/{qp_*,qb_ruth_1}.json`, 6 goldens, `tests/oracle/*.py`, `doc/m0/tool-matrix.md` (D21)

## What
- 2 more real tools (13 of 29; 16 still "not implemented"). `get_word_analysis` returns JSON (testament/book/chapter/verse/original_text/translation/word_count/words[/remark]); `get_parsing_reference` returns Markdown text (D15) built by a Go port of the Python `utils/parsing_format.py` (token ids with b/c/v/w prefix trimmed by `computeAddressTP`, `+w+u+` variant groups, Hebrew RLI/PDI/LRM handling, maqaf spacing, line reversal for OT, `↩`/`↪` newlines). The legend text is generated from the Python source (`legend.go`), not retyped.
- Single verse, same-chapter list/range (`"1-2,5"`, `end_verse`), and cross-chapter ranges (`end_chapter` > chapter: chapter lengths come from a `qb.php` chapter query, like Python) are supported.

## Why
Next M2 slice (original-language tools, both on qp.php).

## How / verification
- Upstream fixtures (8 live requests, 2 s apart, UA `FHL-MCP-Go-fixture-recorder/0.0.0.8`): `qp_john_1_1`, `qp_james_1_1`, `qp_james_1_2`, `qp_gen_1_1`, `qp_ruth_1_22`, `qp_ruth_2_1`, `qb_ruth_1`, `qp_john_99_1`.
- Goldens: `get_word_analysis_john_1_1` and `get_parsing_reference_james_1_1` are the **live Python M0 samples** (byte-identical). The other 4 (`get_word_analysis_gen_1_1`, `get_parsing_reference_{james_1_1_2,gen_1_1,ruth_1_22_to_2_1}`) were produced by running the **Python reference's own formatting code** (a copy of `utils/parsing_format.py` plus a replica of the tool loop) on the recorded fixtures; the oracle was first checked equal to the two live samples. Scripts are kept in `tests/oracle/` (`python -I script <dir-with-parsing_format.py> <fixtures> <out>`); `parsing_format.py` itself is not copied into this repo. Hebrew (OT) and cross-chapter paths are therefore verified against Python code, not against the live Python server. `bind_words` / token-id edge cases are covered by `parsingfmt_test.go` with values from the same oracle.
- Tests: `go vet` clean; `CGO_ENABLED=0 go test ./...` passes (goldens, request mapping incl. gb=1 for `get_word_analysis` use_simplified (D4), validation without upstream calls, empty-data handling, client query/parse, malformed-JSON case).
- Live: replaced my own background server (pid 26244, m2c) with `%TEMP%\fhl-mcp-m2d.exe serve --port 8081` (left running). `get_parsing_reference Jude 1:1` returned the full Markdown (empty `chinesef` falls back to the book's Chinese full name; heading ends with a space because tp=0 gives an empty token id — same as Python); `get_word_analysis Tob 1:1` gave the canon-only error.

## Findings / assumptions
- **Book support**: both tools only accept canon books 1-66; Python reached this via `BookNotFoundError("Invalid book for word analysis: ...")` with an ugly nested message. Go says `Invalid parameter 'book': <input> - <tool> only supports the 66 canon books (book codes 1-66).` (D2 custom message; unknown book uses the unified D3 message).
- **D21 (informational, no approval needed unless user objects)**: (a) for a verse that does not exist (e.g. John 99:1) `qp.php` returns malformed JSON (`"bid":,`), which surfaces as `FHL API request failed - upstream response is not valid JSON` (Python would raise a JSON decode error). (b) `get_parsing_reference` does one upstream request per verse sequentially (plus one qb request per chapter for cross-chapter ranges) with no cap and no cache yet — a large range means many requests; needs the M2 cache/rate-limit work before release. (c) If the upstream returns zero records for a verse, Python crashes with IndexError; Go returns `no original-language data for <book> <c>:<v>`. (d) `get_word_analysis` on an empty record returns the empty-shape JSON (`original_text:""`, `words:[]`), like Python.
- Python quirks kept: `book` is echoed exactly as the caller wrote it in `get_word_analysis`; null fields pass through as null (absent -> `""`); in formatted text a null renders as `None` (`pyStr`); `remark` keys appear only when truthy; `verse` strings are expanded with Python's rules (empty parts skipped, `5-3` yields nothing, an all-empty result is an error).
- Unverified: simplified (`gb=1`) behavior of qp.php, books with multiple `+` variant groups beyond the synthetic oracle cases, Psalms-length ranges.

## Open / next
- Pending approvals: D17, D18, D19. Next M2 slices: commentary (`sc.php`, `ssc.php`, `st.php`: get_commentary, list_commentaries, search_commentary, get_topic_study), then `internal/cache` + retry/rate limit in `fhlapi` (needed before big parsing-reference ranges).
