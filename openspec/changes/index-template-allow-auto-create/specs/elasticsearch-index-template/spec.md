## ADDED Requirements

### Requirement: Schema — `allow_auto_create` attribute (REQ-043)

The `elasticstack_elasticsearch_index_template` resource SHALL expose an **optional** top-level attribute `allow_auto_create` of type boolean. When configured, `allow_auto_create` controls whether auto-creation of matching indices is allowed or denied at the template level, overriding the cluster-level `action.auto_create_index` setting.

The `elasticstack_elasticsearch_index_template` data source SHALL expose a **computed** top-level attribute `allow_auto_create` of type boolean, populated from the API response on read.

The attribute SHALL be placed at the same schema level as `priority`, `version`, and `composed_of`.

**HCL shape (resource):**

```hcl
resource "elasticstack_elasticsearch_index_template" "example" {
  name              = "my-template"
  index_patterns    = ["my-index-*"]
  allow_auto_create = true
}
```

No version gate is required. The `allow_auto_create` field on index templates is available from Elasticsearch 7.11, which is below the provider's minimum supported version (7.17+).

#### Scenario: `allow_auto_create` set to `true`

- GIVEN a configuration with `allow_auto_create = true`
- WHEN create or update runs
- THEN the provider SHALL include `"allow_auto_create": true` in the PUT index template request body
- AND state SHALL contain `allow_auto_create = true` after the read-back

#### Scenario: `allow_auto_create` set to `false`

- GIVEN a configuration with `allow_auto_create = false`
- WHEN create or update runs
- THEN the provider SHALL include `"allow_auto_create": false` in the PUT index template request body
- AND state SHALL contain `allow_auto_create = false` after the read-back

#### Scenario: `allow_auto_create` omitted

- GIVEN a configuration where `allow_auto_create` is not set
- WHEN create or update runs
- THEN the provider SHALL NOT include `allow_auto_create` in the PUT index template request body
- AND state SHALL contain `allow_auto_create = null`

#### Scenario: Data source read populates `allow_auto_create`

- GIVEN an index template in Elasticsearch with `allow_auto_create` set to `true`
- WHEN the data source reads that template
- THEN the data source state SHALL contain `allow_auto_create = true`

### Requirement: Expand `allow_auto_create` into API request (REQ-044)

On create and update, the provider SHALL convert the Terraform `allow_auto_create` attribute to the `allow_auto_create` field on the `models.IndexTemplate` request body when the attribute is non-null. When the attribute is null (not configured), the field SHALL be omitted from the request (`omitempty` serialization).

#### Scenario: Round-trip create

- GIVEN `allow_auto_create = true` configured
- WHEN create runs and the template is read back
- THEN state SHALL contain `allow_auto_create = true`

#### Scenario: Update to explicit `false`

- GIVEN an existing template with `allow_auto_create = true`
- WHEN configuration changes `allow_auto_create` to `false` and apply runs
- THEN the provider SHALL send `allow_auto_create: false` in the updated API request
- AND state SHALL reflect `allow_auto_create = false` after read-back

### Requirement: Flatten `allow_auto_create` from API response (REQ-045)

On read, the provider SHALL map `estypes.IndexTemplate.AllowAutoCreate` from the GET index template API response into the `allow_auto_create` attribute in Terraform state. When the API response does not include `allow_auto_create` (field is nil), the provider SHALL set `allow_auto_create` to null in state.

#### Scenario: API returns `allow_auto_create = true`

- GIVEN the GET index template API returns `allow_auto_create: true`
- WHEN read runs
- THEN state SHALL contain `allow_auto_create = true`

#### Scenario: API omits `allow_auto_create`

- GIVEN the GET index template API does not include `allow_auto_create`
- WHEN read runs
- THEN state SHALL contain `allow_auto_create = null`

### Requirement: Model — `AllowAutoCreate` field in `models.IndexTemplate` (REQ-046)

The internal `models.IndexTemplate` struct SHALL include an `AllowAutoCreate *bool` field serialized as `"allow_auto_create,omitempty"`. When this field is nil, the JSON serialization SHALL omit the key entirely from the request body.

#### Scenario: Nil field omitted from payload

- GIVEN `allow_auto_create` is not configured (field is nil in `models.IndexTemplate`)
- WHEN the struct is serialized to JSON for the PUT index template API
- THEN the JSON payload SHALL NOT include the `allow_auto_create` key

### Requirement: Acceptance tests — `allow_auto_create` coverage (REQ-047)

Acceptance tests for `elasticstack_elasticsearch_index_template` SHALL include a test that covers:

- Creating a template with `allow_auto_create = true` and verifying state after create.
- Updating the template to `allow_auto_create = false` and verifying state after update.
- Removing the attribute (null) and verifying no plan diff after apply.
- Importing the resource and verifying `allow_auto_create` is populated correctly.

#### Scenario: Acceptance test create and update

- GIVEN an acceptance test that creates a template with `allow_auto_create = true`
- WHEN the template is applied and state is read back
- THEN state SHALL contain `allow_auto_create = true`
- AND when the test updates to `allow_auto_create = false`, state SHALL contain `allow_auto_create = false`
