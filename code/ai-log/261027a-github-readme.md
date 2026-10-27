# GitHub README.md

- Date: 2026-10-27
- Plan ref: none (repo presentation)
- Status: done
- Files touched: README.md, code/ai-log/index.md

## What
Added a root `README.md` (正體中文) for the GitHub landing page: what it is, install (npm / Release zip), quick use, 29-tool table by category, build-from-source, doc links, license.

## Why
Repo had no README; user asked to add one for GitHub.

## How
Content condensed from `doc/快速開始.md`, `.claude/CLAUDE.md`, `app_versions.md`. Kept short and linked to the detailed docs instead of duplicating them. No version number in the README (avoids drifting from `app_versions.md`).

## Decisions & assumptions
- Releases link assumes `https://github.com/snowray712000/FHL-MCP-Go/releases` (from `git remote`); whether GitHub Releases are actually published is unverified.
- Original Python repo is named only, not linked (URL not known).
- Not committed; user did not ask for a commit yet.

## Verification
Only checked that linked local files exist (`doc/快速開始.md`, plan, two guides, `LICENSE`, `app_versions.md`). Links/render on github.com not checked.

## Open issues / next
Commit when the user asks. Add a CI/status badge once CI exists (TODO in `app_versions.md`).
