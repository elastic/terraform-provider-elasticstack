## Context

`elasticstack_elasticsearch_security_api_key` migrated from the SDKv2 provider to
the Plugin Framework provider in a previous release. The resource currently declares
schema version 2 and provides two state upgraders:

- **v0 → v1** (`state_upgrade.go:60-73`): converts `expiration = ""` to
  `expiration = null`. Uses `req.State.Get(ctx, &model)`.
- **v1 → v2** (`state_upgrade.go:74-88`): sets `type = "rest"` as the default.
  Also uses `req.State.Get(ctx, &model)`.

Both upgraders decode prior state directly into `apikey.TfModel`, which declares:

```go
attrRoleDescriptors: schema.StringAttribute{
    CustomType: apikey.RoleDescriptorsCustomType(), // embeds jsontypes.Normalized validation
    ...
},
"metadata": schema.StringAttribute{
    CustomType: jsontypes.NormalizedType{},
    ...
},
```

The Plugin Framework validates `jsontypes.Normalized`-typed fields against JSON
syntax as part of `req.State.Get`. An empty string is not valid JSON; any legacy
v0 state with `metadata=""` or `role_descriptors=""` therefore causes
`Invalid JSON String Value` before the existing normalization (expiration, type)
even runs.

The resource predates the SDKv2 → Plugin Framework migration, where
`role_descriptors` and `metadata` were plain optional/computed strings — so legacy
state with `""` for either is plausible for any API key created without one of them
explicitly configured.

The established fix pattern in this codebase (used by ILM, index-template,
alerting-rule) is:

1. Unmarshal `req.RawState.JSON` into `map[string]any` via `stateutil.UnmarshalStateMap`.
2. Call `stateutil.NullifyEmptyString` for the JSON-typed keys before any typed model decode.
3. Apply remaining normalization directly to the raw map.
4. Re-marshal with `stateutil.MarshalStateMap`.

## Goals

- Prevent `Invalid JSON String Value` errors for practitioners upgrading from
  SDKv2 provider builds with empty-string `metadata` or `role_descriptors` in state.
- Bring `elasticstack_elasticsearch_security_api_key`'s state upgraders into parity
  with the pattern established for ILM, template, and alerting-rule resources.
- Cover both affected upgraders (v0 → v1 and v1 → v2) since either can be triggered
  depending on which historical provider version state was written by.

## Non-Goals

- Changing the API key API shape, resource schema, or any user-facing behaviour.
- Fixing any other structural issue that may affect state upgrade beyond the
  empty-string JSON normalization described here.
- Adding a new state schema version — the fix extends existing upgraders in place.

## Decisions

| Topic | Decision |
|-------|-----------|
| Upgrade path coverage | Both v0 → v1 and v1 → v2 upgraders are affected because either can encounter legacy state with `metadata=""` or `role_descriptors=""`. Both must be fixed. |
| Raw-state pattern | Use `stateutil.UnmarshalStateMap` → `stateutil.NullifyEmptyString` → `stateutil.MarshalStateMap`, identical to the alerting-rule and ILM patterns. |
| Fields to nullify | `"metadata"` and `"role_descriptors"` in both upgraders. |
| Expiration normalization (v0 → v1) | Migrate the existing `expiration = ""` → `null` logic from the typed-model path to the raw-map path using `stateutil.NullifyEmptyString(stateMap, "expiration", "metadata", "role_descriptors")`. |
| Type default (v1 → v2) | Migrate the existing `type = "rest"` default logic from the typed-model path to the raw-map path: check `stateMap["type"] == nil \|\| stateMap["type"] == ""` and set `stateMap["type"] = "rest"`. |
| `schemaWithConnection` / `PriorSchema` | These exist to let `req.State.Get` decode prior state; with the raw-state pattern we no longer call `req.State.Get`, so `PriorSchema` can be dropped from both upgraders. `schemaWithConnection` may be removed if unused elsewhere; otherwise leave it. |
| Unit tests | Add a state-upgrade unit test file covering: (a) v0 state with `metadata=""` → null, (b) v0 state with `role_descriptors=""` → null, (c) both simultaneously, (d) valid JSON unchanged, (e) null unchanged; same matrix for v1 → v2. |
| No state version bump | The fix targets existing upgraders only; no new version is introduced. |

## Risks / Trade-offs

- **Existing valid state**: Practitioners already on the Plugin Framework provider
  with valid JSON in `metadata` or `role_descriptors` are unaffected —
  `NullifyEmptyString` only acts when the value is exactly `""`.
- **Over-nullification**: Setting `metadata` or `role_descriptors` to null when the
  SDKv2 stored `""` is correct — the SDKv2 used `""` to represent "not configured",
  which maps to `null` in the Plugin Framework model.
- **Dropping `PriorSchema`**: The `PriorSchema` fields are only needed for
  `req.State.Get`-based upgraders. With the raw-state pattern, omitting them is
  correct and tested behavior (see ILM and alerting-rule upgraders that omit it).

## Open Questions

_(none — the fix is well-understood by analogy with the existing ILM, template,
and alerting-rule state upgrader fixes)_

## Migration / State

No state version increment is needed. The existing v0 → v1 and v1 → v2 upgraders
are extended in place; their semantics remain compatible with all v0/v1 states that
do not have empty-string `metadata` or `role_descriptors`.
