## Context

`elasticstack_elasticsearch_watch` maps `input` and `transform` as opaque JSON strings (`jsontypes.Normalized`) rather than typed HCL blocks, by design (Watcher inputs/transforms are highly polymorphic — `search`, `http`, `simple`, `chain`, `script`, etc.). `jsontypes.Normalized.StringSemanticEquals` (from `terraform-plugin-framework-jsontypes`) treats two JSON strings as equal when they differ only in key order or insignificant whitespace. It does **not** tolerate one side having extra keys the other omits.

Elasticsearch's `WatcherSearchTemplateRequest` (the Java type backing every Watcher `search` input/transform request) defaults `restTotalHitsAsInt` to `true` and `searchType` to `query_then_fetch`, and its serializer always writes both fields on Get Watch — regardless of whether the original Put Watch body included them. So a practitioner who omits these fields (the common case; the fields are rarely set explicitly) sees them appear in the Get Watch response used to populate state. Because `fromAPIModel` marshals that response compactly into `input`/`transform` state as-is, the post-apply state differs from the practitioner's plan value by exactly those injected keys, and the Plugin Framework's consistency check between plan and final state fails the apply with `Provider produced inconsistent result after apply`.

The provider already has a general-purpose solution for this class of problem: `internal/utils/customtypes.JSONWithDefaultsType[TModel]` / `JSONWithDefaultsValue[TModel]`. It wraps `jsontypes.Normalized` and, on `StringSemanticEquals`, unmarshals both the prior and new JSON strings into `TModel`, applies a caller-supplied `PopulateDefaultsFunc[TModel]` to each, re-marshals, and only then defers to `jsontypes.Normalized`'s comparison. This is already used by `internal/elasticsearch/ml/datafeed` (`script_fields`, defaulting `ignore_failure` and script `lang`) and by the Kibana Lens dashboard panel converters. It is the mechanism named in the issue's "Preferred packaging" recommendation.

## Goals

- Make `terraform apply` succeed for watches whose `input`/`transform` search requests omit `rest_total_hits_as_int` and/or `search_type`, and make the following `terraform plan` empty.
- Reuse `customtypes.JSONWithDefaultsType`/`JSONWithDefaultsValue` rather than adding a new custom type or a bespoke comparison path.
- Cover nested `chain` input sub-inputs (each chained `search` sub-input has its own `request` object and is defaulted independently), not just the top-level `input.search`/`transform.search` case from the reported repro.
- Preserve practitioner-authored JSON verbatim in state; default-population only affects the *comparison* used for semantic equality, never the value written to disk or into Terraform state's JSON string itself. (`WithDefaults()` on `JSONWithDefaultsValue` is used internally by `StringSemanticEquals` to build a comparison value — see [`json_with_defaults_value.go`](../../../internal/utils/customtypes/json_with_defaults_value.go) — and does not itself mutate the value returned to the caller of `fromAPIModel`.)
- Keep the existing `mergePreserveRedactedLeaves`/`::es_redacted::` handling in `fromAPIModel` fully intact and running before the new custom type's semantic equality is ever invoked by the framework.

## Non-Goals

- Script `lang: "painless"` re-emission on omitted script `lang` fields (noted in the issue's "Other susceptible surfaces" table). Deferred — see Open Questions.
- `indices_options` defaulting. The issue notes Elasticsearch only serializes `indices_options` when it differs from default, so it is "less likely to drift" and is not part of the reported failure.
- Converting `input`/`transform` from opaque JSON strings to typed nested HCL blocks. The issue explicitly rules this out ("keep them as JSON strings; do not convert to nested HCL blocks").
- Changing `mergePreserveRedactedLeaves` / redaction behavior for `input` or `actions`.

## Decisions

| Topic | Decision |
|-------|----------|
| Custom type | Reuse `customtypes.JSONWithDefaultsType[map[string]any]` / `JSONWithDefaultsValue[map[string]any]` for both `input` and `transform`. `TModel = map[string]any` (not a typed Watcher struct) because Watcher inputs/transforms are polymorphic (`search`, `http`, `chain`, `simple`, `script`, ...) and the resource already treats them as opaque maps elsewhere (`mergePreserveRedactedLeaves` operates on `map[string]any`). |
| Defaults function | One shared `populateWatcherSearchDefaults(model map[string]any) map[string]any` covers both attributes, matching the issue's "one defaults function covers both `input` and `transform`" recommendation. |
| Walk strategy | Recursive, structural walk: whenever an object has a `"search"` key whose value is itself an object containing a `"request"` object, populate `rest_total_hits_as_int`/`search_type` defaults on that `request` object if absent, then continue walking (the walk does not stop at the first match, so multiple chained search sub-inputs are each visited). This is a generic tree walk rather than an enumeration of specific paths (`input.search`, `transform.search`, `input.chain.inputs[].*.search`, ...), so it naturally covers `chain` sub-inputs and any future nesting shape without path enumeration. |
| Where defaults are injected | Only within a `request` object reached via a `search` key, matching `WatcherSearchTemplateRequest`'s shape. The walker SHALL NOT inject these two keys anywhere else in the tree (e.g. not inside an `http` input's `request` object, which has an unrelated shape). |
| Absent vs. explicit-false/other | The walker only sets a key when it is **absent** from the `request` map. An explicit `rest_total_hits_as_int: false` or `search_type: "dfs_query_then_fetch"` is left untouched — the framework's underlying `jsontypes.Normalized` comparison then correctly reports a real difference if the API-returned value differs from an explicit practitioner value. |
| Ordering vs. redaction merge | `fromAPIModel` continues to run `mergePreserveRedactedLeaves` (input) / `mergeActionsPreservingRedactedLeaves` (actions) exactly as today, then marshals the merged map to a compact JSON string, then wraps that string with `customtypes.NewJSONWithDefaultsValue(json, populateWatcherSearchDefaults)` instead of `jsontypes.NewNormalizedValue(json)`. Default-population happens later, only when the framework calls `StringSemanticEquals` during its post-apply/post-read consistency check — it is not part of `fromAPIModel`'s own data flow. |
| `toPutModel` | Unaffected in substance: `customtypes.JSONWithDefaultsValue` embeds `jsontypes.Normalized`, so `.ValueString()` still returns the practitioner's literal JSON for `json.Unmarshal` into the Put Watch request body. No behavior change to what gets sent to Elasticsearch. |
| Script `lang` defaulting | Deferred (see Open Questions). Not wired into `populateWatcherSearchDefaults` in this pass. |

## Risks / Trade-offs

- **Recursive walk cost**: negligible — watch `input`/`transform` documents are small, and the walk only runs during framework semantic-equality checks (not on every plan).
- **Over-broad matching**: a non-Watcher JSON document that happens to have a `search` key containing a `request` object with unrelated semantics (unlikely, but the schema is user-authored free-form JSON) would have these two keys spuriously defaulted for *comparison purposes only*. Given the field is documented as Watcher input/transform JSON and the keys are only for comparison (never rewritten into state), this risk is accepted, matching how `JSONWithDefaultsType` is already used elsewhere in the provider (e.g. `datafeed.script_fields` matches on script-shaped objects generically).
- **Chain input coverage is best-effort structural, not schema-validated**: the walker recognizes chained search sub-inputs by shape (`search.request` present), not by validating against the full Watcher chain-input schema. This matches the issue's recommended approach and the resource's existing opaque-map handling elsewhere.

## Open Questions

1. **Script `lang: "painless"` re-emission**: Elasticsearch also defaults script `lang` to `"painless"` when omitted and re-emits it on Get Watch, per the issue's "Other susceptible surfaces" table (condition/transform/action scripts). This is a real but separate drift source, not exercised by the reported repro (which is specifically about search-request defaults). Deferred to a follow-up change; if a user reports this as a distinct apply-inconsistency bug, `populateWatcherSearchDefaults` (or a sibling function reusing the same `JSONWithDefaultsType` wiring) can be extended to also default `lang` on any object containing a `source`/`id` script body, mirroring `datafeed.populateScriptFieldsDefaults`'s existing script-lang handling.
2. **`indices_options` defaulting**: not currently reported as a drift source (only serialized when non-default per the issue). No action planned; revisit if a future issue reports drift here.

## Migration / State

- No state schema/version change: `input` and `transform` remain `string`-typed attributes at the wire level (`customtypes.JSONWithDefaultsType` embeds `jsontypes.NormalizedType`, itself a `StringType`). Existing state values continue to decode without a state upgrader.
- Behavior change is limited to the Plugin Framework's semantic-equality pass: watches that previously required the workaround JSON (`rest_total_hits_as_int = true`, `search_type = "query_then_fetch"` explicitly set to avoid the apply failure) continue to work unchanged (explicit values are never overwritten). Watches that omit these fields, which previously failed apply, now succeed.
