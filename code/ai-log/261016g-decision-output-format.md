# Decision: tool output = text content + structuredContent{result} + outputSchema

- Date: 2026-10-16
- Plan ref: §9 "輸出與錯誤", §13 M0
- Status: done
- Files touched: `doc/FHL-MCP-Go-重寫計畫書.md` (§9), `.claude/CLAUDE.md` (Output compatibility bullet)

## What
Recorded the user's decision: the Go server emits both `content[0].text` and `structuredContent: {"result": "<same string>"}`
for all 29 tools, and declares the same `outputSchema` (`{result: string}`, required, title `{tool}Output`) as Python.

## Why
M0 export ([261016f](261016f-m0-tools-schema-export.md)) showed Python does exactly this; plan §9 only mentioned text.
User confirmed in chat ("同意：同時輸出 ... 請在計畫書 §9 補上").

## How
Edited plan §9 (added measured facts + decision + error-path note) and the CLAUDE.md bullet that previously said only "keep text content".

## Decisions & assumptions
- Decided by user: both outputs + same outputSchema; never only one.
- Not decided: Go error format (Python `isError:true` + plain text, with a constructor bug on `get_book_info` unknown book). Treated as "explicit bug fix", format TBD in M0 differences table.

## Verification
Doc edits only; nothing run. Client behavior (Claude Code / Codex consuming structuredContent) is NOT verified — plan §9 now requires testing it in M1.

## Open issues / next
- M0 step 3: inventory 19 Prompts / 7 Resources (static read of `server.py`, `prompts/manager.py`, `resources/handlers.py`).
- M0 step 4: per-tool output sampling + differences table; step 5: ask user the remaining §15 items.
