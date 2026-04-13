## ADDED Requirements

### Requirement: Index name validation for static and date math names
The `name` attribute on `elasticstack_elasticsearch_index` SHALL accept either a static index name that matches the existing lowercase index-name rules or a plain Elasticsearch date math index name expression. Validation SHALL keep these paths separate by using `stringvalidator.Any(...)` with the static-name regex `^[a-z0-9!$%&'()+.;=@[\]^{}~_-]+$` and the date-math regex `^<[^<>]*\{[^<>]+\}[^<>]*>$`. Values that satisfy neither regex branch SHALL be rejected during schema validation. When a validated date math name is used to create an index, the provider SHALL URI-encode that name before sending it in the Create Index API path.

#### Scenario: Static index names remain valid
- **WHEN** the configuration supplies a static index name that satisfies the existing lowercase-name rules
- **THEN** schema validation SHALL accept the value without requiring date math syntax

#### Scenario: Plain date math index names are accepted
- **WHEN** the configuration supplies a plain date math expression for the index path
- **THEN** schema validation SHALL accept the value without weakening the static-name validator

#### Scenario: Invalid date math syntax is rejected
- **WHEN** the configuration supplies a value that does not satisfy the static-name validator and is not valid for the dedicated date-math validator
- **THEN** schema validation SHALL reject the value before any API call is made

#### Scenario: Provider encodes date math name for create request
- **WHEN** the configuration supplies a valid plain date math name and the provider constructs the Create Index API request
- **THEN** the provider SHALL URI-encode that name in the request path sent to Elasticsearch

## MODIFIED Requirements

### Requirement: Identity (REQ-005–REQ-006)

The resource SHALL expose a computed `id` in the format `<cluster_uuid>/<concrete_index_name>` and a computed `concrete_name` attribute containing the concrete Elasticsearch index managed by the resource. During create, the resource SHALL compute `id` from the current cluster UUID and the concrete index name returned by Elasticsearch, not from the configured `name`. For imported or legacy state that lacks `concrete_name`, the resource SHALL derive `concrete_name` from `id.ResourceID` during read and store it in state.

#### Scenario: Id and concrete_name set on create

- **WHEN** a Create Index API call succeeds and Elasticsearch returns the created index name
- **THEN** `concrete_name` SHALL be set to that concrete index name and `id` SHALL be set to `<cluster_uuid>/<concrete_index_name>`

### Requirement: Import (REQ-007–REQ-008)

The resource SHALL support import by accepting an `id` value directly via `ImportStatePassthroughID`, persisting the imported `id` to state without validation at import time. Read and delete operations SHALL parse `id` in the format `<cluster_uuid>/<concrete_index_name>` and SHALL return an error diagnostic when the format is invalid. When imported or legacy state lacks `concrete_name`, read SHALL backfill it from `id.ResourceID`. When imported state also lacks `name`, read SHALL backfill `name` from the concrete index name so the resource remains readable without inventing a date math expression.

#### Scenario: Import passthrough backfills concrete identity

- **WHEN** an import command stores a composite `id` and the next read runs
- **THEN** the resource SHALL use the imported resource id as the concrete index identity for subsequent read, update, and delete operations

### Requirement: Create flow (REQ-013–REQ-014)

On create, the resource SHALL build an API model from the plan (including settings, mappings, and aliases) and submit a Create Index request using the configured `name` together with the configured `wait_for_active_shards`, `master_timeout`, and `timeout` parameters. When the configured `name` is a validated date math expression, the provider SHALL URI-encode it before sending the Create Index API request path. After a successful create, the resource SHALL capture the concrete index name returned by the Create Index API response, store it in `concrete_name`, compute `id` from the cluster UUID and that concrete name, and then perform a read to refresh all computed attributes in state. That post-create read SHALL preserve the configured `name` value in state rather than replacing it with the concrete index name.

#### Scenario: Serverless — master_timeout and wait_for_active_shards omitted

- **WHEN** the Elasticsearch server flavor is `serverless` and a create request is issued
- **THEN** `master_timeout` and `wait_for_active_shards` SHALL be omitted from the API call parameters

#### Scenario: Date math create stores configured and concrete names separately

- **WHEN** the configuration uses a plain date math index name and Elasticsearch creates a concrete index from it
- **THEN** state SHALL preserve the configured expression in `name` and store the concrete created index in `concrete_name`

### Requirement: Update flow (REQ-015–REQ-018)

On update, the resource SHALL only call the relevant update APIs when the corresponding values have changed. Alias changes SHALL be applied by deleting aliases removed from config (via Delete Alias API) and upserting all aliases present in plan (via Put Alias API). Dynamic setting changes SHALL be applied by calling the Put Settings API with the diff, setting removed dynamic settings to `null` in the request. Mapping changes SHALL be applied by calling the Put Mapping API when `mappings` has semantically changed. All update APIs SHALL target the persisted concrete index identity from state / `id`, not the configured `name`. After all updates, the resource SHALL perform a read to refresh state while preserving any configured `name` already stored in state.

#### Scenario: Removed alias is deleted

- **WHEN** an alias exists in state but is absent from the plan
- **THEN** the resource SHALL call the Delete Alias API for that alias against the concrete managed index

#### Scenario: Removed dynamic setting set to null

- **WHEN** a dynamic setting is present in state but absent from the plan
- **THEN** the resource SHALL send that setting as `null` in the Put Settings request

### Requirement: Read (REQ-019–REQ-021)

On read, the resource SHALL parse `id` to extract the concrete index name, call the Get Index API with `flat_settings=true`, and if the index is not found (HTTP 404 or missing from response), SHALL remove the resource from state without error. When the index is found, the resource SHALL populate `concrete_name`, all aliases, `mappings`, `settings_raw`, and all individual setting attributes from the API response. When state already contains a configured `name`, read SHALL preserve that configured value and SHALL NOT overwrite it with the concrete index name. When state does not contain `name`, read SHALL backfill `name` from the concrete index name.

#### Scenario: Index not found

- **WHEN** the Get Index API returns 404 or the concrete index name is absent from the response
- **THEN** the resource SHALL be removed from state and no error diagnostic SHALL be added

#### Scenario: Read preserves configured date math name

- **WHEN** state already contains a configured date math expression in `name` and read refreshes the managed concrete index
- **THEN** `name` SHALL remain unchanged and `concrete_name` SHALL reflect the concrete index being managed
