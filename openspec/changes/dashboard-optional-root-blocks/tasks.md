## 1. Generated client (own commit)

- [ ] 1.1 Edit the `PUT /api/dashboards/{id}` request-body `required` list in `generated/kbapi/dashboard-paths.json` to `["title"]`.
- [ ] 1.2 Regenerate `generated/kbapi/kibana.gen.go` (see `dev-docs/high-level/generated-clients.md`) and confirm `PutDashboardsIdJSONBody` now has pointer `TimeRange`, `RefreshInterval`, `Query`, and `Options`. Commit the generated diff separately.

## 2. Provider implementation

- [ ] 2.1 In `internal/kibana/dashboard/schema.go`, make `time_range` (via `timeRangeSingleNestedAttribute(..., false)`), `refresh_interval`, and `query` optional. Keep nested attributes required. Do not bump the schema version.
- [ ] 2.2 Update `dashboardToAPIUpdateRequest` in `internal/kibana/dashboard/models.go` to set `TimeRange`, `RefreshInterval`, `Query`, and `Options` only when the model block is non-nil, adapting to the pointer fields. Fix any other `PutDashboardsIdJSONBody` call sites flagged by the compiler.
- [ ] 2.3 Confirm `dashboardToAPICreateRequest` already omits nil blocks; add no read-side null-preservation or default detection for the three blocks.

## 3. Specs

- [ ] 3.1 Sync the delta in `openspec/changes/dashboard-optional-root-blocks/specs/kibana-dashboard/spec.md` into `openspec/specs/kibana-dashboard/spec.md` (REQ-036, REQ-007, REQ-009).
- [ ] 3.2 In the canonical spec's `## Schema` HCL sketch, mark `time_range`, `refresh_interval`, and `query` as `<optional, object>`.

## 4. Tests

- [ ] 4.1 Unit tests for `dashboardToAPICreateRequest` / `dashboardToAPIUpdateRequest`: nil blocks are omitted (including `options` on update); set blocks are sent.
- [ ] 4.2 Acceptance test: title-only create, empty re-plan, and import verify.
- [ ] 4.3 Acceptance test: add then remove each of `time_range`, `refresh_interval`, and `query`; removal leaves the block null in state.
- [ ] 4.4 Acceptance test: non-default values for each block.
- [ ] 4.5 Optionally cover a `dashboard` panel with `use_time_range` on a title-only dashboard; otherwise note as unverified.

## 5. Examples and docs

- [ ] 5.1 Drop the unneeded root blocks from `examples/resources/elasticstack_kibana_dashboard/panel_no_time_range.tf`; keep other examples explicit.
- [ ] 5.2 Add a title-only example under `examples/resources/elasticstack_kibana_dashboard/`.
- [ ] 5.3 Document that removing `time_range`, `refresh_interval`, or `query` clears it (no Kibana default is applied), then regenerate `docs/resources/kibana_dashboard.md` (`make docs-generate`).

## 6. Verification

- [ ] 6.1 `make build`, `make check-lint`, and `make check-openspec` pass.
- [ ] 6.2 Run the new unit tests and targeted acceptance tests against an available stack.
