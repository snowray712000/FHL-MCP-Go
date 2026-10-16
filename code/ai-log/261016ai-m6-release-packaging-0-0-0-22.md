# M6 release packaging, license, 0.0.0.22

- Date: 2026-10-16
- Plan ref: §11, §10, §12, §13 M6, §15
- Status: partial (Windows-verified only; Mac/Linux archives built, never run)
- Files touched: `tools/release/{main.go,main_test.go}`, `release/NOTICE-head.md`, `LICENSE`, `.gitattributes`, `internal/version/version.go`, `cmd/fhl-mcp/main.go`, `doc/快速開始.md`, `app_versions.md`, `.claude/CLAUDE.md`

## What
- `go run ./tools/release` builds windows/amd64, darwin/arm64, linux/amd64 (`CGO_ENABLED=0 -trimpath`, commit + commit-date stamped via ldflags) into `dist/` as `fhl-mcp-go_<ver>_<os>_<arch>.zip|.tar.gz` + `SHA256SUMS.txt`. Archive = one top folder with binary, LICENSE, THIRD_PARTY_NOTICES.md, 快速開始.md, and per-OS double-click scripts (Windows `.cmd` as CRLF, Mac `.command` 0755, Linux none).
- Output is deterministic (sorted entries, mtime = commit time, uid/gid 0): two runs gave identical `SHA256SUMS.txt`.
- Tool refuses a dirty tree (unless `-allow-dirty`, stamped `-dirty`) and a `Version` / `app_versions.md last` mismatch; runs `version` on the host-platform binary; `audit` enforces a file whitelist and that no file contains the repo path, home dir or TEMP path.
- `THIRD_PARTY_NOTICES.md` = `release/NOTICE-head.md` + LICENSE of every module `go list -deps` links for that target (read from the module cache).
- `internal/version`: `Commit`, `Date`, `Describe()`; `fhl-mcp version` appends them when stamped.
- Release is 0.0.0.22 (user-specified). Not committed (user did not ask).

## Why
Plan §11 (packaging, checksums, no leaks, license/third-party notice) and M6; the user asked to proceed with M6 as 0.0.0.22.

## How
- Stdlib-only Go tool instead of PowerShell/bash so one implementation runs on any dev OS and can be unit-tested.
- Reproducibility chosen over wall-clock build dates: `Date` is the commit date.
- GPL text: official `https://www.gnu.org/licenses/gpl-3.0.txt` (35149 bytes) downloaded once with curl, saved as `LICENSE` (sha256 of the file when downloaded: 3972dc97…6986). No local copy existed.
- Rejected: CI workflow (repo has no remote; host unknown), signing/notarization (plan §15: beta unsigned), a Linux double-click script (plan: Ubuntu gets a no-desktop doc).

## Decisions & assumptions
- User decision (2026-10-16): license GPL-3.0-or-later; copyright holder 財團法人信望愛資訊文化藝術基金會.
- Compatibility check: go-sdk Apache-2.0 (+ some MIT), jsonschema-go/segmentio/* MIT, uritemplate and x/* BSD-3 — all GPLv3-compatible. The Python reference is MIT; its notice (and its "Bible content is not covered" note) is reproduced in the notices. No GPL headers added to source files (not required).
- Assumed: quickstart's Gatekeeper/SmartScreen sentence recommends only the OS's normal confirm-and-allow flow, never disabling protection (plan §11 item 6). Not verified on real machines.
- Assumed: the .exe/binary name stays `fhl-mcp`; archive prefix `fhl-mcp-go` per plan §11 table.

## Verification
- `go vet ./...`, `CGO_ENABLED=0 go test ./...` pass (new `tools/release` tests: audit rejects repo path / stray files, tar modes + determinism, zip entries).
- Windows 10, extracted zip to `%TEMP%\fhl 測試 路徑\` (Chinese + space): `version` OK; `serve --port 18082 --data-dir …`: `/health` ready in 0.6 s, WS 12.5 MB / private 49 MB, 60 s idle CPU 0.000 s; initialize + 100 `get_bible_verse` over MCP all ok (max 625 ms upstream, cache hits ≈20 ms), then WS 20 MB / private 52 MB / 329 handles / 11 threads, 50 cache files; `stop` 0.12 s, no listener, no process, no run-state file left. (First attempt showed 100 failures caused by my PowerShell sending the body in the wrong encoding, not by the server; fixed by sending UTF-8 bytes.)
- NOT verified: any run on macOS/Linux (only archive listing: modes 0755 on binary and `.command`), Windows 11, 5-minute idle (only 60 s), cold start/antivirus scan, race detector, vulnerability scan (govulncheck not installed), SmartScreen/Gatekeeper prompts, non-ASCII `--data-dir` stop on Mac/Linux.

## Open issues / next
- CI + vulnerability check once a remote exists; native Mac/Ubuntu acceptance; Win 11; beta trial and acceptance report (all listed in `app_versions.md` TODOs).
- Plan §17 checklist is still unticked repo-wide (not maintained by earlier steps); left alone rather than ticking unverified items.
- Release assets are in git-ignored `dist/`; nothing was uploaded anywhere.
