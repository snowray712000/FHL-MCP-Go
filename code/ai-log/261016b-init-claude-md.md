# /init: rewrite .claude/CLAUDE.md

- Date: 2026-10-16
- Plan ref: whole plan (digest -> constraints)
- Status: done
- Files touched: `.claude/CLAUDE.md` (rewritten), `code/ai-log/index.md`

## What
Ran `/init`. Rewrote `.claude/CLAUDE.md` (created in 261016a) to the standard `/init` form: required
prefix header, project state, mandatory ai-log workflow, and cross-cutting architecture constraints.

## Why
`/init` asks for a CLAUDE.md with commands + big-picture architecture. The repo has no code, so "commands"
cannot be real; the value is in the plan's constraints that span many future files.

## How
- Checked for existing rules: no README, `.cursor*`, `.github`, project `.codex/`/`.gemini/`, or root
  `CLAUDE.md`. User-level `~/.codex` and `~/.gemini` exist (not read); user was told they can `/import`.
- Kept the file at `.claude/CLAUDE.md` (no root copy) to avoid duplicate rule files.
- Dropped the plan digest's overlap: the digest remains in 261016a; CLAUDE.md keeps only constraints that
  would otherwise require reading several plan sections.
- Explicitly did not invent build/test commands; stated that none exist yet.

## Decisions & assumptions
- Stated plan-intended checks (`go test ./...`, `go vet ./...`, `-race`) as intended, not existing.

## Verification
File written; nothing executable to test. Constraints were transcribed from the plan, not re-checked against
the Python project.

## Open issues / next
- When `go.mod` and first code land (M1), update CLAUDE.md with real build/run/single-test commands.
- Next work item remains M0 (see 261016a).
