# M1 start: local Go toolchain check

- Date: 2026-10-16
- Plan ref: §13 M1, §15 (macOS floor)
- Status: partial
- Files touched: none besides this log and its index row

## What
Checked the local machine before pinning versions. `go` is not on PATH and not in the usual install dirs (`C:\Program Files\Go`, `C:\Go`, `%USERPROFILE%\go`, `%LOCALAPPDATA%\Programs\Go`). `winget` exists; scoop/choco do not. git 2.43.0 present.

## Why
M1 needs Go + go-sdk pinned; nothing can be built or tested without a toolchain.

## How
PowerShell probes + web fetch (via subagent) of go.dev/dl JSON and go-sdk release pages. Subagent output is third-party data, partly model-extracted; re-verify before pinning.

## Decisions & assumptions
- Not decided: installing Go. Downloading/installing a file needs explicit user permission (filename/source/size must be stated first).
- Reported facts (unverified by me beyond the subagent): go.dev lists go1.27.2 stable (also go1.26.9); a web search snippet said go1.26.5 — conflicting, check at install time. go-sdk: v1.8.0 is "Latest" (2026-09-14), v1.7.0 previous stable; v1.7.0 go.mod says `go 1.25.0`; v1.8.0 go.mod not read. SDK negotiates 2025-11-25 and older; 2026-07-28 only when `Stateless=true`; v1.8.0 adds `ServerOptions.SupportedProtocolVersions` to narrow versions. Localhost DNS-rebinding protection is on by default (403 for non-localhost Host on loopback); `CrossOriginProtection` option deprecated, wrap handler with `http.NewCrossOriginProtection()`. Unknown-tool behavior (protocol error vs `isError`) NOT documented — spike (a) must test it.

## Verification
Only the PowerShell probes were run by me. No Go code, no install.

## Open issues / next
- Ask user to approve Go install (winget or official zip/msi), then pin go-sdk (likely v1.8.0 after reading its go.mod) and run spikes (a)–(d).
- Python baseline negotiates 2025-11-25; decide whether Go pins `SupportedProtocolVersions` (needs plan note).

## Update (same day, after user installed Go)
- Downloaded `go1.27.2.windows-amd64.msi` (66154496 B) with user approval; sha256 matched go.dev JSON (`dc673c92...5ed9`). Silent `msiexec /qn` failed (1603, no elevation); user ran the msi manually. `go version` = go1.27.2 windows/amd64 (`C:\Program Files\Go`), GOPROXY default, GOTOOLCHAIN=auto.
- Verified via module proxy: go-sdk versions include v1.7.0 and v1.8.0 (2026-09-04 tag time), v1.8.0 `go 1.25.0`. Pin candidate: v1.8.0 (not yet in a go.mod).
- macOS finding (web search, third-party mirror of golang/go commits + Go wiki): Go 1.26 is the last release running on macOS 12; Go 1.27 requires macOS 13+. Confirm at go.dev/doc/go1.27 before relying on it. Consequence: the macOS 12+ floor (plan §15) holds only if release builds use Go 1.26.x (go1.26.9 listed); with 1.27.x the floor becomes macOS 13+. NOT decided: asked the user.
- Still not done: go.mod, spikes (a)-(d).
