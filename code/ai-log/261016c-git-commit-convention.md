# Adopt git commit convention from FHLNUI

- Date: 2026-10-16
- Plan ref: none (repo process)
- Status: done
- Files touched: `.claude/CLAUDE.md` (new "Git commit convention" section), `code/ai-log/index.md`

## What
Added a "Git commit convention" section to `.claude/CLAUDE.md`, copying the user's habit from the sibling
project `D:\coding\信望愛\FHLNUI\FHLNUI`. No commits were made in this repo.

## Why
User asked to replicate the commit habit of FHLNUI but no longer remembered the original prompt, so it had to
be inferred from that repo.

## How
- User said "FHLNUI 專案"; the actual git repo is the nested `D:\coding\信望愛\FHLNUI\FHLNUI` (outer dir
  has only backups). Read-only inspection: `git log`, `git show --stat`, its `CLAUDE.md`, `ai_wiki/AGENTS.md`,
  `.claude/settings.local.json`. None contains a written commit rule, so the convention is derived from history.
- Observed: one-line Chinese subjects, no body; `{topic}：{detail}` prefix style; `docs:` for doc-only; `發版 X.Y.Z：`
  for versioned releases and `發版 X：整理 A ~ B` roll-ups; `Co-Authored-By: Claude` trailer on newer commits
  (169 of 287 have a Co-Authored trailer); code + CLAUDE.md + docs committed together; main branch `main`.
- FHLNUI-specific bits NOT copied: `app_versions.json` / `currentSWVer` release flow, `vite-build` branch.

## Decisions & assumptions
- (Superseded by 261016d) Originally replaced `發版 {version}` with milestone prefixes; the user reverted that
  to the original `發版 {version}` style.
- Kept "commit only when the user asks" (session rule), since FHLNUI history does not show who triggered commits.

## Verification
History read via git; nothing else to run. The convention is inferred from ~40 sampled subjects plus stats,
not from an authoritative rule file.

## Open issues / next
- Existing files in this repo (`.claude/`, `code/`, `doc/`) are still uncommitted; first commit would be e.g.
  `發版 0.0.0.1：...` (see 261016d).
