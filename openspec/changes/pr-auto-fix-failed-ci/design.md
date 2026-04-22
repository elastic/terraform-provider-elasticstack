## Context

GitHub now supports asking `@copilot` to make changes on an existing pull request, but that documented flow is centered on comments from users with write access, not on comments authored by GitHub Actions. This repository already uses GitHub Agentic Workflows (GH AW) with direct Copilot execution, so a failed-CI remediation workflow can use the same pattern instead of relying on bot-authored comments as an implicit trigger.

The workflow also needs to be reusable across different PR authors. A label-based contract is more stable than "Copilot-owned PR" detection because the same behavior should later apply to Renovate or other trusted automation.

Because `workflow_run` is privileged, the design must keep all repository-authored gating deterministic and must avoid checking out or running pull-request code until the run is known to be safe and in scope.

## Goals / Non-Goals

**Goals:**
- Provide a deterministic, label-gated remediation path for failed CI on same-repository pull requests.
- Feed Copilot structured failure context instead of asking it to discover the failing job from a free-form PR comment.
- Support two high-value remediation profiles first: lint failures and acceptance test failures.
- Give maintainers visible PR feedback when automation skips a run or analyzes a failure without producing a fix.
- Keep the contract reusable for bot-authored PRs beyond Copilot by gating on `auto-fix`, not on PR author identity.

**Non-Goals:**
- Supporting pull requests from forks in the first iteration.
- Attempting to auto-fix every possible CI failure class.
- Using GitHub Actions comments as the primary mechanism for triggering Copilot work.
- Building a general retry scheduler or long-lived state machine for repeated failed runs.

## Decisions

### 1. Use a dedicated `workflow_run` remediation workflow

Create a new GH AW workflow that runs when `Build/Lint/Test` completes, and only continue when the source run concluded with failure for a pull-request event.

Why:
- It gives the remediation workflow direct access to Actions metadata such as failed jobs and job URLs.
- It fits the repository's existing GH AW pattern for direct Copilot execution.
- It avoids the undocumented edge of relying on bot-authored `@copilot` comments to start work.

Alternatives considered:
- Post an `@copilot` PR comment from Actions: rejected because bot-authored comments are not the documented trigger contract.
- Add label-triggered remediation independent of CI results: rejected because the failure context would be weaker and more expensive to reconstruct.

### 2. Use `auto-fix` as the explicit opt-in gate

Eligibility will be controlled by a pull-request label named `auto-fix`. The workflow will not infer eligibility from PR author, commit author, or branch naming.

Why:
- It keeps maintainers in control of which PRs are allowed to self-remediate.
- It makes the workflow reusable for Copilot, Renovate, and future automation without new heuristics.
- It keeps the label semantics simple: present means opted in, absent means do nothing.

Alternatives considered:
- "Copilot-owned PR" detection: rejected because it excludes Renovate and is brittle when humans and bots both push commits.
- Auto-remediate all same-repo bot PRs: rejected because it is too broad and harder to trust operationally.

### 3. Resolve the PR deterministically from the failed run without modifying the main CI workflow

The workflow will resolve the associated PR deterministically from the completed run by requiring a pull-request-triggered source run and querying open same-repository pull requests for the run's head branch. It will require exactly one matching PR and the `auto-fix` label before agent execution starts.

Why:
- It keeps the first version decoupled from changes to `Build/Lint/Test`.
- It avoids introducing artifact handoff purely to recover the PR number.
- It gives deterministic skip reasons for fork, missing-label, or ambiguous-branch cases.

Alternatives considered:
- Upload PR metadata artifacts from the source CI workflow: rejected for v1 because it couples the core CI workflow to remediation-specific plumbing.
- Resolve PR identity after checkout inside the agent: rejected because it moves trusted gating into non-deterministic reasoning.

### 4. Classify failures before prompting Copilot

The deterministic pre-activation path will inspect the failed run's jobs through the Actions API and classify failures into supported remediation profiles:
- `lint`: the `Lint` job failed.
- `acceptance`: one or more `Matrix Acceptance Test` jobs failed, with version-specific context captured from the failed matrix jobs.

The workflow will also capture the source run URL, failed job URLs, and a concise failure summary for the prompt and PR feedback.

Why:
- It produces stable, targeted prompts instead of asking Copilot to reverse-engineer the failing job from the repository state.
- It lets the workflow skip unsupported failures early instead of invoking Copilot on low-signal or risky cases.

Alternatives considered:
- Pass only the run URL and let the agent figure it out: rejected because it wastes context budget and increases failure ambiguity.
- Parse full logs for every job deterministically: rejected for v1 because it is more brittle than job-level classification and links.

### 5. Use profile-specific remediation policy

The agent prompt will follow different rules for each supported failure profile:
- For `lint`, Copilot should fix the lint errors and push the branch update when it can produce a concrete fix.
- For `acceptance`, Copilot should analyze the failing version(s), attempt a fix only when there is a clear path, and otherwise create or update a PR comment summarizing the issue and why it was not auto-fixed.

If both supported profiles fail in the same run, the prompt will include both sets of failures so the agent can address them in one remediation pass.

Why:
- Lint failures are usually localized and safe to remediate directly.
- Acceptance failures are higher variance and often need explanation rather than blind code churn.

Alternatives considered:
- Treat every failed job the same: rejected because acceptance failures need a stricter "clear path" policy.
- Never push fixes for acceptance failures: rejected because some failures are obvious and worth correcting automatically.

### 6. Keep feedback visible and idempotent

The workflow will create or update a marker-based PR comment for non-fix outcomes such as unsupported failures, skipped runs, or analysis-only acceptance outcomes. The marker will be keyed to the source workflow run so reruns update the existing comment instead of spamming the PR.

Why:
- Maintainers need to know why the workflow did nothing or why it stopped short of a fix.
- Marker-based updates match an existing repository pattern for CI comments.

Alternatives considered:
- No PR feedback on skip/analyze-only outcomes: rejected because the workflow would look silent and confusing.
- Always post a new comment: rejected because reruns would spam the PR.

### 7. Support follow-up CI through documented push configuration

When Copilot produces changes, the workflow will update the PR branch through the GH AW push safe output. The design will document the repository configuration needed for those agent-authored pushes to trigger CI again, using the GH AW CI trigger token mechanism when enabled.

Why:
- A remediation workflow is much more useful when the resulting branch update automatically reruns validation.
- The trigger mechanism is a repository-level operational concern and should be explicit in the design and docs.

Alternatives considered:
- Leave CI retriggering entirely manual: rejected because it weakens the self-healing workflow too much.
- Hard-code a workflow-specific PAT contract in the prompt only: rejected because that hides an operational prerequisite in implementation detail.

## Risks / Trade-offs

- [Privileged follow-on workflow after untrusted PR code] -> Mitigation: keep PR resolution, label checks, fork rejection, and failure classification deterministic; do not checkout PR code until those gates pass.
- [Repeated failed reruns could trigger repeated remediation attempts] -> Mitigation: keep the label as an explicit maintainer-controlled gate, dedupe comment output per source run, and document that broader retry budgeting is a follow-up concern if loops become noisy.
- [Failure context may be too weak for some acceptance failures] -> Mitigation: support only targeted profiles in v1, include direct job links and version context, and fall back to PR summaries instead of speculative fixes.
- [Agent-authored pushes may not rerun CI unless repository settings are configured] -> Mitigation: design the workflow around the GH AW CI trigger path and document the required repository secret/setting.
- [Branch-based PR resolution can be ambiguous in unusual repository states] -> Mitigation: require exactly one open same-repo PR match and skip deterministically otherwise.

## Migration Plan

1. Add the authored workflow source, deterministic gating logic, and compiled lock artifact for the new remediation workflow.
2. Add any small helper logic or tests needed to make PR resolution, failure classification, and comment idempotency deterministic.
3. Document the `auto-fix` label contract, supported failure profiles, and the CI-trigger configuration required for follow-up runs after agent pushes.
4. Roll out by applying `auto-fix` only to selected same-repo PRs first, then broaden usage once the workflow behavior is understood.

## Open Questions

- None for the initial proposal. Supporting additional failure classes or adding retry budgets can be proposed later as follow-up changes.
