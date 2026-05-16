## Why

The repository wants to remove unreachable Go code regularly without turning raw `deadcode` output into a blocking CI signal. In this codebase, unreachable-function reports include test infrastructure, interface-contract methods, and other cases where raw reachability alone is not enough to decide whether code should be deleted. A scheduled cleanup workflow can instead use `deadcode` as a deterministic candidate source, attempt one small removal at a time, and rely on build, tests, pull request CI, and human review to decide whether the proposed cleanup is safe.

## What Changes

- Add a scheduled GitHub Agentic Workflow that rotates through `deadcode` candidates in deterministic cooldown-aware order and opens at most one cleanup PR per run.
- Restrict the first iteration to symbols reported dead both with and without tests (`deadcode ./...` and `deadcode -test ./...`).
- Add deterministic pre-activation logic that selects one cooldown-eligible candidate, uses `gopls references` to classify local test references, and rejects acceptance-test-backed candidates before agent execution.
- Instruct the agent to remove the target dead function and, only for prequalified local single-file non-acceptance test cases, remove tests referencing that symbol; otherwise the agent must leave tests untouched or abort if the candidate is invalid.
- Require local verification before PR creation with `make build` (including lint) and unit tests for the impacted package.
- Persist cooldown-only memory for attempted candidates so the workflow avoids retrying the same symbol too frequently.
- Leave merge decisions to humans; maintainers may close incorrect PRs and use that operational feedback to tune the workflow later.

## Capabilities

### New Capabilities
- `ci-deadcode-removal-rotation`: scheduled dead-code cleanup proposal workflow with deterministic candidate selection, conservative local test cleanup, cooldown memory, and PR-based review

### Modified Capabilities
<!-- None. -->

## Impact

- New authored workflow source under `.github/workflows-src/` and compiled workflow artifacts under `.github/workflows/`
- Deterministic helper logic or scripts for dual `deadcode` execution, `gopls` reference classification, cooldown memory, and candidate selection
- A repository memory artifact or file for candidate cooldown tracking
- Maintainer-facing documentation for workflow scope, verification expectations, and review/close semantics
- Small recurring cleanup PRs that remove one dead-code candidate at a time
