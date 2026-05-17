## Context

The `elasticstack_kibana_action_connector` data source is implemented in `internal/kibana/connector_data_source.go` using `terraform-plugin-sdk/v2`. The connector resource in the same domain already uses the Plugin Framework via `internal/kibana/connectors/`. The provider mux (`provider/factory.go`) allows both to coexist, but the SDK dependency should be eliminated from this domain.

All data source requirements in `openspec/specs/kibana-action-connector/spec.md` are preserved exactly. No schema or behavioral changes are introduced.

## Goals / Non-Goals

**Goals:**
- Replace the SDK data source implementation with a PF one using `entitycore.NewKibanaDataSource`
- Co-locate the data source with its resource in `internal/kibana/connectors/`
- Move acceptance tests alongside the implementation
- Add an SDK upgrade test to verify state compatibility from `v0.15.1`

**Non-Goals:**
- Changing the data source schema or behavior
- Migrating any other entities
- Modifying the connector resource

## Decisions

### Use `entitycore.NewKibanaDataSource` directly

The data source has no create/update/delete — only a read that searches by name. `entitycore.NewKibanaDataSource` provides the connection block injection, client resolution, and state persistence, leaving only the search-and-map logic in the read callback. No custom wrapper struct is needed (unlike resources, which use `type Resource struct { *entitycore.KibanaResource[T] }`).

### Place in `internal/kibana/connectors/data_source.go`

The connector resource already lives in `internal/kibana/connectors/`. Adding `data_source.go` alongside it avoids a new package and keeps domain code together. No shared types are needed between the resource and data source models; each uses its own unexported model struct.

### Model struct

The data source model embeds `entitycore.KibanaConnectionField` for `GetKibanaConnection()` and holds all schema attributes as `types.*` fields. The `config` field uses `jsontypes.Normalized` consistent with how other connectors-domain code handles JSON config to suppress semantically-equivalent-but-textually-different diffs.

### Read callback

The read callback calls `kibanaoapi.SearchConnectors()`, filters to a single match (erroring on zero or >1), constructs a `clients.CompositeID`, and populates the model. This mirrors the SDK implementation exactly.

## Risks / Trade-offs

- **State compatibility**: Terraform state written by the SDK implementation will be read by the PF implementation on upgrade. The attribute names and types are identical, so no state migration is required. The SDK upgrade test (`TestAccConnectorsDataSourceFromSDK`) verifies this path explicitly.
- **`config` JSON type**: The SDK used a plain `string` for `config`. Using `jsontypes.Normalized` in PF means semantically-equivalent JSON is treated as equal — this is strictly better and cannot cause a regression.

## Migration Plan

1. Implement `connectors/data_source.go` + add `NewDataSource()` export
2. Wire in `provider/plugin_framework.go`; remove from `provider/provider.go`
3. Move tests; add SDK upgrade test; delete old files
4. `make build` + acceptance tests
