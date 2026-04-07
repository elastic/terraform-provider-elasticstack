## 1. Restore trusted trigger and deterministic classification

- [x] 1.1 Change `.github/workflows-src/openspec-verify-label/workflow.md.tmpl` from `label_command` to `pull_request_target` `labeled` activation and restore deterministic `verify_label` handling.
- [x] 1.2 Add deterministic pre-activation outputs for verification mode, archive/push eligibility, trigger-label cleanup status, and reason strings based on whether the PR head repository matches the base repository.
- [x] 1.3 Update agent job conditions and prompt interpolation so archive/push eligibility is consumed from deterministic outputs rather than inferred by the agent.

## 2. Split same-repository and fork review behavior

- [x] 2.1 Keep the existing workspace bootstrap path for same-repository pull requests, including local OpenSpec tooling and repository setup.
- [x] 2.2 Add a fork-safe review path that can submit the review without requiring trusted workspace bootstrap or archive/push behavior.
- [x] 2.3 Add a deterministic script step that removes only `verify-openspec`, update permissions for that step, and keep label cleanup out of agent safe outputs.

## 3. Restore and extend source helpers and tests

- [x] 3.1 Restore the source helper files for trigger-label verification, add a deterministic trigger-label removal helper, and add helper coverage for same-repository vs fork classification.
- [x] 3.2 Update `.github/workflows-src/lib/openspec-verify-label.test.mjs` to cover the `pull_request_target` trigger, deterministic label-removal step, permission expectations, and deterministic fork archive/push gating.
- [x] 3.3 Regenerate `.github/workflows/openspec-verify-label.md` and `.github/workflows/openspec-verify-label.lock.yml`, then review the compiled workflow for the expected trigger, permissions, bootstrap branching, and safe-output behavior.

## 4. Align requirements and validation

- [x] 4.1 Update `openspec/specs/ci-aw-openspec-verification/spec.md` to match the restored trigger, restored cleanup path, deterministic execution-mode outputs, and same-repository-only archive/push policy.
- [x] 4.2 Run the relevant OpenSpec validation or status checks for `fork-safe-openspec-verify` and resolve any artifact issues needed to make the change apply-ready.
