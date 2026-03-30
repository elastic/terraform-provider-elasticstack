# `elasticstack_elasticsearch_snapshot_repository` — Schema and Functional Requirements

Resource implementation: `internal/elasticsearch/cluster/snapshot_repository.go`

Data source implementation: `internal/elasticsearch/cluster/snapshot_repository_data_source.go`

## Purpose

Define schema and behavior for the Elasticsearch snapshot repository resource and data source: API usage, identity and import, connection, lifecycle, create/update/read/delete semantics, mapping of repository type settings, and data source read-only semantics.

## Schema

### Resource

```hcl
resource "elasticstack_elasticsearch_snapshot_repository" "example" {
  id     = <computed, string>  # internal identifier: <cluster_uuid>/<repo_name>
  name   = <required, string>  # force new
  verify = <optional, bool>    # default: true

  # Exactly one of: fs, url, gcs, azure, s3, hdfs
  # Changing repository type forces replacement

  fs {
    # type-specific
    location = <required, string>
    # common settings
    chunk_size                 = <optional, string>
    compress                   = <optional, bool>    # default: true
    max_snapshot_bytes_per_sec = <optional, string>  # default: "40mb"
    max_restore_bytes_per_sec  = <optional, string>
    readonly                   = <optional, bool>    # default: false
    # standard settings
    max_number_of_snapshots    = <optional, int>     # default: 500; min: 1
  }

  url {
    # type-specific
    url                 = <required, string>  # must start with: file:, ftp:, http:, https:, jar:
    http_max_retries    = <optional, int>     # default: 5; min: 0
    http_socket_timeout = <optional, string>  # default: "50s"
    # common settings
    chunk_size                 = <optional, string>
    compress                   = <optional, bool>    # default: true
    max_snapshot_bytes_per_sec = <optional, string>  # default: "40mb"
    max_restore_bytes_per_sec  = <optional, string>
    readonly                   = <optional, bool>    # default: false
    # standard settings
    max_number_of_snapshots    = <optional, int>     # default: 500; min: 1
  }

  gcs {
    # type-specific
    bucket    = <required, string>
    client    = <optional, string>  # default: "default"
    base_path = <optional, computed, string>
    # common settings
    chunk_size                 = <optional, string>
    compress                   = <optional, bool>    # default: true
    max_snapshot_bytes_per_sec = <optional, string>  # default: "40mb"
    max_restore_bytes_per_sec  = <optional, string>
    readonly                   = <optional, bool>    # default: false
  }

  azure {
    # type-specific
    container     = <required, string>
    client        = <optional, string>  # default: "default"
    base_path     = <optional, computed, string>
    location_mode = <optional, string>  # default: "primary_only"; one of: primary_only, secondary_only
    # common settings
    chunk_size                 = <optional, string>
    compress                   = <optional, bool>    # default: true
    max_snapshot_bytes_per_sec = <optional, string>  # default: "40mb"
    max_restore_bytes_per_sec  = <optional, string>
    readonly                   = <optional, bool>    # default: false
  }

  s3 {
    # type-specific
    bucket                = <required, string>
    endpoint              = <optional, computed, string>  # must be a valid http:// or https:// URL when set
    client                = <optional, string>            # default: "default"
    base_path             = <optional, computed, string>
    server_side_encryption = <optional, bool>             # default: false
    buffer_size           = <optional, computed, string>
    canned_acl            = <optional, string>            # default: "private"
    storage_class         = <optional, string>            # default: "standard"
    path_style_access     = <optional, bool>              # default: false
    # common settings
    chunk_size                 = <optional, string>
    compress                   = <optional, bool>    # default: true
    max_snapshot_bytes_per_sec = <optional, string>  # default: "40mb"
    max_restore_bytes_per_sec  = <optional, string>
    readonly                   = <optional, bool>    # default: false
  }

  hdfs {
    # type-specific
    uri           = <required, string>
    path          = <required, string>
    load_defaults = <optional, bool>  # default: true
    # common settings
    chunk_size                 = <optional, string>
    compress                   = <optional, bool>    # default: true
    max_snapshot_bytes_per_sec = <optional, string>  # default: "40mb"
    max_restore_bytes_per_sec  = <optional, string>
    readonly                   = <optional, bool>    # default: false
  }

  elasticsearch_connection {
    endpoints                = <optional, list(string)>
    username                 = <optional, string>
    password                 = <optional, string>
    api_key                  = <optional, string>
    bearer_token             = <optional, string>
    es_client_authentication = <optional, string>
    insecure                 = <optional, bool>
    headers                  = <optional, map(string)>
    ca_file                  = <optional, string>
    ca_data                  = <optional, string>
    cert_file                = <optional, string>
    key_file                 = <optional, string>
    cert_data                = <optional, string>
    key_data                 = <optional, string>
  }
}
```

### Data source

```hcl
data "elasticstack_elasticsearch_snapshot_repository" "example" {
  id   = <computed, string>  # internal identifier: <cluster_uuid>/<repo_name>
  name = <required, string>
  type = <computed, string>  # repository type returned by the API

  # At most one block is populated, corresponding to the repository type
  fs    { <all settings computed> }
  url   { <all settings computed> }
  gcs   { <all settings computed> }
  azure { <all settings computed> }
  s3    { <all settings computed> }
  hdfs  { <all settings computed> }

  elasticsearch_connection {
    # same fields as resource, all optional
  }
}
```

## Requirements

### Requirement: Snapshot repository CRUD APIs (REQ-001–REQ-004)

The resource SHALL use the Elasticsearch Create or Update Snapshot Repository API (`PUT /_snapshot/<repository>`) to create and update snapshot repositories ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/put-snapshot-repo-api.html)). The resource SHALL use the Elasticsearch Get Snapshot Repository API (`GET /_snapshot/<repository>`) to read snapshot repositories ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/get-snapshot-repo-api.html)). The resource SHALL use the Elasticsearch Delete Snapshot Repository API (`DELETE /_snapshot/<repository>`) to delete snapshot repositories ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/delete-snapshot-repo-api.html)). When Elasticsearch returns a non-success status for any create, update, read, or delete request (other than 404 on read), the resource SHALL surface the API error to Terraform diagnostics.

#### Scenario: API error surfaced

- GIVEN a failing Elasticsearch response (non-404 on read, or any error on write/delete)
- WHEN the provider processes the response
- THEN diagnostics SHALL include the API error message

### Requirement: Identity (REQ-005)

The resource SHALL expose a computed `id` in the format `<cluster_uuid>/<repository_name>`. During create and update, the resource SHALL compute `id` from the current cluster UUID and the configured `name` by calling `client.ID(ctx, repoID)`, then persist `id` to state.

#### Scenario: Id set after create

- GIVEN a successful create
- WHEN the resource is stored in state
- THEN `id` SHALL be `<cluster_uuid>/<name>`

### Requirement: Import (REQ-006)

The resource SHALL support import via `schema.ImportStatePassthroughContext`, persisting the imported `id` directly to state. On subsequent read or delete, the resource SHALL parse the `id` using `CompositeIDFromStr`, which requires the format `<cluster_uuid>/<repository_name>`; if the format is invalid (not exactly two slash-separated parts), the resource SHALL return an error diagnostic and SHALL not call the Elasticsearch API.

#### Scenario: Import by composite id

- GIVEN an existing snapshot repository with a known composite id
- WHEN `terraform import` is run with `<cluster_uuid>/<name>`
- THEN the resource SHALL be added to state with `id` set to the provided value

#### Scenario: Invalid id on read

- GIVEN a malformed `id` in state (missing or extra slash-delimited parts)
- WHEN read or delete runs
- THEN the resource SHALL return an error diagnostic and SHALL not call the Elasticsearch API

### Requirement: Lifecycle — repository type forces replacement (REQ-007)

The `name` attribute SHALL be marked `ForceNew`; changing `name` SHALL require resource replacement. Each repository type block (`fs`, `url`, `gcs`, `azure`, `s3`, `hdfs`) SHALL be marked `ForceNew`; configuring a different repository type SHALL require resource replacement.

#### Scenario: ForceNew on name change

- GIVEN an existing snapshot repository
- WHEN `name` is changed in configuration
- THEN Terraform plan SHALL require replacement

#### Scenario: ForceNew on type change

- GIVEN a repository of type `fs`
- WHEN the configuration changes to use `s3`
- THEN Terraform plan SHALL require replacement

### Requirement: Exactly one repository type (REQ-008)

The schema SHALL enforce `ExactlyOneOf` on the set `{fs, url, gcs, azure, s3, hdfs}`. Each type block SHALL also be in `ConflictsWith` for every other type block. The resource SHALL return a validation error if more than one or none of these blocks is configured.

#### Scenario: Two types configured

- GIVEN a configuration with both `fs` and `s3` blocks
- WHEN Terraform validates the configuration
- THEN a validation error SHALL be returned

### Requirement: Connection (REQ-009)

By default, the resource SHALL use the provider-level Elasticsearch client. When an `elasticsearch_connection` block is configured, the resource SHALL construct and use a resource-scoped Elasticsearch client for all API calls, via `clients.NewAPIClientFromSDKResource`.

#### Scenario: Resource-level connection

- GIVEN an `elasticsearch_connection` block is configured
- WHEN the provider executes any CRUD operation
- THEN the resource-scoped client SHALL be used for that operation

### Requirement: Create and update (REQ-010)

On create and update, the resource SHALL inspect the schema to determine which repository type block is present, flatten its settings into a `map[string]any`, and set them as the `Settings` field of the `models.SnapshotRepository` request body. The resource SHALL set the `Type` field to the name of the detected type block. The resource SHALL marshal the request body and call `PutSnapshotRepository`. After a successful Put, the resource SHALL set `id` in state and call the read function to refresh all attributes.

#### Scenario: Settings sent for fs repository

- GIVEN an `fs` block with `location`, `compress`, and `max_restore_bytes_per_sec`
- WHEN create runs
- THEN the Put API request body SHALL include `type: "fs"` and those settings in `settings`

### Requirement: Read and not-found handling (REQ-011)

On read, the resource SHALL parse `id` via `CompositeIDFromStr` to extract the repository name, then call `GetSnapshotRepository`. When the API returns a 404 (not found), the resource SHALL remove itself from state by setting `id` to empty string and SHALL return without error. When the API returns a repository whose type is not in the supported set `{fs, url, gcs, azure, s3, hdfs}`, the resource SHALL return an error diagnostic with summary "API responded with unsupported type of the snapshot repository." and SHALL not update state. On successful read, the resource SHALL update the `name` attribute and the relevant type-block attributes from the API response.

#### Scenario: Repository deleted out-of-band

- GIVEN the repository was deleted directly in Elasticsearch
- WHEN read runs
- THEN the resource SHALL be removed from state with no error

#### Scenario: Unsupported repository type returned

- GIVEN the API returns a repository type not in the supported set
- WHEN read runs
- THEN an error diagnostic SHALL be returned

### Requirement: Delete (REQ-012)

On delete, the resource SHALL parse `id` via `CompositeIDFromStr`, then call `DeleteSnapshotRepository` with the extracted repository name. Non-success API responses SHALL be surfaced as error diagnostics.

#### Scenario: Destroy calls Delete API

- GIVEN an existing snapshot repository in state
- WHEN `terraform destroy` runs
- THEN the Delete Snapshot Repository API SHALL be called with the repository name

### Requirement: Settings mapping — expand and flatten (REQ-013)

When building the API request body (create/update), the resource SHALL skip any settings value for which `IsEmpty` returns true (i.e., only non-empty settings values SHALL be included in the request). When reading the API response (flatten), the resource SHALL map only settings keys that appear in the type-specific schema; unknown keys SHALL be silently ignored. For `TypeInt` and `TypeFloat` schema fields, the resource SHALL parse the string value returned by the API using `strconv.Atoi`; if parsing fails, the resource SHALL return an error diagnostic. For `TypeBool` schema fields, the resource SHALL parse the string value using `strconv.ParseBool`; if parsing fails, the resource SHALL return an error diagnostic. String-typed schema fields SHALL be stored as-is from the API response.

#### Scenario: Int field parse error

- GIVEN the API returns a non-numeric string for an integer setting
- WHEN read runs and flattening is attempted
- THEN an error diagnostic with summary "Unable to parse snapshot repository settings." SHALL be returned

### Requirement: URL validation (REQ-014)

The `url.url` attribute SHALL be validated against the regular expression `^(file:|ftp:|http:|https:|jar:)`. The `s3.endpoint` attribute SHALL be validated as a valid HTTP or HTTPS URL (scheme must be `http` or `https`, host must be present); empty string SHALL be treated as valid (no endpoint override).

#### Scenario: Invalid URL scheme

- GIVEN `url.url` set to `"sftp://example.com"`
- WHEN Terraform validates the configuration
- THEN a validation error SHALL be returned

#### Scenario: Invalid S3 endpoint

- GIVEN `s3.endpoint` set to `"not-a-url"`
- WHEN Terraform validates the configuration
- THEN a validation error SHALL be returned

### Requirement: Verify parameter (REQ-015)

When `verify` is set to `true` (the default), the resource SHALL include `verify: true` in the request body sent to the Put Snapshot Repository API, causing Elasticsearch to verify the repository is functional. The `verify` attribute is read from the configuration but is not returned by the Get API and therefore SHALL not be updated from the API response on read.

#### Scenario: Verify defaults to true

- GIVEN no explicit `verify` configuration
- WHEN create runs
- THEN the API request body SHALL include `"verify": true`

---

## Data source requirements

### Requirement: Data source read-only semantics (REQ-DS-001)

The data source SHALL support only a read operation. It SHALL NOT perform create, update, or delete operations.

### Requirement: Data source API (REQ-DS-002)

The data source SHALL use the Elasticsearch Get Snapshot Repository API (`GET /_snapshot/<repository>`) to fetch the repository identified by `name`. When the API returns a non-success status, the data source SHALL surface the API error to Terraform diagnostics. When the repository is not found (API returns `nil` with no error), the data source SHALL set `id`, return a warning diagnostic with the message "Could not find snapshot repository [<name>]", and SHALL not attempt to populate type block attributes.

#### Scenario: Repository not found

- GIVEN no repository with the requested name exists
- WHEN the data source is read
- THEN a warning diagnostic SHALL be returned and type block attributes SHALL remain empty

### Requirement: Data source identity (REQ-DS-003)

The data source SHALL set `id` in the format `<cluster_uuid>/<repository_name>` by calling `client.ID(ctx, repoName)` after resolving the client. The `id` SHALL be set regardless of whether the repository was found.

### Requirement: Data source connection (REQ-DS-004)

By default, the data source SHALL use the provider-level Elasticsearch client. When an `elasticsearch_connection` block is configured, the data source SHALL construct and use a resource-scoped client via `clients.NewAPIClientFromSDKResource`.

### Requirement: Data source type block population (REQ-DS-005)

After a successful read, the data source SHALL set the `type` attribute to the repository type string returned by the API. The data source SHALL populate only the type block corresponding to the returned type; all other type blocks SHALL remain empty. The data source SHALL flatten settings from the API response using the same type conversion logic as the resource (string-to-int, string-to-bool, string-as-string). If the `type` returned by the API does not match any of the supported type block names in the schema, the data source SHALL return an error diagnostic.

#### Scenario: GCS repository

- GIVEN a GCS snapshot repository exists in Elasticsearch
- WHEN the data source is read
- THEN `type` SHALL be `"gcs"`, the `gcs` block SHALL be populated with the repository settings, and all other type blocks SHALL remain empty

### Requirement: Data source schema — computed attributes (REQ-DS-006)

All attributes in the data source schema except `name` SHALL be computed. The `name` attribute SHALL be required. The data source schema does NOT include `max_number_of_snapshots` for the `gcs`, `azure`, `s3`, and `hdfs` type blocks; only `fs` and `url` merge `commonStdSettings` and therefore include that attribute. The S3 type block in the data source SHALL NOT include the `endpoint` attribute.

#### Scenario: Name is required

- GIVEN no `name` is provided in the data source configuration
- WHEN Terraform validates the configuration
- THEN a validation error SHALL be returned
