## Tasks

### Task 1: Bump `.custom-gcl.yml` version to `v2.12.2`

**File:** `.custom-gcl.yml`

Change the `version` field from `v2.11.4` to `v2.12.2` to align with the golangci-lint version used by the Makefile and the version to be passed to `golangci/golangci-lint-action`.

---

### Task 2: Add `golangci-lint` job to `provider.yml`

**File:** `.github/workflows/provider.yml`

Add a new job named `golangci-lint` with the following properties:

- `name: Go Lint (golangci-lint)`
- `needs: classify`
- `if: needs.classify.outputs.provider_changes == 'true'`
- `runs-on: ubuntu-latest`
- `permissions: contents: read, pull-requests: read`
- Steps:
  1. `actions/checkout` (pinned SHA, `persist-credentials: false`)
  2. `actions/setup-go` (pinned SHA, `go-version-file: 'go.mod'`, `cache: true`)
  3. `golangci/golangci-lint-action` — use the latest pinned SHA for `v9`, with `version: v2.12.2`

The action will auto-detect `.custom-gcl.yml` and build/run the custom binary.

---

### Task 3: Modify the `lint` job in `provider.yml`

**File:** `.github/workflows/provider.yml`

Replace the single `run: make check-lint` step with a step that calls the remaining targets directly:

```yaml
- name: Lint
  run: make check-openspec check-fmt gen check-docs
```

No other steps in the `lint` job change (Node, Go, Terraform setup remain).

---

### Task 4: Update the `gate` job in `provider.yml`

**File:** `.github/workflows/provider.yml`

1. Add `golangci-lint` to the `needs` list of the `gate` job (alongside `classify`, `build`, `lint`, `test`).
2. Add `PROVIDER_GATE_GOLANGCI_LINT_RESULT: ${{ needs.golangci-lint.result }}` to the `env` block of the `gate` step.

---

### Task 5: Extend `gate-provider.js`

**File:** `.github/scripts/workflows/lib/gate-provider.js`

1. Add `golangciLintResult` to the function parameter destructuring.
2. Include `golangciLintResult` in the `jobResults` array alongside `buildResult`, `lintResult`, and `testResult`.
3. Update the failure message string to include `golangciLintResult` where other results are reported.

---

### Task 6: Extend `gate.js` fields for the provider gate

**File:** `.github/scripts/workflows/lib/runners/gate.js`

Add `'GOLANGCI_LINT_RESULT'` to the `fields` array for the `provider` gate entry. Update the `evaluate` call to pass `golangciLintResult: env.GOLANGCI_LINT_RESULT` to `gateProvider`.
