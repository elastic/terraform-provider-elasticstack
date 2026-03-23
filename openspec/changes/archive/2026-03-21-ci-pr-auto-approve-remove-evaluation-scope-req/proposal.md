## Why

REQ-001 (“Evaluation scope”) required the script to ignore draft pull requests. The implementation still applies an explicit draft filter (`evaluator.go`). We are removing that filter so draft PRs are no longer excluded by the script on scope grounds, and dropping REQ-001 from the canonical spec so requirements match the new behavior. Renumbering keeps requirement IDs contiguous.

## What Changes

- Remove **Requirement: Evaluation scope (REQ-001)** and its **Scenario: Draft PR ignored** from `openspec/specs/ci-pr-auto-approve/spec.md`.
- Renumber remaining requirements (REQ-002 through REQ-014) to REQ-001 through REQ-013.
- **Implementation**: Remove draft-PR exclusion from `scripts/auto-approve` (e.g. logic that appends `"pull request is draft"` and any early exit tied to it).
- **Tests**: Update or remove table-driven cases that expect draft PRs to be rejected for scope, and add/adjust coverage so behavior matches the spec.

## Capabilities

### New Capabilities

- _(none)_

### Modified Capabilities

- `ci-pr-auto-approve`: Drop REQ-001, renumber subsequent requirements, and align the script and tests with drafts no longer being filtered out as ineligible by scope.

## Impact

- **Runtime**: Draft PRs may now be evaluated (and potentially approved) by the auto-approve script when they satisfy categories and gates; previously they were rejected before category routing.
- **Code**: `scripts/auto-approve/` (evaluator and tests).
- **Docs / references**: Any mention of REQ-001 or “draft ignored” for this script should be updated after renumbering.
