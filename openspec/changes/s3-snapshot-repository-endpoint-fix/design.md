## Context

Canonical requirements for this resource live in
[`openspec/specs/elasticsearch-snapshot-repository/spec.md`](../../specs/elasticsearch-snapshot-repository/spec.md).
The write path lives in
[`internal/clients/elasticsearch/snapshot_repository.go`](../../../internal/clients/elasticsearch/snapshot_repository.go),
and the settings-building logic lives in
[`internal/elasticsearch/cluster/snapshot_repository/write.go`](../../../internal/elasticsearch/cluster/snapshot_repository/write.go).

### Root cause

`PutSnapshotRepository` (snapshot_repository.go lines 65–72) unmarshals the settings map into
`types.S3RepositorySettings` from go-elasticsearch v8. That struct omits `endpoint` and
`path_style_access` (confirmed by `go doc`). Go's `encoding/json` silently discards unknown fields
on unmarshal, so both values are lost before the Elasticsearch API call is made.

The `s3ToSettings` function (write.go line 198) correctly adds `endpoint` via `setIfNotEmpty`.
The bug is entirely in the consumer of that map, not in the producer.

### Fix strategy

Mirror the existing HDFS bypass (snapshot_repository.go lines 87–105):

```go
case "s3":
    body := map[string]any{
        "type":     repoType,
        "settings": settings,
    }
    bodyBytes, err := json.Marshal(body)
    if err != nil {
        return diagutil.FrameworkDiagFromError(err)
    }
    req := typedClient.Snapshot.CreateRepository(name).Raw(bytes.NewReader(bodyBytes)).Verify(verify)
    _, err = req.Do(ctx)
    if err != nil {
        return diagutil.FrameworkDiagFromError(err)
    }
    return nil
```

The `settings` map is the value returned by `s3ToSettings` and already contains all fields the
schema defines. Sending it as raw JSON bypasses the lossy unmarshal step entirely.

## Goals / Non-Goals

**Goals:**

- Ensure `endpoint` and `path_style_access` appear in the S3 PUT request body when set.
- Add a unit test asserting `endpoint` IS in the map when provided.
- Investigate and resolve `endpoint` read-back behaviour (plan drift prevention).

**Non-goals:**

- Fixing any non-S3 repository type (no evidence of similar bugs for `fs`, `url`, `gcs`, `azure`).
- Upgrading the go-elasticsearch dependency.
- Changing the schema structure or renaming any S3 block attributes.
- Fixing the HDFS `path` setting or any other unrelated issue.

## Decisions

- **Approach A (raw JSON bypass)** is adopted; Approach B (typed-API struct with manual raw-field
  merge) was evaluated by the research phase and rejected as overly complex and fragile. See research
  comment on issue #3434 for the full comparison.

- The `settingsCannedACL` / `settingsStorageClass` defaults (`"private"` / `"standard"`) are
  written conditionally via `setIfNotEmpty` by `s3ToSettings`, but because the schema provides
  defaults they are always non-empty in practice. The raw-JSON bypass preserves this behaviour
  unchanged; no schema default changes are needed.

- For `path_style_access`: `s3ToSettings` writes it unconditionally as a bool. The Elasticsearch
  GET response echoes back bool settings, so no plan modifier is needed for that field.

## Open questions

- **Does the Elasticsearch GET `/_snapshot/{name}` response for S3 return `endpoint` in the
  settings object?** If not, the state after a successful create will hold `null` for `endpoint`,
  causing a perpetual plan diff on subsequent applies. The raw-JSON GET overlay in
  `GetSnapshotRepository` (snapshot_repository.go lines 155–165) captures raw settings from the GET
  response, so if ES echoes `endpoint` back, it will be preserved. If ES treats `endpoint` as a
  write-only client-level setting (i.e. it is absent from GET), then `settingsToS3` must implement
  read-side state inheritance: when the API omits `endpoint`, the prior state value should be
  preserved instead of overwriting state with `null`. This mirrors the existing `compressFallback`
  pattern in `settingsToFs` and `settingsToURL`. This must be confirmed during implementation; the
  implementer should test a create-then-read cycle with a real S3-compatible endpoint or inspect
  the ES source.

- **Is `path_style_access` echoed by GET?** The debug trace in the issue does not show it in the
  request (evidence it was also being dropped by the typed struct). If ES echoes `false` back for
  `path_style_access`, the fix to the write path is sufficient. If ES omits it when `false`, a
  `UseStateForUnknown` or omit-on-false convention may be needed to avoid drift.

## Risks / Trade-offs

- **Loses typed-API compile-time field validation for S3**: The typed struct currently provides no
  useful validation (it was silently dropping valid fields), so this is no regression.
- **Consistent with HDFS pattern**: Reviewers familiar with the HDFS branch will find the fix
  immediately legible; no new pattern is introduced.
