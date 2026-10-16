# app_versions.md and release-version commit style (0.0.0.1)

- Date: 2026-10-16
- Plan ref: none (repo process)
- Status: done
- Files touched: `app_versions.md` (new), `.claude/CLAUDE.md`, `code/ai-log/261016c-git-commit-convention.md`, `code/ai-log/index.md`

## What
- User reverted the milestone-prefix idea: release commits use the original FHLNUI style `發版 {version}：...`.
- Created `app_versions.md` at repo root as the Markdown counterpart of FHLNUI's `app_versions.json`.
- First version is `0.0.0.1`.

## Why
Keep the user's established workflow: a user-facing changelog plus `發版 X.Y.Z.W：` commits, so the history
reads the same as FHLNUI.

## How
- Read FHLNUI `app_versions.json` (read-only): top `last` field, a `TODOs` block first, then entries newest
  first; each entry is a version + list of items prefixed `🆕 新功能 / ✨ 新增·改進 / 🐞 修正 / ⏳ 等待`;
  roll-up entries (`7.6.1` = 整理 `7.5.16.1 ~ 7.5.16.6`). `index.html` holds `currentSWVer` that must match.
- `app_versions.md` mirrors that: `last`, TODOs (from plan §13/§15), `## 0.0.0.1` with this session's deliverables.
  Written in 正體中文 (user-facing, like the original); ai-log stays English.
- CLAUDE.md commit section now says: `發版 {version}：...`, 4-part versions, changelog file, keep `last` in sync.

## Decisions & assumptions
- Version format `主.次.修.小` is inferred from FHLNUI (`7.5.16.6`); the roll-up rule is copied, not documented by user.
- No Go code exists, so there is no in-binary version to sync yet; recorded as a future requirement
  (plan §11 says version/commit/build date go into the build).
- Emoji set is my guess from the observed entries; `ℹ️ 說明` is an extra I added.

## Verification
Files written. No commit was made, so the `發版 0.0.0.1：...` message has not been exercised.

## Open issues / next
- Not yet committed: `.claude/`, `code/`, `doc/`, `app_versions.md`. Suggested message:
  `發版 0.0.0.1：建立重寫計畫書、.claude 規則、ai-log 與 app_versions.md` — only when the user asks.
