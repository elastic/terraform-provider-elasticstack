## ADDED Requirements

### Requirement: Data source identity

The `elasticstack_kibana_osquery_pack` data source SHALL accept `pack_id` (Required string) and `space_id` (Optional string, default `"default"`) as its identity attributes. `kibana_connection` SHALL be an Optional block for connection override.

#### Scenario: Read by pack_id
- **WHEN** `pack_id = "linux-processes"` is configured
- **THEN** the data source SHALL call `GET /api/osquery/packs/linux-processes` (space-aware)
- **AND** all Computed attributes SHALL be populated from the response

#### Scenario: Missing pack returns error
- **WHEN** the pack does not exist (HTTP 404)
- **THEN** the data source SHALL return an error diagnostic (not silently remove from state)

### Requirement: Data source schema attributes

The data source SHALL expose the following attributes (Computed unless noted as Required/Optional):

- `id` — Computed string; mirrors `pack_id`
- `pack_id` — Required string (input identity)
- `space_id` — Optional string, default `"default"`
- `kibana_connection` — Optional block
- `name` — Computed string
- `description` — Computed string
- `enabled` — Computed bool
- `policy_ids` — Computed list of strings
- `shards` — Computed map(string → number)
- `schedule_type` — Computed string
- `interval` — Computed Int64
- `rrule_schedule` — Computed SingleNestedAttribute (same fields as the resource: `rrule`, `start_date`, `end_date`, `splay`, `timeout`)
- `queries` — Computed MapNestedAttribute (same nested schema as the resource, all fields Computed)
- `read_only` — Computed bool; indicates whether the pack is prebuilt and cannot be managed via the resource

#### Scenario: All Computed attributes populated from API response
- **WHEN** the data source reads a pack with all fields set (name, description, enabled, policy_ids, shards, schedule_type, interval, queries)
- **THEN** all Computed attributes SHALL be populated in state from the API response

### Requirement: Prebuilt pack is readable via data source

The data source SHALL NOT return an error when the API indicates `read_only = true`. The data source IS the intended path for referencing prebuilt Osquery packs.

#### Scenario: Prebuilt pack is read successfully
- **WHEN** the API returns `read_only = true` for a pack
- **THEN** the data source SHALL populate all Computed attributes including `read_only = true`
- **AND** SHALL NOT return an error diagnostic

### Requirement: Space-awareness

The data source SHALL inject space-awareness via `kibanautil.SpaceAwarePathRequestEditor(spaceID)`, identical to the resource. Default space is `"default"`.

#### Scenario: Non-default space used in data source read
- **WHEN** `space_id = "staging"` is configured
- **THEN** the data source SHALL call `GET /s/staging/api/osquery/packs/{id}`

### Requirement: Minimum Kibana version

The data source SHALL declare the same minimum Kibana version requirements as the resource (base packs CRUD floor), via `GetVersionRequirements`.

#### Scenario: Kibana below minimum version returns a version error
- **WHEN** the configured Kibana instance is below the declared minimum version
- **THEN** the provider SHALL return an error diagnostic indicating the unsatisfied version requirement
