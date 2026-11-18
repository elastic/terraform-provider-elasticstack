## Why

`elasticstack_elasticsearch_ml_anomaly_detection_job` fails every apply with "Provider
produced inconsistent result after apply" on `custom_settings` whenever Elasticsearch
(via Kibana) has populated that field and the Terraform config does not set it (issue
#4729). Kibana writes metadata into `custom_settings` — a `created_by` wizard/module
provenance label, or a `custom_urls` drill-down array, sometimes empty (`[]`) — that the
Terraform config never mentions.

Three things combine to turn this into a failed apply, not just a diff:

1. `custom_settings` is `Optional` and **not** `Computed`, so Terraform requires the
   post-apply value to equal the planned value
   (`internal/elasticsearch/ml/anomalydetectionjob/schema.go:398-402`).
2. `UpdateAPIModel.BuildFromPlan` only copies `custom_settings` into the update request
   when the plan value is non-null, so an omitted/null plan never produces an API field
   and `hasChanges` stays `false`
   (`internal/elasticsearch/ml/anomalydetectionjob/models_api.go:234`).
3. The entitycore resource envelope always refreshes from the Get Jobs API after the
   update callback returns, even when the callback made no API call
   (`internal/entitycore/resource_envelope.go:433`).

The update callback returns `WriteResult{Model: plan}` with `custom_settings` null; the
envelope's read-after-write overwrites that with whatever Elasticsearch returns
(`{"created_by":"advanced-wizard"}`, `{"custom_urls":[]}`, …), and the final state no
longer matches the plan Terraform already validated. Nothing is written to
Elasticsearch — the update callback returned before any API call — so this is a planning
consistency bug, not a data-corruption one, but it fails every subsequent apply until the
provider is downgraded to 0.14.5 or the attribute is ignored.

This is issue #1543/#1544's sibling: that fix covered a config that explicitly sets
`custom_settings = null`; this issue is a config that omits the attribute entirely while
Elasticsearch has independently populated it.

## What Changes

Following the investigation and agreed design recorded on issue #4729 (research comment
and the "Agreed path: Path 1" follow-up from `@tobio`), `custom_settings` becomes an
**atomic, hands-off-when-omitted** attribute. The schema stays `Optional` and **not**
`Computed` — this proposal explicitly rejects the `Computed: true` +
`UseStateForUnknown()` fix suggested in the issue body, because that would silently
adopt whatever Elasticsearch/Kibana last wrote and could never express "clear this
field."

Contract:

| Config | Effect |
| --- | --- |
| omitted / `null` | The provider does not send `custom_settings` and does not copy the API value into state; state stays null. Applies to create, update, `terraform import`, and plain read/refresh alike. |
| `"{}"` (empty JSON object) | The only way to wipe an existing bag. The provider sends `{"custom_settings":{}}` and persists `"{}"` in state even when the Get Jobs API omits the field afterward. |
| any other JSON object | Full replace — Terraform owns the whole bag. Kibana/operator keys not present in the object are dropped by the Update Job API (it replaces, not merges) and will not survive a write. If Elasticsearch/Kibana adds keys after the write, the next read shows them, the next plan diffs, and the next apply replaces the bag with exactly the configured object. |

- `internal/elasticsearch/ml/anomalydetectionjob/models_tf.go` (`fromAPIModel`): stop
  unconditionally copying `apiModel.CustomSettings` into state. Only copy the live API
  value when the incoming model's `custom_settings` (the value already carried by the
  plan on the write path, or by prior state on the plain-read path) is a non-empty JSON
  object. When it is null, leave it null. When it is the empty-object sentinel `"{}"`,
  keep `"{}"` regardless of what the API returns.
- `internal/elasticsearch/ml/anomalydetectionjob/models_api.go` (`UpdateAPIModel`,
  `BuildFromPlan`, `toPutJobRequest`): change the wire encoding so an explicit empty
  object is distinguishable from "not set" and is not dropped by `omitempty`, and update
  `BuildFromPlan`'s change-detection guard to send the wipe.
- Add/extend acceptance test coverage for: omitted config over a Kibana-populated job
  (the reported repro), explicit `"{}"` wipe, and re-ownership of a bag that has drifted
  to include extra keys.

## Capabilities

### New Capabilities

- _(none)_

### Modified Capabilities

- **`elasticsearch-ml-anomaly-detection-job`**: `custom_settings` read/write contract
  changes from "always mirror the API value" to "hands-off when omitted, atomic replace
  when set, explicit `\"{}\"` wipe" (see delta spec for full requirements).

## Impact

- **Users**: Fixes the reported failed-apply loop for any job where Elasticsearch has
  populated `custom_settings` and the config omits the attribute. Users who need to clear
  an existing bag must set `custom_settings = "{}"` explicitly (previously there was no
  way to author this, since the field was always mirrored from the API). Users who set a
  partial object on a job that already has Kibana-authored keys will now see those keys
  disappear from the resource on the next apply (Update Job API replace semantics were
  already true; this proposal does not change apply-time behavior for the "field is set"
  case, only the "field is omitted" case).
- **Code**: `internal/elasticsearch/ml/anomalydetectionjob/models_tf.go`,
  `internal/elasticsearch/ml/anomalydetectionjob/models_api.go`,
  `internal/elasticsearch/ml/anomalydetectionjob/acc_test.go` and associated testdata.
  No schema (`schema.go`) changes — `custom_settings` keeps its current `Optional`,
  non-`Computed`, `jsontypes.NormalizedType` declaration.
- **State migration**: none. Existing state that already holds a Kibana-authored bag
  (left over from the pre-0.15.0 perpetual diff, or from any state that previously stored
  the API value while config was omitted) will show one plan of `{...} -> null` the first
  time this fix runs, matching the new hands-off contract; that apply makes no API call
  and the resource goes quiet afterward.
- **Maintenance**: closes the gap between this resource's `Optional`-not-`Computed`
  fields (`custom_settings`) and its `Optional`-plus-`Computed` fields
  (`daily_model_snapshot_retention_after_days`) by making the non-computed contract
  internally consistent end-to-end (write guard, read guard, wire encoding) rather than
  papering over it with `Computed`.
