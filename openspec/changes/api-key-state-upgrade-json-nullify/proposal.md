## Why

`elasticstack_elasticsearch_security_api_key` has two JSON-typed attributes —
`metadata` (`jsontypes.NormalizedType{}`) and `role_descriptors` (a custom type
embedding `jsontypes.Normalized` validation) — that are not handled by the
resource's state upgraders (v0 → v1 and v1 → v2)
(`internal/elasticsearch/security/apikey/resource/state_upgrade.go`).

When the SDKv2 provider persisted `""` for either of those fields (e.g. an API key
created without `metadata` or without `role_descriptors` on a cross-cluster key),
upgrading the provider to a current Plugin Framework build and then running
`terraform plan` produces an `Invalid JSON String Value` error. The Plugin Framework
validates `jsontypes.Normalized`-based custom types against JSON syntax as part of
the `req.State.Get(ctx, &model)` call, so any legacy state with `metadata=""` or
`role_descriptors=""` fails immediately — before any existing normalization runs.

The same class of bug was previously fixed for:
- `elasticstack_elasticsearch_index_template` / `elasticstack_elasticsearch_component_template` (#3914)
- `elasticstack_kibana_alerting_rule` (alerting-rule-state-upgrade-params change)
- `elasticstack_elasticsearch_index_lifecycle` (#4162 / #4167)

This change brings `elasticstack_elasticsearch_security_api_key` into parity.

## What Changes

In both the v0 → v1 and v1 → v2 upgraders in
`internal/elasticsearch/security/apikey/resource/state_upgrade.go`:

- Switch from `req.State.Get(ctx, &model)` to the raw-state pattern:
  unmarshal `req.RawState.JSON` into a `map[string]any` via
  `stateutil.UnmarshalStateMap`, call
  `stateutil.NullifyEmptyString(stateMap, "metadata", "role_descriptors")` to
  nullify empty-string JSON fields, then re-marshal with `stateutil.MarshalStateMap`.
- Retain the existing v0 → v1 logic for `expiration = ""` → `null` (apply it to the
  raw map instead of the typed model).
- Retain the existing v1 → v2 logic for `type = "rest"` default (apply it to the
  raw map instead of the typed model).
- Add unit tests covering the upgraded upgraders with empty-string, null, and
  valid-JSON values for both JSON fields.

No schema changes, no API changes, and no changes to the canonical spec beyond the
state-upgrade invariant.

## Capabilities

### New Capabilities

- _(none)_

### Modified Capabilities

- `elasticsearch-security-api-key`: Fix state upgraders (v0 → v1 and v1 → v2) to
  nullify empty-string `metadata` and `role_descriptors` before decoding into the
  typed model, preventing `Invalid JSON String Value` errors when upgrading from the
  SDKv2 provider to the Plugin Framework provider.

## Impact

- **Specs**: Delta under
  `openspec/changes/api-key-state-upgrade-json-nullify/specs/elasticsearch-security-api-key/spec.md`
  capturing the new state-upgrade invariant.
- **Implementation**: `internal/elasticsearch/security/apikey/resource/state_upgrade.go`
  (both upgraders refactored to use the raw-state pattern); unit test file in the
  same package.
