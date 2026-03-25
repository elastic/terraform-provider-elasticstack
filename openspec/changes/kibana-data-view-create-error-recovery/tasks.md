## 1. Spec

- [ ] 1.1 Validate the change with `./node_modules/.bin/openspec validate kibana-data-view-create-error-recovery`.
- [ ] 1.2 Sync or archive the delta into `openspec/specs/kibana-data-view/spec.md` after implementation is verified.

## 2. Create Reconciliation

- [ ] 2.1 Update `internal/kibana/dataview/create.go` so create can reconcile a managed data view when the create request includes an explicit `data_view.id` and the create call returns an error.
- [ ] 2.2 Reuse the configured `space_id` plus the explicit `data_view.id` for the follow-up read, and populate final state from that read result rather than the mutating response.
- [ ] 2.3 Preserve the current error path when the create cannot be reconciled, including the case where no explicit `data_view.id` is configured.

## 3. Test Harness And Proxy Regression

- [ ] 3.1 Add the minimal acceptance-test helper or wiring needed for one test to use an explicit Kibana proxy endpoint without depending on global `KIBANA_ENDPOINT` mutation.
- [ ] 3.2 Add a dedicated `internal/kibana/dataview/acc_test.go` regression that creates an isolated Kibana space, routes provider traffic through an `httptest` reverse proxy, and rewrites only the first matching data view create response to a synthetic error after forwarding upstream.
- [ ] 3.3 Ensure the regression asserts successful convergence: the apply succeeds, the resource is in state, and a follow-up plan or apply is clean.

## 4. Focused Tests And Verification

- [ ] 4.1 Add a narrower HTTP-level test in `internal/clients/kibanaoapi` or `internal/kibana/dataview` that covers the create-error reconciliation path without running the full acceptance suite.
- [ ] 4.2 Run targeted Go tests for the touched packages and `make build`.
- [ ] 4.3 Run the targeted acceptance test for `internal/kibana/dataview` against a live stack using the environment described in `dev-docs/high-level/testing.md`.
