# app_versions.md: user-facing vs engineer-note style

- Date: 2026-10-16
- Plan ref: repo process (release notes)
- Status: done
- Files touched: `app_versions.md`, `.claude/CLAUDE.md`

## What
User defined the changelog style: `🆕 新功能`, `✨ 優化`, `🐞 Bug`, `ℹ️ 說明` are end-user facing and written in plain language (新功能/優化 short; 說明 longer). Technical detail is `📝工程師筆記:` (optionally `📝工程師筆記-✨新增` etc.), placed at the end of each version. Rewrote `app_versions.md`: added a style section, converted all 0.0.0.1–0.0.0.6 items to `📝工程師筆記-…` (they were all engineering detail; no user-facing entries invented), rewrote TODOs in plain language, added 0.0.0.7. CLAUDE.md commit-convention bullet updated to point at the style.

## Why
Release notes are read by non-engineers; engineering detail was mixed in.

## How
Scripted prefix replacement for old sections; 0.0.0.7 and TODOs written by hand.

## Decisions & assumptions
- Decided by user: the style above. Assumed: old versions had nothing user-visible (no runnable program existed), so they carry only engineer notes; `⏳ 待決定` kept for TODOs.

## Verification
Read the result head and 0.0.0.7/0.0.0.6 sections; 31 `📝工程師筆記` lines. Not checked: rendering.

## Open issues / next
- Once a runnable build exists, add real `🆕 新功能` lines per release in plain language.
