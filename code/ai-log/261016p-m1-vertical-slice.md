# M1 vertical slice: serve + embedded contract + get_bible_verse

- Date: 2026-10-16
- Plan ref: §13 M1, §5, §6, §9 (D1-D6), §12
- Status: partial (client verification incomplete)
- Files touched: `cmd/fhl-mcp/main.go`, `internal/{app,transport,mcpserver,tools,bible,fhlapi,version}/`, `assets/` (embedded `tools-list.json`), `tests/fixtures/{upstream,golden}`, `.gitignore`, `go.mod`/`go.sum`, `.claude/CLAUDE.md`, `doc/m0/tool-matrix.md` (D17)

## What
- `fhl-mcp serve|version|help` (default `serve`, `--host` loopback-only, `--port` default 8081 fails if taken, `--log-level`, Ctrl+C graceful shutdown 5 s). `/mcp` (SDK handler wrapped in `http.NewCrossOriginProtection`), `/health` (no upstream), `/` text hint, everything else 404 (no root-path compat).
- `internal/mcpserver`: registers all 29 tools verbatim from the embedded M0 `tools-list.json` via low-level `AddTool`; text content + `structuredContent{result}`; errors = `isError` text `Error executing tool X: ...` without structuredContent; unknown tool = same via receiving middleware (D2). Capabilities = tools only (`listChanged:false`), no logging/prompts/resources.
- `internal/tools`: arg parsing (null==absent D5, integer strings D6, custom short messages D2) and the only real tool `get_bible_verse`; the other 28 return `isError` "not implemented yet in this build".
- `internal/bible`: 86-entry book table copied from the Python source; `ResolveBookID` (no alias table, matches Python `get_book_id`; aliases = D11/M2); verse service; `MarshalIndent2` = Python `json.dumps(indent=2, ensure_ascii=False)` (no HTML escape so `<WG25>` survives).
- `internal/fhlapi`: injectable `Client` (BaseURL/HTTP/UserAgent), `QB`, typed errors; `use_simplified` really sends `gb=1` (D4). No retries/cache yet (M2).
- Fixtures: 4 real `qb.php` responses (4 upstream requests, one per fixture) + 3 goldens extracted from the Python M0 samples.

## Why
M1 exit: minimal vertical slice reachable by real clients; de-risk SDK behavior against the frozen contract.

## How
Layering as in the plan: mcpserver -> tools -> bible -> fhlapi. Tests are default-offline (httptest mocks, fixtures). End-to-end tests use the official SDK client over a real loopback listener.

## Decisions & assumptions (NOT user-approved unless stated)
- Module path `github.com/snowray712000/FHL-MCP-Go` (assumed).
- serverInfo: name `FHL-MCP-Go`, version = product version (Python reported `FHL Bible MCP Server` / `1.29.0` = mcp SDK version). Assumed; easy to change.
- Stateful session mode (SDK default) kept; not yet decided vs `Stateless`. Server restart invalidates old session ids — check client reconnect behavior.
- No `SupportedProtocolVersions` narrowing; client negotiated 2025-11-25 in tests.
- Validation message format `Invalid parameter 'x': <value> - <reason>` / `Invalid parameter 'x': required parameter is missing` (D2 says "custom, concise"; exact wording mine). Lax bool strings ("true"/"1"/"yes"/"on" ...) accepted like Pydantic; integral floats (3.0) accepted for ints.
- NEW difference D17 (needs approval): Python resolves `Titus`/`提多書` to Tobit via the duplicate short name `多`; Go resolves to Titus (56), and bare `多` -> Titus. Added to `doc/m0/tool-matrix.md`. Only statically derived + offline `get_book_id('多')=101`; not reproduced against live Python.
- Upstream requests made this session: 4 fixture fetches + 1 live smoke call (Titus 1:1 through the Go server). Python output text for non-ASCII on this Windows console prints garbled; compared via files/goldens only.

## Verification
- `go vet ./...` clean; `CGO_ENABLED=0 go test ./...` passes (assets, app, bible, fhlapi, tools, version). Python goldens: 3 outputs byte-identical. `tools/list` over HTTP equals the frozen contract for all 29 tools. Host `evil` -> 403, Origin `evil` -> 403, `Host: localhost` -> 200, taken port and non-loopback host rejected, graceful shutdown.
- Cross-compiled (`CGO_ENABLED=0 -trimpath`) windows/amd64, darwin/arm64, linux/amd64 (about 13-14 MB each); only the Windows binary was run (`version`, `serve`, `/health`, live MCP call returned Titus 1:1, content text == structuredContent.result).
- Claude Code 2.1.218: connected and completed `initialize` against the server (server log: session connected/initialized) via `claude -p --mcp-config ... --strict-mcp-config`, but the model call failed ("OAuth session expired") so tools/list and tools/call from Claude Code were NOT verified. `claude mcp list` with a project `.mcp.json` shows "Pending approval" (interactive).
- Codex: no `codex` CLI on PATH; only the desktop app files under `%LOCALAPPDATA%\OpenAI\Codex` and `~/.codex`. NOT verified, config untouched.
- NOT verified: `-race` (needs cgo, no gcc on this shell), macOS/Linux runtime, reconnect after server restart, concurrent two-client session, long-running stream behavior.

## Open issues / next
- Finish client verification: user decides how (see wrap-up): register at user scope with `claude mcp add --transport http --scope user fhl http://127.0.0.1:8081/mcp`, and Codex by hand or by appending `[mcp_servers.fhl]` to `~/.codex/config.toml` (backup first).
- Approve/deny D17; then M2 resolver reuse. Decide stateless vs stateful after real-client reconnect tests.
- Release as 0.0.0.8 when the user asks (version const is still 0.0.0.7 = `last`).

## Update: Claude Code registration (user-approved, Claude Code only)
- User said to verify with Claude Code first. Ran `claude mcp add --transport http --scope user fhl http://127.0.0.1:8081/mcp` (no prior `fhl` entry existed; it wrote `C:\Users\snow\.claude.json`). Server was started from the scratchpad build (`%TEMP%\fhlbuild\fhl-mcp-windows-amd64.exe serve --port 8081`, background).
- `claude mcp list` -> `fhl: http://127.0.0.1:8081/mcp (HTTP) - ✔ Connected` (health check = initialize). tools/list and a real tool call from inside Claude Code still need the user to open a NEW Claude Code session (running sessions do not pick up new servers) and try `/mcp` + a John 3:16 question; not yet confirmed.
- Codex untouched. D17 not yet approved (user did not answer it).
- To undo: `claude mcp remove fhl --scope user`.

## Update: Claude Code end-to-end confirmed (user test, new session)
- User opened a new Claude Code session and asked for 約 3:16; `mcp__fhl__get_bible_verse` succeeded and returned John 3:16 (unv, record_count 1, prev 3:15 / next 3:17), result wrapped as `{"result": "<JSON string>"}`. Go server log shows the new sessions connecting/initializing on 127.0.0.1:8081 (nothing else listens on that port; the Python server was not running), so the call was served by the Go build.
- The leading `「　神` (unpaired quote + full-width space) is upstream data: present in the recorded `qb.php` fixture (as `\u300c\u3000`) and in the Python golden; Go did not introduce it.
- Per-call tool invocations are not logged at info level (only errors); add request logging later if wanted.
- Remaining for M1: Codex (no CLI found), concurrent two-client test, D17 approval.
