## MODIFIED Requirements

### Requirement: Fleet package policy CRUD, space awareness, and diagnostics (REQ-013)

The resource SHALL use the Kibana Fleet package policy APIs to create, read, update, and delete the underlying package policy. The resource SHALL obtain its Fleet client from provider configuration by default. When `kibana_connection` is configured on the resource, the resource SHALL resolve an effective scoped client from that block and SHALL use the scoped Fleet client for all package policy operations. When `space_ids` is configured or returned, the resource SHALL preserve the operational space needed for subsequent read, update, and delete operations, following the same space-aware lifecycle pattern as the existing Fleet integration policy resource. Transport failures, unexpected response shapes, and API errors SHALL be surfaced as Terraform diagnostics. On read, a not-found response SHALL remove the resource from state.

#### Scenario: Read removes missing Defend policy from state

- GIVEN a Defend package policy that has been deleted outside Terraform
- WHEN the resource refreshes state
- THEN the provider SHALL remove the Terraform resource from state instead of returning a persistent error

#### Scenario: Provider client used by default

- GIVEN `kibana_connection` is not configured on the resource
- WHEN any CRUD operation runs
- THEN the resource SHALL use the provider-configured Fleet client

#### Scenario: Scoped Fleet connection

- GIVEN `kibana_connection` is configured on the resource
- WHEN any CRUD operation runs
- THEN the resource SHALL use the scoped Fleet client derived from that block
