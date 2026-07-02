## MODIFIED Requirements

### Requirement: Role CRUD APIs (REQ-001–REQ-003) — partial update

The `GetRole` implementation SHALL bypass the typed `Security.GetRole` client call and instead fetch `GET /_security/role/<name>` via `typedClient.Transport.Perform`. The PutRole and DeleteRole implementations continue to use the go-elasticsearch Typed API unchanged. The raw response body SHALL be decoded as `map[string]json.RawMessage` to locate the per-role entry. The `global` field SHALL be extracted as `json.RawMessage` and carried through to the model layer without being decoded through the typed `Role.Global` field. All other role fields (applications, cluster, indices, remote_indices, run_as, metadata, description) SHALL continue to be decoded from the API response using the typed `types.Role` struct or equivalent individual field decoders.

This change is required because the go-elasticsearch typed client declares `Role.Global` as `map[string]map[string]map[string][]string`, which cannot decode heterogeneous per-category shapes such as the `"data_source": []` array introduced in Elasticsearch 9.5. Upstream tracking: elasticsearch-specification#6377.

#### Scenario: GetRole decodes global as raw JSON

- GIVEN an Elasticsearch 9.5+ role that includes `"global": {"data_source": [], "application": {}, "profile": {...}}`
- WHEN the provider reads the role
- THEN the provider SHALL successfully decode the response without an unmarshal error
- AND the `global` field in Terraform state SHALL contain the API-returned JSON blob
- AND all other role attributes SHALL be populated from the same API response

#### Scenario: GetRole preserves existing behavior for non-global fields

- GIVEN a role that has indices, cluster, applications, and run_as configured
- WHEN the provider reads the role via the raw transport path
- THEN the provider SHALL populate all non-global fields correctly, matching the behavior of the prior typed-client read path

#### Scenario: GetRole returns not-found for absent role

- GIVEN a role name that does not exist in Elasticsearch
- WHEN the provider calls GetRole
- THEN the provider SHALL return `(nil, nil)` without error, preserving the existing not-found behavior

#### Scenario: GetRole surfaces HTTP errors

- GIVEN Elasticsearch returns a non-200, non-404 status code for the role read request
- WHEN the provider calls GetRole
- THEN the provider SHALL return an error diagnostic with the HTTP status and response body

---

## ADDED Requirements

### Requirement: Write-path global decode uses map[string]any (REQ-039)

When the user configures the `global` attribute, the `toAPIModel` conversion SHALL decode the JSON string into `map[string]any` (not `map[string]map[string]map[string][]string`). This accommodates heterogeneous per-category shapes in user-supplied `global` JSON — for example, categories whose values are arrays rather than nested objects. The decoded value SHALL then be marshaled back to JSON for the PutRole API payload using the existing marshal-to-`map[string]json.RawMessage` path in `security_role.go`.

#### Scenario: Global with standard application privileges encodes correctly

- GIVEN `global = jsonencode({"application": {"manage": {"applications": ["myapp"]}}})`
- WHEN create or update runs
- THEN the provider SHALL marshal `global` without error and include the correct payload in the Put role API request

#### Scenario: Global with array-typed category encodes without error

- GIVEN `global = jsonencode({"data_source": [], "application": {"manage": {"applications": ["*"]}}})`
- WHEN create or update runs
- THEN the provider SHALL marshal `global` without error and include the payload in the Put role API request, forwarding the array as-is to Elasticsearch
