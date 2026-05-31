## Context

The `entitycore` package provides two envelope families that own the Plugin Framework wiring for concrete entities:

- **Resource envelope** (`ElasticsearchResource[T]`, `KibanaResource[T]`): model constraints require identity accessors (`GetID`, `GetResourceID`, `GetSpaceID`); the envelope resolves read identity centrally (`resolveElasticsearchReadResourceID`, `resolveKibanaResourceIdentity`), passes a resolved `resourceID` (and `spaceID`) into the read callback, drives state removal from a `found` bool, supports a `PostRead` hook, and is constructed via an options struct.
- **Data source envelope** (`genericElasticsearchDataSource[T]`, `genericKibanaDataSource[T]`): model constraints require only the connection getter; the read callback receives the raw config model with `func(ctx, client, T) (T, diag.Diagnostics)`; there is no resolved identity, no `found` bool, no `PostRead`, and constructors are positional.

As a result, concrete data sources re-implement identity resolution, composite-ID parsing, default-space handling, computed `id` assignment, and not-found behavior by hand, with divergent not-found semantics across data sources. This design brings the data source contract to parity with the resource contract so concrete data source implementations shrink to entity-specific mapping only.

## Goals / Non-Goals

**Goals:**
- Resolve read identity (`resourceID`, plus `spaceID` for Kibana) once in the envelope and pass it to the read callback, reusing the existing resource resolution helpers.
- Add a `found bool` return to the read callback and centralize a single not-found policy in the envelope.
- Keep composite `id` assignment in the read callback (matching the resource envelope, where the read callback sets `id`), so standard and non-standard id entities are handled uniformly without envelope branching.
- Move data source constructors to an options struct with an optional `PostRead` hook, matching the resource envelope ergonomics.
- Make Kibana space resolution (default space, composite, unscoped opt-out) identical between data sources and resources.
- Let a single entity read function be sharable between a data source and its resource where the mapping is identical.

**Non-Goals:**
- Changing any Terraform schema, attribute names, or the connection-block injection mechanism.
- Reworking the resource envelope contract (it is the reference; only data sources move toward it).
- Adding `Create`/`Update`/`Delete` semantics to data sources.
- Preserving the legacy not-found variations (warning-only, partial-empty-state) as first-class options; standardization is intended.

## Decisions

### 1. Model constraints gain identity accessors
`ElasticsearchDataSourceModel` and `KibanaDataSourceModel` add `GetID() types.String` and `GetResourceID() types.String` (and `GetSpaceID() types.String` for Kibana), matching the resource model constraints. Concrete models already embed the connection field struct and declare an `id` attribute; they add the same value-receiver accessors used by resources.

*Alternative considered:* keep connection-only constraints and pass identity via a separate optional interface. Rejected — optional interfaces reintroduce per-data-source branching and prevent the envelope from owning identity uniformly.

### 2. Read callback signature mirrors the resource read callback
New signatures:
- Elasticsearch: `func(ctx, *clients.ElasticsearchScopedClient, resourceID string, model T) (T, bool, diag.Diagnostics)`
- Kibana: `func(ctx, *clients.KibanaScopedClient, resourceID string, spaceID string, model T) (T, bool, diag.Diagnostics)`

This is intentionally identical to `elasticsearchReadFunc`/`kibanaReadFunc` so a concrete entity can share one read function across its data source and resource.

*Alternative considered:* keep the config-in/config-out shape and only add `found`. Rejected — it leaves identity resolution duplicated in every data source.

### 3. Reuse resource identity-resolution helpers
The envelope resolves identity from the decoded model using the existing `resolveElasticsearchReadResourceID` and `resolveKibanaResourceIdentity` helpers (composite-ID-or-fallback), including the `KibanaUnscopedSpace` opt-out. Data sources are read-only, so resolution runs against config rather than prior state, but the rules are the same.

### 4. Centralized not-found policy: standardized error
When the read callback returns `found == false`, the envelope appends a single standardized "not found" error diagnostic (including component, name, and resolved identity) and does **not** set state. This replaces the current mix of warning-plus-partial-state, manual field-nulling, and ad-hoc errors. A failed data source read is an error because downstream configuration depends on the resolved values.

*Alternative considered:* warning + empty state (snapshot-repository's current behavior). Rejected as the default — it silently yields null attributes that break dependent config; standardizing on an error is the safer contract. See Risks for the behavior-change handling.

### 5. Read callback owns `id` assignment (parity with the resource envelope)
The concrete read callback computes and assigns the model's `id` and returns it on the model `T`; the envelope does not mutate `id`. This matches the resource envelope exactly, where the read callback sets `data.ID` (for example `data.ID = types.StringValue(client.ID(ctx, resourceID).String())`) and the envelope only persists the returned model. The data source model constraint exposes `GetID()` for read-identity resolution but intentionally provides no identity mutator, so a value-typed generic `T` cannot be assigned by the envelope.

Standard entities assign the composite `id` via `client.ID(...)`. Non-standard entities assign an entity-specific `id` directly in the callback with no envelope opt-out: `internal/elasticsearch/cluster/info` derives `id` from `cluster_uuid`, and `internal/elasticsearch/index/indices` uses the target pattern rather than `client.ID(resourceID)`.

*Alternative considered:* have the envelope own `id` assignment via a new `SetID` mutator on the model constraint. Rejected — it would require a pointer-receiver / `*T` constraint pattern the resource envelope does not use (breaking the "match the resource model constraints" parity goal) and could not express non-standard ids (cluster UUID, index target pattern) without an opt-out. Keeping `id` in the callback handles standard and non-standard entities uniformly with no envelope branching.

### 6. Options-struct constructors with optional `PostRead`
Introduce `ElasticsearchDataSourceOptions[T]{ Schema, Read, PostRead }` and `KibanaDataSourceOptions[T]{ Schema, Read, PostRead }`. `NewElasticsearchDataSource[T]`/`NewKibanaDataSource[T]` take the options struct. `PostRead` runs after state is set on a found read, mirroring the resource `PostReadFunc` ordering. The data source signatures are `func(ctx, *clients.ElasticsearchScopedClient, T) diag.Diagnostics` and `func(ctx, *clients.KibanaScopedClient, T) diag.Diagnostics`; they deliberately omit the resource `PostReadFunc`'s trailing `privateState any` argument because `datasource.ReadResponse` has no `Private` field (data sources have no private state).

*Alternative considered:* add `PostRead` as a trailing positional parameter. Rejected — positional growth is exactly the brittleness the resource envelope avoided with an options struct.

## Risks / Trade-offs

- **Breaking envelope API** → All call sites are in-repo; migrate every concrete data source in the same change and rely on `make build` plus existing acceptance tests to catch regressions.
- **Not-found behavior change for data sources that previously warned (e.g. snapshot repository) or returned partial empty state (e.g. security role)** → Audit each migrated data source; where a hard error materially changes documented behavior, capture it in the delta spec scenarios and the data source's own spec, and confirm acceptance tests still reflect intended behavior. If any data source genuinely requires soft semantics, the callback can return `found == true` with explicitly emptied fields rather than reintroducing envelope branching.
- **Models must add identity accessors** → Mechanical addition of value-receiver methods; covered by the compile-time type constraint, so omissions fail the build rather than at runtime.
- **Non-standard `id` derivation** → Because the read callback owns `id`, standard entities call `client.ID(...)` while non-standard entities set their own `id` directly in the callback with no envelope opt-out (`internal/elasticsearch/cluster/info` derives `id` from `cluster_uuid`; `internal/elasticsearch/index/indices` uses the target pattern). Verify each migrated data source still assigns `id` in its read function.

## Migration Plan

1. Land the new options-struct constructors, model constraints, read signatures, shared identity resolution, centralized not-found policy, and `id`/`PostRead` handling in `internal/entitycore`.
2. Migrate concrete data sources package-by-package (Elasticsearch, Kibana, Fleet), deleting now-redundant identity/`id`/not-found boilerplate and adding model identity accessors.
3. Run `make build` and the data source acceptance tests after each package migration.
4. Update the `entitycore-datasource-envelope` spec to reflect the new contract.
