## MODIFIED Requirements

### Requirement: Unset role_descriptors SHALL be valid for create and update (REQ-027-REQ-029 amendment)

The resource SHALL omit `role_descriptors` from the Elasticsearch request when the attribute
is null or unknown, rather than attempting JSON parsing that fails. This extends REQ-027: JSON
parsing SHALL only be attempted when `role_descriptors` is a known, non-null value.

#### Scenario: Create without role_descriptors

- GIVEN a configuration that sets only `name` (no `role_descriptors`)
- WHEN Terraform applies the configuration for the first time
- THEN the resource SHALL create the API key successfully without returning an error
- AND the Elasticsearch create request SHALL NOT include a `role_descriptors` field
- AND the resulting state SHALL contain a valid `id`, `key_id`, `api_key`, and `encoded`

#### Scenario: Update without role_descriptors

- GIVEN an API key resource whose configuration does not set `role_descriptors`
- WHEN Terraform applies a plan that modifies another mutable attribute (e.g. `metadata`)
- THEN the resource SHALL update the API key without returning an error
- AND the Elasticsearch update request SHALL NOT include a `role_descriptors` field

#### Scenario: Restriction validation skips when role_descriptors is absent

- GIVEN a configuration that does not set `role_descriptors`
- WHEN the provider validates whether any role descriptor contains a `restriction` block
- THEN the validation SHALL return no error diagnostics
- AND the provider SHALL NOT attempt to parse an Unknown or Null JSON value

### Requirement: Acceptance coverage SHALL include API key with no role_descriptors

The acceptance test suite SHALL include a test case that creates an
`elasticstack_elasticsearch_security_api_key` resource with only `name` set (no
`role_descriptors`, no `expiration`) and verifies that the apply succeeds and all computed
attributes are populated in state.

#### Scenario: Acceptance test exercises create without role_descriptors

- GIVEN a Terraform configuration with only `name` set on the API key resource
- WHEN the acceptance test runs `terraform apply`
- THEN the test SHALL pass without error
- AND `resource.TestCheckResourceAttrSet` SHALL confirm `id`, `key_id`, `api_key`, and `encoded` are set
