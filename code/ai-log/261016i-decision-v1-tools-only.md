# Decision: first release = 29 Tools only; Prompts/Resources deferred to M4

- Date: 2026-10-16
- Plan ref: §1, §4, §5, §12, §13 (M4), §15 item 2, §17
- Status: done
- Files touched: `doc/FHL-MCP-Go-重寫計畫書.md`, `.claude/CLAUDE.md`, `app_versions.md` (TODOs)

## What
Recorded the user's decision (chat, after [261016h](261016h-m0-prompts-resources-inventory.md)): first release delivers only the 29 Tools;
19 Prompts and 7 Resource URIs move to the next version. Plan updated: §1 (decision + first-release definition of done = M0-M3, M5, M6),
§4 (section retitled "下一版本範圍", added pointer to inventory), §5 (diagram + `internal/resources`, `internal/prompts` marked M4/not created),
§12 (v1 test: `initialize` must not advertise prompts/resources), §13 (M4 row marked deferred, M5/M6 don't wait on it), §15 item 2 (resolved), §17 checklist item struck.

## Why
User choice; also M0 showed these exist only in Python's stdio server and HTTP exposes 0, so no HTTP baseline exists for them.

## How
Targeted edits only; milestone numbers kept (M4 not renumbered) so existing ai-logs/app_versions references stay valid.
CLAUDE.md "project state" paragraph now says v1 = Tools only and forbids building the two packages / advertising the capabilities in v1.

## Decisions & assumptions
- Decided by user: scope above.
- My assumption (made explicit in plan §4): `internal/mcpserver` registration is designed to accept resources/prompts later, but no empty packages and no empty capabilities in v1.
- Still open (§15): platform minimums (Ubuntu/macOS), signing/notarization, Smithery/root-path compat.

## Verification
Doc edits only; re-read not repeated beyond edit success. Nothing run.

## Open issues / next
- M0 step 4: per-tool output sampling for the 27 unsampled tools + differences table (compatible / bug fix / approved change / unresolved).
- M0 step 5: ask the user remaining §15 items, then M0 exit check and M1 start (Go/SDK version pinning).
