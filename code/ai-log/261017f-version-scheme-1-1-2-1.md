# Version scheme (odd/even 3rd number) and 1.1.2.1

- Date: 2026-10-17
- Plan ref: §11
- Status: done
- Files touched: internal/version/version.go, app_versions.md, .claude/CLAUDE.md, tools/npmpack/main.go, tools/npmpack/main_test.go, code/ai-log

## What
User-defined scheme (2026-10-17): version `a.b.c.d`; odd `c` (`1.1.1`, `1.1.3`, `1.2.1`, three-part) = formal release; even `c` (`1.1.2.x`) = small dev changes for R&D, never formally released. Dev builds are rolled up into the next odd release (`1.1.3`, or `1.2.1` for an important feature). This commit is `1.1.2.1`: version bump, changelog entry (npm/Ubuntu stale-stub note, rule text), two TODOs (launcher shows real error; quickstart EACCES recipe), plus the ai-log 261017e.

## Why
npm needs three-part semver, so only odd-`c` releases can be published to npm; dev builds stay GitHub/local only. `tools/npmpack` now rejects non-semver versions (`1.1.2.1`) with a clear message instead of emitting packages npm refuses.

## How
`internal/version.Version = "1.1.2.1"`, `app_versions.md` `last` + rules + `## 1.1.2.1`; CLAUDE.md commit-convention bullet rewritten; `semverRe` guard + `TestRejectsNonSemverVersion`. `tools/release` regex already accepts 3 or 4 parts. No dist rebuilt, no tag.

## Decisions & assumptions
User said "將這以 1.1.2.1 commit"; read as: bump to 1.1.2.1 and commit the pending ai-log with the rule recorded. Launcher/doc fixes are only TODOs (they need an odd-`c` release for npm).

## Verification
`go vet ./...`, `go test ./...` pass. Nothing run on Mac/Ubuntu for this change.
