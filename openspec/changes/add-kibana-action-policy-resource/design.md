# Design

## Context

See proposal.md for motivation and `specs/kibana-action-policy/spec.md` for requirements.

Current state and constraints:

- **Generated client already covers the API.** `generated/kbapi/kibana.gen.go` contains `PostAlertingV2ActionPolicies`, `PutAlertingV2ActionPoliciesId`, `GetAlertingV2ActionPoliciesId`, `DeleteAlertingV2ActionPoliciesId`, `PostAlertingV2ActionPoliciesIdEnable` and `PostAlertingV2ActionPoliciesIdDisable`. The lifecycle paths are `/_enable` and `/_disable`.
  - Request types: `KibanaHTTPAPIsAlertingNewActionPolicy` and `KibanaHTTPAPIsAlertingPutActionPolicy` have identical fields and JSON tags (`kibana.gen.go:38700-38720`, `38786-38806`).
  - Response type: `KibanaHTTPAPIsAlertingActionPolicyResponse`.
  - `destinations`, `grouping` and `throttle` are generated `union json.RawMessage` types with `From…`/`As…` accessors.
  - The Kibana OAS prerequisite (elastic/rna-program#665) is closed, so no regeneration is needed.
- **Kibana API facts** (Kibana checkout `5e44410b25fd`; see Sources):
  - POST creates with a server-generated id.
  - `PUT /{id}` is an upsert: it creates when missing and fully replaces otherwise. Omitted optional fields are cleared; `enabled`, `snoozed_until` and creation metadata are kept.
  - Write bodies are strict and do not accept `enabled`. Lifecycle state changes only through `_enable`/`_disable`. New policies are enabled.
  - Omitted `matcher`/`grouping`/`throttle` are stored as absent and omitted from GET. Kibana applies dispatch defaults only at runtime and never echoes them back.
  - Kibana trims `name`, `description`, `matcher.expression` and `grouping.fields[]`.
  - Every create, PUT or PATCH stores an API key from the caller's credentials. When the caller authenticates with an Elasticsearch API key, that key itself is stored and never invalidated by Kibana. Cloud/UIAM API keys get a 400.
  - The saved object type is space-scoped (`multiple-isolated`); `/s/{space_id}/…` routing works.
  - Errors: 404 `ACTION_POLICY_NOT_FOUND`; 409 `ACTION_POLICY_VERSION_CONFLICT` on concurrent writes; 403 for privileges or license (Enterprise required for create, PUT and enable); 503 `ALERTING_DISABLED` when the `alerting:v2:enabled` setting is off.
  - Routes are `stability: stable, since: 9.6.0`.
  - The 9.5.0 experimental API had an incompatible shape: string matcher, `groupBy`/`groupingMode`, and `version`/`auth` in the response.
  - Alerting v2 is disabled in serverless Elasticsearch, Security and vectordb projects.
- **Envelope behavior.** `entitycore.NewKibanaResource` owns client resolution, the `kibana_connection` block, timeouts, version gating (`WithVersionRequirements`) and read-after-write.
  - `runGenericResourceWrite` (`internal/entitycore/generic_resource_write.go`) returns early on write-callback errors without saving state.
  - `kibana_slo` therefore orphans an SLO when its post-create enable/disable call fails (`internal/kibana/slo/create.go`, `enabled_reconcile.go`). This design must not repeat that.
- **Reference implementation.** `internal/kibana/casetemplate` is the closest analog: envelope, `KibanaSpaceImporter` with `RequireSpaceID`, `SimpleKibanaRead`/`SimpleKibanaDelete`, a `kibanaoapi` wrapper with `SpaceAwarePathRequestEditor`, a `resource-description.md`, and version-gated acceptance tests.
- **Experimental set.** `provider/plugin_framework.go` `experimentalResources()` currently registers only `kibanatag.NewResource`. Acceptance tests always include experimental resources (`AccTestVersion`).

## Goals / Non-Goals

**Goals:**
- Build the resource on the shared Kibana envelope so that connection handling, timeouts, version gating, import and read-after-write behave like the other Kibana resources.
- Keep a policy whose write succeeded tracked in state even when the follow-up enable/disable call fails.
- Keep plan-time validation to the shape of the configuration. Kibana stays the authority for cross-field compatibility, size limits and KQL.

**Non-Goals:**
- Using PATCH, `_snooze`/`_unsnooze`, `_update_api_key`, any bulk endpoint, the execution history or the list endpoint.
- Adopting the new envelope option in `kibana_slo` or other resources.
- Supporting the 9.5.0 experimental API shape.
- Client-side KQL parsing.

## Decisions

### Architecture: envelope with an opt-in persist-on-error write result

| layer | unit | role |
|---|---|---|
| provider registration | `provider/plugin_framework.go` `experimentalResources()` | add `actionpolicy.NewResource` next to `kibanatag.NewResource` |
| resource package | `internal/kibana/actionpolicy/` (new) | embeds `*entitycore.KibanaResource[actionPolicyModel]` and `*entitycore.KibanaSpaceImporter`, following `internal/kibana/casetemplate/resource.go` |
| shared envelope | `internal/entitycore` (change) | new opt-in field on `KibanaWriteResult[T]`, `PersistOnError bool`. When a write callback returns error diagnostics and the field is set, the envelope saves `Model` to state, skips read-after-write, and returns the errors |
| client wrapper | `internal/clients/kibanaoapi/action_policy.go` (new) | thin functions over generated `kbapi`: create, get, put, delete, enable, disable, each with `kibanautil.SpaceAwarePathRequestEditor(spaceID)` |
| generated client | `generated/kbapi` | unchanged |

Boundaries:
- The envelope owns client resolution, `kibana_connection`, timeouts, the version gate and read-after-write.
- The resource package owns the schema, mapping, validation, the existing-policy guard and enabled reconciliation.
- The resource never calls PATCH, `_snooze`/`_unsnooze`, `_update_api_key` or any bulk endpoint.

Why this option over the alternatives:
- **Reconciling inside the callback with the envelope unchanged (the `kibana_slo` pattern)** would orphan a newly created policy whenever `_disable` fails. That violates the spec's partial-failure requirement. A user-chosen `policy_id` would then also be blocked by the existing-policy guard until the user imports it.
- **Overriding Create/Update in the resource** would duplicate client resolution, version gating, timeouts and read-after-write in one package.
- **The opt-in field** is a small, contained change to shared code. It defaults to the current behavior and can later be reused by `kibana_slo`. Its size: `resourceWriteAdapter.invokeWrite`, which is shared with the Elasticsearch adapter, must carry the flag back to `runGenericResourceWrite`. The persist branch must still apply `preserveModelTimeouts` and set the `timeouts` attribute, as the normal path does.

Terraform's own semantics then apply. A create that returns state together with an error is saved **tainted** and replaced on the next apply. An update that returns state together with an error is saved as returned.

### Schema mapping

```hcl
resource "elasticstack_kibana_action_policy" "example" {
  space_id    = <optional, computed, string>   # kbschema.ResourceSpaceIDAttribute(): default "default", RequiresReplace
  policy_id   = <optional, computed, string>   # ^[a-zA-Z0-9_-]{1,150}$; UseStateForUnknown + RequiresReplace
  name        = <required, string>             # non-empty, no surrounding whitespace
  description = <optional, string>             # non-empty, no surrounding whitespace; not computed (PUT clears on omit)
  enabled     = <optional, computed, bool>     # UseStateForUnknown; reconciled only when configured

  destinations = <required, set(object({       # SizeAtLeast(1)
    type = <required, string>                  # OneOf("workflow")
    id   = <required, string>
  }))>

  matcher = <optional, object({                # AtLeastOneOf(tags, expression); not computed
    tags       = <optional, set(string)>       # SizeAtLeast(1)
    expression = <optional, string>            # non-empty, no surrounding whitespace
  })>

  grouping = <optional, object({               # not computed
    mode   = <required, string>                # OneOf("per_alert", "all", "per_field")
    fields = <optional, list(string)>          # required (SizeAtLeast(1)) iff mode = "per_field"; entries non-empty, no surrounding whitespace
  })>

  throttle = <optional, object({               # not computed
    strategy = <required, string>              # OneOf("on_status_change", "per_status_interval", "time_interval", "every_time")
    interval = <optional, string>              # ^\d+(ms|s|m|h|d|w)$; required iff strategy is per_status_interval/time_interval
  })>

  id = <computed, string>                      # "<space_id>/<policy_id>"; UseStateForUnknown

  kibana_connection { ... }                    # injected by the envelope
}
```

- **Collection types:**
  - `destinations` is a set: unique by type and id, and order has no effect on dispatch.
  - `matcher.tags` is a set: any-overlap semantics, so order is irrelevant.
  - `grouping.fields` stays a list, because the order of grouping keys is user intent.
- **No computed optional blocks.** Kibana never echoes defaults, so omitted blocks are read back as absent and map to null. `UseStateForUnknown` is not needed and would hide clear-on-omit.
- **Interval validator.** The existing `validators.ElasticDuration()` accepts decimals and units Kibana rejects, so the interval check is a dedicated regex matching Kibana's `^(\d+)(ms|s|m|h|d|w)$`.
- **Conditional checks.** The `fields` ↔ `per_field` and `interval` ↔ strategy checks use the existing `validators.Condition` helpers where they fit; otherwise a small object-level validator.
- **Not checked at plan time:** grouping ↔ throttle compatibility and size limits stay with Kibana, per the spec decision.

### Component structure

| unit | responsibility |
|---|---|
| `schema.go` | Attributes, validators and plan modifiers above. Description loaded from `resource-description.md`. |
| `models.go` | `actionPolicyModel`: `GetID`, `GetResourceID` (returns `policy_id`), `GetSpaceID`, `GetKibanaConnection`, and `GetVersionRequirements` (minimum `9.6.0`). One `toAPIBody()` builds the write body for both POST and PUT; the two request types are identical, and the result is converted between them. One `populateFromAPI(kbapi.KibanaHTTPAPIsAlertingActionPolicyResponse)` maps the response. |
| `unions.go` | Converts `ActionPolicyDestination`, `ActionPolicyGrouping` and `ActionPolicyThrottle` unions to and from nested objects, dispatching on `type`/`mode`/`strategy` with the generated accessors. |
| `write.go` | `createPolicy`, `updatePolicy` and the shared `reconcileEnabled`. |
| `resource.go` | Envelope wiring (`SimpleKibanaRead`, `SimpleKibanaDelete`, create/update callbacks) and `KibanaSpaceImporter` with `RequireSpaceID` and the `<space_id>/<policy_id>` error message. |
| `resource-description.md` | Covers: technical preview; Kibana 9.6.0+; Enterprise license; Alerting v2 enabled; not available in serverless Elasticsearch or Security; routing-tags explanation; last-writer credentials and the Elasticsearch API key caveat; Cloud API keys rejected; required Kibana privileges (`manage_alerting-v2-action-policies` plus `read_alerting-v2-rules` for writes, `read_alerting-v2-action-policies` for reads). |
| `internal/clients/kibanaoapi/action_policy.go` | Wrapper functions using the repo's typed response handlers. Get maps 404 to "not found" without an error. Delete treats 204 and 404 as success via `diagutil.HandleStatusResponse`. |
| `internal/entitycore` | `KibanaWriteResult.PersistOnError` honored in `runGenericResourceWrite`. |

### Data flow

**Create** (envelope → `createPolicy`):
1. The envelope resolves the scoped client and enforces the 9.6.0 gate.
2. If `policy_id` is set:
   - GET `/{policy_id}`.
   - Found → error "action policy `<policy_id>` already exists in space `<space_id>`; import it with `<space_id>/<policy_id>`". No write.
   - Not found → `PUT /{policy_id}` with `toAPIBody()`.
3. If `policy_id` is unset → POST with `toAPIBody()`, and record the server-generated id.
4. Set `policy_id` and `id = <space_id>/<policy_id>` on the model.
5. Run `reconcileEnabled`. On error, return the model built from the write response (POST or PUT), with `PersistOnError = true` plus the errors.
6. On success, the envelope reads back; the read result becomes state.

**Update** (envelope → `updatePolicy`):
1. `PUT /{policy_id}` with the full `toAPIBody()`. Optional blocks absent from the plan are absent from the body, so the full-replace semantics clear them.
2. Run `reconcileEnabled`. On error after a successful PUT, return `PersistOnError` with the model built from the PUT response: the written fields are in state, `enabled` keeps the server value, and the next plan shows the `enabled` difference.

**`reconcileEnabled`:**
- Decide from the **configuration**, not the plan: if `req.Config.Enabled` is null, do nothing. Because `enabled` is optional, computed and `UseStateForUnknown`, an unconfigured value appears in the plan as the prior state value. A plan-based check would therefore enforce stale state when the plan is run with `-refresh=false`, or when the policy was toggled between refresh and apply.
- If the configured value is unknown, do nothing.
- Otherwise compare the configured value with the `enabled` value in the write response, and call `_enable` or `_disable` only on a mismatch.

**Read** (`SimpleKibanaRead`):
- GET `/{policy_id}`; not found → remove from state.
- `populateFromAPI` maps absent `description`/`matcher`/`grouping`/`throttle` to null and maps unions to nested objects.
- `snoozed_until`, `created_*` and `updated_*` are ignored, so out-of-band snoozes produce no drift.

**Delete** (`SimpleKibanaDelete`): `DELETE /{policy_id}`; 204 or 404 is success.

**Import:** `KibanaSpaceImporter` parses `<space_id>/<policy_id>` into `id`, `space_id` and `policy_id`; a read populates the rest. An id without a space segment is rejected.

**State ownership:**
- Kibana is the source of truth. Final state comes from a read, except in the `PersistOnError` branch, where it comes from the write response.
- `enabled` is optional, computed and `UseStateForUnknown`. When unset, the plan carries the prior server value and Terraform never enforces it.

### Failure handling

| failure | handling |
|---|---|
| Kibana < 9.6.0 | Envelope `EnforceVersionRequirements` returns `Unsupported server version` before any API call |
| 400 (schema, cross-field, limits) | Error diagnostic with HTTP status and Kibana `message`; no retry |
| 403 (privileges, license) | Same |
| 409 `ACTION_POLICY_VERSION_CONFLICT` | Same. No automatic retry: a concurrent writer is a real conflict, resolved by the next plan or apply |
| 503 `ALERTING_DISABLED` | Same |
| Transport error or unexpected status | Generic diagnostic from the typed response handlers |
| `_enable`/`_disable` fails after a successful write | `PersistOnError`: create → tainted; update → saved, with the `enabled` difference on the next plan |
| GET 404 on read | Remove from state |
| DELETE 404 | Success |

### Testing approach

| layer | target | infrastructure |
|---|---|---|
| unit: envelope | `PersistOnError` saves state, skips read-after-write and returns the errors; the default `false` keeps current behavior | existing `kibana_resource_envelope_test.go` harness |
| unit: mapping | `toAPIBody`/`populateFromAPI` round-trip for every grouping mode, throttle strategy and matcher combination; absent blocks map to null | plain Go tests on `kbapi` types |
| unit: validators | `fields`↔`per_field`, `interval`↔strategy, duration format, empty matcher and empty tags, surrounding whitespace and empty strings, `policy_id` pattern | framework validator tests |
| unit: write paths | Guard 200 → error with no PUT; guard 404 → PUT; enabled mismatch → exactly one enable/disable call; reconciliation failure → `PersistOnError` | `httptest` server behind the `kibanaoapi` client, following `kibanaoapi/case_template_test.go` |
| acceptance | Create with server id and with `policy_id`; in-place update; clearing `matcher`/`throttle`; `enabled` toggle; named space; import; destination reorder with an empty plan; existing-id guard error; `ExpectError` for each validation rule; `kibana_connection` override | `acc_test.go` with `versionutils.SkipIfUnsupportedConstraints(>= 9.6.0)` |

Environment notes:
- The docker stack runs a `trial` license (`docker-compose.yml`), which covers the Enterprise gate.
- Alerting v2 is enabled by default in stateful Kibana.
- Kibana does not check that a workflow destination exists at write time, so acceptance tests use a fixed workflow id with no workflow fixture.
- The partial enable/disable failure cannot be reproduced against a real stack, so it is covered only at the `httptest` layer.

### Security and compatibility

- No attribute is sensitive; the schema carries no secrets.
- Every write binds the policy's dispatch credentials to the provider identity. This is documented, not mitigated (user decision: documentation only). The documentation lists the required Kibana privileges so practitioners can scope the provider identity.
- The 9.6.0 gate excludes the incompatible 9.5.0 experimental shape completely, so no dual mapping is needed.
- At most four API calls per create (guard GET, PUT or POST, enable/disable, read-back) and three per update.

### Sources

Kibana checkout `5e44410b25fd` (2026-11-09):
- Routes: `x-pack/platform/plugins/shared/alerting_v2/server/routes/action_policies/*` (`create_action_policy_route.ts`, `upsert_action_policy_route.ts`, `enable_`/`disable_action_policy_route.ts`).
- Schemas: `x-pack/platform/packages/shared/response-ops/alerting-v2-schemas/src/action_policy_data_schema.ts`, `action_policy_response_schema.ts`, `policy_matcher_schema.ts`, `common.ts`.
- Client: `x-pack/platform/plugins/shared/alerting_v2/server/lib/action_policy_client/action_policy_client.ts`, `utils.ts`.
- API keys: `lib/services/api_key_service/api_key_service.ts`.
- Dispatch matching: `lib/dispatcher/steps/evaluate_matchers_step.ts`, `lib/dispatcher/state/policy_matcher.ts`.
- Availability: `server/config.ts`, `settings/advanced_settings.ts`, `config/serverless.*.yml`.
- Commit `53684a94e1fa` (2026-10-19) marked the APIs stable for 9.6.0.

## Risks / Trade-offs

- [Guard race: a policy created by someone else between the guard GET and the PUT is overwritten] → The window is a single round trip, and the spec scopes the guarantee to "exists when create runs". Accepted.
- [An update PUT recreates a policy deleted out of band between refresh and apply] → The result matches the configuration. Accepted.
- [Every write re-binds dispatch credentials to the provider identity; with Elasticsearch API key auth, the key itself is stored and revoking it breaks dispatch] → Documented in the resource description, together with the required privileges.
- [Kibana does not validate the KQL `expression` at write time; an invalid expression silently matches nothing] → Documented as a limitation; no client-side KQL validation (shape-only decision).
- [A failed post-create enable/disable leaves a tainted policy that the next apply replaces, so it gets a new identifier when `policy_id` is unset] → Standard Terraform semantics; preferred over an orphaned policy.
- [The shared envelope gains an option that other resources could misuse] → Defaults to `false`, with envelope unit tests; only `actionpolicy` sets it in this change.
- [Plan-time validation does not mirror grouping ↔ throttle compatibility or size limits, so these fail at apply] → Kibana's 400 message is surfaced verbatim. Accepted to avoid rule drift.

## Migration Plan

- Additive: a new resource type, `schemaVersion` 0, no state upgraders.
- Opt-in: registered in `experimentalResources()`. Users without `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL` see no change.
- The envelope option defaults to the current behavior, so existing Kibana resources are unaffected.
- CI matrix entries below 9.6.0 skip the acceptance tests through the version constraint. The constraint uses `9.6.0-SNAPSHOT` only if pre-release builds need it.
- Docs: generated `docs/resources/kibana_action_policy.md`, an example under `examples/resources/elasticstack_kibana_action_policy/`, and a CHANGELOG "Added" entry marked experimental.
- Rollback: revert the change. The resource is experimental, so no compatibility promise is broken.
- Downstream: the v2 rule resource (elastic/rna-program#673) must expose `metadata.routing_tags`. Nothing in this change depends on it.

## Open Questions

- Was Kibana commit `dd3a598f984d` (destination uniqueness and the future-only `snoozed_until` rule) backported to 9.6? It only changes which invalid inputs Kibana rejects. The gate and schema are unchanged.
