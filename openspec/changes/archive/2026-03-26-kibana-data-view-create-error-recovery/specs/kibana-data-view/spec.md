## MODIFIED Requirements

### Requirement: API and client error surfacing (REQ-002)

For create, read, update, and delete, when the provider cannot obtain the Kibana OpenAPI client, the operation SHALL return an error diagnostic. For read and update, transport errors and unexpected HTTP statuses SHALL be surfaced as error diagnostics. For create, transport errors and unexpected HTTP statuses SHALL be surfaced as error diagnostics unless the provider can deterministically reconcile a managed data view create under REQ-014. Delete SHALL also surface transport errors and unexpected HTTP statuses, except that delete not-found SHALL be treated as success.

#### Scenario: Missing Kibana OpenAPI client

- **GIVEN** the resource cannot obtain a Kibana OpenAPI client from provider configuration
- **WHEN** any CRUD operation runs
- **THEN** the operation SHALL fail with an error diagnostic

#### Scenario: Create error without deterministic reconciliation

- **GIVEN** a create request that does not meet the managed reconciliation conditions in REQ-014
- **WHEN** Kibana returns a transport error or unexpected HTTP status for create
- **THEN** the provider SHALL surface an error diagnostic and SHALL NOT record Terraform state for the resource

#### Scenario: Delete not found

- **GIVEN** a delete request for a data view that is already absent
- **WHEN** Kibana returns HTTP 404
- **THEN** the provider SHALL treat the delete as successful

## ADDED Requirements

### Requirement: Managed create reconciliation after an error response (REQ-014)

When a create request supplies an explicit `data_view.id`, the provider SHALL treat that identifier as the managed identity for create reconciliation. If Kibana persists the create request but returns an error or unexpected HTTP status to the provider, the provider SHALL perform a follow-up read of that same data view id in the target `space_id`. If the read succeeds, the provider SHALL populate Terraform state from the read result and complete create successfully. If the read fails or the data view is not found, the provider SHALL surface the original create failure and SHALL NOT write state.

#### Scenario: Managed create succeeds server-side but returns an error response

- **GIVEN** configuration sets an explicit `data_view.id` and target `space_id`
- **AND** Kibana persists the data view create request
- **AND** Kibana returns an error or unexpected HTTP status for the create call
- **WHEN** the provider handles the create result
- **THEN** the provider SHALL read the data view by that configured id in the same space
- **AND** SHALL populate Terraform state from the read result
- **AND** SHALL complete create without leaving the resource unmanaged

#### Scenario: Managed create error cannot be reconciled

- **GIVEN** configuration sets an explicit `data_view.id`
- **AND** Kibana returns an error or unexpected HTTP status for the create call
- **AND** a follow-up read by that id does not return the created data view
- **WHEN** the provider handles the create result
- **THEN** the provider SHALL surface the original create failure
- **AND** SHALL NOT write Terraform state for the resource

#### Scenario: Create without explicit managed id

- **GIVEN** configuration does not set `data_view.id`
- **WHEN** Kibana returns an error or unexpected HTTP status for the create call
- **THEN** the provider SHALL NOT attempt heuristic reconciliation by title or other mutable fields under REQ-014
- **AND** SHALL surface the create failure as an error diagnostic
