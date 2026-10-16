# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project state

Go rewrite of the FHL (信望愛) Bible MCP server. **As of 2026-10-16 the M1 vertical slice exists and M2-a/b/c/d/e are built (M2-a ~ d released as 0.0.0.9, M2-e as 0.0.0.10, D17 approval as 0.0.0.11, D18 as 0.0.0.12, D19 as 0.0.0.13, D21 cap as 0.0.0.14, the file cache (`internal/cache`, wired into `fhlapi.Client`; `fhl-mcp cache clear`, `--data-dir`, `--cache-max-mb`) as 0.0.0.15, M2-f audio tools as 0.0.0.16, M2-g Apocrypha tools as 0.0.0.17, M2-h Apostolic Fathers tools as 0.0.0.18, M2-i footnote tool as 0.0.0.19, M2-j article tools as 0.0.0.20 — all 29 contract tools are now implemented; M5 CLI `setup`/`status`/`stop`/`doctor`/`disconnect` + double-click scripts in `scripts/` as 0.0.0.21, M6 release packaging (`go run ./tools/release` → `dist/` zip/tar.gz + `SHA256SUMS.txt`, LICENSE GPL-3.0-or-later, `THIRD_PARTY_NOTICES.md`) as 0.0.0.22, setup/doctor npm install hint + Ubuntu `setup.sh`/`start.sh` as 0.0.0.23, then the 0.0.0.x line rolled up into the user-named formal release **1.1.1** (three-part version, code identical to 0.0.0.23; tagged `v1.1.1`))**: `go.mod` pins Go 1.27.2 + `go-sdk` v1.8.0 (module path
`github.com/snowray712000/FHL-MCP-Go` is an assumption); `cmd/fhl-mcp` (`serve`/`setup`/`status`/`stop`/`doctor`/`disconnect`/`cache clear`/`version`/`help`), `internal/{app,transport,mcpserver,tools,bible,fhlapi,cache,clients,runstate,version}`
and `assets/` (embedded frozen `tools-list.json`) are in place. Implemented so far (M2-a/b/c/d/e): `get_bible_verse`, `get_bible_chapter`, `query_verse_citation`, `get_book_list`, `get_book_info`, `list_bible_versions`, `search_available_versions`, `search_bible`, `search_bible_advanced`, `lookup_strongs`, `search_strongs_occurrences`, `get_word_analysis`, `get_parsing_reference`, `get_commentary`, `list_commentaries`, `search_commentary`, `get_topic_study`, `get_audio_bible`, `get_audio_chapter_with_text`, `list_audio_versions`, `get_apocrypha_verse`, `search_apocrypha`, `list_apocrypha_books`, `get_apostolic_fathers_verse`, `search_apostolic_fathers`, `list_apostolic_fathers_books`, `get_bible_footnote`, `search_fhl_articles`, `list_fhl_article_columns` (all 29; `TestEveryContractToolIsImplemented` guards it, the "not implemented yet" path in `mcpserver` stays for future contract changes). The plan is `doc/FHL-MCP-Go-重寫計畫書.md`
(正體中文, "discussion draft v0.1"). Commands that work: `go test ./...`, `go vet ./...`, `CGO_ENABLED=0 go build ./cmd/fhl-mcp`, `go run ./tools/release [-allow-dirty]` (needs a clean tree, a version bump in both `internal/version` and `app_versions.md`)
(Go is at `C:\Program Files\Go\bin`, add it to PATH in fresh shells); `go test -race` needs cgo/gcc and is not verified on this
machine. Real-upstream (FHL API) tests must be opt-in, never part of the default run; default tests use fixtures/mocks
(`tests/fixtures/upstream`, goldens copied from the Python server in `tests/fixtures/golden`).

The plan is the source of truth for scope, CLI, and milestones (M0–M6). **First release = the 29 Tools only**
(user-decided 2026-10-16): the 19 Prompts and 7 Resource URIs are deferred to M4 / next version — do not build
`internal/prompts`, `internal/resources`, or advertise those capabilities in the first release. Plan §15 is fully
decided (2026-10-16): Ubuntu 22.04/24.04 x64 + macOS 13+ arm64 (Go 1.27.x toolchain; macOS 12 dropped, user-approved) + Windows x64;
first release is a beta with no code signing/notarization; no Smithery or root-path `/` compat (`/mcp` only). The
approved Python-vs-Go behavior differences are in plan §9 (D2–D12). Anything the plan still leaves open must not be
silently decided — ask the user, or record the assumption in an ai-log entry.

Reference implementation (read-only, never modify): `D:\coding\信望愛\FHL-MCP-Server` (Python). It has
uncommitted/untracked files that must be included when capturing contracts (`tools/parsing_reference.py`,
`utils/parsing_format.py`, `tests/fixtures/parsing_reference_james_1_1.md`). Its README claims (test counts,
coverage) are not verified facts. Do not touch the user's Claude Code / Codex client config unless asked.

## AI log workflow (mandatory)

1. Before any work, read `code/ai-log/index.md`; open only the logs it says are relevant.
2. After each logical section of work, add a log file in `code/ai-log/` and a row in its index, in the same turn.
   Format, naming (`YYMMDD` + letter + `-title.md`) and template are defined in `index.md`.
3. Logs are for AI readers: English, terse, What / Why / How. Replies to the user are 正體中文.

Version history (user-facing, 正體中文) is `app_versions.md`, not ai-log. `code/` holds only `ai-log/`. Go sources follow the plan's layout at the repo root
(`cmd/fhl-mcp/`, `internal/...`, `assets/`), not under `code/`.

## Git commit convention

Copied from the user's habit in `D:\coding\信望愛\FHLNUI\FHLNUI` (inferred from its history — 287 commits, all by
`snowray712000`; that repo has no written commit rule). Commit only when the user asks.

- **Subject line only, in 正體中文, no body.** The subject itself carries the detail: what changed plus the
  user-visible behavior or symptom fixed, joined with `，` / `；`. Long subjects are normal.
  Example: `複製：繼續選可跨 cell / 跨欄；只有單次選取跨格才展開成整格`.
- **Format `{topic}：{description}`** — short topic prefix (feature/area), full-width `：`. Docs-only commits use
  `docs: ...`. Identifiers, paths, version numbers stay in original form; item lists go in `( ... )` with ` / `.
- **Release commits** use `發版 {version}：...`; a roll-up commit reads `發版 X：整理 A ~ B (summary)`.
  Same as FHLNUI. Versions are 4-part (`0.0.0.1`); the first release is `0.0.0.1`.
  Changelog is `app_versions.md` at the repo root (the Markdown counterpart of FHLNUI's `app_versions.json`):
  newest first, `TODOs` block always on top, `last` field kept in sync.
  Item style (user-decided 2026-10-16, spelled out in `app_versions.md`): `🆕 新功能` / `✨ 優化` / `🐞 Bug` / `ℹ️ 說明` are for
  end users — plain 正體中文 a non-engineer understands (新功能/優化 short, 說明 longer). All technical detail goes in
  `📝工程師筆記:` (sub-categories as `📝工程師筆記-✨新增` / `-🐞發現` / `-ℹ️說明`), placed at the end of the version.
  `⏳ 待決定` is used in TODOs. Update it in the same commit as the release. Once Go code exists, the binary's version string must match `last`.
- **One small logical change per commit.** Include the related doc updates in the same commit (CLAUDE.md edits,
  `doc/`, and the matching `code/ai-log/` entry + index row) rather than separate "update docs" commits.
- **Trailer**: end with the `Co-Authored-By: Claude ...` line required by the session's attribution reminder.
- Never skip hooks, amend published commits, or force-push without being asked.

## Architecture constraints (cross-cutting, from the plan)

- **Transport**: single local Streamable HTTP MCP endpoint `http://127.0.0.1:8081/mcp` via the official
  `github.com/modelcontextprotocol/go-sdk/mcp`. No stdio, no hand-rolled REST in place of MCP. Must validate
  Host/Origin; do **not** copy the Python server's `0.0.0.0` bind + open CORS + disabled DNS-rebinding protection.
- **Layering**: one tool/resource/prompt registration source (`internal/mcpserver`) → domain logic
  (`internal/bible`: book-name normalization, Strong's, etc.) → single injectable FHL API client
  (`internal/fhlapi`, mockable) → file cache (`internal/cache`, built in 0.0.0.15). Keep domain logic independent of the SDK.
  Do not translate Python line-by-line and do not keep two registration paths (Python had `server.py` stdio +
  `http_server.py` HTTP; only the HTTP 29-tool surface is the contract baseline).
- **Output compatibility**: Python HTTP returns JSON strings as MCP text content. Emit BOTH text content and
  `structuredContent: {"result": "<same string>"}` with the same per-tool `outputSchema` (user-confirmed 2026-10-16;
  see plan §9 and `doc/m0/contract/`) — never only one of them; preserve tool names, parameter semantics, null/empty/number-vs-string
  behavior. Freeze the contract from a real `tools/list` export, not from README or code reading alone.
- **Upstream endpoints**: Python config shows both `/api/` and `/json/` plus `https://www.fhl.net/api/json.php`
  for articles — verify per-feature endpoints, no global string replace. Parsers handle HTML/Markdown, not just JSON.
- **Build**: `CGO_ENABLED=0`, pure Go, embed templates/static data; targets windows/amd64, darwin/arm64,
  linux/amd64. Release artifacts must not contain cache, credentials, or local paths.
- **Process lifecycle**: manual foreground start (`serve`), no daemon/auto-start/system service, fixed port
  (fail on conflict, never silently change it). `stop` must not rely on PID alone nor expose an unauthenticated
  HTTP stop endpoint (use local control channel or loopback + token). `/health` must not depend on upstream.
- **Client registration** (`setup`): prefer the official CLIs (`claude mcp add --transport http --scope user fhl <url>`,
  `codex mcp add fhl --url <url>`); same name+URL = success, same name+different URL = report conflict, never
  overwrite; never rewrite whole client config files; registration success ≠ connectivity (run initialize /
  tools/list smoke test).
- **Config precedence**: CLI > `FHL_` env vars > config file > defaults; data in `os.UserConfigDir`/`os.UserCacheDir`
  (not exe dir or CWD); new cache is versioned and does not reuse the Python cache.
