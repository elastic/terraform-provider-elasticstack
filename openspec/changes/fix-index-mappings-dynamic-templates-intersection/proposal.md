## Why

`elasticstack_elasticsearch_index_mappings`'s `mappings` attribute is documented (REQ-004) to keep only the user-declared subset of the API response on read — dynamic extras added by Elasticsearch are supposed to be discarded. This holds for `properties`, which is intersected by field name via `intersectProperties`, but not for `dynamic_templates`. `intersectMappings` (`internal/elasticsearch/index/indexmappings/intersect.go`) falls back to `index.FieldSemanticallyEqual` for every top-level key other than `properties`. That helper type-asserts both sides to `map[string]any`, but `dynamic_templates` is a `[]any`, so the assertion always fails and the entire API array — including index-template-injected entries — is written into state. This contradicts the documented contract and isn't caught by the existing acceptance test, which only asserts a minimum template count.

## What Changes

- Move `intersect.go` (and its test file) from `internal/elasticsearch/index/indexmappings` into `internal/elasticsearch/index`, consolidating with `MappingsValue` and `dynamicTemplatesByName` so the fix reuses that helper in-package instead of exporting new public API surface. `intersectMappings` becomes exported `IntersectMappings`; the `indexmappings/read.go` call site is updated to `index.IntersectMappings(apiMap, stateMap)`.
- Add a `dynamic_templates` branch to the (now relocated) intersection logic that filters the API's `dynamic_templates` array down to the template names declared in state, preserving the state's declared order, and reusing `dynamicTemplatesByName` for name-keyed lookup.
- A declared template name absent from the API (deleted out-of-band) is dropped from state, mirroring `intersectProperties`'s existing omit-on-missing behavior; drift then surfaces on the next plan. This includes the case where the API omits the `dynamic_templates` key entirely (e.g. every declared template was removed out-of-band) — the generic "API key absent → keep declared value" fallback used for other top-level keys does not apply to `dynamic_templates`.
- If either side's `dynamic_templates` array doesn't parse via `dynamicTemplatesByName` (duplicate template name, or a template value that isn't an object — only reachable via manual state edits/import), fall back to today's whole-array passthrough for that key rather than dropping it.
- Extend `dynamicTemplatesSemanticallyEqual` (`mappings_value.go`) to also compare the relative order of the user's declared template names against the API's array (ignoring API-only extras), so that a real reorder of declared templates in Elasticsearch is not masked as "no diff" now that `Read` reconstructs the state's declared order.
- Tighten `checkStateMappingsDynamicTemplates` in `acc_test.go` from a `len >= minCount` assertion to an exact-name-set assertion that also asserts `template_default` is absent, update both existing call sites (`TestAccResourceIndexMappings_allTopLevelKeys` and `TestAccResourceIndexMappings_dynamicTemplatesFromIndexTemplate`), and extend `TestAccResourceIndexMappings_dynamicTemplatesFromIndexTemplate` with a step exercising the drop-on-missing behavior (a declared template name deleted out-of-band).
- Update `openspec/specs/elasticstack-elasticsearch-index-mappings/spec.md` REQ-004 to document that `dynamic_templates` receives the same name-keyed filtering treatment as `properties`.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `elasticstack-elasticsearch-index-mappings`: extend REQ-004 ("Read — user-declared subset only") so the user-declared-subset guarantee explicitly covers `dynamic_templates`, filtered by template name with drop-on-missing and passthrough-on-unparseable-shape semantics.

## Impact

- `internal/elasticsearch/index/indexmappings/intersect.go` and `intersect_test.go` — relocated into `internal/elasticsearch/index` (new file names/paths under that package); `intersectMappings` exported as `IntersectMappings`.
- `internal/elasticsearch/index/indexmappings/read.go` — call site updated to `index.IntersectMappings`.
- `internal/elasticsearch/index/mappings_value.go` — `dynamicTemplatesByName` gains an in-package caller from the relocated intersection logic; `dynamicTemplatesSemanticallyEqual` gains an order check on the shared-name subsequence (see above).
- `internal/elasticsearch/index/indexmappings/acc_test.go` — `checkStateMappingsDynamicTemplates` tightened to exact-name-set plus absence assertion, with both call sites (`TestAccResourceIndexMappings_allTopLevelKeys`, `TestAccResourceIndexMappings_dynamicTemplatesFromIndexTemplate`) updated; the latter also gains a drop-on-missing step.
- `openspec/specs/elasticstack-elasticsearch-index-mappings/spec.md` — REQ-004 wording extended for `dynamic_templates`.
- The rest of the plan/apply consistency path (`MappingsSemanticallyEqual`'s overall shape, `StringSemanticEquals`, `RequiresMappingsUpdate`) is unchanged from #4581 and out of scope here; only `dynamicTemplatesSemanticallyEqual`'s order sensitivity is added.
