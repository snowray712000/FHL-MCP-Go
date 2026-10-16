# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project state

Go rewrite of the FHL (信望愛) Bible MCP server. **As of 2026-10-16 the repo contains no Go code** — only the
plan (`doc/FHL-MCP-Go-重寫計畫書.md`, 正體中文, "discussion draft v0.1") and AI work logs. There is no `go.mod`,
no build, no tests yet. Do not invent build/test commands; the plan's intended checks are `go test ./...`,
`go vet ./...`, and `go test -race` on the native platform. Real-upstream (FHL API) tests must be opt-in, never
part of the default run; default tests use fixtures/mocks.

The plan is the source of truth for scope, CLI, and milestones (M0–M6). Items it marks undecided
(§15: whether 19 Prompts + 7 Resource kinds are in the first release, platform minimums, signing) must not be
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

Version history (user-facing, 正體中文) is `app_versions.md`, not ai-log. `code/` currently holds only `ai-log/`. Go sources follow the plan's layout at the repo root
(`cmd/fhl-mcp/`, `internal/...`), not under `code/`.

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
  newest first, `TODOs` block always on top, `last` field kept in sync, items prefixed `🆕 / ✨ / 🐞 / ⏳`.
  Update it in the same commit as the release. Once Go code exists, the binary's version string must match `last`.
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
  (`internal/fhlapi`, mockable) → file cache (`internal/cache`). Keep domain logic independent of the SDK.
  Do not translate Python line-by-line and do not keep two registration paths (Python had `server.py` stdio +
  `http_server.py` HTTP; only the HTTP 29-tool surface is the contract baseline).
- **Output compatibility**: Python HTTP returns JSON strings as MCP text content. Keep text content for tools
  (don't switch to structuredContent only); preserve tool names, parameter semantics, null/empty/number-vs-string
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
