## Why

The Copilot auto-approve path caps total PR edits at 300 lines (`additions + deletions`), which is tight for larger test-only or Terraform-only changes that still fit the file allowlist. Raising the cap to 1000 reduces false rejections while keeping a bounded review surface.

## What Changes

- Raise the Copilot category diff threshold in the canonical spec from strictly less than 300 to strictly less than 1000 total line edits.
- Align the `scripts/auto-approve` implementation and unit tests with the new threshold and failure messages.

## Capabilities

### New Capabilities

- (none)

### Modified Capabilities

- `ci-pr-auto-approve`: Update REQ-009 (Copilot diff threshold) and the large-PR scenario to use 1000 instead of 300.

## Impact

- `openspec/specs/ci-pr-auto-approve/spec.md` (after apply / sync).
- `scripts/auto-approve/evaluator.go` (`maxEditedLines`).
- `scripts/auto-approve/evaluator_test.go` (boundary and message expectations).
