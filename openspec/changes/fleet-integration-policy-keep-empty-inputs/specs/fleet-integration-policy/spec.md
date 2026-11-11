## MODIFIED Requirements

### Requirement: Create — read-back after create (REQ-013)

After a successful create, the resource SHALL retrieve package info for the created policy's package (name and version) from the Fleet registry cache. The resource SHALL call `populateFromAPI` to set all state fields from the API response. When `inputs` was null, unknown or an empty map in the plan, the resource SHALL not populate `inputs` from the API response (to avoid provider-produced inconsistent result errors): an empty map in the plan SHALL be stored as an empty map, otherwise `inputs` SHALL be null.

#### Scenario: Inputs omitted in plan

- GIVEN `inputs` is null in the plan
- WHEN create completes
- THEN `inputs` in state SHALL be null (not populated from API)

#### Scenario: Empty inputs map in plan

- GIVEN `inputs` is an empty map in the plan
- WHEN create completes
- THEN `inputs` in state SHALL be an empty map (not populated from API, not null)

### Requirement: Update — inputs state preservation (REQ-015)

On update, when neither the prior state nor the plan had inputs configured (both null or empty), the resource SHALL not populate `inputs` from the API response. An empty map in the plan SHALL be stored as an empty map; otherwise `inputs` SHALL be null.

#### Scenario: Inputs not added during update when unconfigured

- GIVEN `inputs` is null in both prior state and plan
- WHEN update completes
- THEN `inputs` in the updated state SHALL remain null

#### Scenario: Empty inputs map applied to a policy without inputs

- GIVEN `inputs` is null in prior state and an empty map in the plan
- WHEN update completes
- THEN `inputs` in the updated state SHALL be an empty map

### Requirement: Read — import detection and inputs population (REQ-017)

During read, when `policy_id` has a value but `name` is null or unknown (indicating an import operation), the resource SHALL always populate `inputs` from the API response. When `inputs` was previously known and non-empty in state, the resource SHALL populate `inputs` from the API. When `inputs` was previously null or an empty map (user did not configure inputs), the resource SHALL not populate `inputs` from the API, leaving it as it was in state.

#### Scenario: Import populates all inputs

- GIVEN an import where only `policy_id` is set in state
- WHEN read runs after import
- THEN `inputs` SHALL be fully populated from the API response

#### Scenario: Empty inputs map survives refresh

- GIVEN `inputs` is an empty map in state
- WHEN read runs
- THEN `inputs` SHALL remain an empty map
