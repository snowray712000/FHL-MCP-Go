# npm install on Ubuntu failed: stale 0.0.0-stage placeholder (not a package bug)

- Date: 2026-10-17
- Plan ref: §11 (distribution)
- Status: done (diagnosis only; no code change)
- Files touched: none (this log + index only)

## What
`sudo npm i -g @fhlnet/fhl-mcp` on Ubuntu (Node 22.22.1, apt npm 9.2.0, prefix /usr/local) left a launcher that printed "找不到執行檔套件 @fhlnet/fhl-mcp-linux-x64". Cause: `/usr/local/lib/node_modules/@fhlnet/fhl-mcp-linux-x64` was the registry placeholder `0.0.0-stage` (package.json `stub:true`, no `bin/`), so `require.resolve` hit MODULE_NOT_FOUND. A fresh user-prefix install and `npx npm@10` both worked. Fix that worked: `sudo npm uninstall -g` both, `sudo rm -rf /usr/local/lib/node_modules/@fhlnet`, `sudo npm cache clean --force`, `sudo npm i -g @fhlnet/fhl-mcp` -> `fhl-mcp version` prints 1.1.1 linux/amd64.

## Why
Registry times (UTC 2026-10-17): every package first had a `0.0.0-stage` placeholder (00:06-00:07, not created by our tooling); main 1.1.1 was published 00:10, linux-x64 1.1.1 only at 00:14, i.e. main before its platform package. The user's early install landed in that gap; later `npm i -g` did not replace the stub (mechanism unproven; root's npm cache showed "cache hit" for the packument).

## How
Diagnosis delegated to Claude on the Ubuntu box via a read-only prompt; report pasted back. Registry facts checked with `npm view`.

## Decisions & assumptions
No code/version change. Candidate for the next version (needs a new npm version, 1.1.1 is immutable): launcher should print `e.code`/`e.message` instead of swallowing the error and hint "uninstall stale platform package and reinstall"; docs: don't install platform packages on their own; publish platform packages BEFORE the main package; `npm deprecate` the `0.0.0-stage` versions (user action). Also possible quickstart note for apt-npm EACCES (user-prefix recipe).

## Verification
Ubuntu end result is the user's pasted output (clean reinstall works). Cause of "stub not replaced" is a hypothesis.
