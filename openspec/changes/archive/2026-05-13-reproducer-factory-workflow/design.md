## Context

Three factory workflows already exist: `research-factory` (produces a sticky research comment), `change-factory` (produces an OpenSpec change proposal), and `code-factory` (produces an implementation PR). All three share a common pre-activation scaffold: event qualification, actor trust, context normalisation, and artifact upload. The agent job then runs with a time-boxed Claude session.

Bug triage today is manual. A maintainer reads an issue, tries to reproduce it locally, writes a test (or doesn't), and eventually determines whether the bug is current, already fixed, or too vague to reproduce. There is no systematic output from that process. The `reproducer-factory` workflow captures that triage logic as a repeatable agent task.

The key constraint is that the PR produced by this workflow must be **mergeable**: it must contain a test that passes in CI. For a bug reproduction that means the test asserts the failure condition rather than the absence of it (i.e. `ExpectError` / `ExpectNonEmptyPlan`).

## Goals / Non-Goals

**Goals:**
- Agent attempts to reproduce the failure condition from an issue report
- If reproduced: creates a PR containing a passing `TestAccReproduceIssue{N}` test + sticky comment with root cause analysis
- If not reproducible: sticky comment with 3 specific, codebase-referenced investigation avenues
- If appears fixed: sticky comment with evidence (test output + git archaeology)
- Always posts a sticky comment (`<!-- gha-reproducer-factory -->`) regardless of outcome
- Reuses the existing factory infrastructure patterns (pre-activation scaffold, scripts, lib helpers)

**Non-Goals:**
- Fixing the bug (that is `code-factory`'s job)
- Reproducing non-acceptance-testable bugs (e.g. documentation issues, config parsing bugs where the agent cannot write a Terraform acceptance test)
- Chaining automatically into `code-factory` or `change-factory`

## Decisions

### Single agent vs two-phase (research then code)

The research-factory → code-factory pattern is valuable when the research output is independently useful to a human before code is written. For bug reproduction the research is entirely in service of writing the test: the agent reads the issue, understands the bug, writes the test, and runs it in one continuous session. Splitting into two jobs would add latency without adding value. **Decision: single agent job.**

### Two safe outputs (comment + PR) vs separate workflows

The agent needs to post a comment on every run (all three outcomes) and optionally create a PR (reproduced outcome only). This maps naturally onto two safe outputs in the same workflow: `update-reproducer-comment` (max 1, always emitted) and `create-pull-request` (max 1, emitted only when test passes). A separate comment-posting workflow would add unnecessary indirection. **Decision: two safe outputs in the same workflow.**

### Test file placement

Placing the test in the relevant resource package (`internal/kibana/alertingrule/issue_{n}_acc_test.go`) is preferred because it co-locates the reproduction with the code it exercises and makes it visible when a fix lands. The agent falls back to `internal/acctest/reproductions/issue_{n}_acc_test.go` when it cannot confidently identify a single resource. **Confident identification** means the issue title or body names a Terraform resource type (`elasticstack_*`) or a clear human description of one (e.g. "dashboard resource", "kibana alerting rules", "SLO resource"). Ambiguous issues, multi-resource interactions, and provider-level bugs go to the fallback path.

### Outcome detection

The outcome is determined by whether the acceptance test passes:

```
WRITE TEST with ExpectError / ExpectNonEmptyPlan
RUN TEST
  ├── PASSES → bug confirmed → create PR + comment (outcome A)
  └── FAILS
        ├── Agent could not form a credible test config → comment with 3 avenues (outcome B)
        └── Test passes WITHOUT ExpectError (error never fires) → appears fixed → comment with evidence (outcome C)
```

The agent runs the test with the full acceptance-test environment (same `host.docker.internal:9200` endpoint as `code-factory`). If the test passes with `ExpectError` matching the reported failure, reproduction is confirmed. If running without `ExpectError` also passes cleanly, the bug appears fixed.

### Timeout

The `code-factory` is 30-minute timeout / 25-minute budget. `research-factory` is 35-minute timeout / 25-minute budget. The reproducer workflow needs to run acceptance tests (3–5 minutes each, potentially multiple iterations) on top of reading the issue and writing the test. A 65-minute timeout / 55-minute budget gives the agent enough headroom for 3–4 test iterations. **Decision: 65-minute timeout, 55-minute budget.**

### Concurrency

Same as other factories: keyed by issue number, queue rather than cancel in-flight runs.

### Duplicate PR suppression

Same as `code-factory`: check for an existing open PR with `Closes #N` on the `reproducer-factory/issue-{n}` branch before running the agent. If a duplicate exists, skip agent activation and use `noop`.

## Risks / Trade-offs

**False "appears fixed" signals** → Mitigation: the agent must explicitly document the test config it used in the comment so a maintainer can assess whether the config correctly mirrors the reported scenario. The comment must include the test code or a link to the branch where the test was attempted.

**Acceptance test environment availability** → Mitigation: same risk as `code-factory`; the workflow uses the same `host.docker.internal` environment. If the environment is unavailable the agent should emit a `cannot reproduce` comment with "environment unavailable" as one of the investigation avenues rather than a silent failure.

**Time budget** → 55 minutes is generous for a single test, but if an issue is complex the agent may exhaust time during test iteration. Partial-output preference (same as research-factory) ensures the comment is always emitted even if investigation is incomplete.

**Investigation avenue quality** → The "cannot reproduce" outcome is only useful if the 3 avenues are specific. The agent prompt must enforce that each avenue names a file path or code symbol, not a vague direction.

## Open Questions

- Should the workflow dispatch to `code-factory` (or suggest it via a comment label) when a bug is confirmed? Deferred: out of scope for this change, can be wired up later.
- Should the `reproducer-factory` label be removed after activation (mirroring other factories)? Yes — same pattern as all other factories.
