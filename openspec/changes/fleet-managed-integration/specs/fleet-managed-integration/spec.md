## ADDED Requirements

### Requirement: Resource type and registration

The `elasticstack_fleet_managed_integration` resource SHALL be registered in `experimentalResources()` in `provider/plugin_framework.go`, matching the upstream tech-preview status of the Elastic Managed Integration feature. The old `elasticstack_fleet_agentless_policy` resource SHALL be removed from the provider and SHALL NOT appear in the provider schema.

#### Scenario: New resource type is discoverable
- **WHEN** `terraform providers schema -json` is run against the provider
- **THEN** `elasticstack_fleet_managed_integration` SHALL appear in the resource schema
- **AND** `elasticstack_fleet_agentless_policy` SHALL NOT appear in the resource schema

#### Scenario: Resource registered as experimental
- **WHEN** the provider is initialised
- **THEN** the resource SHALL be registered via `experimentalResources()` only

### Requirement: CRUD via managed_integrations API

The resource SHALL call the following Kibana Fleet managed_integrations endpoints, using the space-aware path:

| Operation | Endpoint |
|-----------|----------|
| Create    | `POST /api/fleet/managed_integrations` |
| Read      | `GET /api/fleet/managed_integrations/{id}` |
| Update    | `PUT /api/fleet/managed_integrations/{id}` |
| Delete    | `DELETE /api/fleet/managed_integrations/{id}` |

No fallback through `/api/fleet/package_policies/{id}` SHALL exist. On HTTP 404, Read SHALL remove the resource from state without error; Delete SHALL treat 404 as success.

#### Scenario: Create calls managed_integrations endpoint
- **WHEN** a `elasticstack_fleet_managed_integration` resource is applied for the first time
- **THEN** `POST /api/fleet/managed_integrations` SHALL be called with the request body derived from the config
- **AND** state SHALL be populated from the `KibanaHTTPAPIsManagedIntegration` response

#### Scenario: Read calls managed_integrations GET endpoint
- **WHEN** a refresh or plan is run against an existing resource
- **THEN** `GET /api/fleet/managed_integrations/{id}` SHALL be called
- **AND** no call to `/api/fleet/package_policies/{id}` SHALL occur

#### Scenario: Read removes resource on 404
- **WHEN** `GET /api/fleet/managed_integrations/{id}` returns HTTP 404
- **THEN** the resource SHALL be removed from state without error

#### Scenario: Update calls managed_integrations PUT endpoint
- **WHEN** a change to an updatable attribute is applied
- **THEN** `PUT /api/fleet/managed_integrations/{id}` SHALL be called with a full-replace body derived from the plan
- **AND** no call to `/api/fleet/package_policies/{id}` SHALL occur

#### Scenario: Delete calls managed_integrations DELETE endpoint
- **WHEN** the resource is destroyed
- **THEN** `DELETE /api/fleet/managed_integrations/{id}` SHALL be called
- **AND** no call to `/api/fleet/agentless_policies/{id}` SHALL occur

#### Scenario: Delete tolerates 404
- **WHEN** `DELETE /api/fleet/managed_integrations/{id}` returns HTTP 404
- **THEN** the resource SHALL be removed from state without error

### Requirement: Full-replace PUT semantics

The PUT body SHALL be constructed entirely from the current plan (desired state). The body SHALL include all fields from the plan, including `name` and `package.version`. The `cloud_connector` field SHALL be derived from state (preserved association) and always re-sent when a cloud connector is associated; if omitted from the PUT body, the connector detaches. `cloud_connector.name` and `cloud_connector.target_csp` SHALL NOT be included in the PUT body (they are write-only fields that do not round-trip). No echo-current or overlay mechanism SHALL exist for the update path.

#### Scenario: Full-replace body includes all plan fields
- **WHEN** an update is applied with changes to `name` and `vars_json`
- **THEN** the PUT body SHALL include the new `name`, the new `vars_json`, all `inputs` from the plan, `namespace`, `description`, and all other updatable fields
- **AND** the body SHALL NOT selectively merge with the prior server state

#### Scenario: cloud_connector re-sent from state on update
- **WHEN** an update is applied to a resource that has an associated cloud connector
- **THEN** the PUT body SHALL include `cloud_connector: {enabled, cloud_connector_id}` derived from state
- **AND** `cloud_connector.name` and `cloud_connector.target_csp` SHALL NOT appear in the PUT body

#### Scenario: cloud_connector detaches if omitted from update body
- **WHEN** an update is applied and cloud_connector is absent from the PUT body
- **THEN** the upstream API detaches the connector
- **GIVEN** this requirement, the resource MUST always re-send cloud_connector from state when one is present

### Requirement: In-place name and package.version updates

`name` and `package.version` SHALL be updatable in-place (no `RequiresReplace` plan modifier). Changing them SHALL trigger a PUT to the managed_integrations endpoint rather than a destroy-and-recreate cycle. `package.name` SHALL remain `RequiresReplace` (immutable upstream).

#### Scenario: Name updated in-place
- **WHEN** `name` is changed in config from `"old-name"` to `"new-name"`
- **THEN** Terraform SHALL NOT destroy and recreate the resource
- **AND** `PUT /api/fleet/managed_integrations/{id}` SHALL be called with `name: "new-name"`
- **AND** `id` SHALL remain unchanged after the update

#### Scenario: package.version updated in-place
- **WHEN** `package.version` is changed in config from `"1.0.0"` to `"1.1.0"`
- **THEN** Terraform SHALL NOT destroy and recreate the resource
- **AND** `PUT /api/fleet/managed_integrations/{id}` SHALL be called with the new version

#### Scenario: package.name change forces replacement
- **WHEN** `package.name` is changed in config
- **THEN** Terraform SHALL destroy and recreate the resource

### Requirement: `global_data_tags` as MapNestedAttribute

`global_data_tags` SHALL be modelled as a `MapNestedAttribute` keyed by tag name, with item type `{string_value: StringAttribute, number_value: Float32Attribute}`. Exactly one of `string_value` or `number_value` SHALL be set per tag entry, enforced via `ConflictsWith`+`AtLeastOneOf` validators. This mirrors the `global_data_tags` implementation in `internal/fleet/agentpolicy/schema.go`.

#### Scenario: String-valued global_data_tag
- **WHEN** `global_data_tags = { "env" = { string_value = "prod" } }` is applied
- **THEN** the API SHALL receive `global_data_tags: [{"name": "env", "value": "prod"}]`
- **AND** state SHALL reflect `global_data_tags = { "env" = { string_value = "prod" } }`

#### Scenario: Number-valued global_data_tag
- **WHEN** `global_data_tags = { "priority" = { number_value = 1 } }` is applied
- **THEN** the API SHALL receive `global_data_tags: [{"name": "priority", "value": 1}]`
- **AND** state SHALL reflect `global_data_tags = { "priority" = { number_value = 1 } }`

#### Scenario: Both string_value and number_value set is rejected
- **WHEN** `global_data_tags = { "k" = { string_value = "a", number_value = 1 } }` is set
- **THEN** Terraform SHALL reject the plan with a validation error indicating the conflict

#### Scenario: Neither string_value nor number_value set is rejected
- **WHEN** `global_data_tags = { "k" = {} }` is set
- **THEN** Terraform SHALL reject the plan with a validation error indicating at least one must be set

### Requirement: cloud_connector modelling unchanged

`cloud_connector` SHALL remain a `SingleNestedAttribute` with sub-fields `enabled`, `cloud_connector_id`, `name`, and `target_csp`. The `cloud_connector` attribute SHALL carry a single object-level `RequiresReplace` plan modifier (not one per sub-field), which forces replacement when any sub-field changes. On Read, `name` and `target_csp` SHALL be preserved from prior state (they are write-only wire fields that do not appear in the API GET/PUT response).

#### Scenario: cloud_connector name preserved on Read
- **WHEN** a resource with `cloud_connector = { name = "my-conn", target_csp = "aws", enabled = true }` is read
- **THEN** state SHALL retain `cloud_connector.name = "my-conn"` and `cloud_connector.target_csp = "aws"` from prior state
- **AND** `cloud_connector.cloud_connector_id` and `cloud_connector.enabled` SHALL be populated from the API response

#### Scenario: cloud_connector change forces replacement
- **WHEN** any sub-field of `cloud_connector` is changed in config
- **THEN** Terraform SHALL destroy and recreate the resource

### Requirement: Version gate for managed_integrations endpoint

The resource SHALL declare a `GetVersionRequirements` entry that enforces a minimum Kibana version of **9.5.0** (verified against a 9.5.0-SNAPSHOT build; the same version already used as `policyshape.MinVersionCondition`) for `/api/fleet/managed_integrations`. Against a Kibana version older than 9.5.0, the resource SHALL fail with a helpful error message naming the minimum required version. Because this floor is identical to the version that introduced `condition` support, the resource SHALL NOT perform a separate, distinct capability check for `condition` — `condition` is unconditionally supported once the resource-level floor is satisfied, and no dedicated `SupportsCondition`-style gate SHALL exist.

#### Scenario: Older Kibana returns error
- **WHEN** the resource is planned or applied against a Kibana version older than 9.5.0
- **THEN** Terraform SHALL fail with an error message stating the minimum required version
- **AND** no API call to `/api/fleet/managed_integrations` SHALL be made

#### Scenario: Exactly at version floor succeeds
- **WHEN** the resource is planned against a Kibana version exactly equal to 9.5.0
- **THEN** the version check SHALL pass and API calls SHALL proceed

#### Scenario: Same-core Kibana SNAPSHOT build at release floor succeeds
- **WHEN** the resource is planned against a Kibana version of `9.5.0-SNAPSHOT` (same core as the 9.5.0 release floor)
- **THEN** the Kibana `EnforceMinVersion` check SHALL pass via the provider's Kibana SNAPSHOT convention
- **AND** API calls to `/api/fleet/managed_integrations` SHALL proceed when other preconditions are met

### Requirement: Topology and topology skip-check carried over

The resource SHALL carry over the `topology.go` preflight that rejects self-managed stacks (not ECH or Serverless), with the `skip_topology_check` escape hatch attribute. This logic is endpoint-agnostic and carries over unchanged from the agentless policy resource.

#### Scenario: Self-managed stack rejected
- **WHEN** the resource is applied against a self-managed (non-ECH, non-Serverless) Kibana
- **THEN** Terraform SHALL fail with an error identifying the stack type
- **UNLESS** `skip_topology_check = true` is set

### Requirement: Create-only flag preservation on Read

The following attributes are create-only (not returned by GET) and SHALL be preserved from prior state on every Read to avoid spurious diffs:
- `force`
- `create_dataset_templates`
- `skip_topology_check`

`cloud_connector.name` and `cloud_connector.target_csp` SHALL also be preserved from prior state on Read (write-only wire fields).

#### Scenario: force preserved across plan cycles
- **WHEN** `force = true` is set in config and a plan is run after initial creation
- **THEN** `force` SHALL remain `true` in state
- **AND** no spurious diff SHALL be produced for `force`

### Requirement: Create/delete-only flag updates skip managed_integrations API calls

The attributes `force`, `create_dataset_templates`, `force_delete`, and `skip_topology_check` are not part of the managed_integrations GET or PUT body. None carry `RequiresReplace`, so Terraform still invokes Update when only they change. When the diff is confined to those four attributes, the resource SHALL persist the updated plan to Terraform state without calling any `/api/fleet/managed_integrations/{id}` endpoint (no GET, PUT, or DELETE).

#### Scenario: Create/delete-only flag change updates state without API call
- **WHEN** an existing resource is updated and the only changed attributes are among `force`, `create_dataset_templates`, `force_delete`, and `skip_topology_check`
- **THEN** Terraform state SHALL reflect the new flag values
- **AND** no call to `GET`, `PUT`, or `DELETE` `/api/fleet/managed_integrations/{id}` SHALL occur

### Requirement: Space-aware API calls

All API calls SHALL be space-aware, using `SpaceAwarePathRequestEditor(spaceID)`, where `spaceID` is derived from the `space_ids` attribute (mirroring the existing agentless policy resource and other Fleet resources). The `space_ids` attribute SHALL be a `Computed`+`Optional` set of strings, defaulting to `["default"]` when omitted from config. Changing `space_ids` SHALL force resource replacement.

#### Scenario: Non-default space
- **WHEN** `space_ids = ["my-space"]` is set and the resource is created
- **THEN** all API calls SHALL use the path prefix for `my-space`

#### Scenario: space_ids change forces replacement
- **WHEN** `space_ids` is changed in config
- **THEN** Terraform SHALL destroy and recreate the resource

### Requirement: Import via composite ID

The resource SHALL support import via the composite ID `"<space_id>/<policy_id>"`, using the shared `fleet.SpaceImporter`. On import with a composite ID, Read SHALL parse it to derive the space ID and `policy_id`, and SHALL set `space_ids` to the singleton set `[<space_id>]`. On import with a plain (non-composite) ID, `policy_id` SHALL be set from the given ID and `space_ids` SHALL NOT be set.

#### Scenario: Import by composite ID
- **WHEN** `terraform import elasticstack_fleet_managed_integration.x "default/<policy_id>"` is run
- **THEN** `policy_id` SHALL be set to the parsed value
- **AND** `space_ids` SHALL be set to `["default"]`
- **AND** all other attributes SHALL be populated from `GET /api/fleet/managed_integrations/{id}`

### Requirement: Response type cleanup — no PackagePolicy leakage

State SHALL never contain attributes derived from the `PackagePolicy` type (e.g. `policy_ids`, `revision`, `secret_references`, `output_id`, `supports_agentless`, top-level `enabled`). The `KibanaHTTPAPIsManagedIntegration` response type SHALL be the sole source for state population on Read and after Create/Update. The `populateFromPackagePolicy` function SHALL NOT exist in the `managedintegration` package.

#### Scenario: No PackagePolicy fields in state
- **WHEN** a resource is read after creation
- **THEN** state SHALL NOT contain `policy_ids`, `revision`, `secret_references`, `output_id`, or `supports_agentless`

## REMOVED Requirements

### Requirement: elasticstack_fleet_agentless_policy resource — REMOVED

The `elasticstack_fleet_agentless_policy` resource is removed from the provider entirely. There SHALL be no compatibility shim, deprecation warning, or state upgrade function. `elasticstack_fleet_agentless_policy` has never shipped in a release, so no migration guidance for released users is required.

#### Scenario: Old resource type not registered
- **WHEN** a Terraform configuration references `elasticstack_fleet_agentless_policy`
- **THEN** Terraform SHALL report an error indicating the resource type is unknown
- **AND** no provider code SHALL handle the old resource type
