## 1. Bump `.custom-gcl.yml` version to align with the new pinned golangci-lint

- [x] 1.1 Change the `version` field in `.custom-gcl.yml` from `v2.11.4` to `v2.12.2`.

## 2. Add the dedicated `golangci-lint` job to `.github/workflows/provider.yml`

- [x] 2.1 Add a new job `golangci-lint` with `name: Go Lint (golangci-lint)`, `needs: classify`, `if: needs.classify.outputs.provider_changes == 'true'`, `runs-on: ubuntu-latest`, and `permissions: contents: read` plus `pull-requests: read`.
- [x] 2.2 Add a pinned `actions/checkout` step (with `persist-credentials: false`) and a pinned `actions/setup-go` step using `go-version-file: 'go.mod'` and `cache: true`.
- [x] 2.3 Add a pinned `golangci/golangci-lint-action` v9 step with `version: v2.12.2` so the action auto-detects `.custom-gcl.yml` and builds/runs the custom binary.

## 3. Modify the existing `lint` job in `.github/workflows/provider.yml`

- [x] 3.1 Replace the `Lint` step's `run: make check-lint` with a step that calls the remaining targets directly: `make setup-openspec check-openspec check-fmt gen check-docs`. Leave the Node, Go, and Terraform setup steps unchanged.

## 4. Wire the new job into the `gate` job in `.github/workflows/provider.yml`

- [x] 4.1 Add `golangci-lint` to the `needs` list of the `gate` job alongside `classify`, `build`, `lint`, and `test`.
- [x] 4.2 Add `PROVIDER_GATE_GOLANGCI_LINT_RESULT: ${{ needs['golangci-lint'].result }}` to the `env` block of the `gate` step.

## 5. Extend `gate-provider.js` to evaluate the new job result

- [x] 5.1 Add `golangciLintResult` to the parameter destructuring and JSDoc of `gateProvider`.
- [x] 5.2 Include `golangciLintResult` in the `jobResults` array alongside `buildResult`, `lintResult`, and `testResult`.
- [x] 5.3 Update the failure-message strings so `golangciLintResult` is reported wherever the other job results are reported.
- [x] 5.4 Update existing tests in `.github/scripts/workflows/lib/gate-provider.test.mjs` so every call site passes a `golangciLintResult` value consistent with the scenario.
- [x] 5.5 Add new test coverage for `golangciLintResult` success, failure, and skipped scenarios (mirroring the existing scenarios for the other lint job).

## 6. Extend the provider gate runner in `.github/scripts/workflows/lib/runners/gate.js`

- [x] 6.1 Add `'GOLANGCI_LINT_RESULT'` to the `fields` array for the `provider` gate entry.
- [x] 6.2 Update the `evaluate` call to pass `golangciLintResult: env.GOLANGCI_LINT_RESULT` to `gateProvider`.
