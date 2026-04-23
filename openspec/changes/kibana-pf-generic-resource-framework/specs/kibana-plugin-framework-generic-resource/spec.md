## ADDED Requirements

### Requirement: Shared Kibana Plugin Framework CRUD orchestration
The provider SHALL offer a reusable framework for Kibana Plugin Framework CRUD resources that centralizes common Terraform lifecycle orchestration while delegating resource-specific schema, Terraform model mapping, and Kibana transport logic to typed components.

#### Scenario: Shared lifecycle flow is reused
- **WHEN** a Kibana Plugin Framework CRUD resource is implemented on the framework
- **THEN** provider configuration, scoped Kibana client resolution, version enforcement, CRUD orchestration, and state persistence SHALL be handled by the shared framework rather than bespoke per-resource scaffolding

### Requirement: Separation of orchestration, model mapping, and transport
The framework SHALL separate responsibilities into three layers: a generic resource layer for Terraform lifecycle orchestration, a model layer for Terraform request/response mapping and version requirements, and a transport layer for resource-family-specific Kibana API operations.

#### Scenario: Model does not perform transport operations
- **WHEN** a resource model constructs create or update requests or populates Terraform state from remote data
- **THEN** the model SHALL NOT call Kibana API transport helpers directly

#### Scenario: Transport does not own Terraform lifecycle state
- **WHEN** a focused Kibana transport API performs create, get, update, or delete operations
- **THEN** it SHALL NOT decode Terraform plan/state or persist Terraform state directly

### Requirement: Model-driven version requirements
The framework SHALL let each resource model compute the minimum supported Kibana version and the user-facing unsupported-version diagnostic for the concrete configured shape of that resource.

#### Scenario: Version requirement is computed before API calls
- **WHEN** create, read, update, or delete runs for a framework-based resource
- **THEN** the framework SHALL obtain the version requirement from the model and SHALL enforce it before invoking transport operations

### Requirement: Shared read-after-write reconciliation
The framework SHALL perform a follow-up read after every successful create and update, and SHALL populate Terraform state from that authoritative read response.

#### Scenario: Create uses authoritative remote state
- **WHEN** a framework-based resource completes a successful create
- **THEN** the framework SHALL read the remote object again and SHALL write Terraform state from that read response

#### Scenario: Update uses authoritative remote state
- **WHEN** a framework-based resource completes a successful update
- **THEN** the framework SHALL read the remote object again and SHALL write Terraform state from that read response

### Requirement: Shared remote-not-found read behavior
For framework-based resources, when the transport layer reports that the remote object is not found during read, the framework SHALL remove the resource from Terraform state rather than returning a successful populated state.

#### Scenario: Refresh after out-of-band deletion
- **WHEN** a framework-based resource is refreshed after the remote object has been deleted out of band
- **THEN** the framework SHALL remove the resource from Terraform state

### Requirement: Shared composite-ID support for space-aware Kibana resources
The framework SHALL provide reusable helpers for Kibana resources whose canonical Terraform identity is a composite `<space_id>/<resource_id>` string, including parsing stored IDs, restoring `space_id` into the model when needed, and composing canonical IDs written to state.

#### Scenario: Space-aware resource restores space from composite id
- **WHEN** a framework-based Kibana resource reads a remote object using a stored composite id
- **THEN** the framework SHALL make the parsed `space_id` available for transport calls and model state population
