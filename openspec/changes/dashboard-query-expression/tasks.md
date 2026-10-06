## 1. Schema and models

- [ ] 1.1 In `internal/kibana/dashboard/schema.go`, replace the `query.text` / `query.json` attributes with a single `Required: true` `expression` string attribute. Remove the `stringvalidator.ExactlyOneOf` validators on both and update the `query` block's `MarkdownDescription`.
- [ ] 1.2 Bump the resource schema `Version: 1` → `Version: 2` (`internal/kibana/dashboard/resource.go`).
- [ ] 1.3 In `internal/kibana/dashboard/models.go`, change `models.DashboardQueryModel` to `{ Language types.String; Expression types.String }`; drop the `jsontypes` field and, if unused elsewhere in the file, the import.
- [ ] 1.4 Update the write-side mapping (`dashboardQueryToAPI`) to copy `Language` and `Expression` directly; delete the `textKnown`/`jsonKnown` branch and the "exactly one of `query.text` or `query.json`" diagnostic.
- [ ] 1.5 Update the read-side mapping in `models.go` to set `Expression` directly from `data.Data.Query.Expression`; delete the leading-`{` / `json.Unmarshal` sniffing block.
- [ ] 1.6 Remove any now-unreferenced "exactly one of text/json" helper in `config_validators.go`.

## 2. State upgrade

- [ ] 2.1 In `internal/kibana/dashboard/state_upgrade.go`, add `migrateV1ToV2`: for a non-null `query` object in the raw state map, set `expression` from `text` when `text` is a non-null string, otherwise from `json`; delete `text` and `json`. Leave a null/absent `query` untouched.
- [ ] 2.2 Add `migrateV0ToV2`, composing `migrateV0ToV1`'s transform followed by `migrateV1ToV2`'s transform against the same upgrade response.
- [ ] 2.3 Update `UpgradeState` to return `{0: migrateV0ToV2, 1: migrateV1ToV2}`.

## 3. Specs

- [ ] 3.1 Sync the delta in `openspec/changes/dashboard-query-expression/specs/kibana-dashboard/spec.md` into `openspec/specs/kibana-dashboard/spec.md` (REQ-036, REQ-007, new REQ-055).
- [ ] 3.2 In the canonical spec's `## Schema` HCL sketch, change `query` to `{ language, expression }` and drop the `text` / `json` comment.

## 4. Tests

- [ ] 4.1 Unit tests for `migrateV1ToV2`: text-only, json-only, empty-string text, null query, absent query, already-v2-shaped (idempotent), non-object `query`.
- [ ] 4.2 Unit test for `migrateV0ToV2` composition: a v0 state with both a flat `options_list_control` panel and a root `query.text` ends up with `by_field` and `expression` populated in one upgrade call.
- [ ] 4.3 Update `models_dashboard_root_test.go` (remove the "neither/both text and json" cases), `models_optional_root_blocks_test.go`, `create_test.go`, and `pinned_panels_mapping_test.go` for the `expression` field.
- [ ] 4.4 Update schema validation tests: rename `text` cases to `expression`; add a "missing expression" case (now `Required`).
- [ ] 4.5 Rename `text = "..."` to `expression = "..."` in the dashboard `testdata/**/main.tf` fixtures that set a root `query` (mechanical).
- [ ] 4.6 Delete `TestAccResourceDashboardRootQueryJSON` and its testdata.
- [ ] 4.7 Rewrite `TestAccResourceDashboardQueryTransition` as an in-place `expression` update.
- [ ] 4.8 Add a new acceptance upgrade test: step 1 with `ExternalProviders` pinned to the last released version using `query = { language = "kql", text = "..." }`; step 2 with `ProtoV6ProviderFactories: acctest.Providers` using `query = { language = "kql", expression = "..." }`, asserting `query.expression` is populated from the prior `text` value, `query.text`/`query.json` are gone, and the plan is empty. Follow the pattern in `internal/kibana/alertingrule/acc_test.go`.

## 5. Docs and examples

- [ ] 5.1 Update any `examples/resources/elasticstack_kibana_dashboard/` files using `query.text` or the JSON-query example to use `query.expression`; remove the JSON-query example if it exists.
- [ ] 5.2 Regenerate `docs/resources/kibana_dashboard.md` (`make docs-generate`).

## 6. CHANGELOG

- [ ] 6.1 Add a "Breaking changes" entry (following the existing convention, e.g. 0.16.3) describing the `text`/`json` → `expression` rename, the automatic v1 → v2 state upgrade, and that `.tf` source must be updated manually. Reference `#5081`.

## 7. Verification

- [ ] 7.1 `make build`, `make check-lint`, and `make check-openspec` pass.
- [ ] 7.2 Run the new/updated unit tests and targeted acceptance tests (including the upgrade test) against an available stack.
