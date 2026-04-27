## MODIFIED Requirements

### Requirement: Identity and import (REQ-004)

The resource SHALL expose computed `id` and `policy_id` attributes whose values are set from the Kibana package policy id returned by the API. `policy_id` SHALL be the import key. Changes to a configured `policy_id` SHALL require replacement.

The resource SHALL support import with both plain and composite import IDs.

When the import ID is a composite string in the format `<space_id>/<policy_id>`, the resource SHALL set `policy_id` to the parsed resource ID and `space_ids` to `[<space_id>]` in state, so that Defend integration policies in non-default Kibana spaces can be imported successfully.

When the import ID is a plain (non-composite) string — i.e. it contains no `/` that can be parsed as a composite ID — the resource SHALL treat the entire string as `policy_id` and SHALL NOT set `space_ids` from the import ID. This preserves existing behaviour for default-space imports.

On the subsequent read after import (regardless of ID form), the resource SHALL use the space from state to query the Fleet API and populate all remaining attributes, including validating that the resolved package policy belongs to the `endpoint` package.

#### Scenario: Import by composite space/policy ID

- GIVEN an existing Elastic Defend package policy in Kibana space `"my-space"` with policy ID `"abc-123"`
- WHEN `terraform import` is run with the composite ID `"my-space/abc-123"`
- THEN `policy_id` SHALL be `"abc-123"` and `space_ids` SHALL contain `"my-space"`

#### Scenario: Import by plain policy ID (default space)

- GIVEN an existing Elastic Defend package policy in the default Kibana space with policy ID `"abc-123"`
- WHEN `terraform import` is run with the plain ID `"abc-123"` (no `/` separator)
- THEN `policy_id` SHALL be `"abc-123"` and `space_ids` SHALL NOT be set from the import ID

#### Scenario: Import by policy id (read-back)

- GIVEN an existing Elastic Defend package policy id
- WHEN `terraform import` is run for `elasticstack_fleet_elastic_defend_integration_policy`
- THEN a subsequent read SHALL populate the modeled schema fields from the API response
