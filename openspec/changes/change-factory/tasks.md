## 1. Deterministic Intake

- [ ] 1.1 Add repository-local helper logic for `change-factory` trigger qualification, actor trust, duplicate linked pull request detection, and consolidated gate reasons.
- [ ] 1.2 Add focused tests for the deterministic helper logic, covering eligible issue events, ignored labels/events, trusted and untrusted actors, linked PR detection, and unrelated PRs.
- [ ] 1.3 Reuse or adapt existing `code-factory` script include patterns so the workflow source can call the deterministic helpers from pre-activation steps.

## 2. Workflow Source

- [ ] 2.1 Add an authored `.github/workflows-src/change-factory-issue/` workflow template that triggers on issue `opened` and `labeled` events.
- [ ] 2.2 Configure pre-activation outputs for issue context, gate reason, trust status, duplicate PR status, and duplicate PR URL.
- [ ] 2.3 Configure the agent job to run only after deterministic gates pass and to use branch `change-factory/issue-<issue-number>`.
- [ ] 2.4 Configure safe outputs so the agent can create at most one linked pull request labeled `change-factory` and at most one no-op result.

## 3. Agent Prompt and Tooling

- [ ] 3.1 Write the agent prompt to treat the issue title/body as authoritative and create exactly one OpenSpec change under `openspec/changes/<change-id>/`.
- [ ] 3.2 Instruct the agent to create all artifacts required for implementation readiness by the active OpenSpec schema: proposal, design, tasks, and delta specs.
- [ ] 3.3 Instruct the agent to validate OpenSpec artifacts and avoid provider implementation, Elastic Stack setup, Fleet setup, API-key creation, and Terraform acceptance tests.
- [ ] 3.4 Define the no-op behavior for issues that are too ambiguous to propose safely, including a concise clarification reason.

## 4. Workflow Generation and Verification

- [ ] 4.1 Add the new workflow source to the workflow-source manifest so generated workflow artifacts are checked in.
- [ ] 4.2 Generate the workflow markdown and compiled lock artifacts with the repository workflow tooling.
- [ ] 4.3 Run focused workflow helper tests and workflow generation checks.
- [ ] 4.4 Run OpenSpec validation for the new change.
