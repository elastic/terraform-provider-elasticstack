# Design

## Context

See `proposal.md` for motivation and the two delta specs under `specs/` for requirements. This document covers how the two resources map onto the Alerting v2 API and the provider's existing machinery, and why each mapping choice was made.

**Generated client is already complete.** `generated/kbapi/kibana.gen.go` (pinned Kibana commit `23436cd`) exposes every Alerting v2 operation and named component. No change to `generated/kbapi/transform_schema.go` or regeneration is needed. The relevant surface:

| Concern | Generated symbol |
|---|---|
| Rule write body | `KibanaHTTPAPIsAlertingPutRule` (`PutAlertingV2RulesIdJSONRequestBody`), includes `Enabled *bool` |
| Rule response | `KibanaHTTPAPIsAlertingRuleResponse` (`Id`, `Version int`, `Enabled`, `CreatedAt`/`UpdatedAt time.Time`, `CreatedBy`/`UpdatedBy *KibanaHTTPAPIsAlertingActor`, plus every write field) |
| Rule nested objects | `...RuleMetadata` (with inline `Builder *struct{Type string}`), `...RuleSchedule`, `...RuleQuery` + `...RuleBreach`, `...RuleStateTransition` + `...Pending`/`...Recovering`, `...RuleGrouping`, `...RuleArtifact` (`Data map[string]interface{}`) |
| Rule unions | `KibanaHTTPAPIsAlertingRuleRecovery` and `KibanaHTTPAPIsAlertingRuleNoData` are `union json.RawMessage` wrappers with `As<Variant>()`, `From<Variant>()`, `Discriminator()`, `ValueByDiscriminator()`. Variants are `RecoveryNoBreach`/`RecoveryCondition{Segment}`/`RecoveryQuery{Query}`/`RecoveryManual`, `NoDataIgnore`/`NoDataKeepLast{Query *string}`/`NoDataResolve{Query *string}`/`NoDataAlert{Query *string}` |
| Action policy write body | `KibanaHTTPAPIsAlertingPutActionPolicy` (`PutAlertingV2ActionPoliciesIdJSONRequestBody`): `Name`, `Description *string`, `Destinations []...ActionPolicyDestination`, `Matcher *struct{Tags *[]string; Expression *string}`, `Grouping *...ActionPolicyGrouping`, `Throttle *...ActionPolicyThrottle`. **No `Enabled` field yet** (see Decision 12) |
| Action policy response | `KibanaHTTPAPIsAlertingActionPolicyResponse` adds `Id`, `Enabled bool`, `SnoozedUntil *string`, `CreatedAt`/`UpdatedAt`, actors |
| Action policy unions | `...ActionPolicyDestination` (variant `WorkflowActionPolicyDestination{Id, Type}`), `...ActionPolicyGrouping` (`GroupingPerAlert`/`GroupingAll`/`GroupingPerField{Fields}`), `...ActionPolicyThrottle` (`ThrottleOnStatusChange`/`ThrottlePerStatusInterval{Interval}`/`ThrottleTimeInterval{Interval}`/`ThrottleEveryTime`), all with the same accessor pattern |
| Operations | `ClientWithResponses.{Put,Get,Delete}AlertingV2RulesIdWithResponse(ctx, id, body?, reqEditors...)`, same for `AlertingV2ActionPoliciesId`. Paths are `/api/alerting/v2/...` with no space parameter |
| Responses | `PutAlertingV2RulesIdResponse{JSON200, JSON201, JSON400, JSON401, JSON403, JSON409, JSON500, JSON503}`, `Get...{JSON200, JSON404, ...}`, and `Delete...{JSON404, ...}` (204 has no body). Every error body is `KibanaHTTPAPIsAlertingErrorResponse{Code, Error, Message, Details}` with stable codes such as `RULE_NOT_FOUND`, `ACTION_POLICY_NOT_FOUND`, `ALERTING_DISABLED` |

**API semantics that shape the design** (from `@kbn/alerting-v2-schemas` in Kibana `main`, matching the pinned OAS):

- `PUT /{id}` is create-or-replace: 201 when created, 200 when replaced. Ids are `^[a-zA-Z0-9_-]{1,150}$`, chosen by the caller, permanent.
- Request schemas are `strict`: unknown keys and empty arrays are rejected with 400. Omitted optional keys mean "unset".
- `kind` is the only immutable rule field. A replace that changes it answers 409.
- Rule `PUT` accepts `enabled` (create default `true`, and omitting it on replace preserves the stored value). Action policy `PUT` does **not** accept `enabled` yet. The stored value is preserved on replace and `_enable`/`_disable` own the transition.
- Durations are `^\d+(ms|s|m|h|d|w)$`, at most `365d`. `schedule.every` is at least `5s` (deployment-level minimum, default `1m`, is enforced server-side only).
- Every v2 route answers 503 `ALERTING_DISABLED` while the `alerting:v2:enabled` advanced setting is off. Routes are `access: public`, `availability: stable since 9.6.0`. The plugin-level `xpack.alerting_v2.enabled` setting defaults to `true`.
- Responses echo values as sent (snake_case, no normalization), so stored durations and ES|QL strings compare byte-for-byte.

**Provider machinery assumed** (all existing):

- `entitycore.NewKibanaResource[Model]` envelope with `KibanaResourceOptions{Schema, Read, Create, Update, Delete}`, `entitycore.SimpleKibanaDelete`, `entitycore.KibanaSpaceImporter` for `<space_id>/<id>` import, `GetVersionRequirements` on the model for the version gate (serverless bypass is envelope behaviour), and `kbschema.ResourceSpaceIDAttributeRequiresReplaceOnly()`. Closest precedents: `internal/kibana/alertingrule` (composite id, `rule_id` optional-computed, `ValidateConfig`) and `internal/kibana/maintenance_window` (version gate, nested attributes).
- `internal/clients/kibanaoapi` wrappers return `diag.Diagnostics`, call `client.API.<Op>WithResponse(ctx, id, ..., kibanautil.SpaceAwarePathRequestEditor(spaceID))`, and use `HandleGetTypedResponse` / `HandleMutateTypedResponse` / `diagutil.HandleStatusResponse` / `diagutil.ReportUnknownHTTPError`.
- `jsontypes.Normalized` for JSON-string attributes. `kibanacustomtypes.AlertingDuration` exists but accepts units `smhdwMy` and not `ms`, so it does not match the v2 grammar.
- Docs come from `examples/resources/<type>/{resource.tf,import.sh}` through `make docs-generate`. Acceptance tests follow `acctest.NamedTestCaseDirectory` + `testdata/<Test>/<step>/` and `versionutils.SkipIfUnsupported`.

## Goals / Non-Goals

**Goals:**

- One write path per resource (`PUT`) whose body is built entirely from the plan, so a re-apply after any partial failure converges and removing configuration removes it server-side.
- A Terraform schema that mirrors the API body one-to-one in names and nesting, so the Kibana API docs double as resource docs and every API constraint has a plan-time validator.
- Shared helpers between the two resources (id validator, duration validator, error mapping) so the action policy implementation is a thin second instance of the same shape.
- Zero changes to the generated client, `entitycore`, or existing resources.

**Non-Goals:**

- Read-merge-write, optimistic concurrency (the API's `PUT` takes no token), or conflict detection beyond what `RequiresReplace` on `kind` already prevents.
- Reconciling `enabled` through `_enable`/`_disable`. The action policy resource waits for the upstream `PUT` change instead.
- Semantic equality for durations or ES|QL. Values are compared as written because the API echoes them as written.
- Exposing runtime state (`snoozed_until`, actors) or anything outside basic CRUD.

## Decisions

### D1. One `alerting_v2` directory holding two resource packages plus one shared helper package

All Alerting v2 code lives under `internal/kibana/alerting_v2/`, one sub-package per resource plus a shared one. Following the repo convention for underscore directories (for example `internal/kibana/maintenance_window` is package `maintenancewindow`), package names drop the underscore.

- `internal/kibana/alerting_v2/rule` (package `rule`) — `elasticstack_kibana_alerting_v2_rule`
- `internal/kibana/alerting_v2/action_policy` (package `actionpolicy`) — `elasticstack_kibana_alerting_v2_action_policy`
- `internal/kibana/alerting_v2/utils` (package `utils`) — shared, dependency-free schema helpers: entity-id validator, duration validator (with optional minimum), and the "already exists, import instead" diagnostic builder.
- `internal/clients/kibanaoapi/alerting_v2_rule.go`, `alerting_v2_action_policy.go`, and `alerting_v2_errors.go` (shared decoding of `KibanaHTTPAPIsAlertingErrorResponse` into diagnostics, including the `ALERTING_DISABLED` diagnostic that names `alerting:v2:enabled`. It is the single owner of that text).

Each resource package follows the repo layout: `resource.go`, `schema.go`, `models.go` (plus `models_*.go` where the union mapping gets long), `validate.go`, `create.go`, `read.go`, `update.go`, `descriptions.go` + `descriptions/` markdown, `acc_test.go`, `entitycore_contract_test.go`, `testdata/`.

Alternative: one package with both resources. Rejected: the repo convention is one package per resource, and the two have no shared model code, only shared validators.

### D2. `PUT` for create and update, client-side id generation, existence pre-check on create

Both lifecycle writes call `PutAlertingV2{Rules,ActionPolicies}IdWithResponse`. When `rule_id` / `action_policy_id` is not configured, the resource generates a UUID (`github.com/google/uuid`, already a dependency) in Create before the write and stores it, so every write is addressed and idempotent.

Create first issues a `GET`: 200 means the object exists and Create fails with an "already exists, import it" diagnostic without writing. A 404 proceeds to `PUT`. If `PUT` still answers 200 (a race), Create fails with the same diagnostic plus "the object has been replaced". Update expects 200. A 201 means the object was deleted between plan and apply and the `PUT` recreated it. That is an error, not a silent recreation: the plan carries `created_at` as a known prior value (D9), so the read-after-write would fail the apply with an inconsistent-result error anyway. The diagnostic states that the object was deleted outside Terraform and recreated, and that a new plan reconciles state.

The `PUT` response body is never used as state. The envelope performs the read-after-write with the resource's `Read` callback, so final state after create and update always comes from `GET`, as the specs require.

Known limitation, shared with every `entitycore` resource: if the read-after-write fails after a successful `PUT`, the envelope returns the error without saving state, so the object exists on the server but not in state. `Read` includes the composite id in its error diagnostics so the written id is visible. Recovery is `terraform import` with that id. A retried create with a supplied id fails the existence pre-check with the import diagnostic. A retried create with a generated id would create a second object, which is why the id is in the diagnostic. No retry logic is designed in.

Alternatives considered:

- **`POST` for create with a client-supplied id, `PUT` for update.** From Kibana 9.7.0 the single create routes accept an optional `id` and answer 409 on collision ([#1026](https://github.com/elastic/rna-program/issues/1026) for rules, [#1347](https://github.com/elastic/rna-program/issues/1347) for action policies) and accept `enabled` ([elastic/rna-program#1266](https://github.com/elastic/rna-program/issues/1266)), so this shape becomes available. It is rejected because it buys only a server-side 409 in place of the pre-check, whose race window is a single request wide, and it costs a second request struct and a second model-to-API mapping per resource, a second field-accounting test, a second error table, and a 9.7.0 gate on the rule resource that `PUT` does not need. A version branch that picks `POST` on 9.7.0+ and `PUT` below would keep both paths alive until 9.6 leaves support for the same non-gain. The `PUT` design needs nothing from these two issues and is not revisited when they land.
- **`POST` when the id is omitted.** Two write paths and no safe retry after a partial create.
- **Adopting silently on `PUT` 200.** Terraform never takes over unmanaged objects, and it would clobber a UI-authored object with no plan diff.
- **Skipping the pre-check.** Clobbers before failing.

### D3. Schema mirrors the API body one-to-one and unions flatten to `{discriminator + variant fields}`

All nested objects are `SingleNestedAttribute`s (not blocks) with the API's names: `metadata`, `schedule`, `query`, `query.breach`, `recovery`, `no_data`, `state_transition.pending|recovering`, `grouping`. `artifacts` and `destinations` are `ListNestedAttribute`s.

Each discriminated union becomes one object whose discriminator is required and whose variant-specific fields are optional, with plan-time validation in `ValidateConfig` (D11) enforcing the variant contract:

| API union | Terraform object | Variant rule |
|---|---|---|
| rule `recovery` (`strategy`) | `recovery = { strategy, segment?, query? }` | `segment` iff `condition`. `query` iff `query` |
| rule `no_data` (`strategy`) | `no_data = { strategy, query? }` | `query` only when strategy ≠ `ignore`. `alert` rejected on write, tolerated on read |
| policy `destinations[]` (`type`) | `{ type, id }` | `type` enum (`workflow`) |
| policy `grouping` (`mode`) | `grouping = { mode, fields? }` | `fields` iff `per_field` |
| policy `throttle` (`strategy`) | `throttle = { strategy, interval? }` | `interval` iff `per_status_interval`/`time_interval`. Strategy set depends on `grouping.mode` |

Mapping to the API uses the generated `From<Variant>()` setters chosen by the discriminator. Mapping from the API uses `Discriminator()` then `As<Variant>()`. An unrecognised discriminator on read is an error diagnostic (never silently dropped).

Alternatives: one nested attribute per variant (`recovery_condition = {...}` etc.) is verbose and moves the exactly-one-of rule into validation anyway. A JSON blob for the lifecycle objects is the classic resource's `params` problem the v2 API was designed to avoid.

`metadata.builder` is **not exposed**. It marks which Kibana UI form authored a rule (for example `threshold`), is absent for rules authored directly in ES|QL, has no effect on evaluation, and is an unvalidated free string the provider could give no contract for. A Terraform-managed rule is an ES|QL rule by definition, and setting a builder type from HCL would make the UI try to parse the query back into a form it never came from. Leaving it out is reversible (adding an optional attribute later is non-breaking) whereas exposing it and removing it later is not. The cost is that importing a UI-authored rule clears the builder on the first full-replace `PUT`. That is the first apply that changes an attribute, because a matching configuration plans nothing and writes nothing. So `Read` emits a warning diagnostic whenever the response carries `metadata.builder`: it fires on import and on every refresh until a write clears it, and the resource description states that Terraform-managed rules are ES|QL rules. The field is the first entry in the D4 waiver list.

### D4. Write body built entirely from the plan, guarded by a fail-closed field-accounting test

No read-merge-write. The schema models every field of `KibanaHTTPAPIsAlertingPutRule` and `KibanaHTTPAPIsAlertingPutActionPolicy`, so a `PUT` body built from the plan is the whole object. To keep that true across Renovate bumps of the generated client, each package carries a unit test that walks the `PUT` body struct recursively: named nested structs, anonymous inline structs (`matcher`, `metadata.builder`), pointers, and slices. The generated union types (`recovery`, `no_data`, `grouping`, `throttle`, `destinations[]`) wrap an unexported `json.RawMessage`, so a flat scan of JSON tags would stop at them. For each union the test enumerates the generated `As<Variant>()` methods through the type's method set and walks every variant struct. It asserts that the resulting set of JSON paths and union variants equals the modelled set plus an explicit, commented waiver list of paths (`metadata.builder` for rules, per D3). A new upstream field at any depth, or a new strategy, mode, or destination type, turns the regeneration PR red with the path or variant name.

Alternative: `GET`, seed the body from the response, overwrite modelled fields, `PUT`. Rejected: preserves unmodelled fields forever without anyone modelling them, and the fail-closed test gives the same protection more cheaply.

### D5. Durations: plain strings with a v2-specific validator, no semantic equality

A `validator.String` in `internal/kibana/alerting_v2/utils` accepts `^\d+(ms|s|m|h|d|w)$`, enforces at most 365 days, and takes an optional minimum (used as `5s` for `schedule.every`). Attributes stay `types.String`. The API echoes the value as written, so no normalization or semantic-equality type is needed.

Alternative: reuse `kibanacustomtypes.AlertingDuration`. Rejected: it lacks `ms`, accepts `M` and `y` that v2 rejects, and its semantic equality would mask a genuine server change.

### D6. Space handling, identity, import

Requests use `kibanautil.SpaceAwarePathRequestEditor(spaceID)`. `space_id` is `kbschema.ResourceSpaceIDAttributeRequiresReplaceOnly()`. Models satisfy `entitycore.KibanaResourceModel` with `GetResourceID()` returning `rule_id` / `action_policy_id` and `GetID()` the composite. The resource embeds `entitycore.KibanaSpaceImporter` built with `NewKibanaSpaceImporter(path.Root("id"), path.Root("space_id"), path.Root("rule_id"))` (respectively `action_policy_id`), as `internal/kibana/alertingrule` does, but wraps `ImportState`: it validates the import id first (exactly one separator, both segments non-empty) and reports the `<space_id>/<rule_id>` format in its own diagnostic, then calls `KibanaSpaceImporter.SeedState`. The bare importer sets an empty `space_id` for `/my-id` and reports a cluster-oriented format for an id without a separator, neither of which meets the import contract. Read sets `id = <space_id>/<resource_id>`.

### D7. Version gate at 9.6.0 for rules and 9.7.0 for action policies

`GetVersionRequirements` returns `entitycore.SingleVersionRequirement(9.6.0, "Alerting v2 rules require Elastic Stack v9.6.0 or later.")` for rules and the same shape with 9.7.0 for action policies, the first release whose action policy `PUT` accepts `enabled` (D12). The envelope evaluates it before every callback and bypasses it on serverless. Acceptance tests gate with `versionutils.SkipIfUnsupported(t, 9.6.0, versionutils.FlavorAny)`.

### D8. Error handling

`alerting_v2_errors.go` decodes non-2xx bodies as `KibanaHTTPAPIsAlertingErrorResponse` and maps:

| Status / code | Read | Delete | Create / Update |
|---|---|---|---|
| 404 | `(nil, nil)` → removed from state | success | error on the read-after-write (the object vanished mid-apply and the message is surfaced). Create's existence pre-check treats 404 as the normal "free id" path (D2) |
| 503 `ALERTING_DISABLED` | error naming `alerting:v2:enabled` | same | same |
| 400 / 409 / 401 / 403 / 500 | error with `code` and `message` | same | same |
| 200 on create's `PUT` | n/a | n/a | "already exists, import instead" + "replaced" (D2) |
| 201 on update's `PUT` | n/a | n/a | error: deleted outside Terraform between plan and apply and recreated, re-plan to reconcile (D2) |
| empty 2xx body on `GET` or `PUT` | error | n/a (`DELETE` answers 204 with no body) | error |

Diagnostics always include the server `message` and `code`. `details` is not parsed (the API documents it as unstable).

### D9. Server-owned fields, defaults, and plan modifiers

| Attribute | Source | Plan modifier / default |
|---|---|---|
| `version` (rule only) | `Version int` | computed, no `UseStateForUnknown` (changes on every write) |
| `created_at` | `CreatedAt` formatted RFC 3339 | computed, `UseStateForUnknown` |
| `updated_at` | `UpdatedAt` | computed, no `UseStateForUnknown` |
| `id` | composite | computed, `UseStateForUnknown` |
| `rule_id` / `action_policy_id` | path id | optional, computed, `RequiresReplace`, `UseStateForUnknown` |
| `space_id` | `kbschema.ResourceSpaceIDAttributeRequiresReplaceOnly()` | optional, computed, default `default`, `RequiresReplace` |
| `kind` (rule only) | `Kind` | required, `RequiresReplace` (API-immutable, never sent as a change) |
| `enabled` | `Enabled` | optional, computed, schema `Default` `true` (`booldefault.StaticBool`), so the plan always holds a known value and the `PUT` body always carries it |
| `time_field` (rule only) | `TimeField` | optional, computed, schema `Default` `"@timestamp"` (`stringdefault.StaticString`), always sent. On read, an absent `time_field` maps to `"@timestamp"` |

Schema-level defaults, not server defaults read back with `UseStateForUnknown`, are used for `enabled` and `time_field` so that the write body never omits them and a plan shows the effective value. Both are always populated by `Read`, including on import: the response carries `enabled` as a non-optional boolean, and `time_field` is an optional pointer in the response type, so `Read` maps nil to `"@timestamp"`, the server's documented default (`DEFAULT_TIME_FIELD`, which the Kibana create-or-replace schema applies when the field is omitted). A rule created elsewhere without `time_field` therefore imports with no diff. Actors (`created_by`, `updated_by`) and `snoozed_until` are not modelled. The framework ignores response fields it does not know.

### D10. Artifacts `data`

`artifacts[].data` is `jsontypes.Normalized` (so key order and whitespace never diff) with a resource-level validator asserting a JSON object with at most 32 top-level keys of 1-256 characters. Mapping unmarshals into `map[string]interface{}` for the request and marshals the response map back. A `listvalidator`-style uniqueness check on `id` runs at plan time. Type-specific rules (`runbook.content`, `dashboard.dashboard_id`) stay server-side so new artifact types need no provider change.

### D11. Validation placement

- Attribute-level validators (`stringvalidator.OneOf`, `LengthBetween`, `RegexMatches`, `listvalidator.SizeBetween`, `int64validator.Between`, the duration validator) carry every bound in the specs' Schema sections so they appear in generated docs and `terraform validate`.
- `AlsoRequires` and `ConflictsWith` test attribute presence only, so they cannot express the variant contracts, where a field is required or forbidden depending on the discriminator's value (`recovery.segment` iff `strategy = "condition"`, `recovery.query` iff `strategy = "query"`, `no_data.query` forbidden with `ignore`, `grouping.fields` iff `mode = "per_field"`, `throttle.interval` iff an interval strategy). Those are not declared on the attribute. List-level uniqueness (`artifacts[].id`, `destinations[]` `(type, id)` pairs) is a list validator on the collection attribute.
- The variant contracts above and the cross-object rules (the rule spec's seven cross-field rules and the policy's throttle-strategy-by-grouping-mode rule) live in `ValidateConfig` (`resource.ResourceWithValidateConfig`), each check returning early when any participating value is unknown, following `internal/kibana/alertingrule/validate.go`.

### D12. Action policy `enabled` and the upstream prerequisite

The spec requires `enabled` in the `PUT` body. `KibanaHTTPAPIsAlertingPutActionPolicy` has no `Enabled` field on the pinned client, so the action policy resource is **blocked** until Kibana adds it through [elastic/rna-program#1266](https://github.com/elastic/rna-program/issues/1266) (targeted at 9.7.0, which is why the action policy version gate is 9.7.0 in D7) and the regular Renovate `kibana-openapi-spec` bump brings it into `generated/kbapi`. The client is not regenerated by hand for this change. The tasks file gates the whole action policy slice on that. The rule resource ships first. Fallback reconciliation through `_enable`/`_disable` is deliberately not designed in, so the resource never has two write models.

### D13. Testing approach

- **Unit**: model ↔ API mapping in both directions for every union variant and every optional object (present / absent), `GetVersionRequirements`, the fail-closed field-accounting test (D4), the duration and id validators, and the `ValidateConfig` rules with known and unknown inputs.
- **Wrapper**: `httptest`-backed tests of the `kibanaoapi` functions covering 201 vs 200 on `PUT`, 404 on `GET`/`DELETE`, 503 `ALERTING_DISABLED`, 400 with a JSON error body, and the `/s/<space>` path prefix, modelled on `internal/clients/kibanaoapi/alerting_rule_test.go`.
- **Plan-only**: `ExpectError` steps for every validation scenario in the specs. These run without `TF_ACC`.
- **Acceptance** (9.6.0+ stack for rules, 9.7.0+ for action policies): one test per spec acceptance item, each step followed by `plancheck.ExpectEmptyPlan()`. Configs in `testdata/<Test>/<step>/` with unique names through variables. Import steps with `ImportStateVerify`. Action policy tests create their destination workflow in the same configuration through `elasticstack_kibana_agentbuilder_workflow` (the Kibana Workflows API the dispatcher dispatches to). Confirm during implementation that the id it returns is what `destinations[].id` expects.
- **Registration**: `entitycore_contract_test.go` per package and the provider-level resource registration test.

### D14. Documentation

Each resource gets `examples/resources/<type>/resource.tf` (a realistic alert rule with breach, recovery, no-data, grouping, and a routing tag, together with a policy routing those tags to a workflow) and `import.sh`, plus `descriptions/` markdown embedded into `MarkdownDescription`. The rule description states that updates are full replacements (last write wins) and that the engine must be enabled. `make docs-generate` produces the registry pages. No hand-written template is needed unless `tfplugindocs` output needs reordering.

## Risks / Trade-offs

- [Full-replace `PUT` is last-write-wins. A UI edit between refresh and apply is overwritten] → Standard Terraform ownership semantics, stated in the resource description. `kind` changes are the only server-side conflict and `RequiresReplace` prevents sending them.
- [A new upstream write field at any depth, or a new union variant, would be silently cleared or rejected on every apply] → Fail-closed field-accounting unit test (D4) walks nested and anonymous structs and every union variant, and fails the Renovate regeneration PR by path or variant name.
- [Create pre-check adds one `GET` and still has a race window] → One request on create only. The race is closed by treating `PUT` 200 as an error with an explicit "replaced" message.
- [Null ⇄ absent mapping assumes the API never materialises defaults for omitted optional objects such as `grouping` (absent means `per_alert`)] → The Context section records that responses echo values as sent. `time_field` is the one scalar with a server default and D9 maps it explicitly. The "removing grouping/throttle clears it" and "absent tags read back as null" acceptance scenarios catch any drift cheaply, and the mapping would then switch to treating the server default as null.
- [`ms` and `w` durations are accepted by the validator but may be rejected by future server rules] → The validator mirrors the schema regex exactly. Server rejections surface verbatim.
- [Action policy resource depends on an upstream Kibana change landing and being picked up by Renovate] → Tasks are gated. The rule resource is independent and ships first. The design needs no change when the field lands.
- [Serverless availability of Alerting v2 at release time is unverified] → The version gate is bypassed on serverless by the envelope. If the routes are absent, requests fail with a clear 404 diagnostic and acceptance tests are skipped by flavour.
- [`elasticstack_kibana_agentbuilder_workflow` may not be the workflow object action policies dispatch to] → Confirm during the action policy slice. If not, the acceptance test creates the workflow through a provider-supported alternative and the design note is updated.

## Migration Plan

New resources with no prior state. Nothing to migrate. Rollout is two pull requests after this proposal merges: the rule resource first, the action policy resource once the upstream `enabled` field is in the pinned client. Rollback is removing the resource from configuration. Destroy issues a normal `DELETE`.

## Open Questions

- Which Kibana pull request implements [elastic/rna-program#1266](https://github.com/elastic/rna-program/issues/1266) for the action policy `PUT`, and which pinned `generated/kbapi` commit first contains it. Deferrable: it only unblocks the action policy tasks. The design and specs do not change.
- Whether Alerting v2 is enabled on serverless projects at the time the rule resource ships. Deferrable: the envelope already bypasses the version gate on serverless. Only the acceptance-test flavour filter might change.
