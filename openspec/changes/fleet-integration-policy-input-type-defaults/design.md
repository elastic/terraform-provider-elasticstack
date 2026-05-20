## Context

`internal/fleet/integration_policy/models_defaults.go` implements the defaults
extractor. The key types and functions are:

```
apiPolicyTemplate       — one entry per policyTemplates[] element in PackageInfo
apiPolicyTemplateInput  — one entry per inputs[] element (integration-type)
apiDatastreamStream     — one entry per streams[] element in a data-stream
apiVars                 — []apiVar; each var has Name, Default, Multi
packageInfoToDefaults() — top-level driver; calls policyTemplates.defaults() then datastreams.defaults()
apiPolicyTemplates.defaults() — produces map[inputID]jsontypes.Normalized (input-level var defaults)
apiDatastreams.defaults()     — produces map[inputIDSuffix]map[datasetName]inputDefaultsStreamModel
```

**Integration-type shape** (already handled):

```json
{
  "name": "kafka",
  "inputs": [
    { "type": "jolokia/metrics", "vars": [ ... ] },
    { "type": "kafka/metrics",   "vars": [ ... ] }
  ]
}
```

`apiPolicyTemplates.defaults()` iterates `policyTemplate.Inputs`, keys the
defaults as `"{templateName}-{inputType}"`.

**Input-type shape** (currently unhandled):

```json
{
  "name": "gcp_pubsub",
  "input": "gcp-pubsub",
  "vars": [
    { "name": "project_id",        "type": "text",    "default": null },
    { "name": "subscription_name", "type": "text",    "default": null },
    { "name": "subscription_type", "type": "text",    "default": "shared" },
    { "name": "data_stream.dataset","type": "text",   "default": "gcp_pubsub.generic" }
  ]
}
```

For an input-type package Kibana materialises one implicit stream per policy
template. The resulting policy carries the template-level `vars` as the
stream's variable defaults. The policy input ID in state is
`"{templateName}-{inputType}"`, identical in shape to the integration-type key.

## Goals

- After the fix, `apiPolicyTemplates.defaults()` extracts variable defaults
  for both package shapes.
- For integration-type templates (have a non-empty `inputs` array), behaviour is
  unchanged.
- For input-type templates (have `input` string + `vars`), the function produces
  one entry keyed `"{templateName}-{inputType}"` whose `jsontypes.Normalized`
  holds the extracted defaults — exactly the same shape the integration-type path
  emits.
- The fix also ensures that downstream `packageInfoToDefaults()` can then pair
  those var defaults with the stream-level defaults already produced by
  `apiDatastreams.defaults()`.

## Non-Goals

- No changes to the user-facing schema.
- No changes to the Kibana API request/response bodies.
- No changes to the semantic-equality contract itself.

## Decisions

| Topic | Decision |
|-------|----------|
| Struct extension | Add `Input string` and `Vars apiVars` to `apiPolicyTemplate`; Go JSON decode ignores unknown fields so integration-type templates that have no `input` field are unaffected. |
| Defaults key format | Re-use the same `"{name}-{type}"` key as integration-type so the rest of `packageInfoToDefaults()` requires no changes. |
| Null defaults | Input-type vars with no `Default` and `Multi: false` are omitted from the defaults map (existing behaviour from `apiVars.defaults()`). Vars with `Multi: true` and no `Default` produce an empty array — same existing behaviour. |
| Test fixture | Add `testdata/integration_gcp_pubsub.json` with a compact but representative `PackageInfo` payload for `gcp_pubsub`. The fixture must include at least one var with a non-null default (e.g. `subscription_type: "shared"`) and one without a default (e.g. `project_id`) so the unit test can verify which vars appear in the output. |
| Acceptance test | Use `gcp_pubsub` as the integration under test. Apply a policy that sets only the required user-visible vars; assert no inconsistency error and that state values are stable across a plan/apply/plan cycle. Version-gate to `>= 8.10.0` (same as existing Fleet integration policy acceptance tests). |

## Open Questions

- None. The root cause and fix are fully identified from the investigation
  notes in [issue #2436](https://github.com/elastic/terraform-provider-elasticstack/issues/2436).

## Risks / Trade-offs

- The fix extends a struct used in JSON unmarshaling. Go's `encoding/json`
  ignores unknown struct fields by default, so integration-type templates with no
  `input` field will decode to the zero value (`Input: ""`), which is
  distinguishable from a real input-type template (where `Input` will be a
  non-empty type string). The dispatch in `defaults()` should therefore check
  `len(policyTemplate.Inputs) > 0` for integration-type and
  `policyTemplate.Input != ""` for input-type, rather than checking the `vars`
  array (which may be non-empty for integration-type templates too).
