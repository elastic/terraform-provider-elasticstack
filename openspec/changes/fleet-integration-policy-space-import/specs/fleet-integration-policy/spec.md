# Delta spec: `fleet-integration-policy` — space-aware import

Modifies: `openspec/specs/fleet-integration-policy/spec.md`

## MODIFIED Requirements

### Requirement: Import (REQ-006)

The resource SHALL support import with both plain and composite import IDs.

When the import ID is a composite string in the format `<space_id>/<policy_id>` (as
produced by `clients.CompositeIDFromStrFw`), the resource SHALL set `policy_id` to the
parsed resource-ID segment and SHALL set `space_ids` to a single-element set containing the
space-ID segment. The subsequent read SHALL query the package-policy API in the named space,
so that policies created in non-default Kibana spaces can be imported successfully.

When the import ID is a plain (non-composite) string — i.e. it contains no `/` separator
that `clients.CompositeIDFromStrFw` recognises as a composite ID — the resource SHALL treat
the entire string as `policy_id` and SHALL NOT set `space_ids` from the import ID. This
preserves existing behaviour for default-space imports.

On the subsequent read after import (regardless of ID form), the resource SHALL populate all
attributes from the Fleet API response, including inputs.

#### Scenario: Import by composite space/policy ID

- GIVEN a package policy that exists in the Kibana space `"my-space"` with policy ID
  `"abc-123"`
- WHEN `terraform import` is run with the composite ID `"my-space/abc-123"`
- THEN `policy_id` SHALL be `"abc-123"`, `space_ids` SHALL contain `"my-space"`, and a
  subsequent refresh SHALL populate all state fields from the API

#### Scenario: Import by plain policy ID (default space)

- GIVEN a package policy that exists in the default Kibana space with policy ID `"abc-123"`
- WHEN `terraform import` is run with the plain ID `"abc-123"` (no `/` separator)
- THEN `policy_id` SHALL be `"abc-123"`, `space_ids` SHALL NOT be set from the import ID,
  and a subsequent refresh SHALL populate all state fields from the API
