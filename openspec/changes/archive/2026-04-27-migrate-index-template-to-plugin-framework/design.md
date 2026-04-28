## Context

`elasticstack_elasticsearch_index_template` is implemented on `terraform-plugin-sdk/v2` in `internal/elasticsearch/index/template.go` (resource) and `internal/elasticsearch/index/template_data_source.go` (data source). The resource ships several bespoke SDK behaviors that need careful preservation when porting to the Plugin Framework:

1. A `DiffSuppressFunc` (`suppressAliasRoutingDerivedDiff`) that hides spurious diffs when the user sets only `alias.routing` and Elasticsearch echoes the same value into `index_routing` and `search_routing` on read.
2. A complementary post-read "preserve user routing" pass (`extractAliasRoutingFromTemplateState` + `preserveAliasRoutingInFlattenedAliases`) that re-injects the user's `routing` value into refreshed alias state.
3. A custom `Set` hash on `alias` keyed by `name` (`hashAliasByName`) so changes to routing-only fields don't change set membership.
4. An 8.x update workaround for `data_stream.allow_custom_routing` that re-sends `false` when prior state had `true`.
5. JSON normalization with `DiffJSONSuppress` for `metadata`, `mappings`, and `alias.filter`, plus `DiffIndexSettingSuppress` for `template.settings` (dotted-vs-nested key equivalence + `index.` prefix normalization).
6. Version gating: `ignore_missing_component_templates` requires ES ≥ 8.7.0 and `template.data_stream_options` requires ES ≥ 9.1.0.

Several of these workarounds were necessary because the SDK does not have first-class semantic equality. The Plugin Framework does, via `*ValuableWithSemanticEquals` interfaces, which lets us replace plan-time hacks with type-level equality that the framework consults during refresh, plan, and post-apply state comparison. Using semantic equality also avoids a real risk of "Provider produced inconsistent result after apply" errors that a plan-modifier-only solution could introduce.

## Goals / Non-Goals

**Goals:**
- Migrate the resource and data source to the Plugin Framework with no breaking changes to attribute paths, block syntax, identity format, or import behavior.
- Replace SDK diff-suppression hacks with custom types that use semantic equality so behavior is consistent across plan, refresh, and post-apply state checking.
- Land the resource on `resourcecore.Core` to satisfy this provider's PF resource testing contract.
- Keep the SDK-implemented `elasticstack_elasticsearch_component_template` working unchanged.

**Non-Goals:**
- Migrating `elasticstack_elasticsearch_component_template` (tracked separately).
- Reworking the schema shape (e.g. flattening `template { … }` into a nested attribute, or moving `alias` from a block to an attribute).
- Changing the public spec for `data_stream_options`, `ignore_missing_component_templates`, or version gating semantics.
- Removing the `template.alias` set semantics; users continue to write `alias { … }` blocks.

## Decisions

### 1. Use `resourcecore.Core` for the resource and explicit `datasource.DataSourceWithConfigure` for the data source

The `internal/resourcecore` package exposes `Configure`, `Metadata`, and `Client()` for resources only; data sources are wired explicitly. This change follows the same split that `ilm`, `index`, and `datastreamlifecycle` use. Concretely:

```go
type Resource struct {
    *resourcecore.Core
}

func newResource() *Resource {
    return &Resource{
        Core: resourcecore.New(resourcecore.ComponentElasticsearch, "index_template"),
    }
}
```

CRUD methods (`Create`, `Read`, `Update`, `Delete`, `ImportState`, `ValidateConfig`) live on `*Resource`. There is no `UpgradeState`: state from the SDK implementation is forward-compatible (see Decision 8).

### 2. Custom type with `ObjectValuableWithSemanticEquals` for the `alias` element

The alias element type implements `basetypes.ObjectValuableWithSemanticEquals` so that the framework treats two alias values as equal when they differ only in API-derived `index_routing`/`search_routing`. Concretely the element is equal when:

- `name`, `is_hidden`, `is_write_index`, and `routing` are equal between the two values.
- `filter` is JSON-equal (delegates to `jsontypes.Normalized` on the inner attribute).
- For each of `index_routing` and `search_routing`: either the values are equal, or the prior value (`v` receiver) is null/empty AND the new value equals the new value's `routing` field (i.e. it's an API-derived echo).

Crucially, semantic equality is consulted during refresh, plan, *and* post-apply state comparison. That means a plan-modifier-only approach is rejected here: a plan modifier solves the plan diff but does not protect against "Provider produced inconsistent result after apply" when state and refresh disagree. Semantic equality is the framework's documented mechanism for this exact problem.

This decision also resolves the alias-set hashing concern. Plugin Framework set membership is determined by element equality, deferring to `SemanticEquals` when implemented. Two aliases with the same `name` but routing-only differences therefore collapse to one set member, replicating `hashAliasByName` without a custom hasher hook (which Plugin Framework does not expose).

The legacy "preserve user routing on read" workaround (`extractAliasRoutingFromTemplateState`/`preserveAliasRoutingInFlattenedAliases`) is *not* ported to PF; it is unnecessary once semantic equality is in place.

#### Strict vs permissive equivalence

The semantic-equality predicate matches the existing SDK behavior of `suppressAliasRoutingDerivedDiff` (line-for-line):

```text
v.index_routing ≡ new.index_routing  ⇔
    v.index_routing == new.index_routing
    OR (v.index_routing is null/empty AND new.index_routing == new.routing AND new.routing != "")
```

Receiver `v` represents the prior/state value and the argument is the new value, matching the convention used in `internal/utils/customtypes/json_with_defaults_value.go`. The same rule applies to `search_routing`.

### 3. Keep `alias` as `SetNestedBlock`, not `SetNestedAttribute`

`alias { name = "x" }` is the existing HCL syntax. `SetNestedAttribute` would require `alias = [{ name = "x" }]`, which is a breaking change. Plugin Framework permits a `CustomType` on `NestedBlockObject`, so we attach our alias custom type to the block's nested object. This matches the syntax users have today and gives us semantic equality for the element.

### 4. Custom type for `template.settings`

A new type `customtypes.IndexSettingsValue` is added under `internal/utils/customtypes/`. It:

- Embeds `jsontypes.Normalized`.
- Implements `basetypes.StringValuableWithSemanticEquals` whose comparator parses both strings as JSON, runs `flattenMap` (port from `internal/tfsdkutils/diffs.go`) to produce dotted keys, runs `normalizeIndexSettings` (port) to ensure all keys have the `index.` prefix and all values are stringified, and compares with `reflect.DeepEqual`.
- Implements `xattr.ValidateableAttribute` to enforce that the value parses to a JSON object (replaces SDK-side `stringIsJSONObject`).

The `tfsdkutils.DiffIndexSettingSuppress` helper remains in place for `component_template.go`. The flattening/normalization helpers are exported (or duplicated) into the customtype as needed; we keep the SDK copy intact rather than refactoring it.

### 5. JSON-normalized fields use `jsontypes.Normalized`

- `metadata`, `template.mappings`, `template.alias.filter` → `jsontypes.NormalizedType{}`.
- These types already provide `StringSemanticEquals` for JSON normalization.
- For `mappings` we additionally enforce JSON-object shape via `xattr.ValidateableAttribute` (either by extending `jsontypes.Normalized` in a small wrapper or by running validation in `ValidateConfig`); the choice is left to implementation.

### 6. Version gating remains in Create/Update

The two version gates (`ignore_missing_component_templates` ≥ 8.7.0; `template.data_stream_options` ≥ 9.1.0) require the live server version, which is not available at validate time. They stay in Create/Update, return `fwdiag` error diagnostics, and skip the API call on failure — identical observable behavior to the SDK implementation.

### 7. 8.x `allow_custom_routing` workaround in Update

The Plugin Framework equivalent of `d.GetChange("data_stream")` is reading `req.State` in the Update path. The Update method:

1. Loads prior state.
2. Loads plan.
3. If prior state had `data_stream.allow_custom_routing == true`, sets the request body's `allow_custom_routing` even when the planned value is `false`/null. This preserves the existing 8.x compatibility behavior.

### 8. Convert `MaxItems: 1` blocks to `SingleNestedBlock`

The SDK schema uses six `MaxItems: 1` collections that semantically model "an optional/required single object":

| Block path | SDK shape | PF target |
|---|---|---|
| `data_stream` | `TypeList`, Optional, MaxItems 1 | `SingleNestedBlock`, optional |
| `template` | `TypeList`, Optional, MaxItems 1 | `SingleNestedBlock`, optional |
| `template.lifecycle` | **`TypeSet`**, Optional, MaxItems 1 | `SingleNestedBlock`, optional |
| `template.data_stream_options` | `TypeList`, Optional, MaxItems 1 | `SingleNestedBlock`, optional |
| `template.data_stream_options.failure_store` | `TypeList`, **Required**, MaxItems 1 | `SingleNestedBlock`, required-when-DSO-present |
| `template.data_stream_options.failure_store.lifecycle` | `TypeList`, Optional, MaxItems 1 | `SingleNestedBlock`, optional |

`SingleNestedBlock` is the Plugin Framework equivalent of "exactly one object or none". Two consequences worth calling out:

- Plugin Framework `SingleNestedBlock` does not have a `Required` flag. The current `failure_store` Required semantic ("if `data_stream_options` is configured, `failure_store` is required") is enforced via `ValidateConfig` (or a config-level validator) returning an error diagnostic when `data_stream_options` is non-null and `failure_store` is null. This matches REQ-032's existing scenario "data_stream_options without failure_store rejected at plan time".
- HCL syntax `data_stream { hidden = true }` continues to work for `SingleNestedBlock`. There is **no breaking HCL change** for users.

The `template.alias` set remains a `SetNestedBlock` because it is a true many-element collection, not a `MaxItems: 1` collection. Decision 2's alias custom type is unaffected.

### 9. Identity, import, and state migration

- ID format: `<cluster_uuid>/<template_name>`. Computed via `client.ID(ctx, name)` in Create.
- Import: `resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)` — same passthrough behavior as the SDK's `schema.ImportStatePassthroughContext`.
- State schema version is **bumped from `0` to `1`**. The new resource implements `resource.ResourceWithUpgradeState` registering an upgrader for version `0`.

The upgrader rewrites the six single-item collections from list-/set-shaped to object-shaped. Because Plugin Framework `UpgradeState` does not require a v0 schema definition when using `RawState` JSON, the upgrader is implemented as a JSON-level transform:

```text
For each path P in:
  data_stream
  template
  template.0.lifecycle
  template.0.data_stream_options
  template.0.data_stream_options.0.failure_store
  template.0.data_stream_options.0.failure_store.0.lifecycle

Apply (pseudo-code, recursive after the parent has been collapsed):
  raw[P] is null      → unchanged (still null)
  raw[P] is []        → null
  raw[P] is [obj]     → obj
  raw[P] is otherwise → diagnostic error (corrupt prior state)
```

Order matters: collapse the parent (`template`, `data_stream_options`, `failure_store`) before walking into its children, so the path indices `…0.…` no longer apply by the time the inner level is rewritten. The implementation walks the JSON tree top-down to keep this invariant. Set-shaped values (`template.lifecycle` was `TypeSet` in the SDK) serialize identically to lists in tfstate, so the same `[]` / `[obj]` collapse rule applies; no Set-specific handling is needed.

The upgrader preserves all leaf attribute values byte-for-byte; only the surrounding container shape changes. Custom types (`jsontypes.Normalized` for JSON strings, `IndexSettingsValue`, alias element type) accept the same string/object payloads after the shape transform.

If the prior state is malformed (e.g. multi-element list at one of the listed paths), the upgrader returns an error diagnostic identifying the path. This should never occur because the SDK schema enforced `MaxItems: 1`, but failing loudly is preferable to silently dropping data.

`TestAccResourceIndexTemplateFromSDK` covers the upgrade end-to-end: pin the last SDK provider release, create a resource exercising every collapsed block, then re-apply with the new PF provider and assert no diff. Acceptance tests are the source of truth for state forward-compatibility.

### 10. Client diags strategy

`internal/clients/elasticsearch.PutIndexTemplate`, `GetIndexTemplate`, and `DeleteIndexTemplate` are migrated to return `fwdiag.Diagnostics`:

- Use `diagutil.CheckErrorFromFW()` for HTTP error responses.
- Use `diagutil.FrameworkDiagFromError()` or `fwdiag.NewErrorDiagnostic()` for non-HTTP errors.

These three functions have only one caller (`template.go`) plus one test reference (`template_test.go`), both of which move into the new PF package in the same change. No `SDKDiagsFromFramework` shim is needed.

### 11. Data source shares Read with the resource

A package-private `readIndexTemplate(ctx, client, name) (Model, fwdiag.Diagnostics)` helper is the single source of truth. The resource's `Read` and the data source's `Read` both call it. The data source schema is a Computed-only mirror of the resource schema; it is constructed in the same package to keep descriptions aligned.

### 12. Component template entanglement — duplicate, don't share

`expandTemplate`, `flattenTemplateData`, `extractAliasRoutingFromTemplateState`, `preserveAliasRoutingInFlattenedAliases`, `hashAliasByName`, and `stringIsJSONObject` in the SDK package are still used by `component_template.go`. Sharing them with the PF package would require those helpers to operate on both `map[string]any` (SDK) and typed PF model values, which is awkward.

Decision: duplicate the expand/flatten logic inside the PF `template/` package (operating on typed model values), and leave the SDK helpers untouched until component template is migrated separately. Acceptance: temporary duplication is bounded (≈ 200 LOC) and ends when component template lands.

## Risks / Trade-offs

- **Custom-type complexity for the alias element.** Implementing `ObjectValuableWithSemanticEquals` is more code than a simple plan modifier. The trade-off is correctness across plan/apply rather than just plan, plus it reduces supplemental Read-time workarounds. Mitigated by unit tests that exhaustively cover the semantic-equality matrix (null/empty/derived/explicit on both sides).
- **Index settings custom type risk.** The settings comparator must remain byte-equivalent to `DiffIndexSettingSuppress`. Mitigated by lifting the existing test cases from `internal/utils/utils_test.go` and re-running them against the new type.
- **Code duplication with component template.** Two parallel implementations of `expandTemplate`/`flattenTemplateData` exist until CT is migrated. Drift is possible but contained because the SDK copies are frozen except for bug fixes.
- **State migration correctness.** The v0→v1 upgrader rewrites six container shapes; a bug here can silently drop user data. Mitigated by (a) per-path explicit transformation logic with no fallthrough, (b) error diagnostics on unexpected shapes, (c) `TestAccResourceIndexTemplateFromSDK` exercising every collapsed block end-to-end against the last SDK release, and (d) targeted unit tests on the upgrader against synthesized v0 JSON state covering null, empty list, single-element list, and (error case) multi-element list per path.
- **Set semantic equality at scale.** Semantic equality is invoked pairwise during set comparison; the alias set is small (typically <10 elements) so this is not a performance concern.

## Migration Plan

1. Land the new client diag signatures and the new `IndexSettingsValue` custom type behind their existing callers.
2. Land the new PF `template/` package with full schema, models, custom types, and CRUD wired to Plugin Framework.
3. Switch provider registration from SDK to PF (atomic with step 2).
4. Move acceptance tests; add `TestAccResourceIndexTemplateFromSDK`.
5. Delete the SDK files (`template.go`, `template_data_source.go`).

The migration is delivered in a single change because step 3 is necessarily atomic.

## Open Questions

1. Should the `mappings` JSON-object enforcement live in the custom type (a thin wrapper around `jsontypes.Normalized`) or in `ValidateConfig`? Either matches today's behavior; resolved during implementation based on which produces clearer error messages.
2. Should the data source surface `not found` as an error diagnostic (as a data source typically does) rather than removing-from-state semantics inherited from the resource? Today's SDK data source delegates to the resource Read which clears the ID — i.e. silently succeeds with empty state. This is a latent UX bug; preserving today's behavior is the safe choice for this migration. Note for follow-up.
