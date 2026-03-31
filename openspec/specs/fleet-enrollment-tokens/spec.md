# `elasticstack_fleet_enrollment_tokens` — Schema and Functional Requirements

Data source implementation: `internal/fleet/enrollmenttokens`

## Purpose

Define schema and behavior for the Fleet enrollment tokens data source. The data source lists Fleet enrollment tokens, optionally filtered by agent policy ID and Kibana space. It is read-only; tokens cannot be created or deleted via this data source.

## Schema

```hcl
data "elasticstack_fleet_enrollment_tokens" "example" {
  id        = <computed, string>   # policy_id when set, otherwise hash of Kibana URL
  policy_id = <optional, string>   # filter tokens by agent policy ID
  space_id  = <optional, string>   # Kibana space to query

  tokens = <computed, list(object({
    key_id     = string
    api_key    = string (sensitive)
    api_key_id = string
    created_at = string
    name       = string
    active     = bool
    policy_id  = string
  }))>
}
```

## Requirements

### Requirement: Fleet Enrollment Tokens API (REQ-001)

The data source SHALL use the Fleet Enrollment Tokens API to retrieve tokens. When the Fleet API returns an error, the data source SHALL surface it to Terraform diagnostics.

#### Scenario: API error

- GIVEN the Fleet API returns an error
- WHEN the data source read runs
- THEN diagnostics SHALL contain the API error

### Requirement: List all tokens (REQ-002)

When `policy_id` is not configured, the data source SHALL call `GetEnrollmentTokens` with the configured `space_id` to retrieve all enrollment tokens in the given space (or the default space when `space_id` is not set).

#### Scenario: All tokens, default space

- GIVEN neither `policy_id` nor `space_id` is configured
- WHEN read runs
- THEN `GetEnrollmentTokens` SHALL be called with an empty space ID

#### Scenario: All tokens, named space

- GIVEN `space_id = "my-space"` and `policy_id` is not configured
- WHEN read runs
- THEN `GetEnrollmentTokens` SHALL be called with `"my-space"` as the space ID

### Requirement: Filter by policy (REQ-003)

When `policy_id` is configured and `space_id` is not configured (or equals `"default"`), the data source SHALL call `GetEnrollmentTokensByPolicy` using the configured `policy_id`. When both `policy_id` and a non-default `space_id` are configured, the data source SHALL call `GetEnrollmentTokensByPolicyInSpace` using both.

#### Scenario: Filter by policy, default space

- GIVEN `policy_id = "my-policy"` and `space_id` is not set
- WHEN read runs
- THEN `GetEnrollmentTokensByPolicy` SHALL be called with `"my-policy"`

#### Scenario: Filter by policy in named space

- GIVEN `policy_id = "my-policy"` and `space_id = "my-space"`
- WHEN read runs
- THEN `GetEnrollmentTokensByPolicyInSpace` SHALL be called with both `"my-policy"` and `"my-space"`

### Requirement: Computed id (REQ-004)

When `policy_id` is configured, the data source SHALL set `id` to the value of `policy_id`. When `policy_id` is not configured, the data source SHALL set `id` to a hash of the Kibana client URL. If computing the URL hash fails, the data source SHALL return an error diagnostic.

#### Scenario: id from policy_id

- GIVEN `policy_id = "abc123"`
- WHEN read completes
- THEN `id` SHALL equal `"abc123"`

#### Scenario: id from URL hash

- GIVEN `policy_id` is not configured
- WHEN read completes
- THEN `id` SHALL equal a deterministic hash of the Kibana URL

### Requirement: Token state mapping (REQ-005)

The data source SHALL populate the `tokens` list with one entry per token returned by the API. Each entry SHALL include `key_id` (from `Id`), `api_key`, `api_key_id`, `created_at`, `name`, `active`, and `policy_id` from the API response.

#### Scenario: All token fields mapped

- GIVEN the Fleet API returns a token with all fields populated
- WHEN read completes
- THEN the corresponding `tokens` entry SHALL have all fields set from the API response

### Requirement: Sensitive api_key (REQ-006)

The `tokens[*].api_key` attribute SHALL be marked sensitive so Terraform redacts its value from plan and apply output.

#### Scenario: api_key not shown in plan

- GIVEN a token with an api_key
- WHEN the data source is used in a plan
- THEN `api_key` SHALL be displayed as `(sensitive value)` in Terraform output
