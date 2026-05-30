## 1. Write path fix

- [x] 1.1 In `internal/clients/elasticsearch/snapshot_repository.go`, replace the `case "s3":` block
  (lines 65–72) with the raw-JSON bypass pattern used for HDFS (lines 87–105): marshal a
  `map[string]any{"type": repoType, "settings": settings}` and call
  `typedClient.Snapshot.CreateRepository(name).Raw(bytes.NewReader(bodyBytes)).Verify(verify).Do(ctx)`.

## 2. Unit tests

- [x] 2.1 In `internal/elasticsearch/cluster/snapshot_repository/read_write_test.go`, add
  `TestS3ToSettingsWithEndpoint`: construct an `S3Settings` with `Endpoint` set to a non-empty
  value and assert `require.Contains(t, m, "endpoint")` with the correct value. Mirror the
  structure of the existing `TestS3ToSettingsWithDefaults` test.

## 3. Read-back / plan drift investigation

- [x] 3.1 Determine whether the Elasticsearch GET `/_snapshot/{name}` response returns `endpoint`
  in the settings object for S3 repositories. Document the finding inline as a comment or in the
  PR description.
- [x] 3.2 If the GET response does NOT return `endpoint`, add read-side state inheritance to
  `settingsToS3` in `read.go`: pass `state Data` (already available from the existing signature
  pattern) and, when `StrSettingNull(s, settingEndpoint)` is null and the prior state `S3` block
  is non-null, preserve the state value of `endpoint` instead of overwriting it with null. Mirror
  the `compressFallback` pattern in `settingsToFs` and `settingsToURL`.
- [x] 3.3 If the GET response does NOT return `path_style_access` when it is `false`, consider
  whether the same plan-modifier treatment is required for that attribute.

## 4. Acceptance test

- [x] 4.1 Add or update an acceptance test in
  `internal/elasticsearch/cluster/snapshot_repository/` that sets `endpoint` on an S3-type
  repository (using a mock or real S3-compatible target if available in CI) and asserts that the
  attribute value is preserved in state after apply. If a live S3-compatible target is not
  available in CI, add a test that at minimum verifies the resource can be planned and that
  `endpoint` appears in the plan output.

## 5. Delta spec

- [ ] 5.1 Keep the delta spec under
  `openspec/changes/s3-snapshot-repository-endpoint-fix/specs/elasticsearch-snapshot-repository/spec.md`
  aligned with the implementation once tasks 1–3 are complete.
- [ ] 5.2 After merge: sync into `openspec/specs/elasticsearch-snapshot-repository/spec.md` or
  archive this change per project workflow; run `make check-openspec`.
