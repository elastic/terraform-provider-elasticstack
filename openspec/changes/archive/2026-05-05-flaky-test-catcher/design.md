## Context

The repo uses a compiled agentic workflow system where source templates live in `.github/workflows-src/<name>/workflow.md.tmpl` and are compiled to `.github/workflows/<name>.md` by `scripts/compile-workflow-sources/main.go`. The `manifest.json` in `.github/workflows-src/` registers each template/output pair.

Agentic workflows (`*.md` output) follow a two-job pattern: a deterministic **pre-activation** job (pure `actions/github-script` steps) produces gating outputs, and the **agent** job runs conditionally on those outputs. The `schema-coverage-rotation` workflow is the established reference for this pattern.

The repo has a large acceptance test matrix: ~22 Elastic Stack versions × 2 shards, run for every push to `main`. Tests are written in Go (`--- FAIL: TestName`) and named after the Terraform resource they cover (e.g., `TestAccResourceAgentConfiguration`, `TestAccResourceAgentConfiguration_minimal`).

## Goals / Non-Goals

**Goals:**
- Detect broken tests (failing in 100% of recent CI runs on `main`) and flaky tests (failing in ≥ 20% of runs).
- Open one GitHub issue per affected resource, labelled `flaky-test` + `code-factory`.
- Perform commit-based fix detection to surface whether a fix has already been merged.
- No-op cleanly when there are no failures or the issue cap is reached.
- Keep implementation complexity significantly lower than `schema-coverage-rotation` (no Go rotation scripts, no repo-memory).

**Non-Goals:**
- Analysing SNAPSHOT stack versions (CI jobs for SNAPSHOT use `continue-on-error: true` on the test step, which makes them invisible as failures at the job-level API — excluded automatically).
- Fixing flaky tests (delegated to the `code-factory` workflow via label).
- Tracking historical flakiness across runs older than 3 days.
- Per-shard granularity in issue reporting (issues are per resource, not per shard).

## Decisions

### Pre-activation in JS, not in the agent job

**Decision**: Use a deterministic `actions/github-script` pre-activation step to detect CI failures and count issue slots. The agent receives structured outputs rather than discovering this itself.

**Rationale**: Keeps the no-op path cheap and deterministic — if there are no failures, the agent job never starts, burning no agent tokens. Matches the `schema-coverage-rotation` pattern.

**Alternative considered**: Let the agent check for failures itself. Rejected: wastes agent quota on a pure data-fetch operation with no judgement required.

### No Go scripts, no repo-memory

**Decision**: No Go helper programs or repo-memory orphan branch.

**Rationale**: Unlike `schema-coverage-rotation`, there is no rotation state to maintain. Deduplication is handled by checking for existing open `flaky-test` issues via the GitHub API. Fresh CI data is fetched each run. This makes the workflow substantially simpler.

### Fail-rate denominator = total run count from pre-activation

**Decision**: Classify a test as flaky if `fail_count / total_run_count >= 0.20`, where `total_run_count` is the number of workflow runs on `main` in the last 3 days (output from pre-activation).

**Rationale**: The denominator includes all runs (not just those where the test was in scope). Because tests are sharded deterministically, a test that runs in every shard-0 run is compared against the total run count — slightly under-counting the true fail rate. The 20% threshold is a heuristic, so this approximation is acceptable and avoids the complexity of parsing success logs to count per-test attempts.

**Alternative considered**: Parse successful run logs to count actual test attempts per test. Rejected: too expensive (downloading logs from every run, including successful ones).

### Group issues by base test name (`TestAcc[^_]+`)

**Decision**: One issue per base test name (the portion of the test function name before the first `_`), listing all scenario variants within it.

**Rationale**: Issues are actioned by `code-factory` which works at a coherent unit of test logic. One issue per scenario variant would create noise and hit the issue cap quickly. The base test name is derivable purely from the test name string already present in the failure log — no file lookup or package mapping needed.

**Extraction rule**: Apply the regex `TestAcc[^_]+` to the failing test name. `TestAccResourceAgentConfiguration_alternateEnvironment` → `TestAccResourceAgentConfiguration`. Issue title: `[flaky-test] TestAccResourceAgentConfiguration`.

**Alternative considered**: Group by Go package (extractable from the `FAIL <pkg>` line in test output). Rejected for the initial implementation: requires parsing a second line type from the log and the package path is less directly human-readable as an issue title. Can be revisited if the base-name grouping proves too fine-grained.

### Commit-based fix detection: messages + changed file paths

**Decision**: For each affected resource, inspect commits to `main` since the oldest failing run. Check both commit messages and the set of changed file paths for references to the resource or test names.

**Rationale**: Changed file paths (e.g., `internal/elasticsearch/index/resource_test.go`) are a strong signal that a fix was attempted, even when commit messages are terse. Both together cover the common cases without requiring full diff analysis.

**Alternative considered**: Full diff analysis of test files. Rejected: expensive and rarely needed — message + file path is sufficient to flag "this may already be addressed".

### Issue labels: `flaky-test` + `code-factory`

**Decision**: Created issues receive both labels.

**Rationale**: `flaky-test` enables the issue-slot cap query in pre-activation. `code-factory` triggers the existing code-factory workflow to attempt automated remediation.

## Risks / Trade-offs

- **Log size**: CI job logs can be large. The agent must stream or truncate logs rather than loading them entirely into context. Risk: agent hits context limits. Mitigation: agent skill instructs fetching only the relevant test step log section.
- **Fail-rate approximation**: Using total run count (not per-test attempt count) as denominator slightly under-reports true fail rate for sharded tests. Risk: borderline flaky tests (close to 20%) may be mis-classified. Mitigation: the 20% threshold has enough margin; the approximation is consistently in one direction.
- **Issue dedup relies on open issues only**: If a flaky-test issue is closed (e.g., marked as fixed) but the test remains flaky, a new issue will be opened. This is intentional — closed issues represent resolved or dismissed findings.
- **No toolchain setup steps**: The workflow skips Go/Node setup (unlike schema-coverage-rotation). The agent uses only `gh` CLI and `bash`, which are available on standard GitHub Actions runners.
