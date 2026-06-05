## ADDED Requirements

### Requirement: Fleet integration resource migrated to use KibanaResource envelope

The system SHALL migrate `internal/fleet/integration` so that its `integrationResource` struct embeds `*entitycore.KibanaResource[integrationModel]` returned by `NewKibanaResource`. The `integrationModel` SHALL implement `KibanaResourceModel` with `GetResourceID()` returning the package identifier derived from `name`/`version` and `GetSpaceID()` returning `m.SpaceID`, and SHALL implement `KibanaUnscopedSpace` with `IsUnscopedSpace()` returning `true` when `SpaceID` is null or unknown (so the envelope skips space-ID validation in that case). The resource's Terraform schema attributes, CRUD semantics, state upgrade behaviour, and externally observable behaviour SHALL remain functionally equivalent after migration.

#### Scenario: Fleet integration resource satisfies envelope interface after migration

- **WHEN** `NewResource()` is called for the fleet integration resource
- **THEN** the returned value SHALL satisfy `resource.Resource` and `resource.ResourceWithConfigure`
- **AND** the Create, Read, Update, and Delete methods SHALL be owned by the envelope
- **AND** `ResourceWithUpgradeState` SHALL continue to be satisfied by the concrete resource type

#### Scenario: Fleet integration unscoped space skips envelope space validation

- **WHEN** an `integrationModel` value has `SpaceID` null or unknown and the envelope evaluates `KibanaUnscopedSpace`
- **THEN** `IsUnscopedSpace()` SHALL return `true`
- **AND** the envelope SHALL NOT emit an "Invalid space identifier" diagnostic for that write

---

### Requirement: Kibana synthetics monitor resource migrated to use KibanaResource envelope

The system SHALL migrate `internal/kibana/synthetics/monitor` so that its `Resource` struct embeds `*entitycore.KibanaResource[tfModelV0]` returned by `NewKibanaResource`. The `tfModelV0` SHALL implement `KibanaResourceModel` with `GetID()` returning the composite `<space_id>:<monitor_id>` stored in `m.ID`, `GetResourceID()` returning the monitor-UUID portion of that composite identifier, and `GetSpaceID()` returning `m.SpaceID`. The resource's Terraform schema attributes, CRUD semantics, import behaviour, and externally observable behaviour SHALL remain functionally equivalent after migration. The dead `synthetics.ESAPIClient` interface, its compile-time assertion in `monitor/resource.go`, the `synthetics.GetKibanaOAPIClient` helper, and the `GetClient()` method on the concrete resource SHALL be removed as part of the migration.

#### Scenario: Synthetics monitor resource satisfies envelope interface after migration

- **WHEN** `NewResource()` is called for the synthetics monitor resource
- **THEN** the returned value SHALL satisfy `resource.Resource` and `resource.ResourceWithConfigure`
- **AND** the Create, Read, Update, and Delete methods SHALL be owned by the envelope

#### Scenario: ESAPIClient interface and helpers are removed

- **WHEN** the synthetics package is compiled after migration
- **THEN** the identifiers `synthetics.ESAPIClient`, `synthetics.GetKibanaOAPIClient`, and the monitor resource's `GetClient` method SHALL NOT exist
- **AND** no compile-time assertion of the form `_ synthetics.ESAPIClient = newResource()` SHALL remain in `monitor/resource.go`

---

### Requirement: Kibana SLO resource migrated to use KibanaResource envelope

The system SHALL migrate `internal/kibana/slo` so that its `Resource` struct embeds `*entitycore.KibanaResource[tfModel]` returned by `NewKibanaResource`. The `tfModel` SHALL implement `KibanaResourceModel` with `GetID()` returning the composite `<space_id>:<slo_id>` stored in `m.ID`, `GetResourceID()` returning the SLO-UUID portion of that composite identifier, and `GetSpaceID()` returning `m.SpaceID`. The SLO Create and Update write callbacks SHALL perform the enabled-reconcile sequence — intermediate read of the server's `enabled` state, conditional Enable/Disable API call when the plan's `enabled` value differs, and a follow-up read — entirely inside the `KibanaWriteFunc[tfModel]`, not in `PostRead`. The resource's Terraform schema attributes, CRUD semantics, config validators, state upgrade behaviour, and externally observable behaviour SHALL remain functionally equivalent after migration.

#### Scenario: SLO resource satisfies envelope interface after migration

- **WHEN** `NewResource()` is called for the SLO resource
- **THEN** the returned value SHALL satisfy `resource.Resource` and `resource.ResourceWithConfigure`
- **AND** the Create, Read, Update, and Delete methods SHALL be owned by the envelope
- **AND** `ResourceWithConfigValidators` and `ResourceWithUpgradeState` SHALL continue to be satisfied by the concrete resource type

#### Scenario: Enabled reconciliation runs inside the write callback

- **WHEN** the SLO write callback observes a plan `enabled` value different from the server's `enabled` value after the post-create or post-update intermediate read
- **THEN** the write callback SHALL invoke the Kibana Enable or Disable SLO API as appropriate
- **AND** the write callback SHALL re-read the SLO before returning its `KibanaWriteResult`
- **AND** `PostRead` SHALL NOT be involved in this reconciliation
