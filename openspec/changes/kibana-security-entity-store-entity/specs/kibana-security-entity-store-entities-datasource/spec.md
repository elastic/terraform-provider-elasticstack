## ADDED Requirements

### Requirement: Data source exposes the full Entity Store list/search endpoint (REQ-ENT-001)

The system SHALL implement a Terraform data source `elasticstack_kibana_security_entity_store_entities` that calls `GET /api/security/entity_store/entities` and returns the full result set as a normalized JSON string, supporting both page-based and cursor-based pagination modes.

#### Scenario: Page mode query returns results

- **GIVEN** a data source configuration with `page = 1`, `per_page = 20`, and `entity_types = ["host"]`
- **WHEN** the data source read executes
- **THEN** the provider SHALL call `GET /api/security/entity_store/entities?page=1&per_page=20&entity_types=host`
- **AND** SHALL serialize the full response to `results_json` as normalized JSON (sorted keys)

#### Scenario: Cursor mode query returns results

- **GIVEN** a data source configuration with `filter = "entity.type:host"`, `size = 50`, and no page-mode attributes
- **WHEN** the data source read executes
- **THEN** the provider SHALL call `GET /api/security/entity_store/entities` with `filter` and `size` query parameters
- **AND** SHALL serialize the response to `results_json`

#### Scenario: Search-after cursor pagination

- **GIVEN** a data source configuration with `search_after = jsonencode([...])` set to a previously returned cursor
- **WHEN** the data source read executes
- **THEN** the provider SHALL pass the `searchAfter` query parameter to the API
- **AND** SHALL return the next page of results in `results_json`

---

### Requirement: Pagination mode exclusivity (REQ-ENT-002)

The list endpoint does not support combining page-mode and cursor/search-after mode parameters. The data source SHALL enforce mutual exclusivity of the two modes at plan time.

Page-mode parameters: `sort_field`, `sort_order`, `page`, `per_page`, `filter_query`.
Cursor-mode parameters: `filter`, `size`, `search_after`, `source`, `fields`.

#### Scenario: Mixing pagination modes produces plan error

- **WHEN** a data source configuration sets both `page = 1` (page mode) and `filter = "..."` (cursor mode)
- **THEN** the provider SHALL produce a plan-time error indicating the two modes cannot be combined
- **AND** SHALL NOT make any API call

---

### Requirement: Data source schema (REQ-ENT-003)

The `elasticstack_kibana_security_entity_store_entities` data source SHALL expose the following attributes.

**Identity / connection:**
- `id` — computed string; set by the provider to a stable value reflecting the query parameters.
- `space_id` — optional computed string; default `"default"`.

**Single-entity convenience lookup (optional):**
- `entity_id` — optional string; when set, the provider calls the list endpoint with `filter = entity.id:"<entity_id>"`. Conflicts with `filter` and `filter_query`.

**Cursor/search-after mode (optional):**
- `filter` — optional string; KQL filter expression.
- `size` — optional int; number of results to return.
- `search_after` — optional string; JSON-encoded search_after cursor from a previous response.
- `source` — optional list of string; fields to include in response `_source`.
- `fields` — optional list of string; fields to include in response `fields`.

**Page mode (optional):**
- `sort_field` — optional string; field to sort by.
- `sort_order` — optional string; enum `"asc"` or `"desc"`.
- `page` — optional int; 1-indexed page number.
- `per_page` — optional int; number of entities per page.
- `filter_query` — optional string; Elasticsearch query string filter (page mode only).

**Common filter:**
- `entity_types` — optional set of string; values `"user"`, `"host"`, `"service"`, `"generic"`.

**Computed output:**
- `results_json` — computed string; normalized JSON (sorted keys) of the full API response body, including pagination metadata.

#### Scenario: Default read with no filters

- **WHEN** a data source configuration specifies only `space_id` with no additional parameters
- **THEN** the provider SHALL call `GET /api/security/entity_store/entities` with no filter parameters
- **AND** SHALL serialize the raw response to `results_json`

---

### Requirement: Single-entity `entity_id` exclusivity (REQ-ENT-006)

The `entity_id` attribute SHALL be mutually exclusive with `filter` (cursor-mode) and `filter_query` (page-mode) at plan time. When `entity_id` is set, the provider generates an implicit KQL filter and does not accept a user-supplied filter expression.

#### Scenario: Single-entity lookup via entity_id

- **GIVEN** a data source configuration with `entity_id = "host:web-01"` and no `filter` or `filter_query`
- **WHEN** the data source read executes
- **THEN** the provider SHALL call `GET /api/security/entity_store/entities` with `filter = entity.id:"host:web-01"`
- **AND** `entity_types` MAY be passed alongside `entity_id` if also configured
- **AND** SHALL serialize the response to `results_json`

#### Scenario: entity_id conflicts with filter and filter_query

- **GIVEN** a data source configuration with `entity_id = "host:web-01"` and `filter = "entity.type:host"`
- **WHEN** `terraform plan` is executed
- **THEN** the provider SHALL produce a plan-time error indicating `entity_id` cannot be combined with `filter` or `filter_query`
- **AND** SHALL NOT make any API call

---

### Requirement: Version gating at Elastic Stack 9.1.0 (REQ-ENT-004)

The data source SHALL enforce a minimum Elastic Stack version of `9.1.0` via `GetVersionRequirements()` on the data source model.

#### Scenario: Read blocked below minimum version

- **WHEN** the Kibana server is below version `9.1.0`
- **THEN** the provider SHALL return an error diagnostic describing the version requirement
- **AND** SHALL NOT make an API call

---

### Requirement: Acceptance test coverage (REQ-ENT-005)

Acceptance tests for this data source SHALL cover the following scenarios. All tests SHALL be skipped when the test Elastic Stack is below version `9.1.0`.

#### Scenario: Page mode query returns non-empty results

- **GIVEN** one or more entity resources created via `terraform apply`
- **WHEN** the list data source is read with `page = 1` and `per_page = 10`
- **THEN** `results_json` SHALL be a non-empty JSON string representing the API response

#### Scenario: Entity type filter narrows results

- **GIVEN** entity resources of two different types (e.g., host and user) created in the same space
- **WHEN** the list data source is read with `entity_types = ["host"]`
- **THEN** `results_json` SHALL contain only host entities and no user entities

#### Scenario: Mixing pagination modes produces plan error

- **GIVEN** a data source configuration that sets both `page = 1` and `filter = "entity.type:host"`
- **WHEN** `terraform plan` is executed
- **THEN** the provider SHALL produce a plan-time error indicating the modes cannot be combined
