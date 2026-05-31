## Context

The `entitycore` package provides two envelope families that own the Plugin Framework wiring for concrete entities:

- **Resource envelope** (`ElasticsearchResource[T]`, `KibanaResource[T]`): model constraints require identity accessors (`GetID`, `GetResourceID`, `GetSpaceID`); the envelope resolves read identity centrally (`resolveElasticsearchReadResourceID`, `resolveKibanaResourceIdentity`), passes a resolved `resourceID` (and `spaceID`) into the read callback, drives state removal from a `found` bool, supports a `PostRead` hook, and is constructed via an options struct.
- **Data source envelope** (`genericElasticsearchDataSource[T]`, `genericKibanaDataSource[T]`): model constraints require only the connection getter; the read callback receives the raw config model with `func(ctx, client, T) (T, diag.Diagnostics)`; there is no resolved identity, no `found` bool, no `PostRead`, and constructors are positional.

As a result, concrete data sources re-implement identity resolution, composite-ID parsing, default-space handling, computed `id` assignment, and not-found behavior by hand, with divergent not-found semantics across data sources. This design brings the data source contract to parity with the resource contract so concrete data source implementations shrink to entity-specific mapping only.

## Goals / Non-Goals

**Goals:**
- Resolve read identity (`resourceID`, plus `spaceID` for Kibana) once in the envelope and pass it to the read callback, reusing the existing resource resolution helpers.
- Add a `found bool` return to the read callback and centralize a single not-found policy in the envelope.
- Have the envelope compute and assign the composite `id` so read callbacks never touch `config.ID`.
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

### 5. Envelope owns composite `id` assignment
After a successful read, the envelope sets the model's composite `id` from the scoped client and resolved resource identity (the same `client.ID(ctx, resourceID)` path data sources call today). Read callbacks stop computing or assigning `config.ID`.

*Alternative considered:* leave `id` to the callback. Rejected — it is pure boilerplate present in nearly every Elasticsearch data source.

### 6. Options-struct constructors with optional `PostRead`
Introduce `ElasticsearchDataSourceOptions[T]{ Schema, Read, PostRead }` and `KibanaDataSourceOptions[T]{ Schema, Read, PostRead }`. `NewElasticsearchDataSource[T]`/`NewKibanaDataSource[T]` take the options struct. `PostRead` mirrors the resource `PostReadFunc` shape (runs after state is set on a found read).

*Alternative considered:* add `PostRead` as a trailing positional parameter. Rejected — positional growth is exactly the brittleness the resource envelope avoided with an options struct.

## Risks / Trade-offs

- **Breaking envelope API** → All call sites are in-repo; migrate every concrete data source in the same change and rely on `make build` plus existing acceptance tests to catch regressions.
- **Not-found behavior change for data sources that previously warned (e.g. snapshot repository) or returned partial empty state (e.g. security role)** → Audit each migrated data source; where a hard error materially changes documented behavior, capture it in the delta spec scenarios and the data source's own spec, and confirm acceptance tests still reflect intended behavior. If any data source genuinely requires soft semantics, the callback can return `found == true` with explicitly emptied fields rather than reintroducing envelope branching.
- **Models must add identity accessors** → Mechanical addition of value-receiver methods; covered by the compile-time type constraint, so omissions fail the build rather than at runtime.
- **Composite `id` assignment moving into the envelope could differ from a data source's bespoke id logic** → Verify each migrated data source used the standard `client.ID(...)` composite form (most do); any non-standard cases stay in the callback and set their own id before return.

## Migration Plan

1. Land the new options-struct constructors, model constraints, read signatures, shared identity resolution, centralized not-found policy, and `id`/`PostRead` handling in `internal/entitycore`.
2. Migrate concrete data sources package-by-package (Elasticsearch, Kibana, Fleet), deleting now-redundant identity/`id`/not-found boilerplate and adding model identity accessors.
3. Run `make build` and the data source acceptance tests after each package migration.
4. Update the `entitycore-datasource-envelope` spec to reflect the new contract.
