## Why

`elasticstack_elasticsearch_watch` fails apply with `Provider produced inconsistent result after apply` whenever an `input` (or `transform`) `search` request omits `rest_total_hits_as_int` and/or `search_type` ([issue #4522](https://github.com/elastic/terraform-provider-elasticstack/issues/4522)). Elasticsearch's `WatcherSearchTemplateRequest` defaults `rest_total_hits_as_int` to `true` and `search_type` to `query_then_fetch`, and **always serializes** both on Get Watch — even when the practitioner's configuration omitted them. The resource stores the Get Watch response JSON as-is in `input`/`transform`, and `jsontypes.Normalized`'s `StringSemanticEquals` only tolerates key-order/whitespace differences, not API-injected keys. The extra keys therefore register as a real change between plan and the post-apply state, breaking `terraform apply` for any watch whose search input relies on these Elasticsearch defaults.

This reproduces the same category of problem the provider already solves for `transform` via hand-baked test fixtures (`internal/elasticsearch/watcher/watch/acc_test.go`) and for other JSON attributes elsewhere in the provider (`customtypes.JSONWithDefaultsType`, used by `internal/elasticsearch/ml/datafeed` and the Kibana Lens dashboard panels) — but `input` was never covered, and the existing `transform` test fixtures work around the symptom rather than fixing the underlying semantic-equality gap.

## What Changes

- Give `input` and `transform` on `elasticstack_elasticsearch_watch` apply-time semantic equality that accounts for Elasticsearch's Watcher search-request defaults, using the existing `customtypes.JSONWithDefaultsType[TModel]` / `JSONWithDefaultsValue[TModel]` machinery (same pattern as `internal/elasticsearch/ml/datafeed`'s `script_fields` and the Lens dashboard JSON attributes) rather than inventing a new custom type.
- Add a shared `populateWatcherSearchDefaults` function operating on `map[string]any` that walks a parsed `input` or `transform` value and, at every nested object containing a `search.request` object (including `request` objects nested inside `chain` input sub-inputs), fills in:
  - `rest_total_hits_as_int: true` when absent
  - `search_type: "query_then_fetch"` when absent
- Wire this function as the `PopulateDefaultsFunc` for both the `input` and `transform` schema attributes (`CustomType: customtypes.NewJSONWithDefaultsType(populateWatcherSearchDefaults)`), and change `Data.Input` / `Data.Transform` from `jsontypes.Normalized` to `customtypes.JSONWithDefaultsValue[map[string]any]`.
- Leave the existing redaction-preservation merge (`mergePreserveRedactedLeaves`, `fromAPIModel`) untouched: that logic still runs first to reconcile `::es_redacted::` sentinels against prior state/plan, and produces the JSON string that is then wrapped in the new custom type. Semantic equality (default-population) is a distinct, later comparison step performed by the Plugin Framework, not a data-mutation step — the practitioner's authored JSON is never rewritten in state.
- Script `lang` defaulting (Elasticsearch also re-emits `"lang": "painless"` when a script omits it, per the issue's "Other susceptible surfaces" table) is **out of scope** for this change; it is a separate, lower-severity drift source not exercised by the reported bug's repro case. See `design.md` for the deferral rationale.

## Capabilities

### New Capabilities

- _(none)_

### Modified Capabilities

- `elasticsearch-watch`: Add semantic-equality handling for Watcher-injected search-request defaults (`rest_total_hits_as_int`, `search_type`) on the `input` and `transform` attributes, including nested `chain` search sub-inputs (REQ-031).

## Impact

- **Specs**: Delta under `openspec/changes/elasticsearch-watch-search-defaults/specs/elasticsearch-watch/spec.md` until synced into the canonical spec.
- **Implementation** (future): `internal/elasticsearch/watcher/watch/schema.go` (CustomType for `input`/`transform`), `internal/elasticsearch/watcher/watch/models.go` (`Data.Input`/`Data.Transform` field types, construction in `fromAPIModel`), a new `internal/elasticsearch/watcher/watch/search_defaults.go` (the `populateWatcherSearchDefaults` walker), unit tests, and an acceptance test reproducing the issue's repro case (search input omitting `rest_total_hits_as_int`/`search_type`, including a `chain` variant).
