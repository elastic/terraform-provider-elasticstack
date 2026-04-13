## Why

The `schema-coverage-rotation` workflow currently asks the agent to count open `schema-coverage` issues and calculate available issue slots before doing any useful work. That gate is deterministic, cost-sensitive, and should run before agent activation so runs with no remaining capacity skip the agent path entirely.

## What Changes

- Add a deterministic pre-activation script and workflow step that counts open `schema-coverage` issues and computes `issue_slots_available`.
- Require the workflow to be authored from a templated source under `.github/workflows-src/`, following the existing `openspec-verify-label` pattern, with the issue-slot logic extracted into a unit-testable helper module.
- Require the workflow to publish the open-issue count, slot count, and skip reason as pre-activation outputs for downstream job conditions and prompt interpolation.
- Gate the agent job on the computed slot count so the workflow skips agent execution entirely when no issue slots remain.
- Narrow the agent instructions so they consume the precomputed slot information instead of rediscovering GitHub issue state.

## Capabilities

### New Capabilities
- `ci-schema-coverage-rotation-issue-slots`: deterministically compute schema-coverage issue capacity before agent activation and skip the agent job when capacity is exhausted

### Modified Capabilities
<!-- None. -->

## Impact

- `.github/workflows/schema-coverage-rotation.md`
- `.github/workflows/schema-coverage-rotation.lock.yml`
- `.github/workflows-src/schema-coverage-rotation/` and `.github/workflows-src/lib/`
- Deterministic pre-activation workflow logic, extracted helper code, and unit tests for the issue-slot calculation path
- Agent prompt inputs and job conditions for schema-coverage rotation runs
