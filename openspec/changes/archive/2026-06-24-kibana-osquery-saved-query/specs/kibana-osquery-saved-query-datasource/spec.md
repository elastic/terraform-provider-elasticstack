## ADDED Requirements

### Requirement: Data source schema

The `elasticstack_kibana_osquery_saved_query` data source SHALL expose the following input attributes:

- `saved_query_id` — Required string; the ID of the query to look up
- `space_id` — Optional string, default `"default"`
- `kibana_connection` — Optional block (provider-level Kibana connection override)

And the following Computed output attributes (populated from the API response):

- `id` — Computed string; composite identifier in the format `<space_id>/<saved_query_id>`, used to uniquely identify the data source state entry across spaces
- `saved_object_id` — Computed string; Kibana saved object UUID used by the detail API
- `query` — Computed string
- `description` — Computed string
- `platform` — Computed SetAttribute of strings
- `interval` — Computed Int64
- `version` — Computed string
- `snapshot` — Computed bool
- `removed` — Computed bool
- `ecs_mapping` — Computed MapNestedAttribute (same element shape as the resource: `field`, `value`, `values`)
- `prebuilt` — Computed bool

#### Scenario: Required saved_query_id enforced
- **WHEN** the data source is configured without `saved_query_id`
- **THEN** Terraform SHALL reject the configuration with a validation error

### Requirement: Composite id

The data source SHALL expose a computed `id` in the format `<space_id>/<saved_query_id>` to uniquely identify the state entry across spaces.

#### Scenario: Computed id after read
- **WHEN** the data source reads a query with `saved_query_id = "list_all_processes"` in `space_id = "default"`
- **THEN** `id` SHALL equal `"default/list_all_processes"`

### Requirement: Read

The data source SHALL resolve `saved_object_id` from `saved_query_id` via `FindOsquerySavedObjectID`, then call `GET /api/osquery/saved_queries/{saved_object_id}` (space-aware via `SpaceAwarePathRequestEditor`) using `GetOsquerySavedQuery`. The API response wraps the entity in a `data` field; the kibanaoapi helper returns an unwrapped `OsquerySavedQueryGetEntity` for model mapping. When find returns no match or GET returns HTTP 404, the data source SHALL return an error diagnostic (data sources do not remove from state; they error on missing). On success, all Computed attributes SHALL be populated from the entity.

#### Scenario: Successful read of a user-managed query
- **WHEN** a data source with `saved_query_id = "list_all_processes"` is read
- **THEN** find + GET SHALL resolve the query by `saved_query_id` and populate all output attributes from the API response

#### Scenario: Query not found returns error
- **WHEN** the API returns HTTP 404
- **THEN** the data source SHALL return an error diagnostic indicating the query was not found

### Requirement: Prebuilt queries are readable by the data source

Unlike the resource, the data source SHALL NOT error when `prebuilt == true` in the API response. The data source is the recommended Terraform-native way to reference prebuilt queries (e.g., for use in `response_actions[].params.saved_query_id` of a detection rule). The `prebuilt` attribute SHALL be set to `true` in state when the query is prebuilt.

#### Scenario: Reading a prebuilt query
- **WHEN** the API returns `prebuilt: true` for the requested query
- **THEN** the data source SHALL populate all available attributes including `prebuilt = true`
- **AND** no error SHALL be raised

#### Scenario: Reading a user-managed query
- **WHEN** the API returns `prebuilt: false` for the requested query
- **THEN** the data source SHALL populate all available attributes including `prebuilt = false`

### Requirement: ECS mapping normalisation

The data source SHALL apply the same `ecs_mapping` read normalisation as the resource: the API `Value` field SHALL be inspected for string vs array type to populate either `value` or `values` in state. The `ecs_mapping` attribute in the data source is Computed only (no config input).

#### Scenario: ECS mapping populated from API response
- **WHEN** the API response contains `ecs_mapping: { "process.name": { field: "cmdline" } }`
- **THEN** state SHALL contain `ecs_mapping = { "process.name" = { field = "cmdline", value = null, values = null } }`

### Requirement: `interval` and `version` union-type normalisation

The data source SHALL apply the same union-type normalisation as the resource: `interval` is stored as Computed Int64, `version` is stored as Computed string, using the same accessor logic.

#### Scenario: interval populated as Int64
- **WHEN** the API response contains `interval: 3600` or `interval: "3600"`
- **THEN** state SHALL contain `interval = 3600`

#### Scenario: version populated as string
- **WHEN** the API response contains `version: "1.0.0"` or `version: 1`
- **THEN** state SHALL contain `version` as a non-empty string

### Requirement: Connection override and version gating

The data source SHALL obtain its Kibana client via the data source-level `kibana_connection` block when provided, otherwise via the provider-level Kibana configuration. Space-aware requests SHALL use `space_id` via `kibanautil.SpaceAwarePathRequestEditor`. The data source model (or shared base) SHALL implement `GetVersionRequirements` with the same `8.5.0` documented/conservative minimum as the resource.

#### Scenario: Data source kibana_connection override
- **WHEN** `kibana_connection` is configured on the data source
- **THEN** the GET call SHALL use that connection instead of the provider-level Kibana connection

#### Scenario: Pre-minimum Kibana version
- **WHEN** the data source is read against a Kibana version older than `8.5.0`
- **THEN** Terraform SHALL fail with an error message stating the minimum required version
