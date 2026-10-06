## 1. Root `query` uses `expression` end to end

- [ ] 1.1 In `internal/kibana/dashboard/schema.go`, replace the `query.text` / `query.json` attributes with a single `Required: true` `expression` string attribute, remove both `stringvalidator.ExactlyOneOf` validators, and update the `query` block's `MarkdownDescription`.
- [ ] 1.2 In `internal/kibana/dashboard/models/dashboard.go`, change `DashboardQueryModel` to `{ Language types.String; Expression types.String }` and remove the now-unused `jsontypes` import there. Keep the `jsontypes` import in `internal/kibana/dashboard/models.go` (still used for `filter_json`).
- [ ] 1.3 In `models.go`, make `dashboardQueryToAPI` a direct copy of `Language` / `Expression` (delete the `textKnown`/`jsonKnown` branch and the "exactly one of" diagnostic) and set `Expression` directly from the API on read (delete the `{`-sniffing block).
- [ ] 1.4 Update unit tests: `models_dashboard_root_test.go` (remove the "neither/both" cases), `models_optional_root_blocks_test.go`, `create_test.go`, `pinned_panels_mapping_test.go`, and schema validation cases (`text` → `expression`, add "missing expression"). Include `expression = ""` round-trip.
- [ ] 1.5 Acceptance: rename `text =` to `expression =` in the dashboard `testdata/**/main.tf` fixtures that set a root `query`; delete `TestAccResourceDashboardRootQueryJSON` and its testdata; rewrite `TestAccResourceDashboardQueryTransition` as an in-place `expression` update.
- [ ] 1.6 Update `examples/resources/elasticstack_kibana_dashboard/*.tf` and `templates/guides/*.md.tmpl` that use root `query.text`; remove the JSON-query example in `resource.tf`.
- [ ] 1.7 Specs: sync the delta for REQ-036, REQ-007 and REQ-008 into `openspec/specs/kibana-dashboard/spec.md`, and change the `query` entry in the `## Schema` sketch to `{ language, expression }` (drop the `text`/`json` lines).

## 2. Legacy query state upgrades to v2

- [ ] 2.1 Bump the schema `Version: 1` → `Version: 2` in `internal/kibana/dashboard/schema.go`.
- [ ] 2.2 In `state_upgrade.go`, extract the v0 → v1 logic as a pure function over the raw state map and add the v1 → v2 transform: `text`, else `json`, else explicit null `expression`; always delete `text`/`json`; null/absent `query` untouched.
- [ ] 2.3 Add `migrateV0ToV2` (one unmarshal, v0 → v1 then v1 → v2 on the same map, one marshal) and `migrateV1ToV2`; set `UpgradeState` to `{0: migrateV0ToV2, 1: migrateV1ToV2}`.
- [ ] 2.4 Unit tests for `migrateV1ToV2`: text-only, json-only (normalized stored value), empty-string text, `text` beginning with `{`, `query` with neither field (explicit null, not `""`), null query, absent query, already-v2-shaped (idempotent), non-object `query`.
- [ ] 2.5 Unit test for `migrateV0ToV2`: a v0 state with a flat `options_list_control` panel and a root `query.text` ends up with `by_field` *and* `expression` in one call.
- [ ] 2.6 Acceptance upgrade test (required): step 1 `ExternalProviders` pinned to `0.16.5` with `query = { language = "kql", text = "..." }`; step 2 `ProtoV6ProviderFactories: acctest.Providers` with `expression`, asserting `query.expression` carries the prior value, `query.text`/`query.json` are gone, and the plan is empty. Follow the pattern in `internal/kibana/alertingrule/acc_test.go`.
- [ ] 2.7 Specs: sync the delta for REQ-040 and REQ-055 into the canonical spec and update the "does not declare a schema version, custom state upgrader" note near the top of `openspec/specs/kibana-dashboard/spec.md`.

## 3. Breaking change communication

- [ ] 3.1 Add the "Breaking changes" CHANGELOG entry using the draft in `proposal.md` (before/after config example, rationale, state-upgrade note, reference to `#5081`).
- [ ] 3.2 Regenerate `docs/resources/kibana_dashboard.md` (`make docs-generate`) and confirm the guides render the `expression` examples.

## 4. Verification

- [ ] 4.1 `make build`, `make check-lint` and `make check-openspec` pass.
- [ ] 4.2 Run the new/updated unit tests and targeted acceptance tests (including the upgrade test) against an available stack.
