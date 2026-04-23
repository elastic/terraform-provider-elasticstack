## Context

The repository has added a `duplicate-code-detector` GitHub Agentic Workflow that periodically inspects recent repository changes for meaningful code duplication and opens follow-up issues for refactoring. The repository-authored workflow is derived from the upstream `gh-aw` source at `https://github.com/github/gh-aw/blob/main/.github/workflows/duplicate-code-detector.md` and then adapted to fit the existing repository pattern: authored workflow source lives under `.github/workflows-src/`, checked-in generated artifacts live under `.github/workflows/`, and workflow behavior is partly deterministic before the agent starts reasoning.

The design needs to preserve the separation between deterministic gates and agent-driven analysis. Counting existing open issues, enforcing a per-run issue cap, and exposing that gate to the prompt are repository-authored concerns that must remain stable and testable. The agent should only handle the semantic duplicate-detection task and issue writing within those precomputed limits.

## Goals / Non-Goals

**Goals:**

- Define a new CI capability for the duplicate-code detection workflow and its compiled artifacts.
- Capture a deterministic issue-slot gate keyed by the `duplicate-code` issue label and a workflow-configured issue cap.
- Specify the workflow's triggers, analysis scope, and issue-reporting contract in a reviewable requirements change.
- Keep issue creation bounded and focused so the workflow does not flood the repository with broad or duplicate reports.

**Non-Goals:**

- Implementing automated refactoring or code changes as part of this workflow.
- Defining a repository-wide duplicate-detection algorithm outside the workflow contract.
- Persisting historical workflow memory beyond the open-issue slot gate in the initial version.
- Covering non-code duplication concerns such as duplicated documentation or workflow definitions.

## Decisions

### 1. Use a generated GH AW workflow with authored source under `workflows-src`

The workflow will be treated as a GitHub Agentic Workflow whose repository source is authored under `.github/workflows-src/duplicate-code-detector/`, generated to `.github/workflows/duplicate-code-detector.md`, and compiled to `.github/workflows/duplicate-code-detector.lock.yml`. That repository source is expected to stay intentionally aligned with the upstream `gh-aw` duplicate-code detector workflow at `https://github.com/github/gh-aw/blob/main/.github/workflows/duplicate-code-detector.md`, except where this repository needs explicit local adaptations such as issue-slot gating, labels, engine configuration, or generated-artifact handling.

Why:
- It matches the repository's existing pattern for authored workflow sources plus checked-in generated artifacts.
- It keeps the editable source concise while preserving generated outputs for review and execution.
- It records the upstream workflow source of truth while still making repository-specific deltas reviewable.
- It makes regeneration and drift checking compatible with existing `scripts/compile-workflow-sources` and `gh aw compile` workflows.

Alternative considered:
- Author the `.md` workflow directly under `.github/workflows/`: rejected because the repository is already standardizing agentic workflow authoring under `.github/workflows-src/`.

### 2. Keep issue-slot gating deterministic and label-based

Before agent execution, the workflow will count open issues with the `duplicate-code` label, subtract that count from a workflow-configured issue cap, and expose the remaining slot count and gate reason through pre-activation outputs.

Why:
- The repository needs a stable, inspectable cap on how many duplicate-code issues can be open or created at once.
- Label-based counting is a simple contract maintainers can understand and manage directly in GitHub.
- Keeping the gate deterministic avoids asking the agent to discover issue capacity ad hoc.

Alternative considered:
- Let the agent search existing issues and decide whether to create more: rejected because issue budgeting should be deterministic and testable.

### 3. Open one issue per actionable duplication pattern

The workflow will report each significant duplication pattern as its own issue rather than bundling multiple patterns into one report, and it will stop at the number of available issue slots.

Why:
- Separate issues are easier to triage, assign, and remediate.
- Per-pattern issues align with the workflow prompt and prevent one noisy report from hiding unrelated refactors.
- The slot cap remains meaningful only if each issue maps to one actionable pattern.

Alternative considered:
- Bundle all detected duplication into a single periodic report: rejected because the resulting issue would be harder to act on and easier to ignore.

### 4. Scope the analysis to meaningful source-code duplication

The workflow contract will require the agent to focus on recently changed source files, cross-reference the broader codebase, and skip tests, generated artifacts, workflow files, and small or boilerplate snippets.

Why:
- The most valuable findings come from maintainability issues introduced or surfaced by recent changes.
- Excluding well-known noisy areas reduces false positives and preserves issue quality.
- The prompt can still allow broader repository cross-reference without turning the workflow into a full-repo style audit every run.

Alternative considered:
- Analyze the entire repository uniformly on every run: rejected because it increases noise and cost without prioritizing recent, actionable changes.

## Risks / Trade-offs

- [Semantic duplicate detection can over-report weak similarities] -> Mitigation: require significance thresholds, exclude noisy file classes, and demand concrete examples in each issue.
- [Open-issue label counts can be wrong if maintainers relabel issues inconsistently] -> Mitigation: make the `duplicate-code` label part of the workflow contract and use it consistently in safe outputs and gating.
- [Generated workflow artifacts can drift from source] -> Mitigation: require authored source, generated `.md`, and compiled `.lock.yml` to stay paired through repository generation commands.
- [A daily schedule can create issue churn if the cap is too high] -> Mitigation: use a conservative cap and expose the gate reason and available slot count to the agent.

## Migration Plan

1. Add the OpenSpec change for the new workflow capability.
2. Implement or refine the authored workflow source, slot helper logic, and generated workflow artifacts while preserving a clear mapping back to the upstream `gh-aw` workflow source.
3. Add or update focused workflow-source tests for the deterministic slot helper behavior.
4. Regenerate the workflow outputs and validate the resulting change.
5. Enable the workflow through the normal repository workflow rollout path.

Rollback is straightforward: remove or disable the workflow source and generated artifacts, and archive or supersede the capability change if the automation is not kept.

## Open Questions

- None for the initial capability definition. Future changes can refine duplicate-detection heuristics or add repo-memory-based deduplication if needed.
