# Tasks

## 1. A minimal alert rule can be created, read, updated, destroyed, and imported

- [ ] 1.1 Add the shared `internal/kibana/alertingv2` helpers (entity-id validator, duration validator with optional minimum, `ALERTING_DISABLED` and "already exists, import instead" diagnostic builders) with unit tests covering valid and invalid ids, every duration unit, the 365d ceiling, and the `5s` minimum.
- [ ] 1.2 Add `internal/clients/kibanaoapi/alerting_v2_errors.go` and `alerting_v2_rule.go` (get, put, delete) that route through `SpaceAwarePathRequestEditor` and map 404, 503 `ALERTING_DISABLED`, 4xx/5xx with a JSON error body, and 201-vs-200 on `PUT` per design D8; verify with `httptest` wrapper tests asserting each mapping and the `/s/<space>` path prefix.
- [ ] 1.3 Wire `internal/kibana/alertingv2rule` on the `entitycore.NewKibanaResource` envelope with the 9.6.0 version requirement, composite id, `KibanaSpaceImporter` import (including the malformed-id diagnostic), and a schema covering identity (`id`, `rule_id`, `space_id`), `kind`, `enabled`, `time_field`, `metadata.name`, `schedule`, `query.base`, `recovery`, `no_data`, and the server-owned `version`, `created_at`, `updated_at` with the plan modifiers from design D9; `kind`, `rule_id`, and `space_id` require replacement; create generates a UUID when `rule_id` is omitted, pre-checks existence, and both writes send `enabled` in the `PUT` body; verify with unit tests for the model ↔ API mapping of these fields, `GetVersionRequirements`, and the `entitycore_contract_test.go` registration and import contract.
- [ ] 1.4 Register `elasticstack_kibana_alerting_v2_rule` in the provider's standard resource list, add `examples/resources/elasticstack_kibana_alerting_v2_rule/{resource.tf,import.sh}` and the `descriptions/` markdown (full-replace semantics, engine prerequisite), and verify `make docs-generate` produces `docs/resources/kibana_alerting_v2_rule.md` and `make build`/`make lint` pass.
- [ ] 1.5 Add the acceptance test for a minimal alert rule: create, update `metadata.name` and `schedule.every`, assert `version` increments and `created_at` is unchanged, change `kind` and assert the plan shows a replace action, import by `<space_id>/<rule_id>` with `ImportStateVerify`, destroy, and `plancheck.ExpectEmptyPlan()` after every apply step; verify it passes against a 9.6.0+ stack and is skipped below it.

## 2. Every typed rule field round-trips and clears on removal

- [ ] 2.1 Extend the rule schema and mapping with `metadata.description`, `metadata.tags`, `metadata.routing_tags`, `metadata.builder`, `query.breach`, every `recovery` and `no_data` variant, `state_transition.pending|recovering`, `grouping`, and `artifacts` (with `jsontypes.Normalized` data and the object/key-count validator), mapping null ⇄ absent for every optional object and collection and failing with a diagnostic on an unrecognised discriminator on read; verify with unit tests for each union variant in both directions, for present/absent optional objects, and for the unrecognised-discriminator case.
- [ ] 2.2 Add the fail-closed field-accounting unit test that compares the `PUT` body struct's JSON tags to the modelled fields plus a commented waiver list, and verify it fails by name when a field is removed from the modelled set.
- [ ] 2.3 Add the acceptance test for a fully configured alert rule (`time_field`, `schedule.lookback`, tags, routing tags, builder, breach, `recovery = condition`, `no_data = keep_last` with query, both state-transition phases, grouping, runbook and dashboard artifacts): create, read back every attribute, re-write artifact JSON with different whitespace and assert an empty plan, then remove `grouping`, `state_transition`, and `artifacts` and assert they read back null; verify empty plans after each step.
- [ ] 2.4 Add the acceptance tests for a `kind = "signal"` rule with only `metadata`, `schedule`, and `query`, and for `enabled = false` on create followed by toggling it on update with no `_enable`/`_disable` call; verify state and empty plans.
- [ ] 2.5 Update the example and descriptions to show the full lifecycle configuration and the `metadata.routing_tags` link to action policies; verify `make docs-generate` output includes every attribute.

## 3. Invalid rule configurations are rejected before any API call

- [ ] 3.1 Add attribute-level validators for every bound in the spec's Schema section (enums, lengths, list sizes and non-empty collections, integer ranges, id pattern, durations, artifact `id` uniqueness) and the intra-object `AlsoRequires`/`ConflictsWith` rules for `recovery` and `no_data` variant fields; verify with plan-only `ExpectError` test steps that run without `TF_ACC`.
- [ ] 3.2 Implement `ValidateConfig` with the seven cross-field rules from the spec (state_transition and routing_tags only for alert rules, alert rules need both lifecycle objects and signal rules neither, condition recovery needs a breach clause, non-ignore no-data needs breach or presence query, manual recovery forbids a recovering phase, phase thresholds and operator), each skipped on unknown inputs; verify with unit tests for known and unknown inputs and plan-only `ExpectError` steps for each spec scenario.
- [ ] 3.3 Verify a rejected `no_data.strategy = "alert"` at plan time and a tolerated `alert` on read with a unit test, and verify the server-side ES|QL, duplicate-name 409, and deployment-minimum-interval errors surface the server message in acceptance steps.

## 4. Rule failure modes and non-default spaces behave as specified

- [ ] 4.1 Add acceptance tests for a rule in a newly created non-default space (create, import by composite id, change `space_id` and assert a replace action, destroy) and for a supplied `rule_id` versus a generated one (including a `rule_id` change planning a replace); verify the `/s/<space>` routing and stable generated id across re-plans.
- [ ] 4.2 Add acceptance coverage for an externally deleted rule (refresh drops it from state and the next plan recreates it) and for create against a pre-existing id (error naming the id and recommending import, no write); verify with the wrapper tests for the `PUT` 200 race path.
- [ ] 4.3 Add the `kibana_connection` acceptance test proving a resource-level connection block is honoured, and verify the `alerting:v2:enabled` 503 diagnostic text through the wrapper tests.
- [ ] 4.4 Run the `schema-coverage` skill against the rule resource and close any reported gaps; verify `make lint`, `make build`, `go test ./provider/...`, and the targeted acceptance suite pass.

## 5. A minimal action policy can be created, read, updated, destroyed, and imported (blocked until Kibana accepts `enabled` in `PUT /api/alerting/v2/action_policies/{id}` and the Renovate `kibana-openapi-spec` bump that carries it has merged)

- [ ] 5.1 Confirm `KibanaHTTPAPIsAlertingPutActionPolicy` carries `Enabled` in the pinned `generated/kbapi`; if it does not, stop and wait for the Renovate bump rather than regenerating manually, and record the Kibana pull request and pinned commit in design D12 once known.
- [ ] 5.2 Add `internal/clients/kibanaoapi/alerting_v2_action_policy.go` (get, put, delete) reusing `alerting_v2_errors.go`; verify with `httptest` wrapper tests for 201-vs-200, 404, 503, 400, and the space prefix.
- [ ] 5.3 Wire `internal/kibana/alertingv2actionpolicy` on the envelope with the 9.6.0 gate, composite id and import (including the malformed-id diagnostic), UUID generation and existence pre-check on create, and a schema covering identity, `name`, `description`, `enabled`, `destinations`, and the server-owned `created_at`/`updated_at` with the plan modifiers from design D9; `action_policy_id` and `space_id` require replacement; verify with mapping unit tests (including the destination union), `GetVersionRequirements`, and the registration and import contract test.
- [ ] 5.4 Register `elasticstack_kibana_alerting_v2_action_policy`, add its example (a workflow destination created in the same configuration through `elasticstack_kibana_agentbuilder_workflow`, confirming its id is what `destinations[].id` expects and updating design D13 if not) and descriptions, and verify `make docs-generate` produces `docs/resources/kibana_alerting_v2_action_policy.md`.
- [ ] 5.5 Add the acceptance test for a minimal catch-all policy: create, update `name`, toggle `enabled`, assert `created_at` is unchanged across updates, change `action_policy_id` and assert a replace action, import with `ImportStateVerify`, destroy, with empty plans after each step; verify against a 9.6.0+ stack.

## 6. Every typed action policy field round-trips and clears on removal

- [ ] 6.1 Extend the schema and mapping with `matcher` (tags, expression), `grouping` (mode, fields), and `throttle` (strategy, interval), mapping null ⇄ absent and every union variant and failing with a diagnostic on an unrecognised discriminator on read; add the fail-closed field-accounting test; verify with mapping unit tests in both directions.
- [ ] 6.2 Add the acceptance test for a fully configured policy (description, matcher with tags and expression, `per_field` grouping, `time_interval` throttle): create, read back every attribute, then remove `matcher`, `grouping`, and `throttle` and assert they read back null; verify empty plans after each step and that snoozing the policy outside Terraform does not create drift.
- [ ] 6.3 Update the example and descriptions to show matcher, grouping, and throttle; verify `make docs-generate` output includes every attribute.

## 7. Invalid action policy configurations are rejected before any API call

- [ ] 7.1 Add attribute validators for every bound in the spec's Schema section and the intra-object rules (`fields` iff `per_field`, `interval` iff interval strategies, at least one of `tags`/`expression`, unique destinations, supported `type`); verify with plan-only `ExpectError` steps for each spec scenario.
- [ ] 7.2 Implement `ValidateConfig` for throttle-strategy-by-grouping-mode, skipped when the mode is unknown; verify with unit tests for known and unknown inputs and plan-only `ExpectError` steps.

## 8. Action policy failure modes and non-default spaces behave as specified

- [ ] 8.1 Add acceptance tests for a policy in a non-default space with composite-id import and a `space_id` change planning a replace, a supplied versus generated `action_policy_id`, an externally deleted policy, and create against a pre-existing id; verify routing, stable ids, state removal, and the import-recommending error.
- [ ] 8.2 Add the `kibana_connection` acceptance test for the policy resource and verify the 503 diagnostic through wrapper tests.
- [ ] 8.3 Run the `schema-coverage` skill against the action policy resource and close any reported gaps; verify `make lint`, `make build`, `go test ./provider/...`, and the targeted acceptance suite pass.

## Workflow follow-up

- Open the rule implementation PR after groups 1-4; open the action policy PR after groups 5-8 once the upstream prerequisite has landed.
- Run `openspec-verify-change`, then sync or archive the change so both capabilities land in `openspec/specs/`.
