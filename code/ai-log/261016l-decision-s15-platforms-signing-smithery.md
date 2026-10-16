# Decision: §15 remaining items (platform minimums, signing, Smithery)

- Date: 2026-10-16
- Plan ref: §1, §11 item 6, §15, §13 M0 exit
- Status: done
- Files touched: `doc/FHL-MCP-Go-重寫計畫書.md` (§1, §11, §15), `.claude/CLAUDE.md`, `app_versions.md` (TODOs)

## What
Asked the user the three remaining §15 items with recommended defaults; user picked every recommendation:
1. Platforms: Ubuntu 22.04/24.04 x64 + macOS 12+ arm64 (+ Windows x64); final macOS floor locked in M1 after toolchain + real-device check.
2. Signing/notarization: first release is a beta, unsigned and not notarized; docs state SmartScreen/Gatekeeper may block; never advise disabling protection; sign/notarize for the public release.
3. Smithery: out of v1; `/mcp` only, no well-known endpoints, no per-request config middleware, no root `/` -> `/mcp` rewrite; separate work if existing users appear.
Plan §15 rewritten (items 2-6 struck + decided), §1 platform line and §11 item 6 annotated, CLAUDE.md project-state paragraph updated.

## Why
M0 exit requires scope and platform baselines to be fixed; CLAUDE.md forbids silently deciding §15 items.

## How
Used a structured question with the plan's own recommended defaults as the first option. Edited plan via script (scratchpad `upd_s15.py`) because nested shell heredocs failed first time; preserved LF endings.

## Decisions & assumptions
- Decided by user: items 3-6 as listed above.
- Item 1, project name `FHL-MCP-Go`: first treated as an assumption (repo/directory already use it); the user then confirmed it explicitly ("專案名稱 FHL-MCP-Go 確認採用") when releasing 0.0.0.6. Plan §15 item 1 updated to decided.
- macOS 12 is a floor to be validated, not a promise (plan says so).

## Verification
Re-read §15 after the edit (items 1-6 plus closing line). No code run.

## Open issues / next
- M0 exit check: manifest done, schema/contract done, Prompts/Resources inventory done, tool matrix + diffs done and approved, §15 decided -> M0 can be closed.
- M1 start: pin Go + `go-sdk` versions; early spikes: (a) unknown tool returns `isError` result (D2), (b) text + `structuredContent {result}` + `outputSchema` output (D1), (c) Host/Origin validation, (d) macOS 12 support of the chosen Go toolchain.
- (Project name confirmed by the user; nothing pending for M0.)
