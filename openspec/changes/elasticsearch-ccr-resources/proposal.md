## Why

Practitioners orchestrating Cross-Cluster Replication (CCR) today must combine Terraform (for cluster
setup and remote cluster connectivity) with out-of-band REST API calls or SDK scripts to create
follower indices and auto-follow patterns. This gap forces hybrid Terraform+script workflows for
multi-DC ETL pipelines and time-series replication.

The go-elasticsearch v8 typed API (already in `go.mod` at v8.19.6) exposes the full CCR surface
under `typedapi/ccr/`. Remote cluster connectivity is handled by the existing
`elasticstack_elasticsearch_cluster_settings` resource and is not changed here.

## What Changes

### New Resources

#### `elasticstack_elasticsearch_ccr_follower_index`

Manages the lifecycle of a CCR follower index. Key attributes:

- `name` (required, ForceNew): name of the follower index
- `remote_cluster` (required, ForceNew): remote cluster alias
- `leader_index` (required, ForceNew): name of the leader index on the remote cluster
- `data_stream_name` (optional, ForceNew, write-only): local data stream name when following a data stream; not returned by the info API so will be empty after import
- CCR tuning parameters (all optional, int64 for count types): `max_outstanding_read_requests`, `max_outstanding_write_requests`, `max_read_request_operation_count`, `max_read_request_size`, `max_retry_delay`, `max_write_buffer_count`, `max_write_buffer_size`, `max_write_request_operation_count`, `max_write_request_size`, `read_poll_timeout`
- `settings_raw` (optional, write-only): JSON-encoded index settings overrides matching the `elasticstack_elasticsearch_index` convention; not returned by the info API so will be empty after import
- `delete_index_on_destroy` (optional, default `false`): when `true`, the underlying index is deleted on destroy; when `false`, the index is converted to a normal index and left in place
- `status` (optional, default `"active"`): desired replication status — `"active"` or `"paused"`; user-configurable so plan/apply reconciles drift when Elasticsearch pauses the index automatically

**Lifecycle state machine:**
- **Create**: `PUT /{index}/_ccr/follow`; if `status = "paused"`, immediately calls `POST /{index}/_ccr/pause_follow`
- **Read**: `GET /{index}/_ccr/info`; maps `status`, `remote_cluster`, `leader_index`, and all tuning parameters; when paused (`Parameters` is nil in response), preserves prior-state tuning param values
- **Update**: state machine driven by prior and desired `status`:
  - active → active (tuning changed): `pause_follow` → `resume_follow` with new params
  - active → paused: `pause_follow` only
  - paused → active: `resume_follow` with plan tuning params
  - paused → paused: no API call; tuning changes stored in state
- **Destroy** (`delete_index_on_destroy = false`): pause (if active) → `POST /{index}/_close` → `POST /{index}/_ccr/unfollow` → `POST /{index}/_open` (promotes to usable regular index)
- **Destroy** (`delete_index_on_destroy = true`): pause (if active) → `POST /{index}/_close` → `POST /{index}/_ccr/unfollow` → `DELETE /{index}`

#### `elasticstack_elasticsearch_ccr_auto_follow_pattern`

Manages a CCR auto-follow pattern. Key attributes:

- `name` (required, ForceNew): pattern name
- `remote_cluster` (required): remote cluster alias
- `leader_index_patterns` (required, min 1): list of index patterns to match against the remote cluster
- `leader_index_exclusion_patterns` (optional): exclusion patterns
- `follow_index_pattern` (optional): template for follower index name; `{{leader_index}}` is supported
- CCR tuning parameters (same set as follower index resource, all optional, int64 for count types)
- `settings_raw` (optional, write-only): JSON-encoded index settings overrides; not returned by the API so will be empty after import
- `active` (optional, default `true`): whether the pattern should be active; user-configurable so plan/apply reconciles drift when Elasticsearch pauses the pattern automatically

**Lifecycle:**
- **Create**: `PUT /_ccr/auto_follow/{name}`; if `active = false`, immediately calls `POST /_ccr/auto_follow/{name}/pause`
- **Read**: `GET /_ccr/auto_follow/{name}`; maps `active`, `remote_cluster`, `leader_index_patterns`, `leader_index_exclusion_patterns`, `follow_index_pattern`, and all 10 tuning parameters
- **Update**: `PUT /_ccr/auto_follow/{name}` (idempotent upsert); then if `active` changed, call pause or resume as needed
- **Destroy**: `DELETE /_ccr/auto_follow/{name}` — no state machine required

### New Internal Packages

- `internal/elasticsearch/ccr/followerindex/` — CRUD callbacks for the follower index resource
- `internal/elasticsearch/ccr/autofollow/` — CRUD callbacks for the auto-follow pattern resource
- `internal/clients/elasticsearch/ccr.go` — API client functions used by both resources (includes `OpenIndex` for the promote-follower destroy path)

## Capabilities

### New Capabilities

- **`elasticsearch-ccr-follower-index`** — new resource `elasticstack_elasticsearch_ccr_follower_index`
- **`elasticsearch-ccr-auto-follow-pattern`** — new resource `elasticstack_elasticsearch_ccr_auto_follow_pattern`

## Impact

- `internal/clients/elasticsearch/ccr.go` (new ~150 LOC)
- `internal/elasticsearch/ccr/followerindex/` (new resource package)
- `internal/elasticsearch/ccr/autofollow/` (new resource package)
- Provider registration: two new resources added to `provider/provider.go`
- Documentation templates: two new resource pages
- Acceptance tests: require a licensed (trial) Elasticsearch cluster with remote cluster connectivity configured

## Out of Scope

- Remote cluster connectivity (`cluster.remote.*`) — handled by existing `elasticstack_elasticsearch_cluster_settings`
- Dedicated `elasticstack_elasticsearch_remote_cluster` resource — explicitly deferred
- CCR stats / monitoring data sources
- Cross-cluster search (CCS)
- Recreating follower indices after fatal replication errors (operator intervention)
