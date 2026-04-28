## MODIFIED Requirements

### Requirement: Import (REQ-008)

The resource SHALL support import with both plain and composite import IDs.

When the import ID is a composite string in the format `<space_id>/<output_id>`, the resource SHALL set `output_id` to the parsed resource ID and `space_ids` to `[<space_id>]` in state, so that outputs in non-default Kibana spaces can be imported successfully.

When the import ID is a plain (non-composite) string — i.e. it contains no `/` that can be parsed as a composite ID — the resource SHALL treat the entire string as `output_id` and SHALL NOT set `space_ids` from the import ID. This preserves existing behaviour for default-space imports.

On the subsequent read after import (regardless of ID form), the resource SHALL use the space from state to query the Fleet API and populate all remaining attributes.

#### Scenario: Import by composite space/output ID

- GIVEN an existing Fleet output in Kibana space `"my-space"` with output ID `"abc-123"`
- WHEN `terraform import` is run with the composite ID `"my-space/abc-123"`
- THEN `output_id` SHALL be `"abc-123"` and `space_ids` SHALL contain `"my-space"`

#### Scenario: Import by plain output ID (default space)

- GIVEN an existing Fleet output in the default Kibana space with output ID `"abc-123"`
- WHEN `terraform import` is run with the plain ID `"abc-123"` (no `/` separator)
- THEN `output_id` SHALL be `"abc-123"` and `space_ids` SHALL NOT be set from the import ID
