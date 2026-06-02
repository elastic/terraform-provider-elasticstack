# Design: Kibana Security Entity Store

## Context

The Kibana Security Entity Store API is available via the generated kbapi client
(`generated/kbapi/kibana.gen.go`). The relevant generated client operations are:

- `PostSecurityEntityStoreInstallWithResponse` — `POST /api/security/entity_store/install`
- `PutSecurityEntityStoreWithResponse` (via `PutSecurityEntityStoreJSONRequestBody`) — `PUT /api/security/entity_store`
- `GetSecurityEntityStoreStatusWithResponse` — `GET /api/security/entity_store/status`
- `PutSecurityEntityStoreStartWithResponse` — `PUT /api/security/entity_store/start`
- `PutSecurityEntityStoreStopWithResponse` — `PUT /api/security/entity_store/stop`
- `PostSecurityEntityStoreUninstallWithResponse` — `POST /api/security/entity_store/uninstall`

The request bodies are typed as:

- Install: `PostSecurityEntityStoreInstallJSONBody` — `EntityTypes`, `HistorySnapshot.Frequency`,
  and `LogExtraction.*` fields.
- Update: `PutSecurityEntityStoreJSONBody` — `LogExtraction` (required, not a pointer) with the
  same sub-fields as install minus `HistorySnapshot`.
- Start/Stop: `PutSecurityEntityStoreStartJSONBody` / `PutSecurityEntityStoreStopJSONBody` —
  optional `EntityTypes`.
- Uninstall: `PostSecurityEntityStoreUninstallJSONBody` — optional `EntityTypes`.

The status response (`GetSecurityEntityStoreStatusResponse.JSON200`) contains an overall `Status`
(`SecurityEntityAnalyticsAPIStoreStatus`) and a per-engine `Engines` slice with per-engine status,
type, extraction settings (delay, frequency, etc.), and optional component detail.

## Package Layout

```
internal/kibana/security_entity_store/
  resource.go              — resource constructor, interface assertions
  schema.go                — Plugin Framework schema for the resource
  models.go                — tfModel struct (tfsdk tags)
  create.go                — Create callback: install + optional start
  read.go                  — Read callback: GET status, state removal when not_installed
  update.go                — Update callback: PUT log-extraction, reconcile start/stop, add types
  delete.go                — Delete callback: uninstall managed entity types
  data_source.go           — data source constructor
  data_source_schema.go    — schema for the data source
  data_source_models.go    — dsModel struct
  data_source_read.go      — Read callback for data source
  acc_test.go              — acceptance tests (TestAcc prefix)
```

## Resource Schema

### `elasticstack_kibana_security_entity_store`

```
id                         = <computed, string>        — "<space_id>/entity_store"
space_id                   = <optional/computed, string, ForceNew>
entity_types               = <optional/computed, set(string)> — valid: user, host, service, generic
allow_entity_type_shrink   = <optional, bool, default false>
started                    = <optional/computed, bool, default true>
history_snapshot           = <optional, single-nested block>
  frequency                = <optional, string>        — install-only; treated as ForceNew
log_extraction             = <optional, single-nested block>
  additional_index_patterns = <optional, list(string)>
  excluded_index_patterns   = <optional, list(string)>
  delay                     = <optional, string>
  docs_limit                = <optional, int64>
  field_history_length      = <optional, int64>
  frequency                 = <optional, string>
  lookback_period           = <optional, string>
  max_logs_per_page         = <optional, int64>
  max_logs_per_window       = <optional, int64>
  max_logs_per_window_cap_behavior = <optional, string enum: drop, defer>
  max_time_window_size      = <optional, string>
status_json                = <computed, string>        — normalized JSON from last status read
kibana_connection          = <optional, single-nested block>
```

`history_snapshot.frequency` triggers replacement because the install endpoint accepts it but the
update endpoint does not.

`allow_entity_type_shrink` is stored in state but not sent to the API; it is a guard flag that
controls whether removal of entity types from `entity_types` is allowed.

`started` is managed by calling `PUT /api/security/entity_store/start` or
`PUT /api/security/entity_store/stop` after install/update.

## Data Source Schema

### `elasticstack_kibana_security_entity_store_status`

```
space_id           = <optional/computed, string>
include_components = <optional, bool>
installed          = <computed, bool>
overall_status     = <computed, string>
engines_json       = <computed, string>   — normalized JSON of the engines array
status_json        = <computed, string>   — normalized JSON of the full status response
kibana_connection  = <optional, single-nested block>
```

## CRUD Lifecycle

### Create

1. Call `PostSecurityEntityStoreInstallWithResponse` with `EntityTypes` (nil = all default types),
   `HistorySnapshot`, and `LogExtraction` fields from the plan.
2. Accept HTTP 200 (already installed) and 201 (newly installed) as success.
3. If `started == false`, call `PutSecurityEntityStoreStopWithResponse`.
4. Call Read to populate computed fields.

### Read

1. Call `GetEntityStoreStatusWithResponse`.
2. If response indicates `not_installed`, call `resp.State.RemoveResource(ctx)` and return.
3. Populate `entity_types` from the engine types in the response.
4. Populate `started` by checking whether any engine's status is `running`.
5. Populate `log_extraction` from the per-engine extraction settings (using the first engine's
   values as representative, since all engines share the same log-extraction config).
6. Serialize the full status JSON to `status_json`.

### Update

Three concerns reconciled in order:

1. **Log extraction**: If `log_extraction` changed, call `PutSecurityEntityStoreWithResponse`.
2. **Entity type expansion**: If new types appear in `entity_types` that are not present in current
   state, call `PostSecurityEntityStoreInstallWithResponse` with the desired set.
3. **Entity type shrink**:
   - If `allow_entity_type_shrink == false` and any engine types would be removed, return an error
     diagnostic. Do not call the API.
   - If `allow_entity_type_shrink == true`, call `PostSecurityEntityStoreUninstallWithResponse`
     with the removed types.
4. **Start/stop**: If `started` changed, call start or stop accordingly.

### Delete

Call `PostSecurityEntityStoreUninstallWithResponse` with the `entity_types` recorded in state (or
nil to uninstall all). Remove resource from state.

## Import

Import ID format: `<space_id>/entity_store` (matching the computed `id`). The Read callback
populates all computed fields; `allow_entity_type_shrink` defaults to `false` and
`history_snapshot` is left null (install-only, not recoverable from status).

## Version Gate

`EnforceMinVersion("9.1.0")` is applied in Create, Read, and Update via the Kibana resource
envelope's `WithVersionRequirements` interface or an explicit check at the start of each callback.

## Comparison with Existing Patterns

- `internal/kibana/connectors/` — reference for `entitycore.KibanaResource[T]` envelope usage,
  `MinVersion` constants, and `ImportState` delegation.
- `internal/kibana/security_role/` — reference for multi-file CRUD breakdown, `schema.go` +
  `models.go` + per-operation files.
- Log extraction uses `int64` in the Terraform schema but the kbapi uses `*int`; conversion via
  `int64(v)` / `int(v)` is required.

## Open Questions

- **Partial read of log_extraction**: The status endpoint returns per-engine extraction settings.
  If engines have diverged (e.g., after a manual Kibana update), which engine's values should
  be used to populate state? The simplest approach (first engine, or the `user` engine as
  canonical) should be documented in the spec.
- **`entity_types` after install**: The install endpoint with `entity_types = nil` installs all
  default types (user, host, service, generic in 9.x). On first Read, the provider discovers the
  actual set from the status response. The computed default in the schema should be
  `Optional + Computed` so Terraform does not diff if the user omits `entity_types`.
- **Space-scoped API path**: The entity store API under `GET /api/security/entity_store/status`
  uses the Kibana space selected via the client (i.e., the `space_id` on the Kibana connection).
  Confirm whether the generated client passes the space in the URL path or as a header.
- **`started` detection heuristic**: Should `started` be `true` when any engine is running, or
  when all requested engines are running? The issue is ambiguous. Safest default: `true` when
  at least one engine reports `running`.
- **`history_snapshot.frequency` ForceNew**: The install body accepts this field but the update
  body does not. Marking the whole `history_snapshot` block (or just `frequency`) as ForceNew is
  the safe approach. Confirm with acceptance tests whether changing this field triggers replacement
  as expected.
