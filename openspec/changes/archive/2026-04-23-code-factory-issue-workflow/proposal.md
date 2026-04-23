## Why

The repository wants a controlled way to turn selected GitHub issues into implementation pull requests through a GitHub agentic workflow. That workflow needs explicit rules for who may trigger it, when issue labels should activate it, and how it avoids opening duplicate `code-factory` pull requests for the same issue.

## What Changes

- Add a new OpenSpec capability for a `code-factory` issue-intake workflow.
- Define the workflow as repository-authored source under `.github/workflows-src/` that generates checked-in workflow artifacts under `.github/workflows/`.
- Define deterministic pre-activation gating for `issues.opened` and `issues.labeled` events so the workflow only activates when the issue carries the `code-factory` label.
- Define trigger trust rules that allow either GitHub Actions or a human actor with repository permission `write`, `maintain`, or `admin`.
- Define duplicate-prevention behavior so the workflow no-ops when an open linked `code-factory` pull request already exists for the triggering issue.
- Define the agent contract for implementing the issue and creating exactly one linked `code-factory` pull request.

## Capabilities

### New Capabilities
- `ci-code-factory-issue-intake`: issue-driven GitHub agentic workflow that validates trusted `code-factory` triggers, prevents duplicate pull requests, and opens a linked implementation PR

### Modified Capabilities
<!-- None. -->

## Impact

- New authored workflow source under `.github/workflows-src/code-factory-issue/` and generated workflow artifacts under `.github/workflows/`
- Deterministic GitHub-script helper logic under `.github/workflows-src/lib/` and inline workflow scripts under `.github/workflows-src/code-factory-issue/scripts/`
- Workflow tests covering trigger gating, issue-opened-with-label handling, and duplicate-PR detection
- Maintainer expectations for how `code-factory` issue automation is triggered and how linked PR reuse is enforced
