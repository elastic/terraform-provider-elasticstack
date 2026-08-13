## Context

`elasticstack_elasticsearch_watch` maps `input`, `transform`, `condition`, and `actions` as opaque JSON strings (`jsontypes.Normalized`) rather than typed HCL blocks (Watcher payloads are polymorphic: `search`, `http`, `simple`, `chain`, `script`, etc.). `jsontypes.Normalized.StringSemanticEquals` treats two JSON strings as equal when they differ only in key order or insignificant whitespace. It does **not** tolerate extra keys on one side.

Elasticsearch's `WatcherSearchTemplateRequest` (backing every Watcher search request) defaults `restTotalHitsAsInt` to `true` and `searchType` to `query_then_fetch`, and **always serializes** both on Get Watch. Parse also starts `indices` as an empty list; `toXContent` writes `indices` whenever the array is non-null, so an omitted `indices` comes back as `"indices":[]`. The existing transform acceptance fixture already hardcodes all three (`indices:[]`, `rest_total_hits_as_int:true`, `search_type:query_then_fetch`) for that reason.

Those search-request keys are not limited to top-level `input.search`. The same request object appears on `transform.search`, on each `chain` sub-input (`chain.inputs` is an **array** of `{ "<name>": { "search": { "request": ... } } }`), and on search transforms nested under `actions`. Independently, Watcher `Script` defaults omitted `lang` to `"painless"` and re-emits it for `condition`, `transform`, and action scripts. Leaving `lang` out of this change would leave apply failures on `transform` (an attribute whose type we are already changing) and on `condition`/`actions`.

Because `fromAPIModel` marshals the Get Watch payload compactly into those attributes as-is, plan and post-apply state differ by the injected keys and apply fails with `Provider produced inconsistent result after apply`.

The provider already solves this class of problem with `internal/utils/customtypes.JSONWithDefaultsType[TModel]` / `JSONWithDefaultsValue[TModel]`: on `StringSemanticEquals` it unmarshals both values, applies a `PopulateDefaultsFunc`, re-marshals, then defers to `jsontypes.Normalized`. After Create, Update, **and Read**, the Plugin Framework runs `SchemaSemanticEquality`; if equality holds it **replaces new state with the prior/plan value** before Core's consistency check. The same equality also runs during **plan** to suppress drift. This is already used by `internal/elasticsearch/ml/datafeed` (`script_fields`) and Kibana Lens JSON attributes.

`fromAPIModel` does **not** strip injected keys. REQ-023 still marshals the API (redaction-merged) payload into the resource model. Practitioner JSON is preserved in Terraform state only when Framework substitution keeps the prior/plan value.

## Goals

- Make `terraform apply` succeed, and the following `terraform plan` empty, when search requests omit `rest_total_hits_as_int`, `search_type`, and/or `indices`, including nested `chain` sub-inputs and action-level search transforms.
- Make apply and the following plan succeed when Watcher scripts omit `lang` on `condition`, `transform`, or `actions`.
- Reuse `customtypes.JSONWithDefaultsType`/`JSONWithDefaultsValue` rather than adding a new custom type.
- Preserve practitioner-authored JSON in Terraform state via Framework semantic-equality substitution, not by rewriting JSON in `fromAPIModel`.
- Keep `mergePreserveRedactedLeaves` / `mergeActionsPreservingRedactedLeaves` intact and running **before** the custom type is constructed. Semantic equality MUST NOT treat `::es_redacted::` as equal to a concrete secret.

## Non-Goals

- `indices_options` defaulting. Elasticsearch only serializes `indices_options` when it differs from default, so it is not a reported drift source.
- Converting JSON string attributes to typed nested HCL blocks.
- Changing redaction-preservation behavior for `input` or `actions`.
- Changing `JSONWithDefaultsType.Equal` (it compares only the embedded `NormalizedType` and ignores the populate func). All watch attributes that use `JSONWithDefaultsType[map[string]any]` MUST share one populate function so that limitation is harmless.
- `trigger` and `metadata` (no known Get Watch injections of this class).

## Decisions

| Topic | Decision |
|-------|----------|
| Custom type | Reuse `customtypes.JSONWithDefaultsType[map[string]any]` / `JSONWithDefaultsValue[map[string]any]` for `input`, `transform`, `condition`, and `actions`. `TModel = map[string]any` because these attributes are polymorphic JSON and redaction merge already operates on `map[string]any`. |
| Attributes | Wire the same populate function on `input`, `transform`, `condition`, and `actions`. `trigger` and `metadata` remain `jsontypes.Normalized`. |
| Defaults function | One shared `populateWatcherJSONDefaults(model map[string]any) map[string]any`. Because `JSONWithDefaultsType.Equal` ignores the populate func, these four attributes MUST NOT use different populate functions. |
| Copy-on-write | The walker SHALL NOT mutate its input map. It copies maps (and slices, when descending into arrays) before inserting default keys. `JSONWithDefaultsValue.WithDefaults` unmarshals a fresh tree today, but copy-on-write keeps the function safe if called on a live model. (Datafeed's `populateScriptFieldsDefaults` mutates in place; do not copy that pattern here.) |
| Walk strategy | Recurse into **both** JSON objects (`map[string]any`) **and** JSON arrays (`[]any`). Watcher `chain.inputs` is an array of named wrapper objects; walking only maps misses chain sub-inputs. Continue after a match so every nested search request and script is visited. |
| Search-request defaults | When an object has a `"search"` key whose value is an object containing a `"request"` object, copy that `request` and fill absent keys only: `rest_total_hits_as_int: true`, `search_type: "query_then_fetch"`, `indices: []`. Do **not** inject these keys into `http.request` or any other `request` object. |
| Script `lang` | When an object has a `"script"` key whose value is an object, copy that object and fill `lang: "painless"` if `lang` is absent. Do not treat arbitrary objects that happen to have `source` or `id` as scripts (`id` is too common). |
| Absent vs explicit | Only fill keys that are **absent**. Explicit `rest_total_hits_as_int: false`, `search_type: "dfs_query_then_fetch"`, non-empty `indices`, or non-painless `lang` are left untouched so real drift still compares unequal. |
| Ordering vs redaction | `fromAPIModel` still runs redaction merge, marshals compact JSON, then wraps with `customtypes.NewJSONWithDefaultsValue(json, populateWatcherJSONDefaults)`. Default-population is not part of `fromAPIModel`'s data flow. |
| What is stored | `fromAPIModel` still emits API (redaction-merged) JSON into the resource model (REQ-023). After Create/Update/Read, if `StringSemanticEquals` is true, the Framework replaces new state with the **prior** value (plan on apply, prior state on refresh). That is how authored JSON is preserved. On **import**, there is no plan: first state is the API JSON including injected keys; a later plan against config that omits those keys SHALL still be empty because equality fills defaults on the config side. |
| `toPutModel` | Unaffected in substance: `JSONWithDefaultsValue` embeds `jsontypes.Normalized`, so `.ValueString()` still returns the literal JSON sent to Put Watch. |
| When the walker runs | `StringSemanticEquals` (and thus the walk) runs on create, update, read, **and plan**. Watch JSON is small; cost is negligible. |

## Risks / Trade-offs

- **Over-broad `search.request` matching**: user-authored JSON that happens to contain a `search` key wrapping a `request` object would get search-request defaults for comparison only. Accepted: keys are never written into state by the walker, and both sides of the comparison are populated the same way. This is a structural tree walk, not the same as datafeed's `script_fields` helper, which only iterates a top-level map of named fields.
- **Over-broad `script` matching**: a nested object keyed `script` that is not a Watcher Script would get `lang` for comparison only. Same acceptance: comparison-only, both sides.
- **Chain coverage is structural**: the walker recognizes `{ "search": { "request": ... } }` at any depth, including under `chain.inputs[]` name wrappers, without validating the full chain schema.
- **`JSONWithDefaultsType.Equal`**: two `JSONWithDefaultsType[map[string]any]` values compare equal even if their populate funcs differ. Mitigated by using one shared function for every watch attribute that adopts the type.

## Open Questions

1. **`indices_options`**: still not a reported drift source (only serialized when non-default). Revisit if a future issue reports it.
2. **Script shorthand**: if a practitioner sets `"script": "return true"` (string) and Get Watch returns `{ "script": { "source": "return true", "lang": "painless" } }`, that is a shape change this walker does not equalize. Out of scope unless reported.

## Migration / State

- No state schema/version change: the four attributes remain string-typed at the wire (`JSONWithDefaultsType` embeds `NormalizedType`). Existing state decodes via `ValueFromTerraform`, which attaches the schema's populate func. No state upgrader.
- Watches that already set the defaults explicitly continue to work (explicit values are never overwritten). Watches that omit them, which previously failed apply, now succeed.
- Imported watches store API JSON (with injected keys) until the next apply that keeps plan via semantic equality; `terraform plan` after import SHALL be empty when config omits only those defaults.
