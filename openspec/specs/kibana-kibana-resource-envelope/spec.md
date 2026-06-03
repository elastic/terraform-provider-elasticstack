# kibana-kibana-resource-envelope Specification

## Purpose
TBD - created by archiving change kibana-envelope-migration. Update Purpose after archive.
## Requirements
### Requirement: Kibana resources with composite ID state implement GetResourceID via the named resource ID field

Kibana resources that store a composite `<spaceID>/<resourceID>` string in the Terraform `id` attribute SHALL implement `GetResourceID()` by returning the named resource-ID field (e.g. `rule_id`, `item_id`), not the composite `id` field. The `KibanaResource[T]` envelope's `resolveKibanaResourceIdentity` parses `GetID()` as a composite ID and uses `GetResourceID()` only as fallback when no composite is present, ensuring both post-create state and import state route correctly.

#### Scenario: Read after create uses composite ID from state

- **WHEN** a resource stores `id = "<spaceID>/<ruleID>"` in state after create
- **THEN** the envelope SHALL parse `GetID()` as a composite and route Read using the extracted `spaceID` and `ruleID`
- **AND** the `readFunc` callback SHALL receive `resourceID = <ruleID>` and `spaceID = <spaceID>`

#### Scenario: Read without composite ID falls back to named fields

- **WHEN** a resource has `id = "<ruleID>"` (plain, non-composite) and `space_id = "<spaceID>"` in state
- **THEN** the envelope SHALL fall back to `GetResourceID()` and `GetSpaceID()` for routing
- **AND** the `readFunc` callback SHALL receive `resourceID = <ruleID>` and `spaceID = <spaceID>`

### Requirement: Kibana resources with ValidateConfig implement it on the wrapper struct

Kibana resources that validate cross-attribute constraints (e.g. mutual exclusivity between indicator types, rule parameter consistency) SHALL implement `resource.ResourceWithValidateConfig` on the outer wrapper struct that embeds `*entitycore.KibanaResource[T]`. The envelope does not own validation logic; implementing `ValidateConfig` on the wrapper is sufficient for the Plugin Framework to invoke it.

#### Scenario: Validation fires independently of envelope lifecycle

- **WHEN** a Terraform plan is created for a resource with an invalid configuration
- **THEN** `ValidateConfig` SHALL return diagnostics describing the violation
- **AND** this SHALL occur before any Create or Update callback is invoked

### Requirement: Kibana resources with UpgradeState implement it on the wrapper struct

Kibana resources that must migrate prior state schema versions SHALL implement `resource.ResourceWithUpgradeState` on the outer wrapper struct. The envelope does not interfere with state upgrade logic; implementing `UpgradeState` on the wrapper is sufficient.

#### Scenario: State upgrade runs for prior schema versions

- **WHEN** Terraform reads state written by a prior provider version
- **THEN** the `UpgradeState` handler on the wrapper struct SHALL migrate the state to the current schema
- **AND** subsequent Read, Update, and Delete operations SHALL operate on the migrated state

### Requirement: Kibana resources with server-version constraints implement WithVersionRequirements

Kibana resource models whose behaviour depends on a minimum Elastic Stack version SHALL implement `entitycore.WithVersionRequirements` by returning a list of `VersionRequirement`s from `GetVersionRequirements()`. The list MAY be conditional on the model's own field values. The `KibanaResource[T]` envelope evaluates these requirements after client resolution and before invoking lifecycle callbacks; no inline `EnforceMinVersion` calls SHALL remain in CRUD code paths or model-conversion helpers for these resources, and model-conversion helpers SHALL NOT take a `clients.MinVersionEnforceable` parameter solely for version gating.

#### Scenario: security_exception_item gates expire_time

- **WHEN** `GetVersionRequirements()` is called on an `ExceptionItemModel` whose `ExpireTime` field is known and non-null
- **THEN** the returned list SHALL include a `VersionRequirement` with `MinVersion = 8.7.2`

#### Scenario: security_detection_rule gates response actions and alerts filter

- **WHEN** `GetVersionRequirements()` is called on a `Data` model that configures response actions
- **THEN** the returned list SHALL include a `VersionRequirement` with `MinVersion = 8.16.0`
- **WHEN** `GetVersionRequirements()` is called on a `Data` model that configures alerts_filter on any action
- **THEN** the returned list SHALL include a `VersionRequirement` with `MinVersion = 8.9.0`

#### Scenario: alerting_rule gates frequency, alerts_filter, alert_delay, flapping, flapping.enabled, and the notify_when reverse case

- **WHEN** `GetVersionRequirements()` is called on an `alertingRuleModel`
- **THEN** the returned list SHALL include a `VersionRequirement` with `MinVersion = 8.6.0` for any of: (a) one or more actions has `Frequency` set, (b) `NotifyWhen` is null or empty
- **AND** SHALL include a `VersionRequirement` with `MinVersion = 8.9.0` when any action has `AlertsFilter` set
- **AND** SHALL include a `VersionRequirement` with `MinVersion = 8.13.0` when `AlertDelay` is set
- **AND** SHALL include a `VersionRequirement` with `MinVersion = 8.16.0` when `Flapping` is set
- **AND** SHALL include a `VersionRequirement` with `MinVersion = 9.3.0` when `Flapping.Enabled` is set

### Requirement: alerting_rule replaces the bespoke features struct with envelope version requirements

The `alertingRuleFeatures` struct, `alertingRuleFeaturesFromVersion`, `alertingRuleFeaturesAllSupported`, and `resolveAlertingRuleFeatures` (currently in `internal/kibana/alertingrule/features.go`) SHALL be removed. The `toAPIModel` method SHALL NOT take a `features` parameter; the version-gating conditionals inside `toAPIModel` (e.g. `if !features.SupportsAlertDelay`) SHALL be removed. Version enforcement SHALL be performed exclusively by the envelope via `GetVersionRequirements()`.

#### Scenario: toAPIModel no longer threads features

- **WHEN** any caller invokes `toAPIModel` on an `alertingRuleModel`
- **THEN** the signature SHALL be `func (m alertingRuleModel) toAPIModel(ctx context.Context) (models.AlertingRule, diag.Diagnostics)` with no `features` parameter
- **AND** the function body SHALL contain no references to a feature-support struct

### Requirement: Kibana resources fully implement envelope CRUD callbacks

Kibana resources that embed `entitycore.KibanaResource[T]` SHALL supply non-placeholder callbacks for `Create`, `Read`, `Update`, and `Delete` via `KibanaResourceOptions`. The wrapper struct SHALL NOT override the envelope's `Create` or `Update` methods. `PlaceholderKibanaWriteCallback` SHALL NOT be used by `security_exception_item`, `security_detection_rule`, or `alertingrule`.

#### Scenario: Lifecycle dispatch goes through envelope callbacks

- **WHEN** Terraform invokes Create, Read, Update, or Delete on any of the three migrated resources
- **THEN** the envelope's lifecycle dispatcher SHALL invoke the corresponding callback supplied via `KibanaResourceOptions`
- **AND** no wrapper-struct `Create` or `Update` method SHALL be present to shadow the envelope's promoted method

