## 1. Constants and models

- [x] 1.1 In `internal/elasticsearch/snapshot/repository/constants.go`, add
  `settingDisableChunkedEncoding = "disable_chunked_encoding"` and
  `settingAlwaysSignRequests = "always_sign_requests"`.
- [x] 1.2 In `internal/elasticsearch/snapshot/repository/models.go`, add `DisableChunkedEncoding
  types.Bool` and `AlwaysSignRequests types.Bool` to `S3Settings`, with `tfsdk` tags matching the
  new setting keys.
- [x] 1.3 In `internal/elasticsearch/snapshot/repository/attr_types.go`, add both new fields to
  `s3AttrTypes()` as `types.BoolType`.

## 2. Schema

- [x] 2.1 In `internal/elasticsearch/snapshot/repository/schema.go` `s3Block()`, add
  `settingDisableChunkedEncoding` and `settingAlwaysSignRequests` as
  `schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false)}`,
  matching `settingServerSideEncryption`/`settingPathStyleAccess`.
- [x] 2.2 In `internal/elasticsearch/snapshot/repository/data_source.go`, add both attributes to
  the data source `s3` block as computed-only, matching how `path_style_access`/
  `server_side_encryption` are exposed there.

## 3. Write path

- [x] 3.1 In `internal/elasticsearch/snapshot/repository/write.go` `s3ToSettings`, set
  `settingDisableChunkedEncoding: s3.DisableChunkedEncoding.ValueBool()` and
  `settingAlwaysSignRequests: s3.AlwaysSignRequests.ValueBool()` unconditionally in the returned
  map, matching `settingServerSideEncryption`/`settingPathStyleAccess`.

## 4. Read path and drift verification

- [x] 4.1 Attempt to determine empirically (against a live Elasticsearch cluster, e.g. via
  `TestAccResourceSnapRepoS3` or manual `GET _snapshot/{repo}`) whether the API echoes back
  `disable_chunked_encoding` and `always_sign_requests` once set via the repository `settings`
  override mechanism. Record the outcome in the PR description (see design.md Open questions).
- [x] 4.2 If the API does **not** echo the fields back: in
  `internal/elasticsearch/snapshot/repository/read.go` `settingsToS3`, extend the existing
  prior-state fallback pattern (currently applied to `endpoint`/`path_style_access`) to
  `disable_chunked_encoding` and `always_sign_requests` — capture each fallback from
  `state.S3`/`priorS3` before mapping, and use `boolSetting(s, settingX, xFallback)`.
- [x] 4.3 If the API **does** echo the fields back: map them directly with
  `types.BoolValue(boolSetting(s, settingX, false))` (no fallback needed), and note in the PR
  description that the fallback was unnecessary and why.
- [x] 4.4 Update the code comment above the existing `endpoint`/`path_style_access` fallback logic
  in `settingsToS3` to reflect whichever of 4.2/4.3 applies to the two new fields, so the comment
  stays accurate about which S3 attributes need fallback and why.

## 5. Tests

- [x] 5.1 Extend `internal/elasticsearch/snapshot/repository/snapshot_repository_acc_test.go` and
  the relevant `testdata/TestAccResourceSnapRepoS3/**/main.tf` fixtures (already minio-backed,
  i.e. S3-compatible) to set `disable_chunked_encoding = true` and `always_sign_requests = true`
  on at least one step, and assert both round-trip through state on a second `terraform plan`
  with no diff.
- [x] 5.2 Add/extend unit test coverage in `read_write_test.go` (or the nearest existing
  unit-test file covering `s3ToSettings`/`settingsToS3`) for both new fields, including the
  read-side fallback behavior decided in section 4.
- [x] 5.3 Run `go test ./internal/elasticsearch/snapshot/repository/...` (no `TF_ACC`) and `go vet
  ./internal/elasticsearch/snapshot/repository/...`.

## 6. Documentation and spec sync

- [x] 6.1 Run `make docs-generate` and commit the resulting resource/data-source doc updates for
  the two new `s3` block attributes.
- [x] 6.2 Sync this change's delta spec into
  `openspec/specs/elasticsearch-snapshot-repository/spec.md`: update the `s3` block Schema HCL
  sketch, and add REQ-019/REQ-020 (or renumber to the next free REQ id if other changes have
  landed first).
- [x] 6.3 Run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate
  s3-repository-disable-chunked-encoding --type change` and resolve any reported issues.
- [x] 6.4 Run `make check-openspec` after implementation to confirm specs and code stay in sync.
