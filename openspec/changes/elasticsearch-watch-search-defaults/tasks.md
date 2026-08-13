> **Implementation scope note:** This change covers Watcher-injected defaults on JSON attributes that round-trip through Get Watch: search-request `rest_total_hits_as_int`, `search_type`, and `indices` (including `chain` sub-inputs and action-level search transforms) and script `lang: "painless"` on `condition` / `transform` / `actions`. `indices_options` remains out of scope.

## 1. Spec

- [x] 1.1 Keep the delta spec aligned with `proposal.md` / `design.md`; run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate elasticsearch-watch-search-defaults --type change` (or `make check-openspec` after sync).
- [x] 1.2 On completion of implementation, **sync** the delta into `openspec/specs/elasticsearch-watch/spec.md` or **archive** the change per project workflow.

## 2. Implementation

- [x] 2.1 Add `internal/elasticsearch/watcher/watch/json_defaults.go` with `populateWatcherJSONDefaults(model map[string]any) map[string]any`: a copy-on-write recursive walker over maps **and** arrays that:
  - for every object with a `"search"` key whose value has a `"request"` object, fills absent `rest_total_hits_as_int: true`, `search_type: "query_then_fetch"`, and `indices: []` on that `request` only (never `http.request`);
  - for every object with a `"script"` key whose value is an object, fills absent `lang: "painless"`;
  - continues after a match so nested `chain.inputs[]` search sub-inputs and action-level search transforms are each visited;
  - does not mutate the input tree.
- [x] 2.2 In `internal/elasticsearch/watcher/watch/schema.go`, change the `input`, `transform`, `condition`, and `actions` attributes' `CustomType` from `jsontypes.NormalizedType{}` to `customtypes.NewJSONWithDefaultsType(populateWatcherJSONDefaults)`. Leave `trigger` and `metadata` as `jsontypes.NormalizedType{}`. Use the **same** populate function for all four attributes (`JSONWithDefaultsType.Equal` ignores the populate func).
- [x] 2.3 In `internal/elasticsearch/watcher/watch/models.go`, change `Data.Input`, `Data.Transform`, `Data.Condition`, and `Data.Actions` from `jsontypes.Normalized` to `customtypes.JSONWithDefaultsValue[map[string]any]`. Update all construction sites in `fromAPIModel` (`jsontypes.NewNormalizedValue(...)` → `customtypes.NewJSONWithDefaultsValue(..., populateWatcherJSONDefaults)`; null → `customtypes.NewJSONWithDefaultsNull(populateWatcherJSONDefaults)`) for those four fields only.
- [x] 2.4 Confirm `toPutModel` continues to work unchanged: `JSONWithDefaultsValue[map[string]any]` embeds `jsontypes.Normalized`, so `.ValueString()` still returns the literal JSON for Put Watch. Update static type references (including `fromAPIModel`'s `priorActions` / `priorInput` parameters) as needed to compile.
- [x] 2.5 Verify `mergePreserveRedactedLeaves` (input) and `mergeActionsPreservingRedactedLeaves` (actions) in `fromAPIModel` still run exactly as today **before** the merged map is marshaled and wrapped in the custom type. Do not strip injected defaults inside `fromAPIModel`; Framework `StringSemanticEquals` is what preserves authored JSON in Terraform state when equality holds.

## 3. Testing

- [x] 3.1 Add unit tests in `internal/elasticsearch/watcher/watch/json_defaults_test.go` for `populateWatcherJSONDefaults`: top-level `search.request` missing all three search-request keys; missing one key; both search-request keys and `indices` already present and explicit (including `rest_total_hits_as_int: false` and non-empty `indices`); a `chain` input whose `inputs` **array** has multiple named `search` sub-inputs each defaulted independently; an `http` input whose `request` is left untouched; `transform.search.request`; `actions` containing a nested search transform; a `script` object omitting `lang`; copy-on-write (input map unchanged after the call).
- [x] 3.2 Add a unit test exercising `StringSemanticEquals` end-to-end on `customtypes.JSONWithDefaultsValue[map[string]any]` wired with `populateWatcherJSONDefaults`: a prior value omitting the search-request keys (including `indices`) and a new value with Elasticsearch's injected defaults SHALL compare semantically equal; a new value with a genuinely different `search_type` SHALL compare semantically unequal; `::es_redacted::` vs a concrete secret SHALL compare semantically unequal.
- [x] 3.3 Add an acceptance test reproducing the issue's search-`input` case against `.monitoring-es-*` omitting `rest_total_hits_as_int` and `search_type`, asserting apply succeeds and a subsequent plan is empty (`plancheck.ExpectEmptyPlan` or equivalent). Mirror `internal/elasticsearch/watcher/watch/acc_test.go`.
- [x] 3.4 Extend acceptance coverage to: a search `input` that also omits `indices`; a `transform` search request omitting the same defaults; a `chain` input with multiple search sub-inputs; an `actions` search transform omitting the defaults; a `condition` or `transform` script omitting `lang`. Each SHALL apply cleanly with an empty subsequent plan.
- [x] 3.5 Add a test asserting explicit non-default values (`rest_total_hits_as_int = false`, `search_type = "dfs_query_then_fetch"`, non-empty `indices`) round-trip unchanged and a genuine change to any of them still produces a non-empty plan.

## 4. Documentation

- [x] 4.1 If resource documentation (`docs/resources/elasticsearch_watch.md` or embedded schema descriptions) references `input`/`transform`/`condition`/`actions` JSON behavior, note that Elasticsearch-injected search-request defaults (`rest_total_hits_as_int`, `search_type`, `indices`) and script `lang` do not need to be set explicitly and will not cause spurious diffs.
