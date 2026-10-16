# M0 step 3: static inventory of 19 Prompts and 7 Resources

- Date: 2026-10-16
- Plan ref: §4 "Resources 與 Prompts 建議完整範圍", §13 M0, §15 item 2
- Status: done (static; no wire-level sampling of stdio)
- Files touched: `doc/m0/prompts-resources-inventory.md`, `doc/m0/contract/prompts-inventory.json`, `app_versions.md` (TODOs)

## What
Inventoried all 19 prompts (names, descriptions, 45 args, required flags, render behavior) and all 7 resource URIs
(4 schemes), plus how `server.py` exposes them. Human summary in `doc/m0/prompts-resources-inventory.md`; machine data
(arguments + rendered text, required-only and no-args) in `contract/prompts-inventory.json`.

## Why
Plan §15 leaves "Prompts/Resources in first release" undecided; M0 needs verified counts and behavior first.

## How
- Read `prompts/templates.py`, `manager.py`, `base.py`, `resources/handlers.py`, `server.py` lines 583-682.
- Instantiated `PromptManager` with the reference venv (read-only, `PYTHONDONTWRITEBYTECODE=1`) and rendered every prompt with
  required-only args and with no args; scratchpad script `dump_prompts.py`. Markdown table generated from the JSON (no hand copying).
- Key finding: these exist ONLY in stdio `server.py`. HTTP exports (step 2) had empty prompts/resources lists.
- `templates.py` is just a compat re-export of `manager.py`: one real registration path.

## Decisions & assumptions
- None decided for the user. Whether Prompts/Resources ship in v1 stays open (§15 item 2). Documented that "7 類" = 7 URIs in 4 schemes,
  and that Python's `list_resources` advertises example URIs (name/uri swapped), with no resource templates.
- Treated stdio behavior as the M4 baseline since no HTTP baseline exists.

## Verification
Counts confirmed by running code: 19 prompts, 4/3/4/5/3 by prefix, 45 args; all render with required args (249-886 chars);
missing required -> `ValueError`. Resource list confirmed by reading `list_supported_resources` (2+1+1+3).
NOT verified: `resources/read` and `prompts/get` over the wire, tool-return shape feeding `read_resource`, per-`render` handling of
non-string args, stdio vs HTTP tool differences (only noted `server.py` has ~20 tool names vs 29), that `unquote` bug actually bites.

## Open issues / next
- Python defects to classify later: no `unquote` in resource paths; raw `int()` ValueError vs `ResourceError`; `info://books` ignores `simplified`.
- M0 step 4: per-tool output sampling (27 tools unsampled) + differences table. Step 5: ask the user the remaining §15 items
  (scope of Prompts/Resources in v1, platform minimums, signing, Smithery compat).
