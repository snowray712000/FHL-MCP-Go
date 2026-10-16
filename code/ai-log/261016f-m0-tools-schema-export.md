# M0 step 2: export tools/list schema from running Python HTTP server

- Date: 2026-10-16
- Plan ref: §13 M0, §12 "契約與功能", §17 checklist
- Status: partial (schema exported; per-tool output sampling not done)
- Files touched: `doc/m0/contract/{README.md,initialize.json,tools-list.json,prompts-list.json,resources-list.json,resource-templates-list.json,sample-calls.json}`

## What
Started the Python reference's HTTP app, exported `initialize`, `tools/list`, `prompts/list`, `resources/list`,
`resources/templates/list`, and made 3 sample `tools/call`. Stored under `doc/m0/contract/`; findings in its README.

## Why
CLAUDE.md/plan: freeze the contract from a real `tools/list`, not README or code reading.

## How
`.venv` python (3.12.13, mcp 1.29.0) running a scratchpad launcher: `uvicorn.run(create_http_app(), host="127.0.0.1", port=18081)`
(real `main()` binds 0.0.0.0:8081; I changed only bind/port). Client = `mcp.client.streamable_http.streamablehttp_client`,
dumped with `model_dump(by_alias, exclude_none)`. `PYTHONDONTWRITEBYTECODE=1`; server killed by port afterward.
Python repo `git status` unchanged (15 lines before/after). Sample calls: `list_audio_versions`, `get_book_info("約")`, `get_book_info("不存在書卷")`.
First attempt to print results failed only due to Windows console encoding; files are UTF-8.

## Decisions & assumptions
- Sample calls limited to tools that looked static/low-traffic, to avoid hitting the FHL upstream (plan §9).
- Contract files are client-deserialized, not raw wire bytes (documented in README).
- Recommended (not decided by user): Go must emit BOTH text content and `structuredContent {"result": string}` + `outputSchema`, since Python does. This amends the plan §9 wording (text only). Needs user confirmation.

## Verification
Counted 29 tools, names match plan §4; confirmed all outputSchema are `{result:string}`; confirmed success results carry identical text and structuredContent.
NOT verified: other 27 tools' outputs, wire-level JSON-RPC, session/reconnect, behavior without the uncommitted Smithery config middleware.

## Open issues / next
- HTTP Prompts/Resources lists are empty (stdio-only in Python) -> inventory 19 Prompts / 7 Resources must come from `server.py`, `prompts/manager.py`, `resources/handlers.py` (static read).
- Known Python bug: `get_book_info` with unknown book -> `InvalidParameterError.__init__() missing ... 'value'`; classify as bug-fix, define Go error format.
- Next: M0 step 3 (Prompts/Resources inventory), step 4 (per-tool output/fixture sampling, differences table), step 5 (ask user §15 scope).
