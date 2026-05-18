## Context

`role.go` (766 LOC) and `role_data_source.go` (258 LOC) use `terraform-plugin-sdk/v2`. The data source delegates its read directly to `resourceRoleRead()` — there is no independent implementation. The schema is deeply nested: 47+ attributes across `elasticsearch` (with `indices`, `remote_indices`, `field_security`) and `kibana` (with `feature`) blocks, plus 11 expand/flatten helper functions (126 LOC).

The existing `kibanaoapi/security_role.go` client already returns PF-compatible `fwdiag.Diagnostics` for the Get, Put, and Delete helpers. No client migration is needed.

## Goals / Non-Goals

**Goals:**
- Migrate both entities to PF using `entitycore.NewKibanaResource` and `entitycore.NewKibanaDataSource`
- Preserve the old delegation pattern via a shared `fetchRole()` package-level helper
- Preserve all behavioral requirements: version gating, JSON normalization, `base` XOR `feature` validation, field omission rules
- Add SDK upgrade test for the resource
- All existing acceptance tests pass unchanged

**Non-Goals:**
- Schema changes of any kind
- Migrating client functions (already on PF diags)

## Decisions

### Single package for resource and data source

The resource and data source share the same nested schema structure (47+ attrs), the same API client calls, and the same flatten/expand logic. A single `security_role/` package holds both, sharing `models.go`, the flatten helpers, and `fetchRole()`. This mirrors the old delegation, now expressed as a proper package-internal function rather than SDK function reuse.

### `fetchRole()` as shared read helper

```go
// fetchRole calls the Get role API and returns a populated nested role struct.
// Returns (role, found, diags). found=false when the API returns 404/not-found.
func fetchRole(ctx, client, name string) (*kbapi.SecurityRole, bool, diag.Diagnostics)
```

The resource read callback calls `fetchRole()` then populates `resourceModel`. The data source read callback calls `fetchRole()` then populates `dataSourceModel`. Flatten helpers in `flatten.go` accept `*kbapi.SecurityRole` and return the appropriate nested model types — used by both.

### `GetVersionRequirements()` for version gates

Both version gates are conditional on attribute presence:
- `remote_indices` non-empty → require Kibana ≥ 8.10.0
- `description` non-null and non-empty → require Kibana ≥ 8.15.0

The `resourceModel` implements `entitycore.WithVersionRequirements`: `GetVersionRequirements()` inspects both fields and returns whichever requirements apply (zero, one, or both). The entitycore envelope enforces these before calling create or update. This exactly matches the runtime check in the SDK implementation.

The data source does not need version gating (it reads existing roles and these are computed outputs).

### `jsontypes.NormalizedType` for JSON fields

The SDK used `DiffSuppressFunc: tfsdkutils.DiffJSONSuppress` for `metadata` and `query`. In PF, these attributes use `jsontypes.Normalized` (or `types.String` with a `SemanticEquals` implementation). `jsontypes.Normalized` is the idiomatic PF approach and is already used elsewhere in this provider.

### `base` XOR `feature` as schema validator

The SDK validated this at runtime in the expand function. In PF, this becomes an `ObjectValidator` on the `kibana` block (or a resource-level `ConfigValidator`) that checks mutual exclusivity at plan time. An `AtLeastOneOf`-style validator also enforces that at least one of `base` or `feature` is set. The error message matches the existing SDK behavior.

### Model structure

```
security_role/models.go

resourceModel
  entitycore.KibanaConnectionField
  ID           types.String   // = Name (role name is the identity)
  Name         types.String
  Description  types.String
  Metadata     jsontypes.Normalized
  Elasticsearch []ESModel      // max 1
  Kibana        []KibanaModel

  GetID()              → m.Name
  GetResourceID()      → m.Name
  GetSpaceID()         → types.StringValue("") // role API is unscoped; envelope ignores empty space
  GetVersionRequirements() → dynamic based on Description + remote_indices

dataSourceModel
  entitycore.KibanaConnectionField
  Name          types.String   // the filter key
  Description   types.String
  Metadata      jsontypes.Normalized
  Elasticsearch []ESModel
  Kibana        []KibanaModel

  GetKibanaConnection() → embedded

Shared nested:
  ESModel { Cluster, RunAs types.Set; Indices []ESIndexModel; RemoteIndices []ESRemoteIndexModel }
  ESIndexModel { Names, Privileges types.Set; Query jsontypes.Normalized; FieldSecurity *FieldSecurityModel }
  ESRemoteIndexModel { Clusters, Names, Privileges types.Set; Query jsontypes.Normalized; FieldSecurity *FieldSecurityModel }
  FieldSecurityModel { Grant, Except types.Set }
  KibanaModel { Spaces, Base types.Set; Feature []KibanaFeatureModel }
  KibanaFeatureModel { Name types.String; Privileges types.Set }
```

### Role API is unscoped

Like spaces, the Kibana role management API (`/api/security/roles/{name}`) is not scoped to a Kibana space. `GetSpaceID()` returns `types.StringValue("")` — the entitycore envelope accepts an empty space ID for APIs that don't use space context.

### `id` format

The SDK stored `id = name`. The PF implementation does the same: `GetID()` returns `Name`. The entitycore composite-ID parser will not match (role names are `[a-zA-Z0-9_-]+` by Kibana convention — no `/`). Import via name passthrough is preserved.

## Risks / Trade-offs

- **Flatten/expand correctness**: The 11 SDK helper functions involve subtle field-omission rules (empty `cluster`, empty `run_as`, empty `query`). Each must be translated exactly. Unit tests on the flatten/expand helpers are critical.
- **`base` JSON handling**: The SDK stores `kibana.base` as `json.RawMessage` in the client type to work around a union type. The PF implementation reads the API response field as a `[]string` directly from the generated kbapi type rather than deserializing raw JSON — verify this matches the existing client struct behavior.
- **State compatibility**: Both resource and data source state schemas are preserved exactly. The SDK upgrade test for the resource covers the state migration path.
- **`jsontypes.Normalized` vs plain string**: `metadata` and `query` in the SDK stored raw JSON strings. `jsontypes.Normalized` is semantically equivalent but may change how the value appears in state diffs (normalized form). This is strictly better and matches existing practice elsewhere in the provider.

## Migration Plan

1. Create `security_role/` package with models and shared types
2. Implement flatten helpers (one per nested block type)
3. Implement `fetchRole()` and both read callbacks
4. Implement create/update/delete callbacks
5. Implement schema factories (resource + data source)
6. Add `base` XOR `feature` validator
7. Wire provider, remove SDK registrations, delete old files
8. Move and extend tests; add SDK upgrade test
9. `make build` + acceptance tests pass
