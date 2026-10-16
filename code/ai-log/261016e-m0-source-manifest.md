# M0 step 1: Python source manifest

- Date: 2026-10-16
- Plan ref: §3 "來源保存程序", §13 M0, §17 checklist item 2
- Status: done
- Files touched: `doc/m0/source-manifest.tsv`, `doc/m0/source-manifest.md`, `app_versions.md` (TODOs line)

## What
Snapshotted the Python reference working tree: 254 files with git state, byte size, SHA-256
(`doc/m0/source-manifest.tsv`) plus a human summary (`doc/m0/source-manifest.md`).
Breakdown: 237 clean, 7 modified, 10 untracked. HEAD `03e9cb0`, pyproject version `0.1.2`.

## Why
Plan §3 requires a traceable baseline before any contract work, and CLAUDE.md warns the reference has
uncommitted/untracked files (parsing_reference etc.) that HEAD does not contain.

## How
In `D:\coding\信望愛\FHL-MCP-Server` (read-only): `git ls-files --cached --others --exclude-standard`,
per-file `git status` code, `stat -c %s`, `sha256sum`. First run broke on a Chinese file name
(`docs/prompt_example/研經.txt`) because git quotes non-ASCII paths; fixed with `-c core.quotepath=false`.
Gitignored dirs (.venv, .cache, __pycache__, egg-info, .coverage) excluded. File-name scan found no secrets.
Cross-check: `http_server.py` has 29 `@mcp.tool`, names identical to plan §4; `server.py`/`smithery_server.py` have 0.

## Decisions & assumptions
- Manifest lives under `doc/m0/` (plan only says "新專案保存 manifest"; location was my choice).
- Did not copy source files into this repo, only hashes; the Python repo remains the reference.
- Manifest hashes raw working-tree bytes (CRLF/LF as on disk, Windows).

## Verification
Checked line count (254 rows + header), no empty hash fields, state counts match `git status`, Chinese path row present.
NOT done: ran no Python tests, did not start the server, did not check file contents for secrets (names only).

## Open issues / next
- M0 step 2: export real `tools/list` schema from the running Python HTTP server (needs `.venv`/uv; Python 3.12.9 present) and freeze the 29-tool contract.
- M0 step 3: inventory 19 Prompts and 7 Resource kinds (`prompts/manager.py`, `resources/handlers.py`).
- M0 step 4: ask user about §15 scope items (Resources/Prompts in first release, platform minimums).
- Re-run the manifest if the Python repo changes before M1.
