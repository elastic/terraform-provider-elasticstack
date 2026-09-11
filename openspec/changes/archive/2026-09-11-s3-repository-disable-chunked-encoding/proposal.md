## Why

After Elasticsearch's `repository-s3` AWS SDK v2 migration (8.19+/9.1+), S3-compatible storage
backends that do not support chunked transfer encoding fail snapshot uploads with:

```
s3_exception: The provided 'x-amz-content-sha256' header does not match what was computed.
```

Elasticsearch exposes the fix as the S3 client setting `disable_chunked_encoding`
(`s3.client.CLIENT_NAME.disable_chunked_encoding`, "Planned for deprecation"). A related class of
S3-compatible incompatibility — plain-HTTP payload integrity — is addressed by the newer, non-
deprecated client setting `always_sign_requests` (GA since 9.5). Neither setting is exposed on the
`elasticstack_elasticsearch_snapshot_repository` `s3` block
(`internal/elasticsearch/snapshot/repository/schema.go`), so practitioners hitting this class of
error cannot fix it via Terraform; the only workaround is to configure the repository out-of-band
via the raw Elasticsearch API, defeating the point of managing it with this provider (#4875).

Both settings are documented as S3 **client** settings, not native repository settings, but
Elastic's docs confirm non-secure client settings may also be passed inside the repository's
`settings` object on `PUT _snapshot/{repo}` (merged with the named client; repository value wins).
This provider already uses exactly this mechanism to expose `endpoint` and `path_style_access` as
`s3` block attributes, so the same approach applies here.

## What Changes

- Add `disable_chunked_encoding` (bool, optional+computed, default `false`) to the `s3` repository
  type block, matching the existing `server_side_encryption`/`path_style_access` boolean-attribute
  pattern in this file.
- Add `always_sign_requests` (bool, optional+computed, default `false`) to the same block, for the
  same reason and by the same mechanism — confirmed in-scope alongside `disable_chunked_encoding`
  per maintainer direction on the issue.
- Wire both attributes through `S3Settings` (`models.go`), `s3AttrTypes()` (`attr_types.go`),
  `s3ToSettings` (`write.go`, unconditional like `server_side_encryption`/`path_style_access`), and
  `settingsToS3` (`read.go`).
- Apply the same read-side, prior-state fallback already used for `endpoint`/`path_style_access` in
  `settingsToS3`: since Elasticsearch's GET response is not documented to echo client-setting
  overrides back, and this cannot be verified from documentation alone, inherit the prior state
  value when the API omits the key, rather than risk `terraform plan` drift to `false`/`null` after
  every apply.
- Update the data source (`data_source.go`) to expose both attributes as computed, mirroring how
  `path_style_access`/`server_side_encryption` are already exposed there.
- Update `openspec/specs/elasticsearch-snapshot-repository/spec.md` with the new attributes and two
  new requirements (read/write mapping and drift prevention), following the existing REQ-016/REQ-017
  precedent for `endpoint`/`path_style_access`.
- Add acceptance test coverage extending the existing minio-backed (S3-compatible) `s3` fixtures.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `elasticsearch-snapshot-repository`: add `s3.disable_chunked_encoding` and
  `s3.always_sign_requests` to the Schema sketch; add REQ-019 (schema, write-path mapping) and
  REQ-020 (read-side drift prevention), following the REQ-016/REQ-017 pattern already documented
  for `endpoint`/`path_style_access`.

## Impact

- `internal/elasticsearch/snapshot/repository/constants.go` — add
  `settingDisableChunkedEncoding = "disable_chunked_encoding"` and
  `settingAlwaysSignRequests = "always_sign_requests"`.
- `internal/elasticsearch/snapshot/repository/schema.go` — add both `schema.BoolAttribute`s
  (`Optional: true, Computed: true, Default: booldefault.StaticBool(false)`) to `s3Block()`.
- `internal/elasticsearch/snapshot/repository/models.go` — add `DisableChunkedEncoding` and
  `AlwaysSignRequests` (`types.Bool`) to `S3Settings`.
- `internal/elasticsearch/snapshot/repository/attr_types.go` — add both to `s3AttrTypes()` as
  `types.BoolType`.
- `internal/elasticsearch/snapshot/repository/write.go` — set both unconditionally in
  `s3ToSettings`.
- `internal/elasticsearch/snapshot/repository/read.go` — extend `settingsToS3` with the same
  prior-state fallback pattern already used for `path_style_access`.
- `internal/elasticsearch/snapshot/repository/data_source.go` — expose both as computed attributes
  on the data source `s3` block.
- `internal/elasticsearch/snapshot/repository/snapshot_repository_acc_test.go` and
  `testdata/TestAccResourceSnapRepoS3/**` — extend existing minio-backed S3 acceptance fixtures.
- `openspec/specs/elasticsearch-snapshot-repository/spec.md` — schema sketch and REQ-019/REQ-020.
- Generated resource docs via `make docs-generate` (out of scope for this proposal-only change; the
  implementing PR runs it).

Not in scope: `unsafely_incompatible_with_s3_conditional_writes` and the generic S3 client-setting
passthrough map considered and rejected in issue research; the default value of either new
attribute (both remain `false`, matching Elasticsearch); any Elasticsearch server-side behavior
change.
