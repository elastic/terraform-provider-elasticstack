## MODIFIED Requirements

### Requirement: Terraform import

The resource SHALL support `terraform import` by accepting either `<space_id>/<source_id>` or `<source_id>`. The Fleet download source ID SHALL populate `source_id` in state. When a space is provided, the space ID SHALL populate `space_ids` as a single-entry collection. When only `<source_id>` is provided (no `/` separator), `space_ids` SHALL NOT be set from the import ID; the subsequent read cycle SHALL populate all remaining attributes using the default space.

#### Scenario: Import by composite ID

- **WHEN** the practitioner runs import with a valid `<space_id>/<source_id>` identifier
- **THEN** state SHALL contain the parsed resource ID in `source_id`, and `space_ids` SHALL contain exactly the provided `space_id`

#### Scenario: Import by source ID only

- **WHEN** the practitioner runs import with `<source_id>` and no explicit space prefix
- **THEN** state SHALL contain `source_id` set to the provided value, and `space_ids` SHALL NOT be set from the import ID
