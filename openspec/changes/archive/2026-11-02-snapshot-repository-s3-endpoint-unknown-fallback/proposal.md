## Why

Since v0.16.1, applying an `elasticstack_elasticsearch_snapshot_repository` with an `s3` block that does not set `endpoint` fails on create and on every update:

```
Error: Provider returned invalid result object after apply
After the apply operation, the provider still indicated an unknown value for
elasticstack_elasticsearch_snapshot_repository.example.s3.endpoint.
```

Elasticsearch registers the repository, but Terraform taints the resource. Every later apply replaces it and fails the same way, and resources that depend on the repository are never created. Leaving `endpoint` unset is the normal configuration for AWS S3, so most S3 repositories are affected.

The read-side state inheritance added for `endpoint` under REQ-017 carries the prior value into state whenever Elasticsearch omits `endpoint`. During the read that follows create or update, the prior value is the planned value. For an unconfigured Optional + Computed attribute the planned value is unknown, so the unknown value reaches state. REQ-017 does not say what happens when the prior value is unknown.

## What Changes

- The S3 `endpoint` read-back inherits the prior value only when that value is known. When Elasticsearch omits `endpoint` and the prior value is unknown, `endpoint` is null in state.
- Known prior values keep today's REQ-017 behavior: a configured `endpoint` that Elasticsearch does not echo is still preserved in state, with no spurious diff on the next plan.
- Create and update of an S3 repository without `endpoint` succeed, and `s3.endpoint` is null in state.

Non-goals:

- No schema change to `s3.endpoint`. It stays Optional + Computed with no plan modifier.
- No change to the read-side handling of `path_style_access` or of the common settings (`chunk_size`, `max_snapshot_bytes_per_sec`, `max_restore_bytes_per_sec`); they already ignore unknown prior values.
- Removing a previously set `endpoint` from configuration on its own still leaves the prior value in state and in Elasticsearch (Optional + Computed semantics). Making that removal clear the value is out of scope. When the removal is applied together with another change, the apply now succeeds with `endpoint` null instead of failing.
- No change to the `elasticstack_elasticsearch_snapshot_repository` data source, which has no `endpoint` attribute.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `elasticsearch-snapshot-repository`: REQ-017 (S3 endpoint plan drift prevention) states that read-side inheritance of `endpoint` applies only to a known prior value, and that an unknown prior value with no `endpoint` in the GET response maps to null. It also records the existing `path_style_access` behavior (an unknown prior value maps to the schema default) so both inherited fields follow one stated rule; that behavior does not change.

## Impact

- `internal/elasticsearch/snapshot/repository/read.go`: S3 settings read-back.
- `internal/elasticsearch/snapshot/repository/read_write_test.go`: unit coverage for an unknown prior `endpoint`.
- `internal/elasticsearch/snapshot/repository/snapshot_repository_acc_test.go` and new testdata: acceptance coverage for an S3 repository without `endpoint`, on create, re-plan and update.
- `openspec/specs/elasticsearch-snapshot-repository/spec.md`: REQ-017.
- Users: no migration. A configuration that sets `endpoint` only as a workaround can keep it. Removing it on its own is a no-op; the value is cleared on the next apply that also changes another attribute of the repository.
