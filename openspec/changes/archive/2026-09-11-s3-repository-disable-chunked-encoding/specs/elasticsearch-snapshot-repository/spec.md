## ADDED Requirements

### Requirement: S3 disable_chunked_encoding and always_sign_requests schema and write mapping (REQ-019)

The `s3` repository type block SHALL expose `disable_chunked_encoding` and `always_sign_requests`
as `Optional + Computed` boolean attributes with `Default: booldefault.StaticBool(false)`, matching
the existing `server_side_encryption`/`path_style_access` attribute pattern. Both settings are
Elasticsearch S3 **client** settings (`s3.client.CLIENT_NAME.disable_chunked_encoding`,
`s3.client.CLIENT_NAME.always_sign_requests`) exposed on this resource via the same
"overriding client settings in repository settings" mechanism already used for `endpoint` and
`path_style_access`.

The write path (`s3ToSettings`) SHALL include both keys unconditionally in the settings map sent
to the Put Snapshot Repository API, regardless of whether the configured/defaulted value is
`true` or `false`, matching `server_side_encryption`/`path_style_access`. Because the `s3`
repository type already uses the raw-JSON write path (REQ-016), both keys reach Elasticsearch
without being silently discarded.

The data source `s3` block SHALL expose both attributes as computed.

#### Scenario: disable_chunked_encoding present in PUT body

- GIVEN an `s3` repository block with `disable_chunked_encoding` set to `true`
- WHEN the resource creates or updates the repository
- THEN the PUT `/_snapshot/{name}` request body SHALL include `"disable_chunked_encoding": true`
  in `settings`

#### Scenario: always_sign_requests present in PUT body

- GIVEN an `s3` repository block with `always_sign_requests` set to `true`
- WHEN the resource creates or updates the repository
- THEN the PUT `/_snapshot/{name}` request body SHALL include `"always_sign_requests": true` in
  `settings`

#### Scenario: Defaults omit no keys

- GIVEN an `s3` repository block that omits both `disable_chunked_encoding` and
  `always_sign_requests`
- WHEN the resource creates or updates the repository
- THEN the PUT request body `settings` SHALL include both keys with value `false`

#### Scenario: Data source exposes both attributes

- GIVEN an S3 snapshot repository with `disable_chunked_encoding` and `always_sign_requests` set
- WHEN the `elasticstack_elasticsearch_snapshot_repository` data source reads the repository
- THEN both attributes SHALL be populated on the data source `s3` block

### Requirement: S3 disable_chunked_encoding and always_sign_requests read-side drift prevention (REQ-020)

The implementer SHALL determine whether the Elasticsearch GET `/_snapshot/{name}` response
returns `disable_chunked_encoding` and `always_sign_requests` in the S3 settings object.

- If the GET response **does** return a field, no read-side fallback is required for that field
  and it SHALL be mapped directly with `boolSetting(s, settingKey, false)`.
- If the GET response **does not** return a field, `settingsToS3` SHALL implement the same
  read-side state inheritance already applied to `path_style_access` under REQ-017: when the API
  omits the field, the prior state value SHALL be preserved in the refreshed state rather than
  overwritten with the schema default (`false`).

The decision and its rationale SHALL be documented in the implementing PR description, consistent
with REQ-017.

#### Scenario: No spurious diff for disable_chunked_encoding after second apply

- GIVEN a repository created with `disable_chunked_encoding = true`
- WHEN `terraform plan` is run again on the same unchanged configuration
- THEN the plan SHALL show no changes for the `disable_chunked_encoding` attribute

#### Scenario: No spurious diff for always_sign_requests after second apply

- GIVEN a repository created with `always_sign_requests = true`
- WHEN `terraform plan` is run again on the same unchanged configuration
- THEN the plan SHALL show no changes for the `always_sign_requests` attribute
