## MODIFIED Requirements

### Requirement: Kibana Workflow APIs (REQ-001)

The resource SHALL manage workflows through Kibana Workflow API operations corresponding to create, get, update, and delete. The Terraform lifecycle orchestration for these operations SHALL be performed through the shared Kibana Plugin Framework generic resource framework, while workflow-specific request construction, response mapping, and workflow transport remain resource-specific components. After a successful create or update call, the shared framework SHALL perform a follow-up get request and SHALL use that get response as the authoritative source for Terraform state.

#### Scenario: Create then authoritative read

- GIVEN a successful workflow create request
- WHEN the resource completes create
- THEN it SHALL read the created workflow from Kibana and SHALL write state from that read response

#### Scenario: Update then authoritative read

- GIVEN a successful workflow update request
- WHEN the resource completes update
- THEN it SHALL read the workflow from Kibana again and SHALL write state from that read response

### Requirement: Stack version gate (REQ-003)

Before create, read, update, and delete, the resource SHALL verify that the target Elastic Stack version is at least `9.4.0-SNAPSHOT`. The minimum-version requirement and unsupported-version message SHALL be supplied by the resource model and enforced by the shared Kibana Plugin Framework generic resource framework. If the stack version is lower, the resource SHALL fail the operation with an `Unsupported server version` diagnostic stating that Agent Builder workflows require that minimum version or later.

#### Scenario: Stack below minimum version

- GIVEN a target Elastic Stack version below `9.4.0-SNAPSHOT`
- WHEN any workflow resource operation runs
- THEN the operation SHALL fail with an unsupported-version diagnostic before calling the workflow API

### Requirement: Identity and canonical `id` (REQ-005)

The resource SHALL expose a computed canonical `id` in the format `<space_id>/<workflow_id>`. After create, read, or update receives a workflow from Kibana, the resource SHALL set `workflow_id` from the API workflow id and SHALL set `id` from the current `space_id` plus that workflow id. If the working model does not yet hold a `space_id`, it SHALL use `default` when constructing the canonical id. Shared composite-ID helpers from the Kibana Plugin Framework generic resource framework SHALL support parsing and restoring the space-aware identity used by the resource.

#### Scenario: Canonical composite id in state

- GIVEN a workflow read in space `default` with workflow id `workflow-1234`
- WHEN Terraform state is written
- THEN `workflow_id` SHALL be `workflow-1234` and `id` SHALL be `default/workflow-1234`

### Requirement: Import passthrough and composite id expectations (REQ-006)

The resource SHALL support Terraform import by passing the supplied import identifier directly into the `id` attribute. Because read, update, and delete parse `id` as a composite identifier, a functional imported id MUST be in the format `<space_id>/<workflow_id>` with exactly one `/`; if later operations parse a different format, they SHALL return the composite-id diagnostic produced by the shared composite-ID helpers in the Kibana Plugin Framework generic resource framework.

#### Scenario: Composite import id

- GIVEN an import id `observability/workflow-1234`
- WHEN import runs
- THEN the resource SHALL persist that string to `id` for subsequent refresh

#### Scenario: Non-composite imported id

- GIVEN an imported `id` that does not contain exactly one `/`
- WHEN a later read, update, or delete parses the id
- THEN the resource SHALL return an error diagnostic describing the required composite id format

### Requirement: Read behavior and missing resources (REQ-010)

When refreshing state, the resource SHALL parse the composite `id` to obtain the current `space_id` and `workflow_id`, SHALL restore `space_id` into the model before mapping the API response, and SHALL remove the resource from Terraform state when Kibana reports that the workflow is not found. The shared Kibana Plugin Framework generic resource framework SHALL perform the common not-found state-removal behavior.

#### Scenario: Read of deleted workflow

- GIVEN a resource recorded in Terraform state
- WHEN the read operation calls Kibana and the workflow is not found
- THEN the provider SHALL remove the resource from state
