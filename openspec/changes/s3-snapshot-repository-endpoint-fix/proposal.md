## Why

When a user sets `endpoint` (or `path_style_access`) inside an `s3 {}` block on
`elasticstack_elasticsearch_snapshot_repository`, those fields are **silently absent** from the
`PUT /_snapshot/{name}` request body. Elasticsearch falls back to the default AWS S3 endpoint and
fails with `repository_verification_exception`. This regression was introduced by the Plugin SDK →
Plugin Framework migration (PR #2752) when the S3 write path was switched to the go-elasticsearch
typed `S3Repository` / `S3RepositorySettings` struct. That struct does not include `endpoint` or
`path_style_access`, and Go's `encoding/json` silently discards unknown fields on unmarshal, so both
values are dropped before the API call is made.

The existing codebase already contains an identical fix for HDFS: because
`types.HdfsRepository` is not part of the go-elasticsearch typed union, the HDFS case uses a raw
JSON bypass in `PutSnapshotRepository`. The same bypass is the correct fix for S3.

## What Changes

- **Write path fix**: In `internal/clients/elasticsearch/snapshot_repository.go`,
  replace the `case "s3":` typed-struct round-trip with the same raw JSON pattern used for HDFS.
  The settings map from `s3ToSettings` (which already includes `endpoint` and `path_style_access`
  correctly) is marshalled directly to JSON and sent via `.Raw(...)`.

- **Unit test**: Add `TestS3ToSettingsWithEndpoint` in
  `internal/elasticsearch/cluster/snapshot_repository/read_write_test.go` asserting that `endpoint`
  IS present in the settings map when set on `S3Settings`.

- **Read path investigation**: Verify whether `GET /_snapshot/{name}` returns `endpoint` in the
  settings object. If Elasticsearch does not echo `endpoint` back, the `endpoint` attribute needs
  `UseStateForUnknown` or equivalent so a second `apply` does not show a spurious diff.

- **Delta spec**: Amend `openspec/specs/elasticsearch-snapshot-repository/spec.md` to document
  the normative requirement that all S3 settings — including `endpoint` and `path_style_access` —
  MUST appear in the PUT request body, and to capture the raw-JSON bypass strategy for S3.

## Capabilities

### New Capabilities

- _(none)_

### Modified Capabilities

- **`elasticsearch-snapshot-repository`**: S3 write path sends all schema-defined settings
  (including `endpoint` and `path_style_access`) in the PUT request body; unit test coverage for
  the `endpoint` field; potential plan-modifier fix for `endpoint` read-back behaviour.

## Impact

- **Users**: S3-compatible endpoints (MinIO, Ceph, etc.) work correctly. Previously the `endpoint`
  field was silently dropped, causing `repository_verification_exception`.
- **Code**: `internal/clients/elasticsearch/snapshot_repository.go` (one `case` block replaced);
  `internal/elasticsearch/cluster/snapshot_repository/read_write_test.go` (one new test).
- **Maintenance**: S3 write path now consistent with HDFS; any future S3 schema fields are
  automatically preserved without needing to update the go-elasticsearch struct.
