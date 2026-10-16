# M1: version pin, macOS floor change, early spikes (a)-(c)

- Date: 2026-10-16
- Plan ref: §13 M1, §9 (D1, D2), §15 item 4, §12 (Host/Origin)
- Status: partial
- Files touched: `go.mod`, `go.sum`, `doc/FHL-MCP-Go-重寫計畫書.md` (§1 platform line, §15 item 4, release table row), `.claude/CLAUDE.md`, ai-log 261016m (updated)

## What
- Pinned `go 1.27.2` + `github.com/modelcontextprotocol/go-sdk v1.8.0` in `go.mod` (verified on the module proxy; sdk go.mod needs 1.25.0). Module path `github.com/snowray712000/FHL-MCP-Go` is an ASSUMPTION (plan has none, repo has no remote; derived from the user's GitHub name) — rename later if wrong.
- macOS floor changed by user decision: macOS 13+ arm64, built with Go 1.27.x (Go 1.26 was the last to run on macOS 12). Plan §1/§15/release table + CLAUDE.md updated. macOS 13+ real-device check still pending.
- Ran throwaway spike (scratchpad, not in repo): low-level `Server.AddTool` loading all 29 tools verbatim from `doc/m0/contract/tools-list.json`.

## Why
M1 needs pinned toolchain/SDK and the D1/D2/Host-Origin risks de-risked before building the slice.

## How / Results (go-sdk v1.8.0, Streamable HTTP, stateful)
- (b) D1: low-level `AddTool` with raw `mcp.Tool` (inputSchema/outputSchema as parsed from the contract) + returning `CallToolResult{Content:[Text], StructuredContent:{"result":s}}` works. `tools/list` from Go == Python contract for all 29 tools (deep JSON equality). Low-level handler does NO input validation and returns result as-is, so Pydantic-style errors/coercion (D-items) are ours to implement; do not use generic `AddTool[In,Out]` (it would infer schemas / validate).
- (a) D2: SDK default for unknown tool = JSON-RPC protocol error `-32602 unknown tool "x"` (server.go callTool). A receiving middleware (`AddReceivingMiddleware`, method `tools/call`, check name against own set) returning `CallToolResult{IsError:true, text:"Error executing tool X: Unknown tool: X"}` gives the Python-compatible form. Verified over HTTP.
- (c) Host: default protection works — `Host: evil.example.com` on loopback -> 403; `Host: localhost:port` OK. Origin: NOT checked by default (evil Origin -> 200). Wrapping with `http.NewCrossOriginProtection().Handler(h)` gives: no Origin 200, same-origin 200, evil Origin 403, other-port localhost Origin 403, `Sec-Fetch-Site: cross-site` 403. Note: browser-based clients on another localhost port would be rejected unless allowed via `AddTrustedOrigin`; CLI clients (no Origin) fine.
- initialize negotiated 2025-11-25 with a 2025-11-25 client; capabilities were `{logging, tools{listChanged:true}}` — Python's differs (check against `initialize.json` when building the real server; `logging` capability appears by SDK default, prompts/resources absent as required).

## Decisions & assumptions
- User decided: macOS 13+ / Go 1.27.x.
- Assumed: module path; `go` directive `1.27.2`; `// indirect` mark in go.mod is temporary until code imports the SDK.
- Not decided: whether to narrow `ServerOptions.SupportedProtocolVersions` (v1.8.0) to match Python's 2025-11-25; stateful vs `Stateless` mode (spike used default stateful, which negotiates down to 2025-11-25).

## Verification
Spike built with Go 1.27.2 (`go vet` + build OK) and probed with curl against 127.0.0.1:18081; results above. NOT verified: real Claude Code / Codex clients, spike (d) on macOS (only web-search evidence for the 13+ floor, from a third-party mirror — confirm at go.dev/doc/go1.27), race tests, cross-compilation.

## Open issues / next
- Build the real M1 slice: `cmd/fhl-mcp serve`, `internal/mcpserver` (contract-embedded tool list, unknown-tool middleware, CrossOriginProtection), one real tool (verse query via `internal/fhlapi` + mock), then connect Claude Code and Codex.
- Embed `tools-list.json` (or generated schemas) via `embed`; decide where it lives (not `doc/`).
- Compare `initialize` result with `doc/m0/contract/initialize.json` (serverInfo version, instructions, capabilities).
