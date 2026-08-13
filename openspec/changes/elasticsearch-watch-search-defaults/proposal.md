## Why

`elasticstack_elasticsearch_watch` fails apply with `Provider produced inconsistent result after apply` whenever a Watcher **search request** omits fields that Elasticsearch's `WatcherSearchTemplateRequest` always re-serializes on Get Watch ([issue #4522](https://github.com/elastic/terraform-provider-elasticstack/issues/4522)). The reported case was `input.search.request` omitting `rest_total_hits_as_int`; Get Watch returned the same request with `"rest_total_hits_as_int": true`. The resource stores that Get Watch JSON as-is, and `jsontypes.Normalized`'s `StringSemanticEquals` only tolerates key-order/whitespace differences, not API-injected keys.

The same serializer always writes, when omitted at parse time:

- `rest_total_hits_as_int: true`
- `search_type: "query_then_fetch"`
- `indices: []` (parse starts from an empty list; `toXContent` writes the array whenever it is non-null)

Those keys appear on every Watcher search request, not only top-level `input.search`: `transform.search`, `chain` sub-inputs (`chain.inputs[]` → named wrapper → `search.request`), and search transforms nested under `actions`. Independently, Watcher `Script` defaults omitted `lang` to `"painless"` and re-emits it on Get Watch for `condition`, `transform`, and action scripts.

Existing transform acceptance fixtures work around the symptom by hardcoding `indices`, `rest_total_hits_as_int`, and `search_type` (`internal/elasticsearch/watcher/watch/acc_test.go`). That does not fix apply for practitioners who omit the defaults.

## What Changes

- Give `input`, `transform`, `condition`, and `actions` apply-time semantic equality for these Elasticsearch-injected defaults, using existing `customtypes.JSONWithDefaultsType[TModel]` / `JSONWithDefaultsValue[TModel]` rather than a new custom type or nested HCL blocks.
- Add one shared `populateWatcherJSONDefaults(model map[string]any) map[string]any` that **copy-on-write** walks maps **and arrays** and:
  - For every `"search"` key whose value is an object containing a `"request"` object, fills absent `rest_total_hits_as_int: true`, `search_type: "query_then_fetch"`, and `indices: []` on that `request` only (never on `http.request`).
  - For every `"script"` key whose value is an object, fills absent `lang: "painless"` on that object.
- Wire that function as the `PopulateDefaultsFunc` for `input`, `transform`, `condition`, and `actions`. `trigger` and `metadata` stay `jsontypes.Normalized`.
- Leave redaction merge (`mergePreserveRedactedLeaves`, `mergeActionsPreservingRedactedLeaves`) running first in `fromAPIModel`. Semantic equality is a later Framework comparison (`StringSemanticEquals` on create, update, read, **and plan**). `fromAPIModel` still emits API (redaction-merged) JSON; when equality holds, the Framework substitutes the prior/plan value into Terraform state so authored JSON is preserved. Semantic equality MUST NOT treat `::es_redacted::` as equal to a concrete secret.
- Do **not** change `JSONWithDefaultsType.Equal` in this change. That method ignores the populate func, so every watch attribute that uses `JSONWithDefaultsType[map[string]any]` MUST share this same populate function.

## Capabilities

### New Capabilities

- _(none)_

### Modified Capabilities

- `elasticsearch-watch`: Add semantic-equality handling for Watcher-injected search-request defaults (`rest_total_hits_as_int`, `search_type`, `indices`) and script `lang` on `input`, `transform`, `condition`, and `actions`, including nested `chain` search sub-inputs and action-level search transforms (REQ-031).

## Impact

- **Specs**: Delta under `openspec/changes/elasticsearch-watch-search-defaults/specs/elasticsearch-watch/spec.md` until synced into the canonical spec.
- **Implementation** (future): `schema.go` CustomType for `input`/`transform`/`condition`/`actions`; `models.go` field types and `fromAPIModel` construction; `json_defaults.go` walker; unit tests; acceptance tests reproducing omitted search-request defaults (including omitted `indices` and a `chain` variant), an action search transform, and an omitted script `lang`.
