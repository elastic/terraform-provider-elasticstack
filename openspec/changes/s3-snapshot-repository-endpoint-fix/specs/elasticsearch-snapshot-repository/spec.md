## ADDED Requirements

### Requirement: S3 write path uses raw JSON to preserve all settings (REQ-016)

The `PutSnapshotRepository` implementation SHALL use a raw JSON request body for the `"s3"`
repository type, mirroring the existing HDFS bypass. It SHALL marshal a
`map[string]any{"type": "s3", "settings": <settings map>}` to JSON and pass the result to
`Snapshot.CreateRepository(name).Raw(...)`. The `S3Repository` / `S3RepositorySettings` typed
struct from go-elasticsearch SHALL NOT be used for the PUT path, because that struct omits
`endpoint` and `path_style_access`, causing Go's `encoding/json` to silently discard those fields
on unmarshal.

The settings map is produced by `s3ToSettings` which already correctly includes `endpoint` (via
`setIfNotEmpty`) and `path_style_access` (unconditionally). Sending this map as raw JSON ensures
all schema-defined settings reach Elasticsearch without loss.

#### Scenario: endpoint present in PUT body when set

- GIVEN an `s3` repository block with `endpoint` set to a non-empty URL
- WHEN the resource creates or updates the repository
- THEN the PUT `/_snapshot/{name}` request body SHALL include the `endpoint` key in `settings`
  with the configured value

#### Scenario: path_style_access present in PUT body

- GIVEN an `s3` repository block (with or without an explicit `path_style_access` value)
- WHEN the resource creates or updates the repository
- THEN the PUT `/_snapshot/{name}` request body SHALL include the `path_style_access` key in
  `settings`

#### Scenario: other S3 settings unaffected

- GIVEN an `s3` repository block with `bucket`, `client`, `canned_acl`, and `storage_class` set
- WHEN the resource creates or updates the repository
- THEN all those fields SHALL appear in the PUT request body `settings`, unchanged from their
  configured values

### Requirement: S3 endpoint plan drift prevention (REQ-017)

The implementer SHALL determine whether the Elasticsearch GET `/_snapshot/{name}` response
returns `endpoint` in the S3 settings object.

- If the GET response **does** return `endpoint`, no schema change is required and state will
  round-trip correctly.
- If the GET response **does not** return `endpoint` (i.e. Elasticsearch treats it as a
  write-only client-level setting), the `endpoint` attribute in the S3 block schema SHALL carry
  a `UseStateForUnknown` plan modifier so that a second `apply` on an unchanged configuration
  does not produce a spurious diff showing `endpoint` changing from its configured value to
  `null`.

The decision and its rationale SHALL be documented in the implementing PR description.

#### Scenario: No spurious diff after second apply

- GIVEN a repository created with `endpoint` set
- WHEN `terraform plan` is run again on the same unchanged configuration
- THEN the plan SHALL show no changes for the `endpoint` attribute

## AMENDED Requirements

### Requirement: Create and update (REQ-010) — S3 amendment

The existing REQ-010 requirement for create and update is amended for the `"s3"` repository type:
the provider SHALL use the raw JSON bypass (as specified in REQ-016) rather than the
`types.S3Repository` typed struct, to ensure `endpoint` and `path_style_access` are not silently
discarded. All other repository types (fs, url, gcs, azure, source) continue to use their
respective typed structs. The HDFS type continues to use its existing raw JSON bypass unchanged.
