## 1. Library Scripts

- [ ] 1.1 Update `classifyChanges` in `.github/workflows-src/lib/classify-changes.js` to extend the non-impacting path set: `CHANGELOG.md`, `openspec/**`, `.agents/**`, and `.github/**` except `.github/workflows/provider.yml`
- [ ] 1.2 Add unit tests in the corresponding `.test.mjs` file covering the new skip paths (CHANGELOG, `.agents/`, `.github/` non-workflow, `.github/workflows/provider.yml` triggers CI, mixed changes trigger CI)
- [ ] 1.3 Create `.github/workflows-src/lib/gate-provider.js` implementing provider gate logic: succeeds when all jobs passed or all were legitimately skipped (`provider_changes=false`); fails otherwise
- [ ] 1.4 Add unit tests for `gate-provider.js` covering: all-pass, all-skipped-legitimately, unexpected skip, any failure
- [ ] 1.5 Create `.github/workflows-src/lib/gate-workflows.js` implementing workflows gate logic: succeeds when test passed or was legitimately skipped (`workflow_changes=false`); fails otherwise
- [ ] 1.6 Add unit tests for `gate-workflows.js` covering: test passed, test skipped legitimately, unexpected skip, test failed

## 2. Inline Scripts

- [ ] 2.1 Create `.github/workflows-src/provider/scripts/classify_changes.inline.js` that reads PR file list (pull_request) or defaults to `provider_changes=true` (push/dispatch), using `classifyChanges` from `../../lib/classify-changes.js`
- [ ] 2.2 Create `.github/workflows-src/provider/scripts/gate.inline.js` that reads `classify`, `build`, `lint`, and `test` job results and calls `gateProvider` from `../../lib/gate-provider.js`
- [ ] 2.3 Create `.github/workflows-src/workflows/scripts/classify_changes.inline.js` that reads PR file list (pull_request) or defaults to `workflow_changes=true` (push/dispatch), setting `workflow_changes=true` when any file is under `.github/`
- [ ] 2.4 Create `.github/workflows-src/workflows/scripts/gate.inline.js` that reads `classify` and `test` job results and calls `gateWorkflows` from `../../lib/gate-workflows.js`

## 3. Provider Workflow Source Template

- [ ] 3.1 Create `.github/workflows-src/provider/` directory and `workflow.yml.tmpl` with triggers: `push: branches: [main]`, `pull_request: types: [opened, synchronize, reopened]`, `workflow_dispatch`
- [ ] 3.2 Add `classify` job to the template: always runs, uses `classify_changes.inline.js`, outputs `provider_changes`
- [ ] 3.3 Add `build` job to the template: depends on `classify`, condition `provider_changes == 'true'`, sets up Go and Node 24, runs `make vendor` then `make build-ci` (no workflow-test or hook-test steps)
- [ ] 3.4 Add `lint` job to the template: depends on `classify`, condition `provider_changes == 'true'`, sets up Go, Node 24, Terraform (pinned from `.terraform-version`), runs `npm ci` then `make check-lint` with `OPENSPEC_TELEMETRY=0`
- [ ] 3.5 Add `test` matrix job to the template: copy matrix configuration, stack versions, shard strategy, fleet setup, synthetics, wait-readiness, testacc, snapshot PR comment, diagnostics, and teardown steps from `test/workflow.yml.tmpl`; update dependencies to `[classify, build]`; update condition to `provider_changes == 'true'`; remove any reference to `preflight`
- [ ] 3.6 Add `gate` job to the template: `if: always()`, depends on `[classify, build, lint, test]`, uses `gate.inline.js`
- [ ] 3.7 Add `auto-approve` job to the template: `if: always() && github.event_name == 'pull_request' && needs.gate.result == 'success'`, depends on `[gate]`, sets up Go, runs `go run ./scripts/auto-approve` with `contents: read` and `pull-requests: write` permissions; do NOT include the auto-merge step

## 4. OpenSpec Workflow Source Template

- [ ] 4.1 Create `.github/workflows-src/openspec/` directory and `workflow.yml.tmpl` with the same triggers as provider.yml
- [ ] 4.2 Add `validate` job: always runs, sets up Node 24 with npm cache, runs `npm ci`, runs `make check-openspec` with `OPENSPEC_TELEMETRY=0`
- [ ] 4.3 Add `gate` job: `if: always()`, depends on `[validate]`; succeeds when `validate` succeeded; fails when `validate` failed or cancelled

## 5. Workflows Workflow Source Template

- [ ] 5.1 Create `.github/workflows-src/workflows/` directory and `workflow.yml.tmpl` with the same triggers as provider.yml
- [ ] 5.2 Add `classify` job: always runs, uses `classify_changes.inline.js`, outputs `workflow_changes`
- [ ] 5.3 Add `test` job: depends on `classify`, condition `workflow_changes == 'true'`, sets up Go and Node 24, runs `make vendor`, runs `make workflow-test` then `make hook-test`
- [ ] 5.4 Add `gate` job: `if: always()`, depends on `[classify, test]`, uses `gate.inline.js`

## 6. Manifest and Compilation

- [ ] 6.1 Update `.github/workflows-src/manifest.json`: replace the `test/workflow.yml.tmpl → test.yml` entry with three entries mapping `provider/workflow.yml.tmpl → provider.yml`, `openspec/workflow.yml.tmpl → openspec.yml`, and `workflows/workflow.yml.tmpl → workflows.yml`
- [ ] 6.2 Run `make workflow-generate` to generate `.github/workflows/provider.yml`, `.github/workflows/openspec.yml`, and `.github/workflows/workflows.yml`
- [ ] 6.3 Verify the three generated files are syntactically valid YAML and that action references use commit SHAs

## 7. Auto-Approve Script

- [ ] 7.1 Remove the `generated-changelog` category struct and its registration from `scripts/auto-approve/evaluator.go` (or whichever file defines categories)
- [ ] 7.2 Remove all unit test cases for the `generated-changelog` category from `scripts/auto-approve/evaluator_test.go` and `scripts/auto-approve/main_test.go`
- [ ] 7.3 Run `go test ./scripts/auto-approve/...` and confirm all remaining tests pass

## 8. Workflow Tests

- [ ] 8.1 Update `make workflow-test` test fixtures or test cases that reference `test.yml`, the `Build/Lint/Test` workflow name, the `preflight` job, the `test-validation` job, or `ready_for_review` handling to match the new workflow files and job names
- [ ] 8.2 Add workflow test coverage for the new `gate` job logic in `provider.yml` and `workflows.yml` (legitimate skip succeeds, failure fails gate)

## 9. Build Verification

- [ ] 9.1 Run `make build` and confirm the project compiles cleanly
- [ ] 9.2 Run `make check-lint` and confirm no lint errors
- [ ] 9.3 Run `make workflow-test` and `make hook-test` and confirm all pass

## 10. Post-Deploy (Manual)

- [ ] 10.1 After merging to `main`, update GitHub branch protection: remove required checks `Build/Lint/Test / Build`, `Build/Lint/Test / Lint`, `Build/Lint/Test / Test Validation`; add `provider / gate`, `openspec / gate`, `workflows / gate`
- [ ] 10.2 In a follow-up PR, delete `.github/workflows/test.yml` and `.github/workflows-src/test/` and remove `validate_test_result.inline.js` from the test scripts directory
