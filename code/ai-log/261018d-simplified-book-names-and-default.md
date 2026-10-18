# Simplified book names and default-simplified option (1.1.2.4)

- Date: 2026-10-18
- Plan ref: §9 D4, §8 config
- Status: done
- Files touched: internal/bible/{simplified.go,books.go,scripture.go,*_test.go}, internal/tools/{tools.go,simplified_default_test.go}, internal/app/{simplifiedcfg.go,serve.go,cachecfg.go,simplifiedcfg_test.go}, cmd/fhl-mcp/main.go, internal/version/version.go, app_versions.md

## What
(1) Simplified Chinese book names resolve (`约翰福音`, `创`, `启示录`, apocrypha/apostolic names). (2) `serve --simplified` / `FHL_SIMPLIFIED` makes `use_simplified` default true when a call omits it.

## Why
User asked (1.1.2.4) for simplified-Chinese support; decided: book names first, then default option with the "minimal" approach (only upstream content follows `gb=1`; our own fixed messages stay traditional).

## How
- `foldSimplifiedBook` maps ~35 simplified chars to the traditional chars used in the tables; `ResolveBookID` and `ScriptureSet.resolve` try the exact name first, then the folded one, so traditional input is unchanged. Citation parsing goes through `ResolveBookID`.
- Default injected centrally in `Registry.Call` (not per handler): absent/null `use_simplified` -> `true`. Explicit values win. Flag precedence: `--simplified` > env > false. No config file exists in the project, so none was added.

## Decisions & assumptions
- Minimal approach for our own messages (user-approved). Config file not built (TODO already existed).
- Finding: upstream `se.php` picks the simplified/traditional table by `gb`, keyword is not converted (`爱` hits only with gb=1, `愛` only with gb=0). Not fixed: needs a s2t/t2s converter (dependency decision); recorded as TODO.

## Verification
`go vet`, `CGO_ENABLED=0 go test ./...` pass. Live probe of `se.php` through `fhlapi.SE` confirmed the keyword finding. Not verified: real MCP client with `--simplified`.
