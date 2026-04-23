## Context

This change adds the first `code-factory` GitHub agentic workflow for the repository. Unlike the existing issue-opening workflows, this one consumes a single GitHub issue as the source of truth and is allowed to modify the repository by implementing the issue and opening a pull request. That makes the deterministic pre-activation gate more important than usual: the workflow must establish that the event is eligible, that the actor is trusted, and that the issue is not already being served by an open linked `code-factory` pull request before the agent begins making changes.

The repository already has a stable pattern for authored workflow sources under `.github/workflows-src/`, generated workflow artifacts under `.github/workflows/`, and extracted GitHub-script helper logic under `.github/workflows-src/lib/`. This change should follow that pattern so trigger classification and duplicate detection remain reviewable and unit testable outside the compiled workflow.

## Goals / Non-Goals

**Goals:**

- Define one repository-authored GH AW workflow that reacts to trusted `code-factory` issue triggers.
- Support both `issues.opened` and `issues.labeled` so labeling during issue creation activates the same workflow contract.
- Keep actor validation deterministic by treating `github-actions[bot]` as trusted and requiring `write`, `maintain`, or `admin` for human actors.
- Prevent duplicate work by deterministically detecting an existing open linked `code-factory` pull request before agent activation.
- Give the agent a clear contract for creating exactly one linked PR with stable metadata that future reruns can detect.

**Non-Goals:**

- PR follow-up behavior for reviews, comments, or failed CI.
- General issue triage or label management outside the `code-factory` trigger contract.
- A fuzzy or heuristic duplicate detector based only on similar titles or semantic issue text.

## Decisions

### 1. Use one issue-intake workflow authored under `.github/workflows-src/`

The workflow will be authored under `.github/workflows-src/code-factory-issue/`, generated into `.github/workflows/code-factory-issue.md`, and compiled through the repository's existing workflow-generation path.

Why:
- It matches the repository's established GH AW authoring model.
- It keeps deterministic pre-activation steps, helper scripts, and generated outputs aligned with existing review patterns.
- It makes the workflow behavior easy to test and regenerate without editing compiled outputs directly.

Alternative considered:
- Author the generated `.md` workflow directly under `.github/workflows/`: rejected because the repository already standardizes authored workflow sources under `.github/workflows-src/`.

### 2. Accept both `issues.opened` and `issues.labeled`, but only when `code-factory` is present

The workflow will subscribe to `issues` activity for both `opened` and `labeled`. Pre-activation will allow `labeled` only when the incoming label is exactly `code-factory`, and will allow `opened` only when the issue's initial label set already contains `code-factory`.

Why:
- It covers both maintainer labeling after issue creation and automation that applies the label during issue creation.
- It keeps trigger policy deterministic and easy to reason about from the event payload alone.

Alternative considered:
- Support only `issues.labeled`: rejected because it would miss workflows or automations that create the issue with `code-factory` already attached.

### 3. Treat GitHub Actions as a first-class trusted trigger source

Pre-activation will trust `github-actions[bot]` directly. For all other actors, it will query repository collaborator permissions and require `write`, `maintain`, or `admin`.

Why:
- Repository automation already opens issues that should be eligible for downstream automation, such as schema-coverage issues created by GitHub Actions.
- Using explicit repository permission checks for human actors is more precise than `author_association`.
- The bot exception stays narrow because it is limited to the platform-owned GitHub Actions actor rather than any bot account.

Alternative considered:
- Rely on `author_association` alone: rejected because it is less precise than permission-level checks.
- Allow all bot actors: rejected because it would broaden trust without a clear repository policy.

### 4. Define duplicate linkage through stable branch and PR metadata

The workflow will treat a PR as the existing linked `code-factory` PR for an issue when it is open and all of the following are true:
- it carries the `code-factory` label
- it uses the deterministic head branch `code-factory/issue-<number>`
- its title or body includes the triggering issue reference in a deterministic form such as `Closes #<number>`

The agent will be instructed to preserve this branch and linkage format when creating or updating the PR.

Why:
- A stable branch name gives reruns a cheap deterministic key.
- Explicit issue linkage in the PR body makes the relationship visible to maintainers and robust against title edits.
- Requiring both workflow label and linkage metadata avoids reusing unrelated PRs that happen to mention the issue.

Alternative considered:
- Match only on PR title similarity or branch name: rejected because either signal alone is too easy to collide with or edit accidentally.

### 5. Skip agent activation when the deterministic gate fails

Pre-activation will emit scalar outputs describing eligibility, trust, duplicate status, and skip reasons. The agent job will only run when the trigger is valid, trusted, and not already represented by an open linked PR.

Why:
- The workflow should not spend agent capacity re-checking trust or duplicate state.
- Deterministic skip reasons help maintainers understand why a run did not proceed.
- This keeps the agent prompt focused on implementation rather than repository-policy discovery.

Alternative considered:
- Let the agent inspect the issue, actor, and PR list itself before deciding whether to act: rejected because the trust and duplicate policy should be stable, testable, and outside agent judgment.

## Risks / Trade-offs

- [A too-strict duplicate-matching rule could miss a valid previously opened PR] -> Mitigation: require the workflow itself to create the canonical branch and explicit issue reference so reruns converge on one format.
- [Trusting `github-actions[bot]` could allow unintended internally generated issues to trigger work] -> Mitigation: still require the `code-factory` label and issue-event checks before activation.
- [Different issue authors may expect manual relabeling semantics] -> Mitigation: publish deterministic skip reasons so maintainers can see when actor trust or duplicate detection prevented execution.

## Migration Plan

1. Add the new change artifacts for the `ci-code-factory-issue-intake` capability.
2. Implement the authored workflow source, helper logic, and generated workflow artifacts.
3. Add focused tests for event qualification, trust checks, and duplicate-PR detection.
4. Regenerate workflow outputs and validate the OpenSpec change.

Rollback is straightforward: disable or remove the workflow source and generated artifacts, then supersede or archive the capability change if the automation is not retained.

## Open Questions

- None for this initial phase. PR follow-up behavior can be specified in a later change if the repository decides to automate review or CI response flows.
