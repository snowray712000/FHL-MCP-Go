# tools/npmpack: npm packages from the release archives

- Date: 2026-10-17
- Plan ref: §11 (distribution)
- Status: done (not published; not committed)
- Files touched: tools/npmpack/main.go, tools/npmpack/main_test.go, .claude/CLAUDE.md

## What
`go run ./tools/npmpack [-dist dist] [-out dist/npm] [-scope fhlnet] [-version <internal/version>]` writes 4 reproducible `.tgz`: `@fhlnet/fhl-mcp` (launcher `bin/fhl-mcp.js`, `optionalDependencies` pinned to the same version) and `@fhlnet/fhl-mcp-{win32-x64,darwin-arm64,linux-x64}` (binary 0755, `os`/`cpu` fields, LICENSE, THIRD_PARTY_NOTICES). Prints the `npm publish` commands; publishes nothing.

## Why
User wants `npm i -g @fhlnet/fhl-mcp` (npm org `fhlnet`). Chosen layout A (platform sub-packages, no postinstall). User decision: do NOT bump version/docs/changelog for this; npm step runs after the GitHub release is made, from the same archives.

## How
Binaries and notices are extracted from the dist archives after checking SHA256SUMS.txt, so npm and GitHub ship byte-identical executables. Own tarball writer (`package/` root, fixed mtime from the archive, no gzip time) so `npm publish <tgz>` uploads exactly this. Launcher: resolve platform package, chmod 0755 as a safety net (npm may drop exec bits), spawn with inherited stdio, ignore SIGINT (terminal also signals the child), forward SIGTERM, propagate exit code/signal. `publishConfig.access=public` so no flag is needed.

## Decisions & assumptions
Scope `fhlnet` and package name `fhl-mcp` come from the user's npm org; availability on the registry is NOT verified. Version stays 1.1.1 (npm versions are immutable once published).

## Verification
`go vet`, `go test ./tools/npmpack` (contents/modes/package.json, reproducibility, tamper + missing-version rejection). Real packages built from dist 1.1.1; installed on Windows into a temp prefix from the local .tgz files: `fhl-mcp version/status` through the shim OK, `serve`/`status`/`stop` on port 18099 OK (exit codes propagate, no orphan). NOT verified: Mac/Linux install, real registry publish, Ctrl+C behaviour of the shim in an interactive terminal.
