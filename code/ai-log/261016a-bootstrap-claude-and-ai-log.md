# Bootstrap: .claude rules and ai-log system

- Date: 2026-10-16
- Plan ref: whole plan read (§1–17); no milestone work started (M0 not begun)
- Status: done
- Files touched:
  - `.claude/CLAUDE.md` (new)
  - `code/ai-log/index.md` (new)
  - `code/ai-log/261016a-bootstrap-claude-and-ai-log.md` (new, this file)

## What
- Read `doc/FHL-MCP-Go-重寫計畫書.md` in full.
- Created `.claude/CLAUDE.md` with project summary, pointer to the plan, and the mandatory log workflow.
- Created `code/ai-log/` with `index.md` (conventions + entry table) and this first log.
- Repo state at that time: only `.git/` and `doc/` (containing the plan). No Go code, no `go.mod`.

## Why
The user wants every AI work section recorded (what / why / how) so later AI sessions can resume
quickly without re-deriving context. `index.md` lets an AI choose which logs to read instead of
reading all of them.

## How
- User asked for ".Claude"; created lowercase `.claude/` (the directory Claude Code actually reads;
  Windows filesystem is case-insensitive anyway). Rules go in `.claude/CLAUDE.md`. No `settings.json`
  was created because nothing needs configuring yet.
- `ai-log/` is placed at `code/ai-log/` exactly as the user specified, relative to the repo root.
  Note: the plan's planned Go layout (§5) puts source in `cmd/` and `internal/` at the root, not under
  `code/`. `code/` currently only holds `ai-log/`; do not assume Go sources go inside `code/`.
- Naming: `YYMMDD` + letter + `-title.md`. The user's example used `261015a` as a pattern; today's date
  gives `261016`.

## Plan digest (so you need not re-read all 406 lines)
- Goal: Go rewrite of the FHL Bible MCP server. Single binary `fhl-mcp`, local-only Streamable HTTP at
  `http://127.0.0.1:8081/mcp`, no stdio, no Docker/Go/Python needed by end users, manual start (no
  daemon / auto-start). Targets: Windows x64, macOS arm64, Ubuntu 22.04/24.04 x64.
- Clients: Claude Code (`claude mcp add --transport http --scope user fhl <url>`) and Codex
  (`codex mcp add fhl --url <url>`); `setup` must prefer the official CLIs, never overwrite conflicts.
- Scope: 29 Tools (first milestone); 19 Prompts + 7 Resource URI kinds are NOT yet confirmed as
  first-release scope (plan §1, §15) — do not silently drop or promise them.
- Stack: official `github.com/modelcontextprotocol/go-sdk/mcp`, stdlib `net/http`, `slog`, `embed`;
  `CGO_ENABLED=0`; file cache with TTL; Host/Origin validation (do not copy the Python version's
  `0.0.0.0` + open CORS).
- CLI (planned, unimplemented): `serve`, `setup`, `status`, `stop`, `doctor`, `disconnect`,
  `cache clear`, `version`, `help`.
- Milestones: M0 source/contract capture -> M1 vertical slice (verse query + both clients) -> M2 core
  tools + cache -> M3 all 29 tools -> M4 Prompts/Resources -> M5 UX (setup/doctor/stop) -> M6 release.
- Original Python project to mine for contracts: `D:\coding\信望愛\FHL-MCP-Server`
  (`http_server.py` has the 29 tools; includes uncommitted/untracked `tools/parsing_reference.py`,
  `utils/parsing_format.py`, `tests/fixtures/parsing_reference_james_1_1.md` that must be included).

## Decisions & assumptions
- Used `.claude/CLAUDE.md` (not repo-root `CLAUDE.md`) as the rules file; assumption that this is what
  the user meant by ".Claude".
- Log language = English; index `Read when` column added to help AI triage. Both are my choices
  within the user's request.

## Verification
- Files were written; directory listing was not re-run after writing. No code existed to test.
- Plan content was read but its claims (e.g. original project's tool counts) were not verified
  against the original project in this session.

## Open issues / next
- Next natural step is M0: inspect `D:\coding\信望愛\FHL-MCP-Server`, produce source manifest
  (file list + SHA-256 + git status), export the 29 tool schemas, and get the user's decisions on the
  open items in plan §15. Log that as `261016b-...` (or the date it happens).
- Consider whether the repo needs `.gitignore`, `go.mod`, and `LICENSE` at M1 start.
