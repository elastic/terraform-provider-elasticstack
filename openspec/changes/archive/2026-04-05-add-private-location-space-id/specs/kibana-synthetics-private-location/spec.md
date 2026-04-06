## ADDED Requirements

### Requirement: `space_id` attribute (REQ-010)

The resource SHALL expose an optional `space_id` string attribute that selects the Kibana space used for create, read, and delete. When `space_id` is omitted or set to an empty string, the provider SHALL use the default Kibana space. When `space_id` is set to a non-empty value, the provider SHALL use that Kibana space for all Synthetics Private Location API calls. The attribute SHALL use plan modifiers such that changing `space_id` requires resource replacement. The provider SHALL persist `space_id` in state from configuration (and reflect it on read as applicable).

#### Scenario: Default space when `space_id` omitted

- GIVEN configuration does not set `space_id` or sets it to an empty string
- WHEN create, read, or delete runs
- THEN the provider SHALL issue API requests for the default Kibana space

#### Scenario: Non-default space

- GIVEN configuration sets `space_id` to a non-empty Kibana space identifier
- WHEN create, read, or delete runs
- THEN the provider SHALL issue API requests scoped to that space

#### Scenario: Replace on `space_id` change

- GIVEN an existing managed private location
- WHEN `space_id` changes in configuration
- THEN Terraform SHALL plan replacement for the resource

### Requirement: Import with non-default space (REQ-011)

When the practitioner imports a private location that exists in a non-default Kibana space, they SHALL configure `space_id` to match that space. The import identifier SHALL remain the Kibana private location id (including existing composite id parsing rules when the id contains `/`).

#### Scenario: Import requires matching `space_id` for non-default space

- GIVEN a private location exists only in a non-default Kibana space
- WHEN the practitioner runs import with the correct Kibana id but omits `space_id` or uses the wrong space
- THEN subsequent read MAY receive 404 and the provider SHALL apply existing 404 handling (remove from state) or fail as appropriate; the practitioner SHALL set `space_id` to the correct space for a successful read

## MODIFIED Requirements

### Requirement: Synthetics Private Locations API (REQ-001)

The resource SHALL manage private locations through Kibana's Synthetics Private Locations API: create via the legacy Kibana client's `KibanaSynthetics.PrivateLocation.Create`, read via `KibanaSynthetics.PrivateLocation.Get`, and delete via `KibanaSynthetics.PrivateLocation.Delete`. The provider SHALL pass the effective Kibana space derived from `space_id` (per REQ-010) into these operations so that requests use the correct space-scoped API paths.

#### Scenario: CRUD uses Private Locations API

- GIVEN a managed Synthetics private location
- WHEN create, read, or delete runs
- THEN the provider SHALL use the corresponding Kibana Synthetics Private Location API operation with the effective space from `space_id`

### Requirement: All mutable fields require replacement (REQ-007)

Changes to `id`, `label`, `agent_policy_id`, `tags`, or `space_id` SHALL each require resource replacement rather than an in-place update. The `geo` block does not carry `RequiresReplace` independently but changes to it in practice trigger replacement through the interaction with REQ-006 (update not supported).

#### Scenario: Replace on `label` change

- GIVEN an existing managed private location
- WHEN `label` changes in configuration
- THEN Terraform SHALL plan replacement for the resource

#### Scenario: Replace on `agent_policy_id` change

- GIVEN an existing managed private location
- WHEN `agent_policy_id` changes in configuration
- THEN Terraform SHALL plan replacement for the resource

#### Scenario: Replace on `tags` change

- GIVEN an existing managed private location
- WHEN `tags` changes in configuration
- THEN Terraform SHALL plan replacement for the resource

#### Scenario: Replace on `space_id` change

- GIVEN an existing managed private location
- WHEN `space_id` changes in configuration
- THEN Terraform SHALL plan replacement for the resource
