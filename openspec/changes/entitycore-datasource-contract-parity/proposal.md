## Why

The `entitycore` data source envelope (`NewKibanaDataSource`/`NewElasticsearchDataSource`) gives concrete data sources far less than the resource envelope gives concrete resources. The data source read callback receives only the raw config model and must re-implement identity resolution, composite-ID parsing, default-space handling, computed `id` assignment, and not-found behavior by hand. This produces duplicated boilerplate and divergent, inconsistent not-found semantics (warning + partial state, manual field-nulling, or hard error depending on the data source), while the resource envelope already owns all of this centrally.

## What Changes

- Extend the data source model constraints to require identity accessors (`GetID`, `GetResourceID`, and `GetSpaceID` for Kibana), mirroring the resource model constraints. **BREAKING** for the envelope's exported type constraints and read-callback signature.
- The envelope resolves read identity centrally (reusing the resource helpers `resolveElasticsearchReadResourceID` / `resolveKibanaResourceIdentity`) and passes a ready-made `resourceID string` (and `spaceID string` for Kibana) into the read callback.
- Change the read callback return type to `(T, bool, diag.Diagnostics)`. The `found` bool drives a single, centralized not-found policy in the envelope instead of bespoke per-data-source handling.
- The read callback continues to compute and assign the model's `id` (matching the resource envelope, where the read callback sets `id`); the envelope resolves read identity but never mutates `id`, since the model constraint exposes `GetID()` with no setter. Standard entities call `client.ID(...)`, while non-standard entities (`cluster/info` cluster UUID, `index/indices` target pattern) assign their own `id`.
- Replace the positional data source constructors with an options struct (`ElasticsearchDataSourceOptions[T]` / `KibanaDataSourceOptions[T]`) carrying `Schema`, `Read`, and an optional `PostRead` hook — parity with `ElasticsearchResourceOptions`/`KibanaResourceOptions`.
- Reuse Kibana space-identifier resolution (default space, composite `<space>/<id>`, and the `KibanaUnscopedSpace` opt-out) so space handling is identical across data sources and resources.
- Migrate all existing envelope-based data sources to the new contract.

## Capabilities

### New Capabilities
<!-- None: this refines the existing data source envelope contract. -->

### Modified Capabilities
- `entitycore-datasource-envelope`: Read-callback signature gains a resolved identity argument (`resourceID`, plus `spaceID` for Kibana) and a `found` boolean return; the envelope owns identity resolution, centralized not-found handling, and an optional `PostRead` hook; constructors move to an options struct; model constraints require identity accessors. The read callback continues to own `id` assignment, matching the resource envelope.

## Impact

- `internal/entitycore/data_source_envelope.go` and `data_source_envelope_test.go` (envelope contract, constructors, model constraints, shared identity/not-found logic).
- Every concrete envelope data source migrates to the new read signature and options constructor, e.g. `internal/elasticsearch/security/role`, `internal/elasticsearch/security/user`, `internal/elasticsearch/security/rolemapping`, `internal/elasticsearch/cluster/snapshot_repository_data_source.go`, `internal/elasticsearch/cluster/info`, `internal/elasticsearch/synonyms`, `internal/elasticsearch/queryrulesets`, `internal/elasticsearch/index/template`, `internal/elasticsearch/index/indices`, `internal/elasticsearch/enrich`, `internal/kibana/agentbuilderskill`, `internal/kibana/agentbuilderagent`, `internal/kibana/agentbuildertool`, `internal/kibana/agentbuilderworkflow`, `internal/kibana/security_role`, `internal/kibana/spaces`, `internal/kibana/connectors`, `internal/kibana/exportsavedobjects`, `internal/fleet/outputds`, `internal/fleet/integrationds`, `internal/fleet/enrollmenttokens`.
- No change to Terraform schemas or user-facing data source behavior beyond standardized not-found semantics; data source acceptance tests should continue to pass.
