## 1. Workflow implementation follow-up

- [x] 1.1 Add a change-classification job to `.github/workflows/test.yml` that reports `provider_changes=false` only when the workflow diff is limited to `openspec/**`.
- [x] 1.2 Update the matrix acceptance `test` job to depend on the change classifier and run only when both preflight allows execution and `provider_changes=true`.
- [x] 1.3 Add the `Test Validation` job and repoint `auto-approve` so the workflow exposes a stable acceptance-related required check.

## 2. OpenSpec sync

- [x] 2.1 Sync the delta in `openspec/changes/ci-acceptance-gating/specs/ci-build-lint-test/spec.md` into `openspec/specs/ci-build-lint-test/spec.md` or archive the change once the workflow implementation matches the proposed behavior.
- [x] 2.2 Confirm the canonical CI spec consistently describes the new change-classification, validation, auto-approve, and ready-for-review behavior.

## 3. Verification and rollout

- [x] 3.1 Validate the change artifacts with `npx openspec validate ci-acceptance-gating --type change` or an equivalent project OpenSpec validation command.
- [x] 3.2 Verify that an `openspec/**`-only change skips the matrix acceptance job while `Test Validation` succeeds, and that a provider-impacting change fails `Test Validation` when required acceptance coverage does not succeed.
- [x] 3.3 Update GitHub branch protection or rulesets to require `Build/Lint/Test / Test Validation` instead of the per-version `Matrix Acceptance Test (...)` checks after confirming the new check name from a live run.
