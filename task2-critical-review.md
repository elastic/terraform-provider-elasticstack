# Task 2 Critical Review — `remove-7x-support`

## Findings

### 1. Medium — live OpenSpec requirement still required the removed `7.17` fallback
- **Evidence:** After the Makefile change, the live spec at `openspec/specs/makefile-workflows/spec.md` still said REQ-017 applied when `STACK_VERSION` matched `7.17.%`, `8.0.%`, or `8.1.%`, and its scenario was still titled `Older 7.17 / 8.0 / 8.1 line`.
- **Why it matters:** Task 2.2 explicitly required updating current OpenSpec wording so the fallback no longer mentions `7.17`. Leaving the live spec stale meant the implementation no longer matched the documented requirement, and the task was not actually complete.
- **Fix:** Updated `openspec/specs/makefile-workflows/spec.md` to:
  - require the Docker Hub fallback only for `8.0.%` and `8.1.%`
  - rename the scenario accordingly
  - add an explicit scenario that unsupported `7.x` versions do not get a special fallback

## Verified / no further actionable findings

- **Makefile `filter` syntax is correct.**
  - `ifneq (,$(filter 8.0.% 8.1.%,$(STACK_VERSION)))` is valid GNU Make syntax.
  - Verified with `make -pn`:
    - `STACK_VERSION=8.0.0` → `FLEET_IMAGE := elastic/elastic-agent`
    - `STACK_VERSION=8.1.2` → `FLEET_IMAGE := elastic/elastic-agent`
    - `STACK_VERSION=7.17.13` → no Makefile fallback set
    - `STACK_VERSION=8.2.0` → no Makefile fallback set

- **No other `7.17` references remain in `Makefile`.**
  - A targeted search of `Makefile` found no remaining `7.17` matches.

- **No CI regression found for `8.0` / `8.1`.**
  - The workflow matrix still explicitly uses `fleetImage: elastic/elastic-agent` for `8.0.1` and `8.1.3` in both `.github/workflows-src/test/workflow.yml.tmpl` and generated `.github/workflows/test.yml`.
  - The Makefile fallback still covers `8.0.x` and `8.1.x` for local Compose-based flows.

- **No additional risky regression identified.**
  - The change is narrowly scoped to removing unsupported `7.17.x` fallback behavior while preserving the existing `8.0.x` / `8.1.x` path.
