# M5: lifecycle commands, client setup, double-click entries (0.0.0.21)

- Date: 2026-10-16
- Plan ref: §6, §7, §13 M5 (setup/doctor/status/stop/disconnect, double-click entries)
- Status: done (Windows-verified only; see Verification)
- Files touched: cmd/fhl-mcp/{main,lifecycle,setup,doctor}.go (+ m5_test.go), internal/{runstate,clients}/ (new), internal/app/{lifecycle,smoke}.go (+ lifecycle_test.go), internal/app/{serve,cachecfg}.go, internal/transport/{control,http}.go, scripts/*, .gitattributes, doc/快速開始.md, doc/…計畫書 §7 wording, app_versions.md, .claude/CLAUDE.md, internal/version/version.go

## What
New subcommands `setup`, `status`, `stop`, `doctor`, `disconnect`; `serve` now publishes a run-state file and authenticated control endpoints and reports "already running" instead of failing on its own port. Four double-click scripts in `scripts/`. Quick-start doc.

## Why
Plan §1 defines first-release done as including setup/doctor/status/stop/disconnect; §6/§7 spell out the safety rules (no unauthenticated stop, identity not just port, never PID-only, never overwrite client config, read back after add).

## How
- Control channel (plan allows "local channel OR loopback + token"): same listener, `GET /_control/info` + `POST /_control/stop`, bearer token from `<data>/run/state-<port>.json` (0600, atomic, removed on exit only if instance matches). Guards: constant-time token compare, loopback Host check, `http.CrossOriginProtection`. Rejected: separate unix socket/named pipe (platform-specific, CGO_ENABLED=0 pure-Go parity cost) and PID signalling.
- Identity (`app.Probe`): `/health` must say `name == FHL-MCP-Go`; then state file + token-authenticated `/info` must return the same instance id -> `running`; otherwise `running_unverified` (stop refuses) / `port_in_use` (foreign) / `stopped` (+ stale state). Health is probed BEFORE reading state: Serve writes state before it answers, reading in the other order is a race (hit it in tests).
- State file keyed by port inside the data dir, so different ports coexist; status/stop/doctor must be given the same `--data-dir`/`--port` as serve (documented).
- Client registration only via official CLIs (`internal/clients`, `Runner` seam). Add is read back via `get`; "not registered" is recognised only by a not-found message, any other `get` failure is an error. Disconnect removes only loopback `/mcp` entries.
- `.cmd` files: `chcp 65001` must precede any non-ASCII line (cmd parses with the OEM code page first; a Chinese `rem` swallowed the newline). Found by running them.

## Decisions & assumptions
- Assumed (plan left open): control = same-port token endpoints; state dir = `<data>/run`; `serve` when already running exits 0; `stop` when nothing runs exits 0; `status` exits 1 when not running; setup without `--client` asks on a TTY, otherwise = all; missing CLI is not a failure under "all" unless nothing succeeded. `doctor` fails (exit 1) only on ✘.
- Codex `mcp get --json` shape is inferred (recursive `url` lookup + text fallback); Codex CLI is not installed here. Recorded in TODOs.
- `--config` / config file / `FHL_` vars (plan §8) NOT done; flagged in TODOs and plan §7.

## Verification
`go vet`, `CGO_ENABLED=0 go test ./...` pass; builds for windows/amd64, darwin/arm64, linux/amd64. Real runs on Windows: serve/status/second serve/doctor/stop/status on port 18081 with a temp data dir; real `claude` CLI `setup --client claude` (already registered -> "already" + smoke test OK); `stop` against the user's still-running old 0.0.0.20 server on 8081 correctly refused; doctor against real upstream OK; both `.cmd` entries run. NOT verified: Codex CLI, macOS/Ubuntu, `.command` scripts, real `disconnect` against the user's actual registration (deliberately not run), ACLs of the state file on Windows, IPv6 host.

## Open issues / next
- The user's running server on 8081 is the old build: it must be closed by Ctrl+C once, restarted with 0.0.0.21; only then `stop` works.
- Next per plan: M6 packaging (scripts copied next to the binary, tar.gz/zip, CI, resource measurement, native acceptance on Mac/Ubuntu), upstream retry/rate limit, config file/env vars, Codex verification.
