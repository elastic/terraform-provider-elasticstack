> **Status.** `elasticstack_kibana_rule` ships as an **experimental** technical-preview resource: registered via `experimentalResources()`, gated by `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL=true`, no generated registry docs. Graduation is [rna-program#678](https://github.com/elastic/rna-program/issues/678).

## 1. Upstream prerequisites and open questions

- [ ] 1.1 File the three contingent API requests from `proposal.md` against the Alerting v2 team, linking each to this change. Prioritise **3 (two fields named "version")** before graduation ([#678](https://github.com/elastic/rna-program/issues/678)) — it is the only contingent ask that would force a Terraform schema rename — and **1 (writable `enabled`)** for apply-semantics correctness (not a schema change).
- [ ] 1.2 Track [kibana#279519](https://github.com/elastic/kibana/pull/279519) to merge (Check OAS Snapshot is currently red). Implementation of section 3 onwards is blocked on it; nothing in sections 1–2 is.
- [ ] 1.3 Resolve `design.md` Open Question 1: determine the first Kibana version shipping `/api/alerting/v2/rules` as a public route, and confirm serverless behaviour. Record the value in the delta spec as a compatibility requirement.
- [ ] 1.4 Raise `design.md` Open Question 2 (server-side v1 → v2 migration vs. the classic Terraform resource) with the Alerting v2 team and open a follow-up issue for whatever the classic resource needs to do about it.

## 2. Proposal review

- [ ] 2.1 Get sign-off on the resource name `elasticstack_kibana_rule` (`proposal.md`, Decision 2), explicitly resolving the overload and product-meaning objections.
- [ ] 2.2 Get sign-off on `PUT`-only writes (`design.md` D1) and on the two-block query mapping (`design.md` D3) — these are the two decisions that are expensive to reverse after implementation starts.
- [ ] 2.3 Confirm Decision 2 naming sign-off includes the settled stance that classic stays `elasticstack_kibana_alerting_rule` (rename is not viable; do not open a follow-up rename change).
- [ ] 2.4 Validate the change: `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate add-kibana-alerting-v2-rule-resource --type change`.

## 3. Generated client

- [ ] 3.1 Bump `github_ref` in `generated/kbapi/Makefile` to a Kibana commit that includes [kibana#279519](https://github.com/elastic/kibana/pull/279519), then run `make -C generated/kbapi all`.
- [ ] 3.2 Confirm the `alerting_v2_*` components generate as named, reusable Go types, that `alerting_v2_rule_query` produces a discriminated union with typed accessors for the composed and standalone variants, and that the diff is purely additive.
- [ ] 3.3 Add a `transform_schema.go` transformer that collapses `anyOf: [T, {nullable: true}]` into `T` with `nullable: true` (`design.md` D6), registered in the `transformers` pipeline near `transformRemoveAnyOfWhenOneOfPresent`.
- [ ] 3.4 Add a unit test in `generated/kbapi/transform_schema_test.go` covering the collapse, a nested occurrence, and a no-op on an `anyOf` that is not the two-element nullable shape.
- [ ] 3.5 Regenerate and diff: confirm all 37 occurrences resolve to pointers, that no unrelated component's Go type changed, and that `oapi-config.yaml`'s `nullable-type` option is still not required.
- [ ] 3.6 If 3.3–3.5 prove too invasive, fall back to per-field unwrap helpers in the client wrapper (task 4.2) and record the reason in `design.md` D6.

## 4. Client wrapper

- [ ] 4.1 Add `internal/clients/kibanaoapi/rule_v2.go` with create-or-replace (`PUT`), get, delete, enable, and disable functions, all returning `diag.Diagnostics`, using `kibanautil.SpaceAwarePathRequestEditor(spaceID)` for space scoping (`design.md` D7).
- [ ] 4.2 Implement request and response mapping between `internal/models` and the generated types, including the query union accessors and any nullable unwrapping still needed after section 3.
- [ ] 4.3 Return `(nil, nil)` for 404 on get, and treat 404 as success on delete (spec REQ-002).
- [ ] 4.4 Detect HTTP 503 carrying error code `ALERTING_DISABLED` and return a diagnostic naming the `alerting:v2:enabled` advanced setting (spec REQ-013).
- [ ] 4.5 Implement `enabled` reconciliation after every write: compare desired against the value returned by `PUT`, call `_enable`/`_disable` only on mismatch, and on failure return a diagnostic stating the rule was written but its enabled state was not applied (spec REQ-011).
- [ ] 4.6 Unit-test the wrapper against recorded responses: 404 handling, the 503 diagnostic, both reconciliation directions, the no-op reconciliation case, and reconciliation failure.
- [ ] 4.7 Make the write path read-before-write (`design.md` D9): get the rule, seed the `PUT` body from the response, then overwrite every Terraform-modelled field unconditionally including nulls. Skip the read when the rule does not exist. `internal/fleet/agentpolicy/update.go` is the in-repo precedent.
- [ ] 4.8 Unit-test the merge in both directions: a field present in the generated struct but absent from the Terraform model survives a write untouched, and a modelled field set to null is sent as null rather than dropped.

## 5. Resource

- [ ] 5.1 Create `internal/kibana/rule/` and wire `resource.go` with `entitycore.NewKibanaResource[Model](entitycore.ComponentKibana, "rule", ...)` plus `entitycore.KibanaSpaceImporter` for `<space_id>/<rule_id>` import (spec REQ-003).
- [ ] 5.2 Write `schema.go` for the identity and top-level scalars: `id`, `rule_id`, `space_id`, `kind`, `time_field`, `enabled`, `recovery_strategy`, `no_data_strategy`. Apply `RequiresReplace` to `rule_id`, `space_id`, and `kind`; restrict `no_data_strategy` to the three writable values (spec REQ-014).
- [ ] 5.3 Add the `metadata`, `schedule`, `grouping`, and `artifacts` blocks with their length, range, and collection-size validators. For `artifacts`, model `data` as a JSON object string (not the legacy `value` string) and enforce the known-type rules in REQ-012 (`runbook` / `dashboard`) at plan time (spec REQ-005, REQ-008, REQ-012).
- [ ] 5.4 Add the `composed_query` and `standalone_query` blocks with nested `breach` / `recovery` / `no_data` blocks, and the exactly-one-of validator between them (spec REQ-006).
- [ ] 5.5 Add the `state_transition` block with the `AND`/`OR` enums and 0–1000 count bounds.
- [ ] 5.6 Add the duration validator used by `schedule.every` (min `5s`, max `365d`), `schedule.lookback`, and the two `state_transition` timeframes (spec REQ-009). Reuse or extend the existing alerting duration type if its accepted forms match; otherwise add a v2-specific validator and note the divergence.
- [ ] 5.7 Add the computed attributes `version`, `created_at`, `created_by`, `updated_at`, `updated_by`, and `metadata.version`, with `UseStateForUnknown` on `created_at` and `created_by` only (`design.md` D4, spec REQ-010).
- [ ] 5.8 Write `models.go`: the Terraform model satisfying `entitycore.KibanaResourceModel`, mapping to and from the API model, deriving `query.format` from the configured block, omitting empty `metadata.tags` from the request, and mapping absent tags back to a null set (spec REQ-008).
- [ ] 5.9 Add `GetVersionRequirements` with the minimum version from task 1.3 (`design.md` D8).
- [ ] 5.10 Write `validate.go` implementing the seven cross-field checks in spec REQ-007, each skipped when a participating value is unknown.
- [ ] 5.11 Write `create.go`, `read.go`, `update.go`, `delete.go`. Create and update both call the `PUT` wrapper; create generates a `rule_id` when one is not configured (`design.md` D2); both re-read authoritatively afterwards.
- [ ] 5.12 Add `descriptions/` markdown and schema `MarkdownDescription` values. Cover: the technical-preview status, the last-write-wins consequence of full-replace `PUT`, and the distinction between `metadata.version` and `version`.
- [ ] 5.13 Register `rule.NewResource` in `Provider.experimentalResources(...)` in `provider/plugin_framework.go` (spec REQ-001).
- [ ] 5.14 Add `entitycore_contract_test.go` following the pattern in neighbouring packages.
- [ ] 5.15 Add a fail-closed field-accounting test (`design.md` D9): reflect over the `PUT` body struct's JSON tags and assert the set equals the fields the schema maps plus an explicit waived list, each waiver carrying a comment saying why omitting it is safe. The failure message must name the unaccounted field, since the reader will be looking at a Renovate regeneration diff.

## 6. Tests

- [ ] 6.1 Plan-only `ExpectError` tests for every cross-field validation in spec REQ-007 and for the exactly-one-of query constraint in REQ-006. These must not require a live stack (spec REQ-016).
- [ ] 6.2 Plan-only tests for the schema-level bounds in spec REQ-005 (over-long `metadata.name`, >16 `grouping.fields`, >100 `artifacts`, `state_transition.pending_count` above 1000), for `no_data_strategy = "emit"` being rejected (REQ-014), and for artifact `data` validation (invalid JSON, >32 keys, `runbook` missing `content`) per REQ-012.
- [ ] 6.3 Unit test that a `no_data_strategy` of `"emit"` arriving from the API maps into state without error (spec REQ-014).
- [ ] 6.4 Acceptance test: `kind = "alert"` with `composed_query` — create, update, and assert the state round-trip, followed by a step asserting an empty plan (spec REQ-016 items 1 and 7).
- [ ] 6.5 Acceptance test: `kind = "signal"` with `standalone_query` (item 2).
- [ ] 6.6 Acceptance test: `recovery_strategy = "query"` with a `recovery` block, and a standalone rule with `no_data_strategy` plus a `no_data` block (item 3).
- [ ] 6.7 Acceptance test: `state_transition`, `grouping`, and `artifacts`, including a step that removes each block and asserts the value is cleared server side under full-replace `PUT` (item 4, spec REQ-012). This is also the regression test for the read-before-write merge in task 4.7 — a merge that skips nulls would leave these blocks in place and fail here.
- [ ] 6.8 Acceptance test: `enabled = false` on create, and toggling `enabled` on update, asserting the reconciliation in spec REQ-011 (item 5).
- [ ] 6.9 Acceptance test: import via `<space_id>/<rule_id>`, and a rule in a non-default space (item 6).
- [ ] 6.10 Gate all acceptance tests on the minimum version from task 1.3. Follow the repository's acceptance-test isolation conventions so parallel runs do not collide on rule ids or tags.
- [ ] 6.11 Run the `schema-coverage` skill against the resource and close any gaps it reports.

## 7. Verification and close-out

- [ ] 7.1 `make build` and `make lint` clean.
- [ ] 7.2 `go test ./provider/...` — confirms the resource satisfies the `*entitycore.ResourceBase` registration contract under `AccTestVersion`.
- [ ] 7.3 `make docs-generate` produces no `docs/resources/kibana_rule.md` and `make check-docs` stays green (spec REQ-001).
- [ ] 7.4 Targeted acceptance run against a live stack with `TF_ACC=1` and `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL=true`, per `dev-docs/high-level/testing.md`.
- [ ] 7.5 CHANGELOG entry noting the new experimental resource and the `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL` prerequisite.
- [ ] 7.6 Run the `openspec-verify-change` skill, then sync or archive the change so `kibana-rule` lands in `openspec/specs/`.
- [ ] 7.7 Re-check the outcome of every contingent API request from task 1.1 and fold any that landed into the spec before graduation.
