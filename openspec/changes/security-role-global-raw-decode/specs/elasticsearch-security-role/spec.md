## MODIFIED Requirements

### Requirement: Role CRUD APIs (REQ-001–REQ-003) — partial update

The `GetRole` implementation SHALL bypass the typed `Security.GetRole` client call and instead fetch `GET /_security/role/<name>` via `typedClient.Transport.Perform`. The PutRole and DeleteRole implementations continue to use the go-elasticsearch Typed API unchanged. The raw response body SHALL be decoded as `map[string]json.RawMessage` to locate the per-role entry. The `global` field SHALL be extracted as `json.RawMessage` and carried to the model layer **out-of-band** (not assigned to `types.Role.Global`, which is typed `map[string]map[string]map[string][]string` and cannot represent array-typed categories). All other role fields (applications, cluster, indices, remote_indices, run_as, metadata, description) SHALL continue to be decoded from the API response using the typed `types.Role` struct or equivalent individual field decoders.

This change is required because the go-elasticsearch typed client declares `Role.Global` as `map[string]map[string]map[string][]string`, which cannot decode heterogeneous per-category shapes such as the `"data_source": []` array introduced in Elasticsearch 9.5. Upstream tracking: elasticsearch-specification#6377.

#### Scenario: GetRole decodes global as raw JSON

- GIVEN an Elasticsearch 9.5+ role that includes `"global": {"data_source": [], "application": {}, "profile": {...}}`
- WHEN the provider reads the role
- THEN the provider SHALL successfully decode the response without an unmarshal error
- AND the `global` field SHALL be carried to the model layer as raw JSON (not via `types.Role.Global`)
- AND all other role attributes SHALL be populated from the same API response

#### Scenario: GetRole preserves existing behavior for non-global fields

- GIVEN a role that has indices, cluster, applications, and run_as configured
- WHEN the provider reads the role via the raw transport path
- THEN the provider SHALL populate all non-global fields correctly, matching the behavior of the prior typed-client read path

#### Scenario: GetRole returns not-found for absent role

- GIVEN a role name that does not exist in Elasticsearch
- WHEN the provider calls GetRole
- THEN the provider SHALL return a not-found result (nil role) without error, preserving the existing not-found behavior

#### Scenario: GetRole surfaces HTTP errors

- GIVEN Elasticsearch returns a non-200, non-404 status code for the role read request
- WHEN the provider calls GetRole
- THEN the provider SHALL return an error diagnostic with the HTTP status and response body

---

### Requirement: Typed client implementation for security role — narrowed

The `elasticstack_elasticsearch_security_role` resource and data source SHALL manage roles using the go-elasticsearch Typed API for PutRole and DeleteRole (`elasticsearch.TypedClient.Security.PutRole`, `Security.DeleteRole`). **GetRole is narrowed**: because the typed client's `Role.Global` field (`map[string]map[string]map[string][]string`) cannot decode heterogeneous per-category shapes such as the ES 9.5 `"data_source": []` array (upstream: elasticsearch-specification#6377), GetRole SHALL fetch `GET /_security/role/<name>` via `typedClient.Transport.Perform` and decode `global` as `json.RawMessage`, carrying it to the model layer out-of-band. All non-`global` fields continue to use the typed `types.Role` struct. The typed API response SHALL be used directly for PutRole/DeleteRole without manual JSON decoding into an intermediate `models.Role` type.

#### Scenario: Typed API for write/delete

- GIVEN a valid Elasticsearch connection
- WHEN the resource performs create, update, or delete
- THEN the provider SHALL call the typed Security PutRole/DeleteRole APIs
- AND role data for the write payload SHALL be returned as `*types.Role`

#### Scenario: Raw transport for GetRole global field

- GIVEN a role that includes `global` privileges (including ES 9.5+ array-typed categories)
- WHEN the resource or data source reads the role
- THEN the provider SHALL fetch the role via raw `GET /_security/role/<name>` transport
- AND SHALL decode `global` as `json.RawMessage` (not through `types.Role.Global`)
- AND SHALL decode all other fields into the typed `types.Role` struct

---

## MODIFIED Requirements

### Requirement: Global defaults normalization

`populateGlobalPrivilegesDefaults` SHALL strip server-injected empty `global` defaults from the API-returned `global` blob before it is written to Terraform state, so state matches user intent rather than the raw API response. This SHALL include stripping `role` when it is an empty object (`{}`) and `data_source` when it is an empty array (`[]`), and SHALL be generalized to strip future server-injected empty defaults (empty objects or empty arrays) as they appear.

#### Scenario: Empty role object is stripped

- GIVEN an API-returned `global` of `{"application": {}, "profile": {...}, "role": {}}`
- WHEN the provider normalizes `global` for state
- THEN state SHALL contain `{"application": {}, "profile": {...}}` (empty `role` stripped)

#### Scenario: Empty data_source array is stripped

- GIVEN an Elasticsearch 9.5+ API-returned `global` of `{"application": {}, "profile": {...}, "data_source": []}`
- WHEN the provider normalizes `global` for state
- THEN state SHALL contain `{"application": {}, "profile": {...}}` (empty `data_source` stripped)

#### Scenario: Non-empty data_source is preserved

- GIVEN an API-returned `global` of `{"application": {}, "data_source": ["foo"]}`
- WHEN the provider normalizes `global` for state
- THEN state SHALL contain `{"application": {}, "data_source": ["foo"]}` (non-empty `data_source` preserved)
