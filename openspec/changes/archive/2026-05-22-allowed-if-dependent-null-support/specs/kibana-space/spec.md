## MODIFIED Requirements

### Requirement: Create and update behavior (REQ-011–REQ-013)
On create, the resource SHALL call the Kibana Create Space API and then immediately read the space back to refresh state. On update, the resource SHALL call the Kibana Update Space API and then immediately read the space back to refresh state. For both create and update, the resource SHALL build the request body from the configured `space_id`, `name`, and any optional fields (`description`, `disabled_features`, `initials`, `color`, `image_url`, `solution`), omitting optional fields that are not set in configuration. The resource SHALL allow `disabled_features` to be configured when `solution` is `classic`, when `solution` is omitted, or when `solution` is unknown during plan-time validation, and SHALL reject `disabled_features` only when `solution` has a known non-`classic` value.

#### Scenario: Post-create refresh
- GIVEN a successful Create Space API response
- WHEN the resource finishes creating
- THEN it SHALL read the space and populate state

#### Scenario: disabled_features allowed for classic solution
- **WHEN** configuration sets `disabled_features` and `solution = "classic"`
- **THEN** plan-time validation SHALL accept the configuration

#### Scenario: disabled_features allowed when solution is omitted
- **WHEN** configuration sets `disabled_features` and does not set `solution`
- **THEN** plan-time validation SHALL accept the configuration

#### Scenario: disabled_features rejected for non-classic solution
- **WHEN** configuration sets `disabled_features` and `solution` has a known value other than `classic`
- **THEN** plan-time validation SHALL return a validation error
