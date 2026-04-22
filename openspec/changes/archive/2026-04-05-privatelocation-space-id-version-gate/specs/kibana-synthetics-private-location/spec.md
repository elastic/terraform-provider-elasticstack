## MODIFIED Requirements

### Requirement: `space_id` attribute (REQ-010)

The resource SHALL expose an optional `space_id` string attribute that selects the Kibana space used for create, read, and delete. When `space_id` is omitted or set to an empty string, the provider SHALL use the default Kibana space. When `space_id` is set to a non-empty value, the provider SHALL use that Kibana space for all Synthetics Private Location API calls. **When the effective Kibana space for API calls is not the default space (non-empty effective `space_id` after composite import resolution as defined for this resource), the provider SHALL require Elastic Stack 9.4.0-SNAPSHOT or higher and SHALL surface an error diagnostic that states the minimum version if the connected stack is older.** When the effective space is the default space, no such version requirement applies for this attribute. The attribute SHALL use plan modifiers such that changing `space_id` requires resource replacement. The provider SHALL persist `space_id` in state from configuration (and reflect it on read as applicable). The `space_id` attribute documentation SHALL mention the minimum Elastic Stack version for non-default space usage.

#### Scenario: Default space when `space_id` omitted

- GIVEN configuration does not set `space_id` or sets it to an empty string
- WHEN create, read, or delete runs
- THEN the provider SHALL issue API requests for the default Kibana space

#### Scenario: Non-default space on a supported stack

- GIVEN configuration sets `space_id` to a non-empty Kibana space identifier
- AND the connected Elastic Stack is at least 9.4.0-SNAPSHOT
- WHEN create, read, or delete runs
- THEN the provider SHALL issue API requests scoped to that space

#### Scenario: Non-default space on an unsupported stack

- GIVEN the effective Kibana space for API calls is not the default space
- AND the connected Elastic Stack is older than 9.4.0-SNAPSHOT
- WHEN create, read, or delete runs
- THEN the provider SHALL return an error diagnostic that includes the minimum required version

#### Scenario: Replace on `space_id` change

- GIVEN an existing managed private location
- WHEN `space_id` changes in configuration
- THEN Terraform SHALL plan replacement for the resource

### Requirement: Import with non-default space (REQ-011)

When the practitioner imports a private location that exists in a non-default Kibana space, they SHALL use a composite import identifier in the format `<space_id>/<private_location_id>` so the provider can read from the correct Kibana space, **and the connected stack SHALL be at least Elastic Stack 9.4.0-SNAPSHOT for that non-default space resolution to be supported**. After a successful read, the provider SHALL persist `space_id` in state.

#### Scenario: Import requires matching composite `space_id` for non-default space

- GIVEN a private location exists only in a non-default Kibana space
- WHEN the practitioner runs import with a bare Kibana id or a composite id whose `space_id` segment does not match the location's space
- THEN subsequent read MAY receive 404 and the provider SHALL apply existing 404 handling (remove from state) or fail as appropriate

#### Scenario: Import composite id on an unsupported stack

- GIVEN an import identifier resolves to a non-default Kibana space
- AND the connected Elastic Stack is older than 9.4.0-SNAPSHOT
- WHEN read or delete runs
- THEN the provider SHALL return an error diagnostic that includes the minimum required version
