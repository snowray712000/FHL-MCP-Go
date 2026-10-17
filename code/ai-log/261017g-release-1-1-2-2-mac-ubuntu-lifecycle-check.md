# 1.1.2.2: Mac/Ubuntu lifecycle command check + quickstart npm section

- Date: 2026-10-17
- Plan ref: §7, §13 M5
- Status: done
- Files touched: internal/version/version.go, app_versions.md, doc/快速開始.md, code/ai-log

## What
Bumped to 1.1.2.2. Recorded user-supplied transcripts from Ubuntu (linux/amd64) and Mac M2 (darwin/arm64) running `status`, `doctor`, `doctor --offline`, `stop` (twice), `disconnect` (twice) against fhl-mcp 1.1.1 installed via npm. Included the user's quickstart edit: "取得與驗證" split into "從 github 下載" / "使用 npm 安裝". Narrowed the matching TODO.

## Why
The TODO asked for per-command confirmation on Mac/Ubuntu. Results matched designed behavior, so the verified part is removed from TODOs and the unverified remainder kept explicit.

## How
Fixed a typo in the user's edit: `fhl-map setup` -> `fhl-mcp setup` (also reworded "登入 claude 或 openai 的 codex" to "連接 Claude Code 或 Codex" to match the rest of the doc). Changelog entry is `說明`/`優化` for users; transcript detail under the engineer note.

## Decisions & assumptions
- Read "commit 到 1.1.2.2" as: bump version, changelog, ai-log, commit (no tag, no dist, no npm publish).
- Mac `doctor` output showed Claude Code and Codex both registered, `disconnect` removed both (`claude mcp list` then empty); treated as real-removal verified on Mac only.
- Ubuntu had no registrations and no codex CLI, so only the "nothing to remove / skip" paths were seen there.
- macOS version of the M2 was not stated; the "macOS 13 vs newer" TODO is kept.

## Verification
Transcripts are user-reported, not re-run by the AI. AI ran `go vet ./...` and `go test ./...` for the version bump (see commit turn). NOT verified: Mac `.command` double-click entries, Ubuntu real `disconnect` removal, Ubuntu Codex, port-occupied failure on Mac/Ubuntu, release archives (tests used npm installs).

## Open issues / next
Remaining TODO line in `app_versions.md` lists the unverified items.
