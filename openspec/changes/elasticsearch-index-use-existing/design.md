## Context

`elasticstack_elasticsearch_index` currently has a single "create" path: build the API model from the plan and call the Create Index API. If the index already exists in Elasticsearch, the API returns `resource_already_exists_exception` and the apply fails.

Two production scenarios surface this:

1. **Replacement race ([#966](https://github.com/elastic/terraform-provider-elasticstack/issues/966))** — A change to a static setting (e.g. `mapping_coerce`) forces replacement. Terraform destroys the old index, but indexing load or `auto_create_index` template behavior can recreate the same name before the provider's create call lands. Maintainer triage on #966 noted there is "not much the provider can do" without an opt-in mechanism. `use_existing` is that mechanism.
2. **Adopt-without-import** — Practitioners want to manage an index that was created by another tool, a bootstrap job, or a matching index template, without taking it out of state with `terraform import` first.

Both reduce to the same primitive: an opt-in fallback that, on create, tolerates an already-existing index by reconciling the resource state against it.

The resource already has all the building blocks needed:

- `Get Index` helper that returns `nil` for 404.
- An update flow that reconciles aliases, dynamic settings, and mappings against any prior state, and that already understands template-injected mapping supersets.
- A `staticSettingsKeys` list defining which settings cannot be changed on an existing index.
- A composite `id` derived from cluster UUID + concrete name.

The design reuses these to keep the new code path small and predictable.

## Goals / Non-Goals

**Goals:**

- Provide an explicit, opt-in mechanism to tolerate an already-existing index at create time.
- Reuse the existing update-time reconciliation for aliases, dynamic settings, and mappings — including the existing template-aware mapping semantics.
- Fail loudly and refuse to mutate the cluster when adoption would silently lie about static settings.
- Be a no-op at the schema and behavioral level when `use_existing = false`, preserving full backward compatibility.
- Keep the surface area auditable: a single new attribute, a single new branch in `Create`, and a single new requirement in the spec.

**Non-Goals:**

- Tolerating "already exists" errors when `use_existing = false` (no implicit fallback).
- Closing the `Get Index → 404 → Create Index → 409` TOCTOU race that can still occur if the index appears between the existence check and the create call. The existing 409 error keeps surfacing in that narrow window.
- Adopt by date math expression, wildcard, or alias.
- A non-destructive-destroy / "release without delete" mode. Adoption implies full Terraform ownership; `deletion_protection` (already true by default) remains the safety net.
- Coercing static-setting values silently between config and the existing index.

## Decisions

### A new opt-in boolean attribute, no plan modifier

A single optional boolean `use_existing` is added to the schema with default `false`. It has no `RequiresReplace` plan modifier and no `UseStateForUnknown` modifier — flipping it after the resource exists is a no-op (the resource is already in state, so there is no "create" to short-circuit).

Alternative considered: making `use_existing` `ForceNew`. Rejected because that would force replacement (which can fail again for the same reasons that motivated the feature) when the user is simply correcting documentation of intent.

Alternative considered: making the create flow always tolerate `resource_already_exists_exception`. Rejected because adoption is a meaningful state change with surprising consequences; it must be opt-in and visible.

### Pre-flight existence check, not catch-on-409

The adopt path uses a `Get Index` call before the `Create Index` call. If the index exists, the create call is skipped entirely; if it does not, the provider falls through to the normal create path.

Alternative considered: try `Create Index` first and on `resource_already_exists_exception` fall back to adopt. This would close the TOCTOU window, but it produces uglier error paths (the provider has to distinguish 409s caused by adoption-eligible existence from 409s caused by other races) and the body of an Elasticsearch 409 is not always shaped consistently across versions. The pre-flight check is simpler and the residual race is acceptable for v1; it can be tightened in a follow-up by adding a 409 fallback layered on top of the pre-flight check.

### Adopt by reusing the update flow over a synthetic prior state

When the existence check succeeds, the create handler builds a synthetic `tfModel` from the Get Index API response (the same code path that powers `Read`) and treats it as the "prior state" passed to the existing update helpers. Concretely:

- Aliases are reconciled by the existing `updateAliases` helper, which deletes aliases that are in synthetic state but not in the plan and upserts aliases present in the plan. This keeps adoption symmetric with normal update behavior — adopted aliases not in config are removed.
- Dynamic settings are reconciled by the existing `updateSettings` helper, which only sends the diff and explicitly nulls dynamic settings present on the existing index but absent from config.
- Mappings are reconciled by the existing `updateMappings` helper, which already treats template-injected mapping supersets as semantically equal to user intent.

After reconciliation, the create handler computes `id` from the cluster UUID and the existing concrete index name, sets `concrete_name`, and performs the standard post-create read.

Alternative considered: write a fresh "merge" implementation that compares the plan against the API response inline. Rejected because it would duplicate behavior already encoded in the update helpers, including the template-injected mapping semantics. Reusing the update flow keeps adoption semantics tied to normal update semantics by construction.

### Static settings are strict — error, do not mutate

Before any reconciliation, the adopt path compares each static setting attribute that is *explicitly* set in config (i.e. `Known()` and not null) against the value reported by Get Index for the same key, with light normalization (string-or-typed equality, since Elasticsearch returns most settings as strings).

If any explicitly-configured static setting differs from the existing index, the provider:

1. Collects every mismatch into a single error diagnostic listing the attribute name, the configured value, and the actual value.
2. Returns immediately. No alias, setting, or mapping changes are applied.

Static setting attributes that are not explicitly set in config are not compared — the user has no opinion, so any value on the existing index is acceptable.

Alternative considered: warn and adopt anyway, persisting the existing index's values in state. Rejected because it produces permanent silent drift between config and state for static fields, surprising users when they re-read their own config months later.

Alternative considered: delete the existing index and recreate. Rejected because adoption may target indices with real data; destructive mutation must never be implicit.

The strict list mirrors the existing `staticSettingsKeys` slice in `models.go`: `number_of_shards`, `number_of_routing_shards`, `codec`, `routing_partition_size`, `load_fixed_bitset_filters_eagerly`, `shard.check_on_startup`, `sort.field`, `sort.order`, `mapping.coerce`. Analysis settings (`analysis_*`) are also create-only at the API level but are not in `staticSettingsKeys` and are not mass-comparable through the existing settings map; they are out of scope for the strict comparison and are simply not sent on adopt (mirroring the current update flow which also does not send them).

### Date math interaction — runtime skip with a warning

When `use_existing = true` and the configured `name` matches `DateMathIndexNameRe`, the existence check is skipped, a warning diagnostic is emitted explaining that `use_existing` does not apply to date math names, and the create proceeds along the normal path.

Alternative considered: a plan-time configuration validator that errors on `use_existing = true` + date-math `name`. Rejected because the user explicitly requested the more lenient runtime-skip behavior, which keeps configurations portable when `name` switches between static and date math values across environments.

### Adoption is observable via a warning diagnostic

Whenever the adopt branch executes (the existence check returns the index), the provider adds a warning diagnostic on the resource indicating that the index was adopted rather than created, and including the concrete name. This makes adoption visible in `terraform apply` output without raising an error and without requiring users to dig into trace logs.

### Destroy semantics — full ownership after adopt

After adoption the resource state is identical to a freshly-created resource: same `id`, same `concrete_name`, same fields populated. `terraform destroy` calls the Delete Index API, gated by `deletion_protection` (default `true`). No new attribute or knob is introduced for adoption-specific destroy behavior.

Alternative considered: a `delete_on_destroy` flag so adopted indices can be released without deletion. Rejected for v1 because it complicates state semantics and `deletion_protection = true` (the default) already protects accidental deletion. The flag can be added later if there is demand.

## Risks / Trade-offs

- [Risk] **TOCTOU race between existence check and create.** If the index appears between Get Index returning 404 and Create Index returning, the existing 409 still surfaces. Mitigation: explicitly documented as a known limitation; can be closed in a follow-up by chaining a 409 fallback after the pre-flight check.
- [Risk] **Static-setting comparison may be overly strict for value normalization edge cases** (e.g. `"1"` vs `1`, `null` vs absent, ES-server-side defaults like `index.number_of_shards = "1"` filling in unspecified config values). Mitigation: normalize to the same JSON-ish representation used by the existing settings round-trip (`map[string]any`), include focused unit tests for each static setting type.
- [Risk] **Adoption deletes aliases on the existing index that aren't in config** (because it reuses the update flow's symmetric semantics). Mitigation: this matches normal update semantics, is documented in the design, and is the principle-of-least-surprise choice — users who want to keep an alias must declare it.
- [Risk] **Users may expect `use_existing` to also adopt during failed `Update` runs**, e.g. when an update somehow targets a missing index. Mitigation: scope is explicitly create-time-only; document the limitation in the attribute description.
- [Risk] **Operators may flip `use_existing` to `true` on already-managed resources thinking it changes runtime behavior.** Mitigation: the attribute is no-op after create (the resource is already in state), and the design accepts this. Documentation should call this out.
- [Risk] **Static-setting strictness can block legitimate adoption** of an index with extra non-configured settings the user did not anticipate. Mitigation: the strict comparison only fires for settings the user explicitly set, not for any static setting that happens to be set on the existing index. Users can drop the offending attribute from config to adopt and let the existing value pass through.

## Migration Plan

1. Add the schema attribute, its documentation, and the model field; rebuild docs.
2. Add the existence-check + adopt branch in `Create`, reusing the existing update helpers behind a small refactor that exposes them as package-private functions taking the relevant arguments.
3. Add the static-setting strict comparison helper and wire it into the adopt branch ahead of any update calls.
4. Add the date-math runtime-skip + warning logic at the top of the `use_existing = true` path.
5. Add unit tests for the static-setting comparison helper across all `staticSettingsKeys`.
6. Add acceptance tests:
   - adopt an empty index (template-injected only) and verify a clean apply with the warning;
   - adopt with config matching existing → success, idempotent second apply;
   - adopt with a static-setting mismatch → error diagnostic, no mutation;
   - `use_existing = true` + date math name → warning + normal create.
7. Add a `CHANGELOG.md` entry referencing #966.

## Open Questions

- None. The TOCTOU window after the pre-flight existence check is intentionally left open in v1, with a clear path to close it in a follow-up if needed.

## Testing: `use_existing` create-branch coverage

Acceptance tests exercise the full Create branching semantics end-to-end: `TestAccResourceIndexUseExistingDateMath` (date-math gate + normal create), `TestAccResourceIndexUseExistingFallthrough` (404 fall-through to Create Index), `TestAccResourceIndexUseExistingMismatch` (static mismatch + no cluster mutation), `TestAccResourceIndexUseExistingAdopt` / `TestAccResourceIndexUseExistingAdoptAliasReconcile` / `TestAccResourceIndexUseExistingTemplateNoMappingDrift` (successful adoption, symmetric aliases, template mapping superset tolerance). Unit tests cover extractable helpers (`compareStaticSettings`, `formatStaticSettingMismatchesDetail`, the date-math regex gate, and `populateFromAPI` synthetic state). Direct unit tests of each `Create` branch would need an injectable Elasticsearch client in this package; that refactor is intentionally deferred.
