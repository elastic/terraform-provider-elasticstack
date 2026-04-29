# Task 2 Implementation Report: Docker Workflow Support Floor

## Completed Subtasks

- [x] **2.1** Update `Makefile` Fleet image fallback logic so it matches `8.0.%` and `8.1.%`, but not `7.17.%`.
- [x] **2.2** Update comments and current OpenSpec wording for Makefile workflow behavior so older-version fallback language no longer mentions 7.17.

## Changes Made

### `Makefile`
- Removed `7.17.%` from the `filter` pattern in the Fleet image fallback logic.
- Updated the comment above the fallback block to remove the `7.17.x` mention.

**Before:**
```makefile
# Fleet Server image repository. Some older stack versions (notably 7.17.x, 8.0.x, 8.1.x)
# do not publish elastic-agent images to docker.elastic.co, so fall back to Docker Hub.
ifneq (,$(filter 7.17.% 8.0.% 8.1.%,$(STACK_VERSION)))
FLEET_IMAGE := elastic/elastic-agent
endif
```

**After:**
```makefile
# Fleet Server image repository. Some older stack versions (notably 8.0.x, 8.1.x)
# do not publish elastic-agent images to docker.elastic.co, so fall back to Docker Hub.
ifneq (,$(filter 8.0.% 8.1.%,$(STACK_VERSION)))
FLEET_IMAGE := elastic/elastic-agent
endif
```

### `openspec/changes/remove-7x-support/tasks.md`
- Marked subtasks 2.1 and 2.2 as complete (`- [x]`).

## Commits Created

1. `62fc14de` — `build: remove 7.17 Fleet image fallback from Makefile`
2. `9b8e142e` — `chore: mark remove-7x-support subtasks 2.1 and 2.2 as complete`

## Validation

- **Build:** `make build` completed successfully with no errors.
- **Makefile logic:** Verified with a standalone Makefile test that the `FLEET_IMAGE` variable is set correctly:
  - `STACK_VERSION=8.0.0` → `FLEET_IMAGE=elastic/elastic-agent` ✓
  - `STACK_VERSION=8.1.5` → `FLEET_IMAGE=elastic/elastic-agent` ✓
  - `STACK_VERSION=7.17.13` → `FLEET_IMAGE=` (empty, no fallback) ✓
  - `STACK_VERSION=8.2.0` → `FLEET_IMAGE=` (empty, no fallback) ✓
- **OpenSpec spec:** Checked `openspec/changes/remove-7x-support/specs/makefile-workflows/spec.md` — it already described the correct 8.0/8.1-only behavior with no `7.17` mention in the requirement body, so no spec edit was required.

## Blockers / Open Questions

None. Tasks 2.1 and 2.2 are complete and ready for handoff.
