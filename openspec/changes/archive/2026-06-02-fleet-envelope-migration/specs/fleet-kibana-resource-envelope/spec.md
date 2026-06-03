## ADDED Requirements

### Requirement: Fleet resources with optional space assignment implement KibanaUnscopedSpace

Fleet resources that have an optional `space_ids types.Set` field (where null/empty means "default Fleet space") SHALL implement `entitycore.KibanaUnscopedSpace` by returning `true` from `IsUnscopedSpace()`. This allows the `KibanaResource[T]` envelope to accept an empty space identifier without error, since an empty string is a valid routing value meaning "default space" in Fleet API endpoints.

#### Scenario: Resource with null space_ids routes to default space

- **WHEN** a Fleet resource model has `SpaceIDs` as null or empty
- **THEN** `GetSpaceID()` SHALL return `types.StringValue("")`
- **AND** `IsUnscopedSpace()` SHALL return `true`
- **AND** the `KibanaResource[T]` envelope SHALL NOT return an error for the empty space ID

#### Scenario: Resource with populated space_ids routes to first space

- **WHEN** a Fleet resource model has `SpaceIDs` containing one or more space ID strings
- **THEN** `GetSpaceID()` SHALL return the first element of the set as a `types.String`
- **AND** the `KibanaResource[T]` envelope SHALL use that space ID for API routing

### Requirement: Fleet resource Update callbacks use prior-state space for API routing

When a Fleet resource's `space_ids` is updated (values change between plan and state), the Update write callback SHALL use `req.Prior.GetSpaceID()` as the operational space for the API call — the space where the resource currently exists — rather than the plan's space ID. The plan's `space_ids` change is communicated to the Fleet API via the request body, not the URL routing.

#### Scenario: space_ids changes during update

- **WHEN** a Fleet resource update changes `space_ids` from `["space-a"]` to `["space-b"]`
- **THEN** the Update callback SHALL call the Fleet API using `space-a` as the URL routing space (from `req.Prior.GetSpaceID()`)
- **AND** the request body SHALL reflect the new `space_ids` value `["space-b"]`
- **AND** after update the resource SHALL be readable from `space-b`

### Requirement: Fleet resources with a distinct resource ID field expose it via GetResourceID

Fleet resource models that hold an API-assigned ID in a named field (distinct from the Terraform `id` attribute) SHALL return that field from `GetResourceID()`. Resources whose only stable write identity is derived from user-provided fields SHALL derive `GetResourceID()` from those fields.

#### Scenario: server_host exposes HostID as resource identity

- **WHEN** `GetResourceID()` is called on a `serverHostModel`
- **THEN** it SHALL return the value of the `HostID` field

#### Scenario: output exposes OutputID as resource identity

- **WHEN** `GetResourceID()` is called on an `outputModel`
- **THEN** it SHALL return the value of the `OutputID` field

#### Scenario: custom_integration derives identity from package name and version

- **WHEN** `GetResourceID()` is called on a `customIntegrationModel` with known `PackageName` and `PackageVersion`
- **THEN** it SHALL return `types.StringValue("<PackageName>/<PackageVersion>")`

### Requirement: Fleet resources with server-version constraints implement WithVersionRequirements

Fleet resource models whose behaviour depends on a minimum Elastic Stack version SHALL implement `entitycore.WithVersionRequirements` by returning a list of `VersionRequirement`s from `GetVersionRequirements()`. The list MAY be conditional on the model's own field values. The `KibanaResource[T]` envelope evaluates these requirements after client resolution and before invoking lifecycle callbacks; no inline `EnforceMinVersion` calls SHALL remain in CRUD code paths or model-conversion helpers for these resources.

#### Scenario: custom_integration always requires 8.2.0

- **WHEN** `GetVersionRequirements()` is called on a `customIntegrationModel`
- **THEN** it SHALL return a single `VersionRequirement` with `MinVersion = 8.2.0`
- **AND** the envelope SHALL fail the lifecycle call with the requirement's error message when the server version is below `8.2.0`

#### Scenario: output emits kafka requirement only when type is kafka

- **WHEN** `GetVersionRequirements()` is called on an `outputModel` with `Type = "kafka"`
- **THEN** the returned list SHALL include a `VersionRequirement` with `MinVersion = 8.13.0`

#### Scenario: output emits ssl verification mode requirement only when the field is set

- **WHEN** `GetVersionRequirements()` is called on an `outputModel` whose `Ssl` object has `verification_mode` set
- **THEN** the returned list SHALL include a `VersionRequirement` with `MinVersion = 8.10.0`

### Requirement: Fleet resources fully implement envelope CRUD callbacks

Fleet resources that embed `entitycore.KibanaResource[T]` SHALL supply non-placeholder callbacks for `Create`, `Read`, `Update`, and `Delete` via `KibanaResourceOptions`. The wrapper struct SHALL NOT override the envelope's `Create` or `Update` methods. `PlaceholderKibanaWriteCallback` SHALL NOT be used.

#### Scenario: Update callback uses prior-state space for operational routing

- **WHEN** an Update is invoked for a Fleet resource whose `space_ids` differs between plan and prior state
- **THEN** the Update write callback SHALL call the Fleet API using `req.Prior.GetSpaceID()` as the URL routing space
- **AND** the request body SHALL reflect the planned `space_ids` value
