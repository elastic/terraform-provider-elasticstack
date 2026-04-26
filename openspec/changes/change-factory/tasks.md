## 1. Deterministic Intake

- [x] 1.1 Add repository-local helper logic for `change-factory` trigger qualification, actor trust, duplicate linked pull request detection, and consolidated gate reasons.
- [x] 1.2 Add focused tests for the deterministic helper logic, covering eligible issue events, ignored labels/events, trusted and untrusted actors, linked PR detection, and unrelated PRs.
- [x] 1.3 Reuse or adapt existing `code-factory` script include patterns so the workflow source can call the deterministic helpers from pre-activation steps.

## 2. Workflow Source

- [x] 2.1 Add an authored `.github/workflows-src/change-factory-issue/` workflow template that triggers on issue `opened` and `labeled` events.
- [x] 2.2 Configure pre-activation outputs for issue context, gate reason, trust status, duplicate PR status, and duplicate PR URL.
- [x] 2.3 Configure the agent job to run only after deterministic gates pass and to use branch `change-factory/issue-<issue-number>`.
- [x] 2.4 Configure safe outputs so the agent can create at most one linked pull request labeled `change-factory` and `no-changelog`, at most one `add-comment` on the triggering issue, and at most one no-op result.

## 3. Agent Prompt and Tooling

- [x] 3.1 Write the agent prompt to treat the issue title/body as authoritative and create exactly one OpenSpec change under `openspec/changes/<change-id>/`.
- [x] 3.2 Instruct the agent to create all artifacts required for implementation readiness by the active OpenSpec schema: proposal, design, tasks, and delta specs.
- [x] 3.3 Instruct the agent to validate OpenSpec artifacts and avoid provider implementation, Elastic Stack setup, Fleet setup, API-key creation, and Terraform acceptance tests.
- [x] 3.4 Define the ambiguous-issue path: a single `add-comment` on the triggering issue listing required facts (mandatory before `noop`), then `noop`, without speculative proposals, `noop`-only completion, or interactive exploration.

## 4. Workflow Generation and Verification

- [x] 4.1 Add the new workflow source to the workflow-source manifest so generated workflow artifacts are checked in.
- [x] 4.2 Generate the workflow markdown and compiled lock artifacts with the repository workflow tooling.
- [x] 4.3 Run focused workflow helper tests and workflow generation checks.
- [x] 4.4 Run OpenSpec validation for the new change.
