# `elasticstack_fleet_agent_download_source` — Schema and Functional Requirements

Resource implementation: `internal/fleet/agentdownloadsource`

## Schema

```hcl
resource "elasticstack_fleet_agent_download_source" "example" {
  # Required identity/config
  name = "My Agent Download Source"
  host = "https://artifacts.example.com/elastic-agent"

  # Optional identity
  source_id = "custom-download-source-id" # optional, string; if omitted, Kibana generates an ID

  # Optional arguments
  default  = false          # optional, bool; when true, marks this as the default download source
  proxy_id = "proxy-123456" # optional, string; references a Fleet proxy by ID

  # Space scoping
  space_ids = ["default"] # optional+computed, set(string); when set, the resource is managed in the first space ID
}
```

## Requirements

- **[REQ-001] (API)**: The resource shall use the Kibana Fleet **Agent binary download sources** APIs ([docs](https://www.elastic.co/docs/api/doc/kibana/operation/operation-get-fleet-agent-download-sources)) to manage Agent Binary Download Source objects:
  - **Create**: `POST /api/fleet/agent_download_sources`
  - **Read (list)**: `GET /api/fleet/agent_download_sources`
  - **Read (by ID)**: `GET /api/fleet/agent_download_sources/{source_id}`
  - **Update**: `PUT /api/fleet/agent_download_sources/{source_id}`
  - **Delete**: `DELETE /api/fleet/agent_download_sources/{source_id}`.
- **[REQ-002] (API)**: The resource shall use the generated Kibana OpenAPI client in `generated/kbapi` via the Fleet wrapper client (`internal/clients/fleet`) for all HTTP interactions, including the space-aware request editor where applicable.
- **[REQ-003] (Identity)**: The Terraform state `id` attribute shall store the Fleet download source ID returned by Kibana. The `source_id` attribute shall mirror this value and be used as the path parameter for read, update, delete calls.
- **[REQ-004] (Identity: user-specified ID)**: When `source_id` is set in configuration, the create request shall pass it through to Kibana so that the download source is created with that ID. When omitted, the provider shall accept the server-generated ID from the create response and persist it to both `id` and `source_id` in state.
- **[REQ-005] (Schema: required arguments)**: The resource shall require `name` (string) and `host` (string) and map them to the corresponding fields in the Fleet API request/response bodies.
- **[REQ-006] (Schema: optional arguments)**: The resource shall expose the following optional attributes:
  - `default` (bool) mapped to `is_default` in the API.
  - `proxy_id` (string) mapped to `proxy_id` in the API.
- **[REQ-007] (Schema: spaces)**: The resource shall expose `space_ids` as an optional+computed `set(string)`:
  - When `space_ids` is set, the first value shall determine the Kibana space used for create, read, update, and delete operations via the space-aware Fleet client helpers.
  - When `space_ids` is unset or empty, the resource shall operate in the default space.
  - Since the Fleet download sources API does not return space information, `space_ids` shall be preserved from Terraform state and not derived from the API.
- **[REQ-008] (Import)**: The resource shall support `terraform import` by accepting a Fleet download source ID and populating both `id` and `source_id` with that value.
- **[REQ-009] (Error handling)**: When the Fleet API returns a non-success status code for create, update, or delete, the resource shall surface the error (HTTP status and message/body) via Terraform diagnostics. For read operations:
  - A `404 Not Found` for a specific `source_id` shall cause the resource to be removed from state.
  - Other non-success statuses shall be reported as diagnostics.
- **[REQ-010] (Plan/State)**: The resource shall support in-place updates for changes to `name`, `host`, `default`, and `proxy_id`, issuing a `PUT` to the Fleet API. Changes to `source_id` shall force replacement.
- **[REQ-011] (Auth/secrets: out of scope v1)**: The initial implementation shall not expose the Fleet download source `auth` and `secrets` fields (API key, username/password, SSL key material) as Terraform schema attributes. These remain managed by Fleet, and are out of scope for the first version of this resource.
- **[REQ-012] (Compatibility)**: The resource shall be guarded by a minimum Kibana/Fleet version that supports the Agent Binary Download Sources API (TBD based on product docs). If the connected Kibana is below that version, the provider shall emit a clear diagnostic during plan/apply indicating that the resource is not supported.

