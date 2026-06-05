## Context

Cross-Cluster Replication (CCR) requires a Platinum or Enterprise license. Trial mode on the
Elasticsearch cluster activates enterprise features, so existing CI acceptance test infrastructure
should be sufficient. The team must confirm which CI job handles licensed-feature tests and document
the skip guard (e.g. version-skip annotation) so CCR acceptance tests are not silently omitted on
basic-licensed builds.

The go-elasticsearch v8 typed client at v8.19.6 exposes the full CCR surface:

- `typedapi/ccr/follow/` — `PUT /{index}/_ccr/follow`
- `typedapi/ccr/followinfo/` — `GET /{index}/_ccr/info` (returns `types.FollowerIndex`)
- `typedapi/ccr/pausefollow/` — `POST /{index}/_ccr/pause_follow`
- `typedapi/ccr/resumefollow/` — `POST /{index}/_ccr/resume_follow`
- `typedapi/ccr/unfollow/` — `POST /{index}/_ccr/unfollow`
- `typedapi/ccr/putautofollowpattern/` — `PUT /_ccr/auto_follow/{name}`
- `typedapi/ccr/getautofollowpattern/` — `GET /_ccr/auto_follow/{name}` (returns the `patterns` list)
- `typedapi/ccr/pauseautofollowpattern/` — `POST /_ccr/auto_follow/{name}/pause`
- `typedapi/ccr/resumeautofollowpattern/` — `POST /_ccr/auto_follow/{name}/resume`
- `typedapi/ccr/deleteautofollowpattern/` — `DELETE /_ccr/auto_follow/{name}`

## Goals / Non-Goals

**Goals:**
- Two new Terraform resources covering the full CCR practitioner lifecycle
- No out-of-band scripting required for follower index and auto-follow pattern management
- User-controlled pause/resume for both resources, with plan-visible drift when Elasticsearch pauses automatically
- `delete_index_on_destroy` attribute giving operators control over destroy-time index fate
- `settings_raw` JSON string for index settings (matching `elasticstack_elasticsearch_index`)

**Non-Goals:**
- Remote cluster resource or cluster settings changes
- CCR stats/monitoring data sources
- CCS resources
- Fatal error recovery (recreate) — operator must intervene

## Decisions

### `settings_raw` over typed `settings` block

`follow.Request.Settings` is `*types.IndexSettings` (a deeply nested struct); the
`putautofollowpattern.Request.Settings` is `map[string]json.RawMessage`. Using `settings_raw` (a
JSON string) for both resources avoids the schema mismatch and is consistent with
`elasticstack_elasticsearch_index`, which already uses `settings_raw`. The practitioner encodes
settings as `jsonencode({...})` in HCL. The provider deserialises via the respective request's
`FromJSON` path or manual JSON unmarshalling.

### Write-only attributes

`settings_raw` and `data_stream_name` are write-only on `elasticstack_elasticsearch_ccr_follower_index`:
neither field is present in the `types.FollowerIndex` response returned by `GET /{index}/_ccr/info`.
`settings_raw` is likewise write-only on `elasticstack_elasticsearch_ccr_auto_follow_pattern` (absent
from `types.AutoFollowPatternSummary`). Both attributes are therefore `Optional` only (not
`Optional/Computed`). After a `terraform import`, these fields will be empty in state; practitioners
must add them manually to avoid a spurious diff (for `data_stream_name`, an empty value would trigger
a ForceNew replace, so the import note is particularly important).

**`data_stream_name` import sharp edge**: when importing a data stream follower index, `data_stream_name`
will be null in state because the API does not return it. If the practitioner subsequently sets
`data_stream_name` to a non-null value in their configuration, Terraform will plan a ForceNew
replacement — even though the follower already tracks the correct data stream. This is an inherent
API limitation: the provider has no way to verify the current value, so any non-null config value
appears as a change from null. To avoid unintended replacement, practitioners importing a data stream
follower MUST add `data_stream_name` to their configuration with the correct value before running
`terraform import`, or accept that adding it post-import will trigger recreation. Acceptance tests
SHALL verify this behaviour: import a data stream follower, confirm `data_stream_name` is null in
state, then apply a config with `data_stream_name` set and confirm a replacement plan is produced.

### Auto-follow pattern tuning parameters — API read limitation

`GET /_ccr/auto_follow/{name}` returns only one tuning parameter: `max_outstanding_read_requests`.
The other nine (`max_outstanding_write_requests`, `max_read_request_operation_count`,
`max_read_request_size`, `max_retry_delay`, `max_write_buffer_count`, `max_write_buffer_size`,
`max_write_request_operation_count`, `max_write_request_size`, `read_poll_timeout`) are accepted by
the PUT API but never returned by the GET API. All ten are exposed in the schema as `Optional` only.
During Read, the provider updates `max_outstanding_read_requests` from the API response and preserves
the prior-state values for the remaining nine unchanged. This prevents perpetual diffs while still
allowing practitioners to set, update, or remove any tuning parameter via the normal plan/apply cycle.

### ForceNew scope

`name`, `remote_cluster`, and `leader_index` are `ForceNew` on `elasticstack_elasticsearch_ccr_follower_index`
because the API does not allow renaming or re-pointing an existing follower index. `data_stream_name`
is also `ForceNew` for the same reason. Tuning parameters, `settings_raw`, and `status` are not
`ForceNew` — tuning/settings updates go through pause → resume; `status` is user-configurable.

For `elasticstack_elasticsearch_ccr_auto_follow_pattern`, only `name` is `ForceNew`; all other
attributes are updated in-place via `PUT /_ccr/auto_follow/{name}` (idempotent).

### User-configurable `status` and `active`

`status` on `elasticstack_elasticsearch_ccr_follower_index` and `active` on
`elasticstack_elasticsearch_ccr_auto_follow_pattern` are user-configurable optional attributes
(defaults: `"active"` and `true` respectively). Making them configurable rather than purely computed
allows Terraform's standard plan/apply cycle to surface and reconcile drift: when Elasticsearch
automatically pauses a resource, Read writes the actual state, the plan shows the diff against the
configured desired value, and Apply calls the appropriate pause or resume API.

**Follower index `status` state machine (Update callback):**

| Prior state | Desired state | Tuning changed | Action |
|---|---|---|---|
| `active` | `active` | yes | `pause_follow` → `resume_follow` with new params |
| `active` | `paused` | any | `pause_follow` only; tuning changes stored in state |
| `paused` | `active` | any | `resume_follow` with plan tuning params |
| `paused` | `paused` | any | no API call; tuning changes stored in state for next resume |

Create always starts active (the CCR API has no create-as-paused option). If `status = "paused"` is
configured at creation time, the provider creates the index then immediately pauses it.

Delete always requires the index to be paused before closing and unfollowing. The provider checks
prior state: if `status == "active"`, it calls `pause_follow`; if already `"paused"`, it skips the
pause call (idempotency guard).

**Auto-follow pattern `active` state machine (Update callback):**

The `PUT /_ccr/auto_follow/{name}` upsert is always called with the full plan configuration. After
the PUT, if prior state `active != plan active`:
- `true → false`: call `pause_auto_follow_pattern`
- `false → true`: call `resume_auto_follow_pattern`

Create always starts active. If `active = false` is configured at creation time, the provider calls
`pause_auto_follow_pattern` immediately after the PUT.

### Paused state and tuning parameter read-back

`GET /{index}/_ccr/info` omits `Parameters` when the follower index is paused (`types.FollowerIndex`
documents this: "If the follower index's status is paused, this object is omitted"). When Read
encounters a paused index, it MUST preserve the prior-state tuning parameter values rather than
zeroing them out, to avoid spurious diffs on attributes the practitioner has not changed.

### `delete_index_on_destroy`

Mirrors the `disable_on_destroy` pattern used in `internal/kibana/security_enable_rule/`. Default
`false` — destroy converts to a normal index (safe default, data preserved). When `true`, destroy
also sends `DELETE /{index}` after unfollowing.

The destroy sequence for `delete_index_on_destroy = false` is: pause (if active) → close → unfollow
→ **open**. The open step promotes the former follower to a usable regular index; without it the
index would remain closed after destroy. For `delete_index_on_destroy = true` the sequence is:
pause (if active) → close → unfollow → delete (open is skipped since the index is removed).
The close step is required by Elasticsearch before unfollow is allowed.

### `settings_raw` update path on follower index

`resumefollow.Request` exposes only the ten tuning parameters — it has no `Settings` field and no
identity fields (`remote_cluster`, `leader_index`). Settings cannot be applied via the resume call.

When `settings_raw` changes on an active follower index, the update sequence is:
1. `POST /{index}/_ccr/pause_follow`
2. `PUT /{index}/_settings` via the existing `elasticsearch.UpdateIndexSettings` helper (accepts
   `map[string]any`; flat dotted keys are passed through as raw bytes, so no normalisation needed)
3. `POST /{index}/_ccr/resume_follow` with the plan tuning params only

If `pause_follow` fails before step 2, Terraform state is unchanged. On the next plan, Read will
fetch the actual Elasticsearch state and the standard reconciliation cycle recovers without manual
intervention.

### `settings_raw` flat-key normalisation

The follower index `follow.Request.Settings` is `*types.IndexSettings`, whose JSON unmarshaller
expects nested format (`{"index": {"refresh_interval": "30s"}}`). The AFP uses
`map[string]json.RawMessage`, which accepts flat keys natively. The `elasticstack_elasticsearch_index`
convention uses flat dotted keys (`{"index.refresh_interval": "30s"}`), so practitioners writing
CCR config will naturally use that format. To provide a uniform user-facing format, the provider
normalises flat dotted keys to nested form before unmarshalling `settings_raw` into
`*types.IndexSettings` on the follower index resource. The AFP resource passes the value through
unchanged. Implementors SHALL write a small helper function for this transformation.

### `status` vs `active` naming asymmetry

The follower index uses `status` (string: `"active"`/`"paused"`) and the auto-follow pattern uses
`active` (bool: `true`/`false`). This directly mirrors the Elasticsearch API surface: the follow
info endpoint returns `status` as a string enum, while the AFP summary returns `active` as a
boolean. The naming difference is intentional and API-driven; documentation for both resources
should explain this to avoid practitioner confusion.

### `int64` for all count-type tuning parameters — per-API type mapping

All ten count-type tuning parameters use `int64` in the Terraform schema for consistency. Each API
call requires different conversions at the boundary. Implementors MUST follow the table below:

| Field | `follow.Request` (create) | `resumefollow.Request` (update) | `putautofollowpattern.Request` (AFP) | Read (`FollowerIndexParameters`) |
|---|---|---|---|---|
| `max_outstanding_read_requests` | `*int64` | `*int64` | `*int` | `*int64` |
| `max_outstanding_write_requests` | `*int` | `*int64` | `*int` | `*int` |
| `max_read_request_operation_count` | `*int` | `*int64` | `*int` | `*int` |
| `max_write_buffer_count` | `*int` | `*int64` | `*int` | `*int` |
| `max_write_request_operation_count` | `*int` | `*int64` | `*int` | `*int` |
| `max_read_request_size` | `types.ByteSize` | `*string` | `types.ByteSize` | — |
| `max_write_buffer_size` | `types.ByteSize` | `*string` | `types.ByteSize` | — |
| `max_write_request_size` | `types.ByteSize` | `*string` | `types.ByteSize` | — |
| `max_retry_delay` | `types.Duration` | `types.Duration` | `types.Duration` | — |
| `read_poll_timeout` | `types.Duration` | `types.Duration` | `types.Duration` | — |

Schema values (`int64` for counts, `string` for byte-size and duration) are converted at each API
boundary: narrow `int64 → int` or `int64 → *int64` (no-op cast) as required; byte-size strings map
to `types.ByteSize` (raw JSON string value) for `follow`/`putautofollowpattern`, and to `*string`
directly for `resumefollow`. Values that would overflow `int` are rejected at apply time with a
diagnostic error.

### `data_stream_name` support

`follow.Request.DataStreamName` is exposed in v1 as an optional ForceNew attribute to support data
stream followers. The implementation is low-risk since it is just an additional string field in the
create request and does not affect the lifecycle state machine. It is write-only (not returned by
the info API) — see the Write-only attributes decision above.

## Open questions

- **CI acceptance tests**: Confirm which CI job runs licensed-feature tests and whether trial mode
  on the existing stack is sufficient or a separate cluster is needed. The acceptance tests should
  carry an appropriate skip annotation if a licensed cluster is not available.
