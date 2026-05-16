## Context

Raw `go tool deadcode` output is useful as a source of unreachable-function candidates, but it is not a trustworthy blocking signal in this repository. Reports include shared test helpers, interface-contract methods, and other patterns where code should not be removed automatically based on callgraph reachability alone. The workflow design therefore treats `deadcode` as candidate generation, not as proof.

The initial scope is intentionally narrow:
- only consider functions reported dead both without tests and with tests
- open at most one candidate PR per run
- never auto-merge
- keep test deletion highly conservative
- use build, unit tests, PR CI, and human review as layered safety checks

This repository already uses GitHub Agentic Workflows, so the dead-code rotation can follow the same authored-workflow plus compiled-artifact pattern.

## Goals / Non-Goals

**Goals:**
- Regularly surface small dead-code cleanup PRs without blocking normal CI.
- Keep deterministic pre-activation logic outside the agent wherever possible.
- Start with the highest-confidence deadcode category: symbols dead both with and without tests.
- Allow companion test deletion only when references are confined to one local non-acceptance test file.
- Record candidate attempts in cooldown memory so the workflow does not thrash on the same symbol repeatedly.
- Record deterministic attempt outcome reason codes so the workflow produces useful operational learning over time.
- Provide maintainers a compact periodic summary of recent attempt outcomes and sticky areas.
- Require human merge and allow humans to close incorrect PRs.

**Non-Goals:**
- Removing every symbol that is only justified by tests in the first iteration.
- Building a fully general dead-code classifier for interface contracts, reflection, plugin entrypoints, or test intent.
- Auto-merging cleanup PRs.
- Reconciling long-lived workflow memory against merged PR state beyond simple cooldown behavior.
- Automatically adapting selection policy based on historical outcomes.
- Supporting multi-symbol cleanup batches in one run.

## Decisions

### 1. Use dual `deadcode` runs but only act on `dead-with-tests` candidates

The deterministic pre-activation step will run both:
- `go tool deadcode ./...`
- `go tool deadcode -test ./...`

The first iteration will only consider symbols that appear in both outputs.

Why:
- These are the highest-confidence candidates because tests do not keep them alive.
- This avoids the harder first-version problem of deleting symbols whose only remaining references are tests.
- It aligns with the goal of small, safe cleanup PRs rather than maximum removal throughput.

Alternatives considered:
- Also act on symbols alive only because of tests: rejected for v1 because test-companion inference is riskier.
- Use only one `deadcode` mode: rejected because the dual run gives a cleaner confidence split even if only one side is acted on initially.

### 2. Use `gopls references` as the deterministic reference source

The pre-activation step will use `gopls references` on the candidate symbol position and collect unique referring file paths. This is used to determine whether any surviving test references are tightly local.

Why:
- `gopls` is the maintained successor to `guru` and already provides a semantic references query.
- The workflow only needs a narrow structural signal: how many unique files refer to the symbol.
- Using `gopls` avoids writing a custom reference finder in the first iteration.

Trade-off:
- The `gopls` CLI is an experimental interface, but it is acceptable for the initial workflow given the narrow use and the workflow's other safety layers.

### 3. Deterministic pre-activation excludes acceptance-style local test cleanup candidates

For deterministic candidate shaping, a test file is treated as an acceptance test file if its filename matches `acc_*test.go`. If the reference analysis indicates a single local test file and that file is acceptance-style by filename, the workflow will not pass the candidate to the agent as eligible for test cleanup.

Why:
- Filename matching is cheap and deterministic.
- The repository's acceptance tests commonly follow this naming pattern.
- Acceptance suites are too valuable and too broad for first-version automated test deletion.

Alternatives considered:
- Deterministically parse test bodies for `resource.Test` usage only: rejected as the sole deterministic gate because filename matching is cheaper and simpler for pre-activation.

### 4. The agent gets a second acceptance-test backstop

The agent will be instructed that if the relevant local test file contains `resource.Test` or `resource.ParallelTest`, then the candidate is not valid for automatic removal and the agent must stop without making changes.

Why:
- This catches acceptance-style suites that are not filtered by the deterministic filename rule.
- It provides a second guard close to the actual edit step.

Alternatives considered:
- Rely only on deterministic filename filtering: rejected because it is an incomplete proxy for acceptance-style tests.

### 5. Companion test cleanup is allowed only for a single local non-acceptance test file

The workflow will only ever hand the agent candidates whose test references, if any, are confined to exactly one local `*_test.go` file. The agent may remove tests referencing the target symbol only in that case, and only after the acceptance-test backstop passes. If references span multiple files or packages, the candidate is not eligible for automatic test cleanup.

Why:
- This makes test deletion a narrow locality-based optimization rather than a general proof of test exclusivity.
- It keeps PRs small and easy to review.
- It avoids distributed or shared test coverage in the first iteration.

### 6. Verification before PR uses build plus impacted-package unit tests

Before opening a PR, the workflow must run:
- `make build` with a 10 minute tool timeout
- unit tests for the impacted package

The impacted package is the package containing the removed symbol. If the agent also removes a local companion test file from a different package directory, that test package is also impacted and must be tested too.

Why:
- `make build` includes lint and provides a strong compile-time and static-check baseline.
- Package-local unit tests are the cheapest next layer of verification before PR CI.
- PR CI remains the place for broader and slower checks such as acceptance coverage.

Alternatives considered:
- Build only: rejected as too weak.
- Full repository test suite before PR: rejected as too expensive for the scheduled loop.

### 7. Memory is cooldown-only but records deterministic attempt outcome reasons

The workflow will maintain cooldown memory for candidate attempts. Every attempted candidate updates memory regardless of whether the run ends in build failure, agent abort, invalid-candidate detection, or PR creation. Memory is still used only to defer reselection for a configured window, but each attempt record also stores a deterministic outcome reason code and small structured context such as package path or reference-file count.

Why:
- Cooldown-only memory avoids having to reconcile external PR state in the first iteration.
- Recording attempts regardless of outcome prevents rapid reselection thrash.
- Deterministic reason codes turn failed or skipped attempts into actionable operational feedback without requiring fuzzy agent-authored explanations.
- The memory model remains simple enough for v1 while becoming more useful for future tuning.

Alternatives considered:
- Permanent outcome tracking with PR-state reconciliation: rejected for v1 as unnecessary complexity.
- Free-form agent-authored explanations in memory: rejected because deterministic reason codes are easier to aggregate and trust.

### 8. Periodic summaries expose recent workflow outcomes to maintainers

The workflow will produce a compact maintainer-facing periodic summary of recent attempt outcomes using the deterministic reason codes stored in memory. The summary may be emitted as a markdown artifact every run and/or as a marker-based issue or comment updated on a slower cadence. It will include counts by reason code and the most common sticky packages or areas.

Why:
- It turns routine workflow exhaust into reusable operational knowledge.
- It helps maintainers decide whether future refinements should focus on acceptance exclusion, distributed references, verification failures, or other recurring blockers.
- It improves trust by making the workflow's non-PR outcomes visible instead of burying them in cooldown state.

Alternatives considered:
- No reporting beyond cooldown memory: rejected because it hides the reasons the workflow is not producing PRs.
- Free-form narrative summaries: rejected for v1 because deterministic aggregation is cheaper and more trustworthy.

### 9. Selection is deterministic and cooldown-aware without function-size weighting

The pre-activation step will select exactly one cooldown-eligible candidate per run using a deterministic ordering based on candidate class and cooldown age, with a stable tie-breaker such as symbol identity or path order. The first iteration will not weight candidates by function size or span.

Why:
- One candidate per run keeps the workflow low-noise and easy to review.
- Deterministic selection makes runs reproducible.
- The workflow naturally trends toward lower dead-code volume over time, so adding size-weighting complexity is not justified in the first iteration.

### 10. Build/test failure or invalid-candidate detection records memory and stops the run

If the agent cannot safely proceed, if verification fails, or if the candidate is discovered to be invalid during the backstop checks, the workflow records cooldown memory for that candidate and stops without opening a PR.

Why:
- This keeps each run bounded and predictable.
- It avoids cascading retries or fallback behaviors in the first iteration.

### 11. Human review is the final safety layer

Successful local verification leads to a PR, but not to automatic merge. Maintainers review and merge the PR manually. If CI passes but the cleanup is still semantically undesirable, a human may close the PR.

Why:
- The workflow is intended to learn from operational experience.
- Human review is the right final backstop while confidence is still being established.

## Risks / Trade-offs

- [Dual `deadcode` output still contains non-removable symbols] -> Mitigation: restrict v1 to `dead-with-tests`, verify with build and package tests, and require human review.
- [`gopls` CLI behavior changes or workspace inference is imperfect] -> Mitigation: keep usage narrow, deterministic, and replaceable later if it proves awkward in CI.
- [Acceptance-test detection by filename misses some cases] -> Mitigation: add the agent backstop checking for `resource.Test` and `resource.ParallelTest` before any test deletion.
- [Simple deterministic ordering may not always pick the highest-value cleanup first] -> Mitigation: accept this trade-off in v1 and revisit selection heuristics later if the workflow yield is poor.
- [Cooldown memory accumulates stale entries] -> Mitigation: accept this in v1; add deterministic cleanup later only if it becomes operationally noisy.
- [Outcome summaries could become noisy or too detailed] -> Mitigation: keep reason codes finite and deterministic, and keep maintainer-facing summaries compact and periodic.
- [Some dead candidates that need broader test cleanup will not be removed] -> Mitigation: accept conservative false negatives in the first iteration.

## Migration Plan

1. Add the authored GH AW workflow source and compiled workflow artifacts for the scheduled dead-code rotation.
2. Add deterministic helper logic for dual `deadcode` execution, `gopls` reference classification, acceptance exclusion, cooldown memory, deterministic outcome reason codes, periodic outcome summaries, and single-candidate selection.
3. Add the agent prompt and workflow contract for conservative symbol removal, local test cleanup, invalid-candidate aborts, and verification.
4. Document maintainer expectations for review, merge, PR closure, and interpretation of periodic outcome summaries.
5. Roll out on a schedule with human merge only, then revisit scope expansion after observing several runs.

## Open Questions

- None for the initial proposal.
