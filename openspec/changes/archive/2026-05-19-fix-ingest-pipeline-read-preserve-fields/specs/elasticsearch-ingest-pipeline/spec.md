## MODIFIED Requirements

### Requirement: Read (REQ-015–REQ-017)

On read, the resource SHALL parse `id` as `<cluster_uuid>/<pipeline_name>`, call the Get pipeline API with the pipeline name, and remove the resource from state (set `id` to `""`) when the pipeline is not found (HTTP 404). On a successful get, the resource SHALL set `name`, `description` (when present in the response), `processors`, `on_failure` (when present), and `metadata` (when present) from the API response. The provider SHALL decode the Get pipeline response in a way that preserves every field returned by the API for each processor and `on_failure` handler, including fields that are not modeled by the go-elasticsearch typed client.

#### Scenario: Pipeline not found on refresh

- GIVEN the pipeline has been deleted outside of Terraform
- WHEN read runs
- THEN the provider SHALL remove the resource from state by setting `id` to `""`

#### Scenario: State populated from API response

- GIVEN a successful Get pipeline response
- WHEN read completes
- THEN `name`, `processors`, and any present optional fields SHALL be set in state from the API response

#### Scenario: Read preserves processor fields unmodeled by the typed client

- GIVEN a processor body containing a field that is not present on the corresponding go-elasticsearch typed processor struct (for example, `override = true` on a `rename` processor)
- WHEN read runs after a successful create or update
- THEN the resulting `processors` state element SHALL contain that field with the value returned by the Elasticsearch Get pipeline API
- AND a subsequent plan SHALL be empty (no drift)

### Requirement: JSON mapping for processors and on_failure (REQ-020–REQ-022)

Each element of `processors` and `on_failure` SHALL be validated as a JSON string by schema. On create/update, each element SHALL be decoded from its JSON string into a `map[string]any` before being included in the API request body. On read, each processor and on_failure handler object received from the API SHALL be preserved as an opaque object (e.g. `map[string]any`) end-to-end — without any intermediate decode into a typed processor struct that could drop unmodeled fields — and SHALL be marshalled back to a JSON string and stored as the corresponding list element in state.

#### Scenario: Invalid processor JSON

- GIVEN a `processors` element that is not valid JSON
- WHEN the configuration is applied
- THEN Terraform validation SHALL reject it before calling the API

#### Scenario: Round-trip processor JSON

- GIVEN a `processors` list with JSON strings
- WHEN create runs and then read runs
- THEN the `processors` state SHALL contain JSON strings representing the same objects as returned by the API, including every field returned by the API
