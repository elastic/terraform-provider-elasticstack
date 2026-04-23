## MODIFIED Requirements

### Requirement: State mapping — SSL null when no SSL in API response (REQ-019)

When the API response contains no SSL block (nil), the resource SHALL set `ssl` to null in state. When the API response contains an SSL block but all fields (`certificate`, `certificate_authorities`, `key`, `verification_mode`) resolve to empty/nil, the resource SHALL also set `ssl` to null.

#### Scenario: No SSL in API response

- GIVEN the output has no SSL configured in Fleet
- WHEN read runs
- THEN `ssl` SHALL be null in state

#### Scenario: SSL block with only verification_mode

- GIVEN the API returns an SSL block containing only `verification_mode = "none"` and no certificate fields
- WHEN read runs
- THEN `ssl` SHALL be non-null in state with `verification_mode = "none"`

## ADDED Requirements

### Requirement: SSL verification_mode attribute (REQ-025)

The `ssl` block of `elasticstack_fleet_output` SHALL expose a `verification_mode` optional string attribute. Valid values are `"certificate"`, `"full"`, `"none"`, and `"strict"`. The schema validator SHALL reject any other value at plan time. When `verification_mode` is not configured, it SHALL be null in state and SHALL NOT be sent to the API.

#### Scenario: Set verification_mode to none

- WHEN a resource configuration sets `ssl.verification_mode = "none"`
- THEN schema validation SHALL accept the value

#### Scenario: Set verification_mode to invalid value

- WHEN a resource configuration sets `ssl.verification_mode = "invalid"`
- THEN schema validation SHALL return an error

#### Scenario: verification_mode written on create

- WHEN a resource with `ssl.verification_mode = "none"` is created
- THEN the create API request SHALL include `ssl.verification_mode = "none"`

#### Scenario: verification_mode written on update

- WHEN a resource with `ssl.verification_mode = "certificate"` is updated
- THEN the update API request SHALL include `ssl.verification_mode = "certificate"`

#### Scenario: verification_mode read back into state

- GIVEN the Fleet API returns `ssl.verification_mode = "strict"` in the output
- WHEN read runs
- THEN state SHALL contain `ssl.verification_mode = "strict"`

#### Scenario: verification_mode null when unset

- GIVEN no `ssl.verification_mode` is configured
- WHEN the resource is created and read
- THEN `ssl.verification_mode` SHALL be null in state and no `verification_mode` field SHALL appear in the API request

#### Scenario: verification_mode applies to all output types

- WHEN `ssl.verification_mode` is configured on an output of type `elasticsearch`, `remote_elasticsearch`, `logstash`, or `kafka`
- THEN the attribute SHALL be sent and read back for each output type
