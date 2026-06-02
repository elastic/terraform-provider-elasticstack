## Context

[Elasticsearch content connectors](https://www.elastic.co/docs/reference/search-connectors) are Python-based integrations (PostgreSQL, MySQL, MSSQL, GitHub, SharePoint, S3, Salesforce, ServiceNow, Confluence, Jira, Notion, OneDrive, Dropbox, Box, MongoDB, Google Drive, ...) that sync third-party data into Elasticsearch. They run as a separate connector service (Docker or from source) that talks to Elasticsearch via the [connector APIs](https://www.elastic.co/docs/api/doc/elasticsearch/group/endpoint-connector) introduced in 8.12.

The API surface is split across ~30 endpoints. The narrow create body (`PUT /_connector/{id}`) accepts only `name`, `description`, `index_name`, `is_native`, `language`, `service_type`. Everything else is set via dedicated partial-update endpoints (`_pipeline`, `_scheduling`, `_features`, `_configuration`, `_name`, `_index_name`, `_service_type`, `_api_key_id`, `_native`, `_status`, `_error`, `_filtering`, `_filtering_validation`, `_active_filtering`). The full document is read back via `GET /_connector/{id}`, which returns a much larger response including configuration schemas registered by the running connector service, runtime telemetry (`status`, `last_seen`, `last_synced`, `error`), and filtering/sync-rules state.

Configuration values are particularly subtle: the connector service registers its per-service-type schema (host, port, password, ...) on first boot. Until then, `PUT /_connector/{id}/_configuration` will fail because the schema is not yet present. The schema's `value` field is later overwritten by user-supplied values; the same schema also exposes a per-field `sensitive` boolean that flags fields containing secrets (PostgreSQL `password`, GitHub `personal_access_token`, S3 `aws_secret_access_key`, etc.).

Sync jobs are documents in an internal index created via `POST /_connector/_sync_job`. They have a fire-and-forget lifecycle (`pending` → `in_progress` → `completed`/`cancelled`/`error`) driven by the running connector service, not by the API caller.

The provider already supports the `WriteOnly` + private-state hash pattern in `kibana_action_connector` and `security_user` (PF v1.19.0, Terraform 1.11+). The shared `internal/utils/writeonlyhash` helper is being introduced by [PR #3415](https://github.com/elastic/terraform-provider-elasticstack/pull/3415); this change is the second consumer.

## Goals / Non-Goals

**Goals:**

- Full lifecycle (Create, Read, Update, Delete, Import) of an Elasticsearch content connector via `elasticstack_elasticsearch_connector`.
- Faithful representation of the connector envelope (`name`, `description`, `service_type`, `index_name`, `is_native`, `language`, `api_key_id`, `api_key_secret_id`) plus the four most-used sub-aspects (`pipeline`, `scheduling`, `features`, `configuration_values`).
- A `configuration_values` schema that supports every realistic value type (string, number, bool, nested JSON, write-only secret) without coercion ambiguity and with per-element sensitivity.
- Drift detection for write-only secret values without requiring user-managed version companions, via the shared `internal/utils/writeonlyhash` helper.
- A companion `data.elasticstack_elasticsearch_connector` data source that exposes the full read-time shape including runtime telemetry the resource omits.
- A `elasticstack_elasticsearch_connector_sync_job_create` provider-defined action that fires a sync job and (optionally) waits for completion, matching the existing snapshot-create action pattern.
- Provider-wide documentation for the `internal/utils/writeonlyhash` helper landing alongside this change, regardless of which change ships the helper implementation.

**Non-Goals:**

- Filtering / sync rules management (`_filtering`, `_filtering_validation`, `_active_filtering`). The draft → validate → activate flow is complex enough to deserve its own change.
- Connector secrets API (`_connector/_secret/*`). Used internally by Elastic-managed connectors; users almost never set these directly.
- Custom scheduling (`custom_scheduling`). Per-document cron is niche; surface it via the data source only.
- Connector-service-side endpoints (`_check_in`, `_last_sync`, `_status`, `_error`, `_update_stats` for sync jobs). These are written by the running connector service, not by Terraform operators.
- A connector sync job *resource*. Sync jobs have a fire-and-forget lifecycle that doesn't fit Terraform's desired-state model; an action is the right pattern (see Decision 7).
- Listing data source for sync jobs (`data.elasticstack_elasticsearch_connector_sync_jobs`). Sync jobs are runtime telemetry; if needed, exposed later.
- A per-service-type typed resource (`elasticstack_elasticsearch_connector_postgresql` etc.). The generic resource is sufficient and avoids ~30 near-duplicate resources.
- Retrofitting other secret-bearing resources to the `writeonlyhash` helper. Migration is a follow-up.

## Decisions

### Decision 1: One resource for the whole connector envelope, not one per sub-aspect

The connector API has separate endpoints for each sub-aspect (`_pipeline`, `_scheduling`, `_features`, ...). The original issue suggested three resources (connector / sync job / pipeline). I reject that split:

- The Kibana UI presents a single connector with nested forms for each aspect. Users think of "a connector" as one thing.
- Splitting into 4–10 resources would create coordination problems: which resource sets `index_name` first? What if pipeline depends on connector existing? Cross-resource dependencies would proliferate.
- Implementing as one resource with fan-out updates is well-supported by the API (every sub-update endpoint is idempotent) and matches how `kibana_action_connector` (single resource, fan-out under the hood) is structured.

The pipeline-update endpoint specifically becomes the `pipeline` nested attribute on the single resource, not a separate `elasticstack_elasticsearch_connector_pipeline` resource.

**Considered:** one resource per sub-aspect, glued by `connector_id`. Rejected as above.

### Decision 2: Resource layout mirrors `internal/elasticsearch/queryrulesets`

Use `entitycore.NewElasticsearchResource[ContentConnectorData]` with the standard file split (`resource.go`, `schema.go`, `create.go`, `read.go`, `update.go`, `delete.go`, `models.go`, `acc_test.go`). Thin client wrappers in `internal/clients/elasticsearch/connector.go`, mirroring `internal/clients/elasticsearch/queryrulesets.go`.

**Why:** Query rulesets are the most recent canonical example of an Elasticsearch resource that uses the typed client and the entitycore envelope. Reusing that shape keeps the provider internally consistent and reviewable in small chunks. The snapshot-create action lives at `internal/elasticsearch/cluster/snapshot_create/` and will be the canonical example for the sync-job-create action.

### Decision 3: Identity is `<cluster_uuid>/<connector_id>`; `connector_id` is optional and immutable

`id` is computed and composite. `connector_id` is `Optional + Computed` so users can either supply a stable ID (`PUT /_connector/{id}`) or let Elasticsearch auto-generate one (`POST /_connector` → returns an id we capture). Changing `connector_id` triggers replacement because Elasticsearch does not support renaming connectors.

`ImportState` accepts the bare connector ID (or the composite form) and resolves the cluster UUID via the connection's `client.ID(ctx, id)` helper, same as `query_ruleset`.

**Why:** Composite IDs are the standard provider convention for Elasticsearch entities. Optional + Computed for `connector_id` matches `kibana_action_connector`.

### Decision 4: Fan-out updates with idempotent sub-endpoints

Create flow:

```
1. POST /_connector              (if no connector_id)   → capture returned id
   OR
   PUT /_connector/{id}          (if connector_id given) — narrow body only
   (name, description, index_name, is_native, language, service_type)

2. For each sub-aspect set in the plan, call the corresponding endpoint:
   - PUT /_connector/{id}/_pipeline       (if pipeline set)
   - PUT /_connector/{id}/_scheduling     (if scheduling set)
   - PUT /_connector/{id}/_features       (if features set)
   - PUT /_connector/{id}/_api_key_id     (if api_key_id or api_key_secret_id set)
   - PUT /_connector/{id}/_configuration  (if configuration_values set, see D6)

3. GET /_connector/{id} — derive final state from the read response.
```

Update flow is identical to Create step 2 + 3, plus:

- `PUT /_connector/{id}/_name` if `name` or `description` changed
- `PUT /_connector/{id}/_index_name` if `index_name` changed
- `PUT /_connector/{id}/_service_type` if `service_type` changed
- `PUT /_connector/{id}/_native` if `is_native` changed

We skip a sub-update if the corresponding plan attribute is `null` AND was `null` in prior state. Configuration-value updates always send the full merged map (the API replaces values atomically).

**Why over a single mega-PUT:** the API doesn't offer one. Fan-out is the only path.

**Why idempotent sub-calls every time:** simpler and safe. Each endpoint is idempotent; sending an unchanged scheduling block is a no-op server-side. The cost (a handful of extra HTTP calls per update) is acceptable.

### Decision 5: Resource omits runtime telemetry; data source exposes everything

The resource exposes only attributes users can *set* plus minimal computed identity (`id`, `connector_id` when auto-generated). It deliberately omits `status`, `last_seen`, `last_synced`, `last_sync_status`, `last_indexed_document_count`, `last_deleted_document_count`, `last_*_sync_error`, `last_*_sync_scheduled_at`, `error`, `filtering`, `custom_scheduling`, the full `configuration` schema document, and `sync_cursor`.

The data source exposes the entire `GET /_connector/{id}` response including all of the above, marked Computed.

**Why:** Resource state should reflect *desired* state. Runtime telemetry is read-only and would either trigger spurious drift on every refresh or pollute plan output. Users who need telemetry can use the data source against the same connector ID. This matches how Fleet resources separate config-time vs runtime fields.

### Decision 6: `configuration_values` is a typed nested map with one value-branch per element

```hcl
configuration_values = map(object({
  # Exactly one of these branches must be set per element (validated)
  string       = optional(string)
  number       = optional(number)              # NumberAttribute (covers int + float)
  bool         = optional(bool)
  json         = optional(string)              # jsontypes.NormalizedType{} for objects/arrays
  secret_value = optional(string)              # WriteOnly + Sensitive; drift via private-state hash
}))
```

Wire encoding: each element compiles to a single JSON value sent in `PUT /_connector/{id}/_configuration` under `{"values": {<key>: <value>}}`. The branch chosen drives the JSON type:

- `string`  → JSON string
- `number`  → JSON number
- `bool`    → JSON boolean
- `json`    → raw JSON pass-through (must be syntactically valid; enforced by `jsontypes.NormalizedType`)
- `secret_value` → JSON string (treated as a string by the connector framework; users JSON-encode if they need nested secret content via `secret_value = jsonencode({...})`)

A per-element `ObjectValidator` enforces exactly one of the five branches.

**Why this shape over the alternatives:**

| Approach | Verdict | Reason |
|---|---|---|
| `map(string)` with parse-with-fallback coercion | Rejected | Ambiguous for nested objects (`'{"a":1}'` is a JSON object or a literal string?); ambiguous for numeric-looking strings (`"555"`). |
| Two top-level maps (`configuration_values` + `configuration_secret_values`) | Rejected | The connector API has a single `values` map; the split would be provider-invented, not API-driven. Doesn't support nested values cleanly. |
| `DynamicAttribute` for per-element value | Rejected | PF dynamic attributes can't host sensitivity per element or write-only sub-attributes. |
| Typed nested map (this decision) | Selected | Faithful 1:1 to the API; supports per-element write-only + sensitive; supports all real value types without coercion. |

**On the read path**, `populateFromAPI` chooses the branch in the model based on the prior-state branch for that key (so `port` stays `{ number = 5432 }` not `{ string = "5432" }`). If the key is new in the API response (no prior state), the provider picks the branch matching the registered field `type` from the schema (`str` → string, `int` → number, `bool` → bool, `list` → string with comma-separated value). Sensitive fields (`configuration.<k>.sensitive == true`) are skipped on read — we never pull a redacted/hashed API value into state.

### Decision 7: Sync jobs are a provider-defined action, not a resource

`elasticstack_elasticsearch_connector_sync_job_create` is a Terraform 1.14+ provider-defined action (mirroring `elasticstack_elasticsearch_snapshot_create`). It accepts `connector_id`, `job_type` (`full` / `incremental` / `access_control`, default `full`), `trigger_method` (`on_demand` / `scheduled`, default `on_demand`), and `wait_for_completion` (default `false`). When `wait_for_completion = true`, it polls `GET /_connector/_sync_job/{id}` until the job reaches a terminal status or `timeouts.invoke` elapses.

**Why an action over a resource:**

| Concern | Resource | Action |
|---|---|---|
| API surface fits | Borderline — sync jobs persist as docs, but have no updatable fields | Clean — fire-and-forget matches `Invoke` semantics |
| State drift | Status fields change without user input → noisy plans | None — actions have no state |
| Lifecycle fit | "Create" is real, "update" is impossible, "delete" is "delete history" | "Invoke" is exactly what users want |
| Precedent | None | `elasticstack_elasticsearch_snapshot_create` (same shape) |

**On `wait_for_completion`:** the action polls every 5 seconds (constant; not exponential) up to `timeouts.invoke` (default 30m). Terminal statuses are `completed`, `cancelled`, `error`, `suspended`. On `error` or `suspended`, the action returns a diagnostic including the job's `error` field. On `cancelled`, the action returns a diagnostic indicating cancellation.

**Cancellation / cleanup:** the action does NOT delete the sync job document on success or failure. Sync job history is useful for operator debugging; users wanting to clean up can call the delete API out-of-band or wait for connector-service-side retention.

### Decision 8: Write-only secret drift detection via `internal/utils/writeonlyhash`

For each map element where `secret_value` is set in config, the provider stores a bcrypt hash of the value in the resource's private state under the key:

`secret_hash:configuration_values["<map_key>"].secret_value`

(matching the path convention established by the [fleet-cloud-connector change](https://github.com/elastic/terraform-provider-elasticstack/pull/3415)).

`ModifyPlan`:

1. For each `configuration_values` element with `secret_value` set in config:
   - Compute the hash with `Hasher{resourceTypeName: "elasticsearch_connector"}.Compute(value)`.
   - Read the stored hash from private state via `PrivateStateKey("configuration_values[\"" + key + "\"].secret_value")`.
   - If no stored hash exists, this is first-apply (or post-import) — no drift signal.
   - If stored hash exists and matches: no drift.
   - If stored hash exists and does NOT match: mark the resource as needing update and emit a warning diagnostic: `"Detected a change to write-only attribute configuration_values[\"<key>\"].secret_value; the resource will be updated."` (Value never logged.)
2. For each `configuration_values` element that is removed from the new configuration: clear the private-state hash (the element was removed; document that removal doesn't unset server-side).

`Create` / `Update`:

- After a successful API write, store the bcrypt hash of each `secret_value` into private state.

`Read`:

- Does NOT touch private state. The stored hash represents the last *applied* value; refresh does not change that.

**Post-import behaviour:** matches `random_password.bcrypt_hash`. First refresh after `terraform import` produces no drift (no stored hash). First subsequent apply baselines the hash. Documented in the resource help text.

**Why over `_wo_version` companions:**
- Detects silent in-config secret edits (`var.password` change with no version bump). The version-companion pattern requires user discipline; this approach catches it automatically.
- One-time implementation cost is amortised across every secret-bearing resource via the shared helper.

**Trade-off — plan opacity:** when only a secret changed, the plan shows nothing visible (write-only attributes don't appear in diffs). The `ModifyPlan` warning diagnostic names the attribute (path only, no value), restoring "what changed" visibility without leaking the secret.

### Decision 9: `writeonlyhash` documentation lands in this change

Regardless of which change ships the helper implementation first, this change adds:

- `dev-docs/high-level/writeonly-secret-hashing.md` covering: the threat model (state-file leaks; offline brute-force resistance via bcrypt), the per-resource-type salt rationale, the `ModifyPlan` contract, the `PrivateStateKey` path convention, the post-import behaviour, and a worked example showing how a resource adopts the helper.
- A reference to that doc from `dev-docs/high-level/coding-standards.md` so future resource authors find it when they design secret-bearing attributes.
- Per-method Godoc on the exported helper API (whichever change builds it owns the Godoc, but this change *adds tests* for the doc-example code path if the helper is already present).

**Why split docs ownership from implementation ownership:** the docs are valuable regardless of which change merges first. Pinning them to one change creates an artificial dependency on merge order.

### Decision 10: Minimum Elasticsearch version 8.12

Connector APIs are GA-marked from 8.12 onward. The typed Go client (go-elasticsearch v8.19.3, already on the dependency tree) supports all required endpoints. The `entitycore` envelope already supports version gating via `GetVersionRequirements`.

Pinning at 8.12 is the most permissive defensible floor. If acceptance tests against any specific 8.12.x patch reveal blocking issues (e.g. an endpoint not yet present in that patch), we bump as needed in a follow-up.

### Decision 11: Configuration-schema-not-yet-registered error

If the user sets `configuration_values` but `GET /_connector/{id}.configuration` returns an empty object (the connector service has not yet booted and registered the per-service-type schema), `PUT /_connector/{id}/_configuration` will fail with a server-side error.

Rather than letting the raw API error bubble up cryptically, the provider performs `GET /_connector/{id}` immediately before the `_configuration` PUT and, if the registered schema is empty, returns a structured diagnostic: `"Connector configuration schema has not been registered yet. The connector service must boot and write a schema for service_type '<x>' before configuration_values can be applied. See https://www.elastic.co/docs/reference/search-connectors/api-tutorial for setup steps."`

**Why surface this explicitly:** schema-registration is the #1 source of user confusion in the API tutorial. Catching it at apply time with a useful message saves real support load.

**Considered:** a `wait_for_configuration_schema` boolean that polls. Rejected for the first cut — the connector service is operator infrastructure that should be running before `terraform apply` rather than spun up by the apply. Add later if there's demand.

## Risks / Trade-offs

| Risk | Mitigation |
|---|---|
| Connector configuration schema isn't registered when the user runs apply | Pre-flight check with a clear diagnostic (Decision 11). Document the connector-service-first ordering in the resource help text and examples. |
| `configuration_values` map verbosity (`host = { string = "x" }` vs `host = "x"`) | The verbosity cost is the price of type safety, per-element sensitivity, and write-only support. Examples in `examples/resources/elasticstack_elasticsearch_connector/` demonstrate idiomatic usage. |
| Branch coercion on Read could mismatch prior state for keys initialised in different branches across applies | The read path uses prior-state branch as the authoritative choice; for new keys, it uses the registered schema's `type`. Acceptance tests cover all branch round-trips. |
| Plan opacity when only a write-only secret changed | `ModifyPlan` warning diagnostic names the attribute path (no value) so users know an update is happening and why. |
| Post-import drift signal for secrets is silent (matches `random_password` baseline behaviour) | Documented in resource help text. First post-import apply baselines the hash. |
| Removing a key from `configuration_values` does NOT unset the value server-side (the API only writes; it has no per-key delete) | Documented. Users who need to truly clear a value can set the branch's value to an empty string per the connector's own semantics. |
| Sync-job action polling consumes connection requests when `wait_for_completion = true` | 5-second poll interval is conservative; `timeouts.invoke` defaults to 30m. Document the cost. |
| `writeonlyhash` helper might land first via PR #3415 OR via this change | Tasks list explicitly handles both orderings: tasks tagged `(only if helper not yet present)` and `(only if helper already present)`. Docs land regardless (Decision 9). |
| Fan-out updates issue multiple PUTs per apply | Acceptable. All sub-endpoints are idempotent; cost is small compared to provisioning latency. |
| Salt format change in the helper would invalidate all stored hashes across the provider | The shared helper documents salt-stability as a hard backwards-compatibility contract. Any future helper change must be additive. |

## Open Questions

1. **Exact 8.12 patch floor**: are any of the partial-update endpoints we depend on only in 8.12.x+N? Verify against the Elasticsearch CHANGELOG or by running acceptance tests against a clean 8.12.0 snapshot during implementation. Pin in `GetVersionRequirements`.
2. **`bcrypt` cost parameter for the helper**: PR #3415 defaults to cost 10 (~100ms). Confirm this is acceptable across all environments where the provider runs (CI, large applies with many secrets). Easy to tune later.
3. **Sync-job action: terminal-status set**: confirmed against go-elasticsearch as `completed`, `cancelled`, `error`, `suspended`. Double-check during implementation that `suspended` is reachable from an on-demand job (or only from scheduled-cancelled mid-flight).
4. **Whether the sync-job action should support a `cancel` mode alongside `create`**: the API exposes `PUT /_connector/_sync_job/{id}/_cancel`. Out of scope for the first cut; add as `elasticstack_elasticsearch_connector_sync_job_cancel` if there's demand.
