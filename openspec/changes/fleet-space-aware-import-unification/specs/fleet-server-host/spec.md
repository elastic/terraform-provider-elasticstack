## MODIFIED Requirements

### Requirement: Import (REQ-007)

The resource SHALL support import with both plain and composite import IDs.

When the import ID is a composite string in the format `<space_id>/<host_id>`, the resource SHALL set `host_id` to the parsed resource ID and `space_ids` to `[<space_id>]` in state, so that server hosts in non-default Kibana spaces can be imported successfully.

When the import ID is a plain (non-composite) string — i.e. it contains no `/` that can be parsed as a composite ID — the resource SHALL treat the entire string as `host_id` and SHALL NOT set `space_ids` from the import ID. This preserves existing behaviour for default-space imports.

On the subsequent read after import (regardless of ID form), the resource SHALL use the space from state to query the Fleet API and populate all remaining attributes.

#### Scenario: Import by composite space/host ID

- GIVEN an existing Fleet server host in Kibana space `"my-space"` with host ID `"abc-123"`
- WHEN `terraform import` is run with the composite ID `"my-space/abc-123"`
- THEN `host_id` SHALL be `"abc-123"` and `space_ids` SHALL contain `"my-space"`

#### Scenario: Import by plain host ID (default space)

- GIVEN an existing Fleet server host in the default Kibana space with host ID `"abc-123"`
- WHEN `terraform import` is run with the plain ID `"abc-123"` (no `/` separator)
- THEN `host_id` SHALL be `"abc-123"` and `space_ids` SHALL NOT be set from the import ID
