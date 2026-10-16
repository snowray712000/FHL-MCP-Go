# ai-log index

Read this file first. It lists every AI work log in this directory so you can decide which ones to open.
Logs are written for AI readers, in English. Humans' docs live in `../../doc/`.

## Conventions

**File name:** `YYMMDD` + letter + `-{kebab-title}.md`
- `YYMMDD` = date of the work (e.g. 2026-10-16 -> `261016`).
- Letter = `a`, `b`, `c`, ... in order of creation within the same day (several per day are normal).
- Example: `261016a-bootstrap-claude-and-ai-log.md`, `261016b-m0-source-manifest.md`.

**One log = one logical section of work** (a milestone step, a feature, a decision, an investigation).
Do not batch unrelated work into one file; do not split one coherent step across many files.

**Required template for each log:**

```markdown
# {Title}

- Date: YYYY-MM-DD
- Plan ref: {plan section / milestone, e.g. "§13 M0"}
- Status: done | partial | blocked
- Files touched: {paths}

## What
What the AI did (concrete outcome).

## Why
Reason / requirement / plan section that motivated it.

## How
Approach, commands, key design choices, alternatives rejected.

## Decisions & assumptions
Anything decided or assumed (esp. items the plan marks undecided). "None" if none.

## Verification
What was actually run/checked and the result. State what was NOT verified.

## Open issues / next
Follow-ups, risks, what the next section should pick up.
```

**Maintaining this index:** append one row per new log (newest at the bottom of the table).
`Read when` must say, in one line, for which kind of task the log is worth opening.
If a later log supersedes an earlier one, mark the old row's Status as `superseded by <file>`.

## Entries

| File | Date | Plan ref | Status | Summary | Read when |
| --- | --- | --- | --- | --- | --- |
| [261016a-bootstrap-claude-and-ai-log.md](261016a-bootstrap-claude-and-ai-log.md) | 2026-10-16 | §1–17 (read), repo setup | done | Read the rewrite plan; created `.claude/CLAUDE.md` and this `ai-log/` system. No Go code yet. | Starting a new session; need project rules, plan digest, or log conventions. |
| [261016b-init-claude-md.md](261016b-init-claude-md.md) | 2026-10-16 | whole plan | done | `/init`: rewrote `.claude/CLAUDE.md` with project state, log workflow, cross-cutting architecture constraints. | Need to know what CLAUDE.md covers / why no build commands exist yet. |
| [261016c-git-commit-convention.md](261016c-git-commit-convention.md) | 2026-10-16 | repo process | done | Added commit convention to CLAUDE.md, inferred from the FHLNUI repo (one-line Chinese `{topic}：{detail}`, no body, docs in same commit). | Before making any git commit; want to know where the convention came from. |
| [261016d-app-versions-md.md](261016d-app-versions-md.md) | 2026-10-16 | repo process | done | Created `app_versions.md` (changelog, first version `0.0.0.1`); commits use `發版 {version}：` again, not milestone prefixes. | Preparing a release/commit; need version format and changelog rules. |
| [261016e-m0-source-manifest.md](261016e-m0-source-manifest.md) | 2026-10-16 | §3, §13 M0 | done | Hashed 254 files of the Python reference (237 clean / 7 modified / 10 untracked) into `doc/m0/source-manifest.tsv`; confirmed 29 tool names match plan. | Need the reference baseline, its HEAD/dirty state, or what M0 still lacks (schema export, prompts/resources inventory). |
| [261016f-m0-tools-schema-export.md](261016f-m0-tools-schema-export.md) | 2026-10-16 | §13 M0, §12 | partial | Exported real `tools/list` (29 tools) + initialize from Python HTTP server into `doc/m0/contract/`; found outputSchema `{result:string}` + structuredContent, empty HTTP prompts/resources, a Python error bug. | Need the frozen tool schemas, output format (text + structuredContent), or known Python bugs. |
| [261016g-decision-output-format.md](261016g-decision-output-format.md) | 2026-10-16 | §9 | done | User-confirmed: Go emits text content AND `structuredContent{result}` with same `outputSchema`; plan §9 and CLAUDE.md updated. | Implementing tool result encoding or checking what was decided about output format. |
| [261016h-m0-prompts-resources-inventory.md](261016h-m0-prompts-resources-inventory.md) | 2026-10-16 | §4, §13 M0, §15 | done | Static inventory of 19 Prompts / 7 Resource URIs (4 schemes); they exist only in stdio `server.py`, not HTTP; rendered prompts saved as JSON. | Deciding or implementing Prompts/Resources (M4); need arg lists, Python quirks, or why HTTP has none. |
| [261016i-decision-v1-tools-only.md](261016i-decision-v1-tools-only.md) | 2026-10-16 | §1, §4, §13, §15 | done | User-decided: first release = 29 Tools only; Prompts/Resources deferred to M4; plan, CLAUDE.md, TODOs updated. | Need the v1 scope / definition of done, or whether to build prompts/resources packages now. |
| [261016j-m0-tool-sampling-and-diff-list.md](261016j-m0-tool-sampling-and-diff-list.md) | 2026-10-16 | §9, §12, §13 M0 | done | Sampled all 29 tools (55 calls, 32 upstream requests); `doc/m0/tool-matrix.md` has per-tool endpoint/shape table and differences D1–D16 (bugs, `use_simplified` ignored, error format). | Implementing or testing any tool; need upstream endpoints, output shapes, known Python bugs, or pending approvals. |
| [261016k-decision-approve-diffs-d2-d12.md](261016k-decision-approve-diffs-d2-d12.md) | 2026-10-16 | §9, §13 M0 | done | User approved diffs D2–D12 (D10 as-is); written into plan §9; D12 root cause found (upstream `record count > 500` swallowed) and explicit-message behavior approved; D3 static audit widened. | Implementing error handling, simplified-Chinese, null/number coercion, or search_commentary; need what was approved vs still pending. |
| [261016l-decision-s15-platforms-signing-smithery.md](261016l-decision-s15-platforms-signing-smithery.md) | 2026-10-16 | §1, §11, §15 | done | User decided remaining §15: Ubuntu 22.04/24.04 x64 + macOS 12+ arm64, beta unsigned/un-notarized, no Smithery/root-path compat; project name assumed. M0 can close. | Release/packaging, platform support, or compat scope questions; M1 spike list. |
| [261016m-m1-toolchain-check.md](261016m-m1-toolchain-check.md) | 2026-10-16 | §13 M1, §15 | blocked | Local machine has no Go on PATH; collected go/go-sdk release facts (unverified) and spike notes; waiting for user OK to install Go. | Starting M1; need toolchain/go-sdk version facts or why M1 has not begun. |
| [261016n-m1-spikes-and-pin.md](261016n-m1-spikes-and-pin.md) | 2026-10-16 | §13 M1, §9 D1/D2, §15 | partial | Pinned Go 1.27.2 + go-sdk v1.8.0; macOS floor -> 13+ (user); spikes: low-level AddTool reproduces contract tools/list exactly, unknown-tool needs middleware for `isError`, Host protection default on, Origin needs `CrossOriginProtection` wrapper. | Building mcpserver/transport; need SDK behavior facts, version pins, or the macOS floor rationale. |
| [261016o-app-versions-user-vs-engineer-style.md](261016o-app-versions-user-vs-engineer-style.md) | 2026-10-16 | repo process | done | User-set changelog style: plain-language 新功能/優化/Bug/說明 for users, everything technical under `📝工程師筆記` at version end. | Writing `app_versions.md` entries or a release commit. |
| [261016p-m1-vertical-slice.md](261016p-m1-vertical-slice.md) | 2026-10-16 | §13 M1, §5, §6, §9 | partial | M1 slice: `serve`, embedded 29-tool contract (equal to Python tools/list), real `get_bible_verse` (byte-identical to Python goldens), D2 middleware, Host/Origin protection, fetch fixtures; Claude Code only reached `initialize`, Codex unverified; new diff D17 (Titus/多) pending approval. | Implementing any tool/transport code, adding fixtures, or finishing M1 client verification. |
| [261016q-m2a-chapter-citation-books.md](261016q-m2a-chapter-citation-books.md) | 2026-10-16 | §13 M2, §9 | done | M2-a: `get_bible_chapter`, `query_verse_citation`, `get_book_list`, `get_book_info` (5/29 tools); 6 Python goldens byte-identical; D11 resolved (Python aliases unused by HTTP tools); new D18 (testament `NT` for apocrypha) pending. | Implementing the next M2 tools, book-name resolution, or checking what is built/approved. |
| [261016r-m2b-versions.md](261016r-m2b-versions.md) | 2026-10-16 | §13 M2, §9 D4/D5 | done | M2-b: `list_bible_versions`, `search_available_versions` via new `fhlapi.AB` (7/29 tools); goldens byte-identical; upstream ab.php ignores `gb` (simplified has no effect on names); no cache yet. | Implementing the next M2 tools, `Service.API` interface, or version-list behavior/filters. |
| [261016s-m2c-search-strongs.md](261016s-m2c-search-strongs.md) | 2026-10-16 | §13 M2, §9 D7/D12 | done | M2-c: `search_bible`, `search_bible_advanced`, `lookup_strongs`, `search_strongs_occurrences` via new `fhlapi.SE`/`SD` (11/29 tools); 5 goldens byte-identical; new D19 (range search rejects documented greek/hebrew types) pending, D20 (`total_count` = returned count). | Implementing the next M2 tools, Strong's input parsing, search parameter surface, or checking pending approvals. |
| [261016t-m2d-word-analysis-parsing-reference.md](261016t-m2d-word-analysis-parsing-reference.md) | 2026-10-16 | §13 M2, §9 D15 | done | M2-d: `get_word_analysis`, `get_parsing_reference` via new `fhlapi.QP` and a Go port of `parsing_format.py` (13/29 tools); goldens = 2 live Python samples + 4 from running Python's own formatter on fixtures (`tests/oracle/`); D21 findings (invalid JSON for missing verses, unbounded per-verse requests). | Implementing the next M2 tools, parsing-reference formatting, or deciding request limits/caching. |
| [261016u-release-0-0-0-9.md](261016u-release-0-0-0-9.md) | 2026-10-16 | repo process | done | Release 0.0.0.9 roll-up of M2-a ~ M2-d (13/29 tools); version constant and `app_versions.md` bumped. | Checking what 0.0.0.9 contains or how releases are recorded. |
| [261016v-m2e-commentary-topics.md](261016v-m2e-commentary-topics.md) | 2026-10-16 | §13 M2, §9 D12 | done | M2-e: `get_commentary`, `list_commentaries`, `search_commentary`, `get_topic_study` via new `fhlapi.SC/SCList/SSC/ST` (17/29 tools); 4 goldens = live Python samples + 7 from running Python's own `commentary.py`; D12 applied to `ssc.php`; D22 findings. | Implementing the next M2 tools, commentary/topic behaviour, or the D12 error wording. |
| [261016w-release-0-0-0-10.md](261016w-release-0-0-0-10.md) | 2026-10-16 | repo process | done | Release 0.0.0.10 = M2-e (17/29 tools); version constant and `app_versions.md` bumped. | Checking what 0.0.0.10 contains or how releases are recorded. |
| [261016x-release-0-0-0-11-d17.md](261016x-release-0-0-0-11-d17.md) | 2026-10-16 | §9 D17 | done | D17 approved (Titus -> 56, Tobit via `Tob`/`101`); recorded in plan §9/matrix, regression test added, version 0.0.0.11. | Checking whether D17 is approved or how book-name collisions are handled. |
