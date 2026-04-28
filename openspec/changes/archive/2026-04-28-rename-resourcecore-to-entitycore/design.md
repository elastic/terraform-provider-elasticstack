## Context

`internal/resourcecore` was introduced as a narrowly-scoped substrate for Plugin Framework resources: an embedded `Core` carrying the provider client factory, a `Configure` that funnels through `clients.ConvertProviderDataToFactory`, and a `Metadata` that builds Terraform type names from a typed `Component` plus a literal resource-name suffix. The package has been adopted across ~30 Plugin Framework resources spanning Elasticsearch, Kibana, Fleet, and APM. The `resourcecore_contract_test.go` pattern (per-resource reflection-level assertion that the resource embeds `*resourcecore.Core` anonymously) has stabilized that contract.

Plugin Framework data sources in this provider re-implement the same wiring inline. Today there are ~13 PF data sources across Elasticsearch (`enrich`, `index/indices`, `security/rolemapping`, `index/template`), Kibana (`spaces`, `exportsavedobjects`, three `agentbuilder*`), and Fleet (`outputds`, `integrationds`, `enrollmenttokens`). APM has zero. The duplicated boilerplate is small per data source but uniform across all of them, and it diverges in trivial ways (`fmt.Sprintf` vs string concatenation; whether `Configure` early-returns on nil `ProviderData`; whether interface assertions are present) that are not worth maintaining as variation.

The existing package name `resourcecore` actively misdirects against extending the substrate to data sources: a future reader would reasonably interpret an `entitycore.DataSource` symbol living under `package resourcecore` as a smell. The rename is a precondition for the data-source extension, not an afterthought.

## Goals / Non-Goals

**Goals:**

- Rename the Go package `internal/resourcecore` → `internal/entitycore` and rename the canonical OpenSpec capability `provider-framework-resource-core` → `provider-framework-entity-core`, in one change, with no behavior changes to any concrete entity.
- Rename the embedded resource type from `Core` to `ResourceBase` so the package surface reads as `entitycore.ResourceBase` / `entitycore.DataSourceBase` from the call site.
- Introduce `entitycore.DataSourceBase` with a Configure/Metadata/Client contract that mirrors `ResourceBase` for Plugin Framework data sources.
- Prove `DataSourceBase` on one minimal-diff data source per stack component that has Plugin Framework data sources today (elasticsearch, kibana, fleet), with parallel `entitycore_contract_test.go` files.

**Non-Goals:**

- Migrating every Plugin Framework data source to `DataSourceBase` (the proposal explicitly scopes the migration to one pilot per component).
- Including APM in the data-source pilot (no Plugin Framework data source exists in `internal/apm/` today).
- Any change to Terraform resource type names, data source type names, schemas, identity formats, import behavior, or CRUD logic.
- Sharing a single Configure/Metadata implementation between `ResourceBase` and `DataSourceBase` via generics or wrapper interfaces. The `datasource.ConfigureRequest`/`MetadataRequest` and `resource.ConfigureRequest`/`MetadataRequest` types are distinct, and folding them under a generic adds more indirection than the duplication costs.
- Touching archived change artifacts. Archived `resourcecore` references are left in place as historical record.

## Decisions

### Rename `Core` to `ResourceBase` and add `DataSourceBase` (the `XBase` shape)

The two embedded shapes are named `entitycore.ResourceBase` and `entitycore.DataSourceBase`. This was chosen specifically over keeping `Core` for resources and adding `DataSource`/`DataSourceCore`.

**Why `XBase` rather than `Resource`/`DataSource`:** concrete entity types in this provider are conventionally named `Resource` and `DataSource` in their package (e.g., `package template; type Resource struct{ ... }`). Embedding a type also named `Resource` creates a struct of type `Resource` with an embedded field whose promoted name is `Resource`. The Go compiler accepts this — type-names and field-names are in disjoint namespaces — but the contract test assertion `reflect.TypeFor[*entitycore.Resource]()` reads as if it is asserting against the outer type. `ResourceBase` removes that visual collision while still reading naturally at the embed site (`*entitycore.ResourceBase`).

**Alternative considered:** keep `Core` for resources and add only `DataSourceCore`. Rejected because once the package is `entitycore`, retaining `Core` for the resource arm is asymmetric (`entitycore.Core` for resources vs `entitycore.DataSourceCore` for data sources reads as if the resource arm is the default and the data-source arm is an addendum), and that asymmetry is exactly what the rename is trying to remove.

**Alternative considered:** shared private `configurable` substrate that both `ResourceBase` and `DataSourceBase` embed. Rejected because the `Configure` bodies on each side bind to different framework request/response types and cannot share an implementation; the shareable surface (the `Component`, the literal name string, and the stored `*ProviderClientFactory`) is two fields plus one accessor, which is below the threshold where indirection pays back.

### `DataSourceBase` is the strict mirror of `ResourceBase`

`DataSourceBase` carries the same fields (`component`, `dataSourceName`, `client`), exposes the same accessor (`Client() *clients.ProviderClientFactory`), implements `Configure` against `datasource.ConfigureRequest`/`ConfigureResponse`, and implements `Metadata` against `datasource.MetadataRequest`/`MetadataResponse`. The constructor is `NewDataSourceBase(component Component, dataSourceName string) *DataSourceBase`, parallel to `NewResourceBase`.

The Configure diagnostics rule is identical to `ResourceBase`:

- Append the diagnostics returned from `clients.ConvertProviderDataToFactory`.
- If the response has any error-level diagnostics after appending, return without assigning a factory; leave any factory previously stored by an earlier successful Configure unchanged.
- Otherwise, assign the conversion result (which may be a nil factory when `ProviderData` is nil), replacing any prior stored value.

This is the same wording as the existing resource-core spec requirement, ported across.

The Metadata format is `<provider_type_name>_<component>_<data_source_name>`, the same shape as the resource side. The literal `data_source_name` suffix is passed in unmodified to preserve any existing legacy spellings, identical to the resource-side decision.

`DataSourceBase` does NOT define `Schema`, `Read`, or `ConfigValidators`. Concrete data sources retain full ownership of their schema and read paths, mirroring how `ResourceBase` deliberately does not define `ImportState`, `Schema`, or CRUD.

### Pilot picks: minimum diff, broad shape coverage

One pilot per stack component that already has Plugin Framework data sources. Picks were chosen for minimum diff (smallest single-file impact, no extra interface surface to align):

- **Elasticsearch:** `internal/elasticsearch/enrich/data_source.go` (`enrichPolicyDataSource`). The Configure/Metadata methods are short; `Schema` and `Read` live in the same file and stay untouched. Component `elasticsearch`, literal name `enrich_policy`.
- **Kibana:** `internal/kibana/spaces/data_source.go` (`dataSource`). 63 lines total; the Configure block does an explicit nil-`ProviderData` early return today, which is subsumed by `DataSourceBase`'s `clients.ConvertProviderDataToFactory` flow. Component `kibana`, literal name `spaces`.
- **Fleet:** `internal/fleet/enrollmenttokens/data_source.go` (`enrollmentTokensDataSource`). 54 lines total. Component `fleet`, literal name `enrollment_tokens`.

These three together cover the three permutations of how the existing data sources spell `Configure` (one with `fmt.Sprintf` Metadata, two with string concatenation; one without an explicit `ProviderData == nil` check, two with), so reviewing the three diffs is sufficient to validate that `DataSourceBase`'s shape absorbs the inline patterns without behavioral change.

### Spec capability is renamed, not duplicated

The OpenSpec capability `provider-framework-resource-core` is renamed to `provider-framework-entity-core`. The delta in this change is authored at the new capability path (`openspec/changes/rename-resourcecore-to-entitycore/specs/provider-framework-entity-core/spec.md`) and contains the full intended post-rename canonical spec, with the existing resource-core requirements ported into entity-core wording (resource-side requirements remain effectively unchanged, just renamed) and the new data-source requirements added.

The archive step for this change moves `openspec/specs/provider-framework-resource-core/spec.md` to `openspec/specs/provider-framework-entity-core/spec.md` and overwrites it with the delta contents.

**Alternative considered:** keep both capability names alive (resource-core for the resource arm, entity-core for the new data-source arm). Rejected because the substrate is one package; splitting its specification across two capability names would force every future cross-arm requirement to duplicate, and the package rename already commits to a single canonical name.

### Mechanical scope of the rename

The rename is intentionally scoped to:

1. The package directory move (`internal/resourcecore/` → `internal/entitycore/`) and the `package` declarations within it.
2. Symbol renames inside the package: `Core` → `ResourceBase`, `New` → `NewResourceBase`. `Component`, the four `Component*` constants, and `Client` are unchanged.
3. Every import path in the repo from `.../internal/resourcecore` to `.../internal/entitycore`.
4. Every embed site spelling from `*resourcecore.Core` to `*entitycore.ResourceBase`.
5. Every constructor call from `resourcecore.New(...)` to `entitycore.NewResourceBase(...)`.
6. The six `resourcecore_contract_test.go` files: rename to `entitycore_contract_test.go`, change the embed assertion, change the asserted field name from `Core` to `ResourceBase`.
7. Active OpenSpec change artifacts that mention `resourcecore` are updated; archived changes are not touched.

Everything else (resource implementations, schemas, CRUD, validators, state upgraders, acceptance tests, generated docs) is unchanged.

### Non-decision: ordering vs. other in-flight work

Ordering against any in-flight branches that reference `resourcecore` is left to the implementer. Whichever lands second performs the trivial spelling fix in its own diff.

## Risks / Trade-offs

**Risk: silent interface-promotion regression.** Renaming `Core` to `ResourceBase` and changing the field's promoted name in every embedding struct could in principle break a method-set that relies on a specific embedded-field spelling. Mitigation: the existing `resourcecore_contract_test.go` files (renamed to `entitycore_contract_test.go`) assert `field.Anonymous` and the embed type, which transitively guarantees the standard method promotion. No code in this repo accesses the embedded core via its field name; it always uses promoted methods or the `Client()` accessor.

**Risk: large mechanical diff is hard to review.** ~35 files touched, mostly one-line embed changes plus import path updates. Mitigation: the substantive surface (the new `DataSourceBase` and the three pilot rewrites) is concentrated in five files and can be reviewed independently from the rename. The PR description should explicitly invite reviewers to start from `internal/entitycore/data_source_base.go` and the three pilot diffs.

**Trade-off: duplicated `Configure`/`Metadata` bodies between `ResourceBase` and `DataSourceBase`.** Generics could collapse them; the cost would be a generic embed shape that's harder to read at the call site. Choosing duplication keeps both arms transparent and aligned with how the Plugin Framework itself separates the two interface families.

**Trade-off: pilot-only data-source migration leaves ~10 PF data sources still re-implementing the boilerplate.** This is intentional. A subsequent change can sweep them once `DataSourceBase` is proven. Including all of them here would obscure the substrate decision behind a much larger diff.

## Migration Plan

This change has no user-visible migration. For internal consumers (the provider's own codebase):

1. Land this change in a single PR.
2. After merge, any in-flight branch that references `resourcecore` rebases and applies the spelling fixes (`*resourcecore.Core` → `*entitycore.ResourceBase`, `resourcecore.New` → `entitycore.NewResourceBase`, import path swap).
3. The remaining ~10 Plugin Framework data sources can be migrated incrementally in follow-up changes, each adopting the `entitycore_contract_test.go` pattern.

## Open Questions

None. Open threads from the explore session are resolved by the proposal: single change (A+B+C), `XBase` naming, APM excluded from the data-source pilot, minimal-diff pilots (`enrich_policy`, `spaces`, `enrollment_tokens`), spec capability renamed, ordering vs. other in-flight work deferred to the implementer.
