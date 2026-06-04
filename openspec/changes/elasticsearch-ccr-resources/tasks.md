## 1. API client layer

- [x] 1.1 Create `internal/clients/elasticsearch/ccr.go` with the following exported functions (all accept `ctx context.Context` and `apiClient *clients.ElasticsearchScopedClient`):
  - `CreateFollowerIndex(ctx, apiClient, indexName string, req *follow.Request) fwdiags.Diagnostics`
  - `GetFollowerIndex(ctx, apiClient, indexName string) (*types.FollowerIndex, fwdiags.Diagnostics)` — returns `nil, nil` when not found (404)
  - `PauseFollowerIndex(ctx, apiClient, indexName string) fwdiags.Diagnostics`
  - `ResumeFollowerIndex(ctx, apiClient, indexName string, req *resumefollow.Request) fwdiags.Diagnostics`
  - `CloseIndex(ctx, apiClient, indexName string) fwdiags.Diagnostics` — calls `POST /{index}/_close`
  - `UnfollowIndex(ctx, apiClient, indexName string) fwdiags.Diagnostics`
  - `DeleteIndex(ctx, apiClient, indexName string) fwdiags.Diagnostics` — reuse or delegate to existing `DeleteIndex` in `internal/clients/elasticsearch/index.go` if available; add here if not
  - `OpenIndex(ctx, apiClient, indexName string) fwdiags.Diagnostics` — calls `POST /{index}/_open`; used to promote a former follower to a usable regular index after unfollow
  - `PutAutoFollowPattern(ctx, apiClient, name string, req *putautofollowpattern.Request) fwdiags.Diagnostics`
  - `GetAutoFollowPattern(ctx, apiClient, name string) (*types.AutoFollowPatternSummary, fwdiags.Diagnostics)` — returns `nil, nil` when not found (404); filters by name from the `patterns` slice
  - `PauseAutoFollowPattern(ctx, apiClient, name string) fwdiags.Diagnostics`
  - `ResumeAutoFollowPattern(ctx, apiClient, name string) fwdiags.Diagnostics`
  - `DeleteAutoFollowPattern(ctx, apiClient, name string) fwdiags.Diagnostics`
- [x] 1.2 Use `typedClient.Ccr.*` from the go-elasticsearch v8 typed client in all functions (mirroring `ilm.go` which uses `typedClient.Ilm.*`)
- [x] 1.3 Wrap errors with `diagutil.FrameworkDiagFromError`; handle 404 (not found) using the existing `IsNotFoundElasticsearchError` helper from `internal/clients/elasticsearch/`

## 2. `elasticstack_elasticsearch_ccr_follower_index` resource

- [ ] 2.1 Create directory `internal/elasticsearch/ccr/followerindex/`
- [ ] 2.2 Create `models.go` with `Model` struct embedding `entitycore.ElasticsearchConnectionField` and all Terraform-mapped attributes:
  - Required/ForceNew: `Name`, `RemoteCluster`, `LeaderIndex` (string)
  - Optional/ForceNew: `DataStreamName` (string, nullable) — write-only; not readable from the info API
  - Optional: `SettingsRaw` (string, nullable) — write-only; not readable from the info API
  - Optional tuning attrs (all int64 for count types, all string for duration/byte-size types, all nullable): `MaxOutstandingReadRequests`, `MaxOutstandingWriteRequests`, `MaxReadRequestOperationCount`, `MaxReadRequestSize`, `MaxRetryDelay`, `MaxWriteBufferCount`, `MaxWriteBufferSize`, `MaxWriteRequestOperationCount`, `MaxWriteRequestSize`, `ReadPollTimeout`
  - Optional: `DeleteIndexOnDestroy` (bool, default `false`)
  - Optional: `Status` (string, default `"active"`) — valid values: `"active"`, `"paused"`
- [ ] 2.3 Create `schema.go` returning the `schema.Schema` with all attributes documented using descriptions from the API spec:
  - Mark ForceNew attributes with `PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}`
  - `Status` uses `stringdefault.StaticString("active")` and a string validator restricting values to `["active", "paused"]`
- [ ] 2.4 Create `create.go` implementing the Create callback:
  - Build `follow.Request` from plan model; narrow int64 tuning values to int for `MaxOutstandingWriteRequests`, `MaxReadRequestOperationCount`, `MaxWriteBufferCount`, `MaxWriteRequestOperationCount`
  - If `settings_raw` is set, normalise flat dotted keys to nested form (e.g. `"index.refresh_interval"` → `{"index":{"refresh_interval":…}}`), then unmarshal into `*types.IndexSettings` and set `req.Settings`; this normalisation is only required here — `follow.Request.Settings` is `*types.IndexSettings` which expects nested format
  - Call `ccr.CreateFollowerIndex`
  - If plan `status == "paused"`, call `ccr.PauseFollowerIndex`
  - Store plan `status` in state
- [ ] 2.5 Create `read.go` implementing the Read callback:
  - Call `ccr.GetFollowerIndex`; if nil, call `resp.State.RemoveResource` and return
  - Map `types.FollowerIndex` fields to model: `status` (as string), `remote_cluster`, `leader_index`
  - If `Parameters != nil`, map all ten tuning fields back to state, widening `*int` fields to `int64`
  - If `Parameters == nil` (follower is paused), preserve prior-state tuning param values unchanged — do not zero them out
  - Do not attempt to read `settings_raw` or `data_stream_name` from the response (write-only)
- [ ] 2.6 Create `update.go` implementing the Update callback using the following state machine (prior state = `state.Status`, desired = `plan.Status`):
  - NOTE: `resumefollow.Request` exposes only the ten tuning parameters — it has no `Settings` field and no identity fields. Do not attempt to set `remote_cluster`, `leader_index`, or `settings` on it.
  - `active → active` and tuning params and/or `settings_raw` changed:
    1. Call `ccr.PauseFollowerIndex`
    2. If `settings_raw` changed: unmarshal `settings_raw` as `map[string]any` and call `elasticsearch.UpdateIndexSettings` (reuse existing helper; flat keys are accepted, no normalisation needed)
    3. Build `resumefollow.Request` with plan tuning params only and call `ccr.ResumeFollowerIndex`
  - `active → paused`: call `ccr.PauseFollowerIndex` only; if `settings_raw` also changed, call `elasticsearch.UpdateIndexSettings` before resuming would be skipped — store updated `settings_raw` in state for application on next resume transition
  - `paused → active`: if `settings_raw` changed, call `elasticsearch.UpdateIndexSettings` first; then build `resumefollow.Request` with plan tuning params and call `ccr.ResumeFollowerIndex`
  - `paused → paused`: no API calls; store updated tuning and `settings_raw` in state
- [ ] 2.7 Create `delete.go` implementing the Delete callback:
  - If prior state `status == "active"`, call `ccr.PauseFollowerIndex` (skip if already paused)
  - Call `ccr.CloseIndex`
  - Call `ccr.UnfollowIndex`
  - If `delete_index_on_destroy == true`: call `ccr.DeleteIndex`
  - If `delete_index_on_destroy == false`: call `ccr.OpenIndex` to promote the index to a usable regular index
- [ ] 2.8 Create `resource.go` wiring the resource via `entitycore.NewElasticsearchResource[Model]` with appropriate `ElasticsearchResourceOptions`
- [ ] 2.9 Write unit tests in `followerindex_test.go` covering:
  - Schema validation (valid/invalid `status` values)
  - All four branches of the Update state machine
  - Delete sequence with `status = "active"` vs `status = "paused"` prior state
  - Delete with `delete_index_on_destroy = true` (ends in DeleteIndex) and `false` (ends in OpenIndex)
  - `settings_raw` flat-key normalisation at create: verify nested and flat formats both unmarshal correctly into `*types.IndexSettings`
  - `settings_raw` update path: verify `UpdateIndexSettings` is called between pause and resume when `settings_raw` changes, and is NOT called when only tuning params change

## 3. `elasticstack_elasticsearch_ccr_auto_follow_pattern` resource

- [ ] 3.1 Create directory `internal/elasticsearch/ccr/autofollow/`
- [ ] 3.2 Create `models.go` with `Model` struct embedding `entitycore.ElasticsearchConnectionField`:
  - Required/ForceNew: `Name` (string)
  - Required: `RemoteCluster` (string)
  - Required: `LeaderIndexPatterns` (list of string)
  - Optional: `LeaderIndexExclusionPatterns` (list of string)
  - Optional: `FollowIndexPattern` (string, nullable)
  - Optional: `SettingsRaw` (string, nullable) — write-only; not readable from the API
  - Optional tuning attrs (all int64 for count types, all string for duration/byte-size types, all nullable): `MaxOutstandingReadRequests`, `MaxOutstandingWriteRequests`, `MaxReadRequestOperationCount`, `MaxReadRequestSize`, `MaxRetryDelay`, `MaxWriteBufferCount`, `MaxWriteBufferSize`, `MaxWriteRequestOperationCount`, `MaxWriteRequestSize`, `ReadPollTimeout`
  - Optional: `Active` (bool, default `true`)
- [ ] 3.3 Create `schema.go` returning the `schema.Schema` with all attributes documented:
  - `LeaderIndexPatterns` uses `listvalidator.SizeAtLeast(1)` to enforce at least one entry
  - `Active` uses `booldefault.StaticBool(true)`
- [ ] 3.4 Create `create.go` implementing the Create callback:
  - Build `putautofollowpattern.Request` from plan; narrow int64 tuning values to int where required
  - If `settings_raw` is set, unmarshal into `map[string]json.RawMessage` and set `req.Settings`
  - Call `ccr.PutAutoFollowPattern`
  - If plan `active == false`, call `ccr.PauseAutoFollowPattern`
  - Store plan `active` in state
- [ ] 3.5 Create `read.go` implementing the Read callback:
  - Call `ccr.GetAutoFollowPattern`; if nil, call `resp.State.RemoveResource` and return
  - Map readable `AutoFollowPatternSummary` fields to model: `active`, `remote_cluster`, `leader_index_patterns`, `leader_index_exclusion_patterns`, `follow_index_pattern`, and `max_outstanding_read_requests` (widen `int` to `int64`)
  - For the nine tuning params not returned by the API (`max_outstanding_write_requests`, `max_read_request_operation_count`, `max_read_request_size`, `max_retry_delay`, `max_write_buffer_count`, `max_write_buffer_size`, `max_write_request_operation_count`, `max_write_request_size`, `read_poll_timeout`): copy prior-state values unchanged — do not zero them out
  - Do not attempt to read `settings_raw` from the response (write-only)
- [ ] 3.6 Create `update.go`:
  - Build `putautofollowpattern.Request` from plan and call `ccr.PutAutoFollowPattern` (idempotent upsert)
  - If prior state `active == true` and plan `active == false`: call `ccr.PauseAutoFollowPattern`
  - If prior state `active == false` and plan `active == true`: call `ccr.ResumeAutoFollowPattern`
- [ ] 3.7 Create `delete.go`:
  - Call `ccr.DeleteAutoFollowPattern`
- [ ] 3.8 Create `resource.go` wiring via `entitycore.NewElasticsearchResource[Model]`
- [ ] 3.9 Write unit tests in `autofollow_test.go` covering:
  - Schema validation (`leader_index_patterns` empty list rejected)
  - Create with `active = false` calls pause after PUT
  - All three branches of the Update active state machine
  - Read maps `max_outstanding_read_requests` from API and preserves prior-state for the other nine unreadable params

## 4. Provider registration

- [ ] 4.1 Add both new resources to the provider's resource list in `provider/plugin_framework.go`:
  - `elasticstack_elasticsearch_ccr_follower_index`
  - `elasticstack_elasticsearch_ccr_auto_follow_pattern`
- [ ] 4.2 Add imports for the two new packages

## 5. Acceptance tests

- [ ] 5.1 Create `internal/elasticsearch/ccr/followerindex/acctest/` with at least:
  - Happy path: create follower index, verify `status == "active"`, update tuning params, verify, destroy with `delete_index_on_destroy = false` (verify index exists and is open as a regular index)
  - Destroy with `delete_index_on_destroy = true` verifies the index is gone
  - Status management: create with `status = "paused"`, verify paused; update to `status = "active"`, verify resumed
  - `data_stream_name` import path: create a data stream follower with `data_stream_name` set, import the resource, verify `data_stream_name` is null in state, then apply a config with `data_stream_name` set and verify a replacement plan (ForceNew) is produced — confirm that adding `data_stream_name` to config post-import triggers recreation
  - Add license/feature skip annotation if the test cluster does not support CCR (check CI policy)
- [ ] 5.2 Create `internal/elasticsearch/ccr/autofollow/acctest/` with at least:
  - Happy path: create auto-follow pattern, verify `active == true`, update exclusion patterns, verify, destroy
  - Active management: create with `active = false`, verify paused; update to `active = true`, verify resumed
  - Add license/feature skip annotation

## 6. Documentation

- [ ] 6.1 Create documentation template `templates/resources/elasticsearch_ccr_follower_index.md.tmpl` with description, example HCL, and import instructions (noting that `settings_raw` and `data_stream_name` must be added manually after import)
- [ ] 6.2 Create documentation template `templates/resources/elasticsearch_ccr_auto_follow_pattern.md.tmpl` with description, example HCL, and import instructions (noting that `settings_raw` must be added manually after import)
- [ ] 6.3 Run `make docs-generate` (or equivalent) to generate rendered docs pages under `docs/resources/`

## 7. Validation

- [ ] 7.1 Run `make build` — verifies compilation of new packages and provider registration
- [ ] 7.2 Run `go test ./internal/elasticsearch/ccr/...` and `go test ./internal/clients/elasticsearch/...` (unit tests only)
- [ ] 7.3 Run `make check-lint`
- [ ] 7.4 Run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate elasticsearch-ccr-resources --type change`
- [ ] 7.5 Run targeted acceptance tests for both new resources against a licensed trial cluster (requires `TF_ACC=1` and two configured Elasticsearch clusters with remote cluster connectivity)
