# setup/doctor print an npm install hint when claude/codex is missing (0.0.0.23)

- Date: 2026-10-17
- Plan ref: §7, §13 M5
- Status: done
- Files touched: internal/clients/clients.go, cmd/fhl-mcp/setup.go, cmd/fhl-mcp/doctor.go, cmd/fhl-mcp/m5_test.go, internal/version/version.go, app_versions.md, doc/快速開始.md, scripts/setup.sh, scripts/start.sh, tools/release/main.go

## What
`Client.InstallHint()` returns `npm i -g @anthropic-ai/claude-code` / `npm i -g @openai/codex` plus "open a new terminal, re-run; else npm global dir not on PATH". Printed by `setup` (CLIMissing) and `doctor`. Quickstart gained an Ubuntu `tar -xzf` section.

## Why
User hit "CLI not in PATH" on Windows and Mac; installing both via npm -g and re-running `setup` fixed it on both. The message should say so.

## How
Text only; nothing is installed automatically (plan: setup never installs). Asserted in `TestSetupReportsEachClientAndExitCodes` and `TestDoctor`.

## Decisions & assumptions
Version bumped to 0.0.0.23 (not committed). Not done: fallback probing of common dirs outside PATH (~/.local/bin, /opt/homebrew/bin, nvm); user did not ask, PATH + npm fix sufficed.

## Verification
`go vet ./...`, `go test ./cmd/... ./internal/clients/...` pass. Mac result is the user's report; Ubuntu not run.

## Addendum: Ubuntu wrapper scripts
User unpacked the linux tarball and found no setup/start script (only mac got `.command`). Added `scripts/setup.sh` and `scripts/start.sh` (thin `exec ./fhl-mcp ...`, no pause prompt), included for the linux target and allowed in `audit`. Kept version 0.0.0.23 (code already bumped; release tool enforces version == app_versions `last`). Test build made with `-allow-dirty` (tree uncommitted).
