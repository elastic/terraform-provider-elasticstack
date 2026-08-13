> **Implementation scope note:** This change covers `rest_total_hits_as_int` and `search_type` defaulting on Watcher `search` requests within `input` and `transform` (including nested `chain` sub-inputs). Script `lang: "painless"` re-emission (noted in the issue's "Other susceptible surfaces" table) is deferred — see `design.md` Open Questions — and is tracked as unchecked below for visibility only; it is not required for this change to be considered complete.

## 1. Spec

- [ ] 1.1 Keep the delta spec aligned with `proposal.md` / `design.md`; run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate elasticsearch-watch-search-defaults --type change` (or `make check-openspec` after sync).
- [ ] 1.2 On completion of implementation, **sync** the delta into `openspec/specs/elasticsearch-watch/spec.md` or **archive** the change per project workflow.

## 2. Implementation

- [ ] 2.1 Add `internal/elasticsearch/watcher/watch/search_defaults.go` with `populateWatcherSearchDefaults(model map[string]any) map[string]any`: a recursive walker that, for every object with a `"search"` key whose value has a `"request"` object, fills `rest_total_hits_as_int: true` and `search_type: "query_then_fetch"` when absent, without overwriting present keys. The walk SHALL continue past a match so multiple/nested `chain` sub-inputs are each visited.
- [ ] 2.2 In `internal/elasticsearch/watcher/watch/schema.go`, change the `input` and `transform` attributes' `CustomType` from `jsontypes.NormalizedType{}` to `customtypes.NewJSONWithDefaultsType(populateWatcherSearchDefaults)`.
- [ ] 2.3 In `internal/elasticsearch/watcher/watch/models.go`, change `Data.Input` and `Data.Transform` from `jsontypes.Normalized` to `customtypes.JSONWithDefaultsValue[map[string]any]`. Update all construction sites in `fromAPIModel` (`jsontypes.NewNormalizedValue(...)` → `customtypes.NewJSONWithDefaultsValue(..., populateWatcherSearchDefaults)`; `jsontypes.NewNormalizedNull()` → `customtypes.NewJSONWithDefaultsNull(populateWatcherSearchDefaults)`) for the `input` and `transform` fields only — `trigger`, `condition`, `actions`, and `metadata` are unaffected and remain `jsontypes.Normalized`.
- [ ] 2.4 Confirm `toPutModel` continues to work unchanged: `customtypes.JSONWithDefaultsValue[map[string]any]` embeds `jsontypes.Normalized`, so `.ValueString()` still returns the literal configured/state JSON for `json.Unmarshal` into the Put Watch body. Update the field's static type references (e.g. any explicit `jsontypes.Normalized` parameter/return types touching `Input`/`Transform`) as needed to compile.
- [ ] 2.5 Verify `mergePreserveRedactedLeaves` (input) and `mergeActionsPreservingRedactedLeaves` (actions) in `fromAPIModel` continue to run exactly as today, before the merged map is marshaled and wrapped in the new custom type — no change to redaction-preservation ordering or behavior.

## 3. Testing

- [ ] 3.1 Add unit tests in `internal/elasticsearch/watcher/watch/search_defaults_test.go` for `populateWatcherSearchDefaults`: top-level `search.request` missing both keys; missing one key; both keys already present and explicit (including `rest_total_hits_as_int: false`); a `chain` input with multiple `search` sub-inputs each defaulted independently; a non-`search` input (e.g. `http`) left untouched; `transform` with a `search.request`.
- [ ] 3.2 Add or extend a unit test exercising `StringSemanticEquals` end-to-end on `customtypes.JSONWithDefaultsValue[map[string]any]` wired with `populateWatcherSearchDefaults` for `input`: a "prior" value omitting the two keys and a "new" value with Elasticsearch's injected defaults SHALL compare semantically equal; a "new" value with a genuinely different `search_type` SHALL compare semantically unequal.
- [ ] 3.3 Add an acceptance test reproducing the issue's repro case: a watch `input` search request against `.monitoring-es-*` omitting `rest_total_hits_as_int` and `search_type`, asserting `terraform apply` succeeds and a subsequent `terraform plan` is empty (`plancheck.ExpectEmptyPlan` or equivalent). Mirror the existing acceptance test structure in `internal/elasticsearch/watcher/watch/acc_test.go`.
- [ ] 3.4 Extend acceptance/unit test coverage to a `transform` search request omitting the same defaults, and to a `chain` input with multiple search sub-inputs, following the same empty-subsequent-plan assertion pattern.
- [ ] 3.5 Add a test asserting explicit non-default values (`rest_total_hits_as_int = false`, `search_type = "dfs_query_then_fetch"`) round-trip unchanged and a genuine change to either value still produces a non-empty plan.

## 4. Documentation

- [ ] 4.1 If resource documentation (`docs/resources/elasticsearch_watch.md` or embedded schema descriptions) references `input`/`transform` JSON behavior, note that Elasticsearch-injected `search`-request defaults (`rest_total_hits_as_int`, `search_type`) do not need to be set explicitly and will not cause spurious diffs.

## Deferred (script `lang` defaulting)

- [ ] 4.2d Extend or add a sibling `PopulateDefaultsFunc` to also default script `lang: "painless"` on condition/transform/action script bodies that omit it, mirroring `internal/elasticsearch/ml/datafeed`'s existing script-lang defaulting. Out of scope for this change; see `design.md` Open Questions.
