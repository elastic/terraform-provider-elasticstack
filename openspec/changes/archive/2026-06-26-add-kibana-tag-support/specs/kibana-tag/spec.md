# kibana-tag Specification

## Purpose

Define requirements for `elasticstack_kibana_tag` (resource) and `elasticstack_kibana_tags` (data source), which manage Kibana tags via the public `/api/tags` REST API introduced in Kibana 9.5.0.

## ADDED Requirements

### Requirement: Resource schema attributes (REQ-001)

The `elasticstack_kibana_tag` resource SHALL expose the following attributes:

- `name` — Required string; the display name of the tag.
- `tag_id` — Optional string, ForceNew; a client-specified UUID for the tag. When set, the provider uses `PUT /api/tags/{id}` semantics to create the tag. When absent, the provider uses `POST /api/tags` and stores the server-minted UUID.
- `color` — Optional+Computed string; a hex color value (e.g. `#772299`). When absent from configuration, the server generates a random color and `UseStateForUnknown` prevents spurious diffs on subsequent plans. When set in configuration, the provider SHALL validate that `color` matches `#RRGGBB` (six hexadecimal digits).
- `description` — Optional string; a free-text description of the tag.
- `space_id` — Optional+Computed string, ForceNew; the Kibana space. Defaults to `"default"`. A change to `space_id` forces replacement.
- `id` — Computed string; the composite identity `"<space_id>/<tag_id>"`.
- `created_at` — Computed string; ISO 8601 timestamp from `meta.created_at`.
- `updated_at` — Computed string; ISO 8601 timestamp from `meta.updated_at`.

The `managed` field from `meta.managed` SHALL NOT be exposed as a schema attribute.

#### Scenario: Required name enforced

- GIVEN a resource configuration with no `name` attribute
- WHEN Terraform validates the configuration
- THEN the provider SHALL return a validation error

#### Scenario: color UseStateForUnknown

- GIVEN a resource created without a `color` attribute (server assigns a random color)
- WHEN Terraform runs a subsequent plan with no configuration changes
- THEN the plan SHALL be empty (no diff on `color`)

#### Scenario: color hex format validated

- GIVEN a resource configuration with `color = "red"`
- WHEN Terraform validates the configuration
- THEN the provider SHALL return a validation error

### Requirement: Resource identity and composite ID (REQ-002)

The `elasticstack_kibana_tag` resource SHALL set `id` to the composite form `"<space_id>/<tag_id>"` after every Create and Update. `tag_id` represents the raw UUID assigned by the server (or provided by the practitioner). `space_id` defaults to `"default"` when not specified.

#### Scenario: Composite ID after create without tag_id

- GIVEN a resource created without an explicit `tag_id`
- WHEN create completes
- THEN `tag_id` SHALL be set to the server-minted UUID
- AND `id` SHALL equal `"<space_id>/<server-minted-uuid>"`

#### Scenario: Composite ID after create with tag_id

- GIVEN a resource with `tag_id = "my-custom-uuid"` and `space_id = "ops"`
- WHEN create completes
- THEN `id` SHALL equal `"ops/my-custom-uuid"`

### Requirement: Create without client-specified ID (REQ-003)

When `tag_id` is absent, the provider SHALL create the tag by calling `POST /api/tags` with `name`, `color` (if set), and `description` (if set). The response `id` field SHALL be stored as `tag_id` in state.

#### Scenario: POST create stores server-minted ID

- GIVEN a resource with `name = "staging"` and no `tag_id`
- WHEN create runs
- THEN the provider SHALL call `POST /api/tags`
- AND `tag_id` in state SHALL be the UUID returned in the response `id` field

### Requirement: Create with client-specified ID (REQ-004)

When `tag_id` is set, the provider SHALL first call `GET /api/tags/{id}`:

- If the response is **404**, the provider SHALL call `PUT /api/tags/{id}` to create the tag (the API's upsert semantics guarantee creation when the resource does not exist). A `JSON201` response confirms creation.
- If the response is **200**, the tag already exists; the provider SHALL return an error diagnostic instructing the practitioner to use `terraform import` instead.

#### Scenario: tag_id specified, tag does not exist → PUT creates

- GIVEN `tag_id = "my-uuid"` and `GET /api/tags/my-uuid` returns 404
- WHEN create runs
- THEN the provider SHALL call `PUT /api/tags/my-uuid`
- AND on `JSON201` response, `tag_id` SHALL be stored as `"my-uuid"` in state

#### Scenario: tag_id specified, tag already exists → error

- GIVEN `tag_id = "existing-uuid"` and `GET /api/tags/existing-uuid` returns 200
- WHEN create runs
- THEN the provider SHALL return an error diagnostic
- AND the diagnostic SHALL mention `terraform import`
- AND no state SHALL be written

### Requirement: Tag already exists on POST (REQ-005)

If `POST /api/tags` returns an error indicating a name conflict (duplicate tag name in the space), the provider SHALL surface the API error as a Terraform diagnostic. No retry or silent recovery SHALL occur.

#### Scenario: Duplicate name on POST

- GIVEN a tag named `"production"` already exists in the space
- WHEN `POST /api/tags` with `name = "production"` is called
- THEN the provider SHALL return the API error as a Terraform error diagnostic

### Requirement: Read path (REQ-006)

On every Read (including read-after-write and import), the provider SHALL call `GET /api/tags/{id}` and update state with the current values of `name`, `color`, `description`, `created_at`, and `updated_at`. The provider SHALL apply the managed-tag guard (REQ-009) before writing state.

If the tag is not found (404), the provider SHALL remove the resource from state (Terraform will plan a recreation).

#### Scenario: Read populates state

- GIVEN a tag exists with `name = "prod"`, `color = "#FF0000"`, and `description = null`
- WHEN Read runs
- THEN `name`, `color` SHALL be set in state
- AND `description` SHALL be null in state

#### Scenario: Read on missing tag removes from state

- GIVEN `GET /api/tags/{id}` returns 404
- WHEN Read runs
- THEN the resource SHALL be removed from state

### Requirement: Update path (REQ-007)

On Update, the provider SHALL call `PUT /api/tags/{id}` with the configured `name` and the planned `color` value (preserving prior state when unknown via `UseStateForUnknown`) so that omitting `color` from configuration does not cause server-side color regeneration. The provider SHALL include `description` only when set in configuration. The provider SHALL apply the managed-tag guard (REQ-009) before calling the API.

Both `JSON200` (updated) and `JSON201` (upserted, in a race where the tag was deleted between plan and apply) responses SHALL be treated as success.

#### Scenario: Update changes name

- GIVEN a tag with `name = "staging"` in state
- WHEN the practitioner updates `name = "staging-v2"` and applies
- THEN the provider SHALL call `PUT /api/tags/{id}` with the new name
- AND state SHALL reflect the updated name after apply

### Requirement: Delete path (REQ-008)

On Delete, the provider SHALL call `DELETE /api/tags/{id}`. The provider SHALL apply the managed-tag guard (REQ-009) before calling the API.

A 404 response on Delete SHALL be treated as a no-op (the tag is already gone); the provider SHALL remove the resource from state without returning an error.

#### Scenario: Delete succeeds

- GIVEN a tag exists and is not managed
- WHEN Delete runs
- THEN the provider SHALL call `DELETE /api/tags/{id}`
- AND the resource SHALL be removed from state

#### Scenario: Delete on already-deleted tag

- GIVEN `DELETE /api/tags/{id}` returns 404
- WHEN Delete runs
- THEN the provider SHALL NOT return an error
- AND the resource SHALL be removed from state

### Requirement: Managed-tag guard (REQ-009)

When the API response includes `meta.managed = true`, the provider SHALL return an error diagnostic on Read, Update, and Delete, and SHALL NOT modify Terraform state. The error diagnostic SHALL clearly explain that the tag is managed by Kibana and cannot be controlled by this resource.

The resource cannot produce a managed tag via Create (Kibana sets `managed` server-side only), so the guard is not applied on Create.

#### Scenario: Read of a managed tag

- GIVEN `GET /api/tags/{id}` returns a response where `meta.managed = true`
- WHEN Read runs
- THEN the provider SHALL return an error diagnostic
- AND state SHALL NOT be updated

#### Scenario: Update of a managed tag

- GIVEN the current API state of the tag has `meta.managed = true`
- WHEN Update runs
- THEN the provider SHALL return an error diagnostic
- AND the PUT request SHALL NOT be sent

#### Scenario: Delete of a managed tag

- GIVEN the current API state of the tag has `meta.managed = true`
- WHEN Delete runs
- THEN the provider SHALL return an error diagnostic
- AND the DELETE request SHALL NOT be sent

### Requirement: Import format (REQ-010)

The `elasticstack_kibana_tag` resource SHALL support import using the composite format `"<space_id>/<tag_id>"`. The import handler SHALL parse this string, set `space_id` and `tag_id` in state, then trigger a normal Read (which applies the managed-tag guard). Importing a managed tag SHALL fail with the managed-tag error diagnostic.

#### Scenario: Successful import

- GIVEN `terraform import elasticstack_kibana_tag.example default/abc-123`
- WHEN the import runs
- THEN `space_id` SHALL be set to `"default"` and `tag_id` to `"abc-123"` in state
- AND Read SHALL populate all remaining attributes

#### Scenario: Import of managed tag fails

- GIVEN the tag at `abc-123` has `meta.managed = true`
- WHEN the import runs
- THEN the provider SHALL return an error diagnostic from the managed-tag guard
- AND no state SHALL be written

### Requirement: Version gate — Kibana ≥ 9.5.0 (REQ-011)

All CRUD operations on `elasticstack_kibana_tag` and all Read operations on `elasticstack_kibana_tags` SHALL fail with an error diagnostic when the connected Kibana version is below **9.5.0**. The error diagnostic SHALL name the minimum required version and reference the Kibana 9.5 release.

#### Scenario: Version too low

- GIVEN the provider is connected to a Kibana instance running version 9.4.x or earlier
- WHEN any resource CRUD or data source Read operation executes
- THEN the provider SHALL return an error diagnostic
- AND the diagnostic SHALL state that Kibana ≥ 9.5.0 is required

### Requirement: Data source schema (REQ-012)

The `elasticstack_kibana_tags` data source SHALL expose:

- `query` — Optional string; an Elasticsearch `simple_query_string` expression applied to `name` and `description` fields. When absent or empty, all tags in the space are returned.
- `space_id` — Optional string, default `"default"`.
- `tags` — Computed list of objects. Each element SHALL contain:
  - `id` (string) — the tag UUID
  - `name` (string)
  - `color` (string)
  - `description` (string, may be null)
  - `managed` (bool) — `true` when Kibana manages the tag
  - `created_at` (string)
  - `updated_at` (string)

#### Scenario: Data source returns all tags when query is absent

- GIVEN three tags exist in the space and `query` is not set
- WHEN the data source is read
- THEN `tags` SHALL contain all three entries

#### Scenario: Data source returns empty list when no match

- GIVEN no tags match the given `query`
- WHEN the data source is read
- THEN `tags` SHALL be an empty list (not an error)

### Requirement: Data source query filtering (REQ-013)

When `query` is set, the provider SHALL pass it verbatim as the `query` parameter to `GET /api/tags`. The API applies Elasticsearch `simple_query_string` semantics against `name` and `description`. The provider SHALL NOT perform additional client-side filtering.

#### Scenario: query filters by name prefix

- GIVEN tags named `"prod"`, `"production"`, and `"staging"` exist
- AND `query = "prod*"` is configured
- WHEN the data source is read
- THEN `tags` SHALL contain only `"prod"` and `"production"`

### Requirement: Data source auto-pagination (REQ-014)

The data source SHALL auto-paginate across all server-side pages and return the complete set of matching tags. When the total result count exceeds a single page, the provider SHALL continue fetching pages until `len(collected) >= meta.total`. The provider SHALL use a `per_page` value of at most 100 per request (or lower if the server enforces a lower cap).

#### Scenario: Results span multiple pages

- GIVEN 150 tags exist in the space
- AND the server returns at most 100 results per page
- WHEN the data source is read with no query filter
- THEN `tags` SHALL contain all 150 entries
