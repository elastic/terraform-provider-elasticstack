# entitycore-ephemeral-envelope Specification

## Purpose
TBD - created by archiving change entitycore-ephemeral-envelope. Update Purpose after archive.
## Requirements
### Requirement: Constructors produce shared Elasticsearch and Kibana ephemeral envelopes (REQ-ENVE-001)

The system SHALL provide generic constructors `NewElasticsearchEphemeralResource[T ElasticsearchEphemeralModel, S any](name string, opts ElasticsearchEphemeralOptions[T, S]) ephemeral.EphemeralResource` and `NewKibanaEphemeralResource[T KibanaEphemeralModel, S any](name string, opts KibanaEphemeralOptions[T, S]) ephemeral.EphemeralResource`.

Each constructor SHALL take a name suffix and an options struct (matching the call shape of `NewElasticsearchResource[T]` and `NewKibanaResource[T]`). The returned value SHALL implement `ephemeral.EphemeralResource`, `ephemeral.EphemeralResourceWithConfigure`, and `ephemeral.EphemeralResourceWithClose`. It SHALL NOT implement `ephemeral.EphemeralResourceWithRenew` (out of scope for this change).

#### Scenario: Constructor returns a complete ephemeral envelope

- **GIVEN** non-nil `Schema`, `Open`, and `Close` callbacks in `opts`
- **AND** a close-state type `S` containing only plain Go types
- **WHEN** `NewElasticsearchEphemeralResource[T, S]("security_api_key", opts)` is called
- **THEN** the returned value SHALL implement `ephemeral.EphemeralResource`, `ephemeral.EphemeralResourceWithConfigure`, and `ephemeral.EphemeralResourceWithClose`
- **AND** the returned value SHALL NOT implement `ephemeral.EphemeralResourceWithRenew`

#### Scenario: Constructor rejects nil required callbacks

- **GIVEN** an options struct with `Schema`, `Open`, or `Close` set to nil
- **WHEN** the constructor is invoked
- **THEN** the constructor SHALL `panic` with a configuration-error message naming the missing callback

### Requirement: Constructors enforce plain-Go close-state at construction time (REQ-ENVE-002)

The constructors SHALL validate that the type parameter `S` contains no fields whose type's `PkgPath()` is `github.com/hashicorp/terraform-plugin-framework/types`. The validation SHALL be recursive across embedded structs, pointers, slices (element type), maps (value type), and arrays.

When `S` contains any disallowed field, the constructor SHALL `panic` with a message of the form `entitycore: ephemeral close state <S type name> has field <dotted path> of plugin-framework type <PkgPath>/<Name>; Close state must be plain Go types only`.

When `S` contains only plain Go types, the constructor SHALL return normally.

#### Scenario: Close state with a tfsdk field is rejected at construction

- **GIVEN** `S` defined as `struct { KeyID types.String; InvalidateOnClose bool }`
- **WHEN** `NewElasticsearchEphemeralResource[T, S]("name", opts)` is called
- **THEN** the constructor SHALL `panic`
- **AND** the panic message SHALL contain the field name `KeyID`
- **AND** the panic message SHALL contain the path `github.com/hashicorp/terraform-plugin-framework/types`

#### Scenario: Close state with a tfsdk field inside an embedded struct is rejected

- **GIVEN** `S` defined as `struct { Inner struct { Field types.Bool } }`
- **WHEN** the constructor is called
- **THEN** the constructor SHALL `panic`
- **AND** the panic message SHALL include the dotted path `Inner.Field`

#### Scenario: Close state with only plain Go types is accepted

- **GIVEN** `S` defined as `struct { KeyID string; InvalidateOnClose bool; Endpoints []string; Headers map[string]string; Insecure *bool }`
- **WHEN** the constructor is called
- **THEN** the constructor SHALL return without panicking

### Requirement: Envelopes own Configure and Metadata (REQ-ENVE-003)

The system SHALL implement `Configure` on each ephemeral envelope using `clients.ConvertProviderDataToFactory`, appending diagnostics and assigning the converted factory only when no error diagnostics are produced.

The system SHALL implement `Metadata` on the Elasticsearch envelope to set the type name to `<provider_type_name>_elasticsearch_<name>` and on the Kibana envelope to `<provider_type_name>_kibana_<name>`.

#### Scenario: Elasticsearch envelope metadata composition

- **WHEN** the Elasticsearch envelope is constructed with name `"security_api_key"`
- **THEN** `Metadata` SHALL set the type name to `<provider_type_name>_elasticsearch_security_api_key`

#### Scenario: Kibana envelope metadata composition

- **WHEN** the Kibana envelope is constructed with name `"service_account_token"`
- **THEN** `Metadata` SHALL set the type name to `<provider_type_name>_kibana_service_account_token`

#### Scenario: Configure preserves prior factory on error

- **GIVEN** a prior successful Configure has stored a factory
- **AND** a subsequent Configure call receives provider data that fails conversion
- **WHEN** `Configure` runs
- **THEN** the response SHALL contain error diagnostics
- **AND** the previously stored factory SHALL NOT be overwritten

### Requirement: Envelopes inject connection blocks into the schema (REQ-ENVE-004)

The Elasticsearch envelope SHALL inject an optional `elasticsearch_connection` block into the schema returned by the user schema factory. The Kibana envelope SHALL inject an optional `kibana_connection` block. In both cases the user-supplied attributes and any other user-supplied blocks SHALL be preserved unchanged.

The connection block SHALL use the same block shape as the equivalent resource envelopes (reusing the existing `providerschema` helpers, with new ephemeral-namespaced helpers added if not yet present).

#### Scenario: Elasticsearch schema gains the connection block

- **GIVEN** a user schema factory returning a schema with no `elasticsearch_connection` block
- **WHEN** the envelope's `Schema` method is invoked
- **THEN** the returned schema SHALL contain an `elasticsearch_connection` block
- **AND** all other user-supplied attributes and blocks SHALL be present unchanged

#### Scenario: Kibana schema gains the connection block

- **GIVEN** a user schema factory returning a schema with no `kibana_connection` block
- **WHEN** the envelope's `Schema` method is invoked
- **THEN** the returned schema SHALL contain a `kibana_connection` block

### Requirement: Open prelude resolves the client, enforces version requirements, then invokes the user callback (REQ-ENVE-005)

When `Open` is invoked, the envelope SHALL execute the following prelude steps in order, short-circuiting on any error diagnostics:

1. Decode `req.Config` into a value of `T`.
2. Resolve the scoped client (Elasticsearch or Kibana, per envelope flavor) using the connection accessor on `T` (`GetElasticsearchConnection` or `GetKibanaConnection`).
3. Call `EnforceVersionRequirements(ctx, client, &model)` and short-circuit on any error diagnostics.
4. Invoke the user `Open` callback with `OpenRequest[T]{Config: model}`.
5. On user-callback success, snapshot the connection block from the model into the envelope-owned private-state connection slot.
6. JSON-marshal the user-returned close state `S` into the envelope-owned private-state user-state slot.
7. Call `resp.Result.Set(ctx, &result.Model)`.

The user `Open` callback SHALL NOT receive `req.Private` or be required to touch private state directly.

#### Scenario: Open invokes the user callback with the decoded model and a resolved client

- **GIVEN** a valid configured envelope and a request whose config decodes cleanly
- **AND** the user's `Open` callback that returns a populated `OpenResult[T, S]` with no diagnostics
- **WHEN** `Open` runs
- **THEN** the user callback SHALL receive a scoped client resolved from the model's connection block
- **AND** `resp.Result` SHALL contain the user-returned model
- **AND** the envelope-owned private slots SHALL contain the snapshotted connection and the JSON-marshaled close state

#### Scenario: Open short-circuits on decode error

- **GIVEN** a request whose config decode produces error diagnostics
- **WHEN** `Open` runs
- **THEN** the user `Open` callback SHALL NOT be invoked
- **AND** the decode error diagnostics SHALL propagate to `resp.Diagnostics`

#### Scenario: Open short-circuits on client resolution error

- **GIVEN** a connection block that the factory cannot resolve to a client
- **WHEN** `Open` runs
- **THEN** the user `Open` callback SHALL NOT be invoked
- **AND** the client-resolution error diagnostics SHALL propagate

#### Scenario: Open short-circuits on version-requirement error

- **GIVEN** a model whose `VersionRequirement` is not satisfied by the connected cluster
- **WHEN** `Open` runs
- **THEN** the user `Open` callback SHALL NOT be invoked
- **AND** an `Unsupported server version` (or equivalent) diagnostic SHALL propagate

#### Scenario: Open propagates user-callback diagnostics

- **GIVEN** a user `Open` callback that returns error diagnostics
- **WHEN** `Open` runs
- **THEN** the user diagnostics SHALL propagate to `resp.Diagnostics`
- **AND** the envelope SHALL NOT call `resp.Result.Set`
- **AND** the envelope SHALL NOT write to the private-state slots

### Requirement: Close prelude restores the connection-scoped client, then invokes the user callback (REQ-ENVE-006)

When `Close` is invoked, the envelope SHALL execute the following prelude steps in order, short-circuiting on any error diagnostics:

1. Load the envelope-owned connection slot from `req.Private`.
2. Decode the connection snapshot back into a `types.List` value.
3. Resolve the scoped client using the restored connection.
4. Load the envelope-owned user-state slot from `req.Private`.
5. JSON-unmarshal the user-state bytes into a value of `S`.
6. Invoke the user `Close` callback with `CloseRequest[S]{State: state}`.

When either private-state slot is missing, the envelope SHALL return cleanly without invoking the user callback (representing a Close after an Open that did not complete successfully).

The user `Close` callback SHALL NOT receive `req.Private` or be required to touch private state directly.

#### Scenario: Close restores the connection and calls the user callback

- **GIVEN** a request whose `Private` contains an envelope-written connection slot and user-state slot
- **AND** a user `Close` callback that returns no diagnostics
- **WHEN** `Close` runs
- **THEN** the user callback SHALL receive a scoped client resolved from the snapshotted connection
- **AND** the user callback SHALL receive the unmarshaled `S` value

#### Scenario: Close with missing private state is a no-op

- **GIVEN** a request whose `Private` does not contain the envelope's connection slot or user-state slot
- **WHEN** `Close` runs
- **THEN** the user callback SHALL NOT be invoked
- **AND** no error diagnostics SHALL be produced

#### Scenario: Close propagates user-callback diagnostics

- **GIVEN** a user `Close` callback that returns error diagnostics
- **WHEN** `Close` runs
- **THEN** the user diagnostics SHALL propagate to `resp.Diagnostics`

### Requirement: Connection round-trip is lossless for the documented connection fields (REQ-ENVE-007)

The envelope's connection-snapshot codec SHALL round-trip the following Elasticsearch connection fields without loss when present and known: `endpoints`, `username`, `password`, `api_key`, `bearer_token`, `es_client_authentication`, `insecure`, `ca_file`, `ca_data`, `ca_fingerprint`, `cert_file`, `cert_data`, `key_file`, `key_data`, `headers`.

The Kibana variant SHALL round-trip the equivalent set of `clientconfig.KibanaConnection` fields.

The snapshot codec SHALL NOT use `encoding/json` to marshal `terraform-plugin-framework/types` values directly. The snapshot struct SHALL use plain Go types (`string`, `[]string`, `map[string]string`, `*bool`).

#### Scenario: Boolean fields round-trip both true and false

- **GIVEN** an `elasticsearch_connection` block with `insecure = false`
- **WHEN** the connection is snapshotted and restored
- **THEN** the restored value SHALL preserve `insecure = false` (not null, not absent)

#### Scenario: List and map fields round-trip without flattening loss

- **GIVEN** an `elasticsearch_connection` block with `endpoints = ["https://a", "https://b"]` and `headers = { "X-Foo" = "bar" }`
- **WHEN** the connection is snapshotted and restored
- **THEN** the restored value SHALL preserve both endpoints in order and the header map verbatim

#### Scenario: CA fingerprint round-trips without loss

- **GIVEN** an `elasticsearch_connection` block with `ca_fingerprint` set to a SHA-256 certificate fingerprint
- **WHEN** the connection is snapshotted during Open and restored during Close
- **THEN** the restored value SHALL preserve the `ca_fingerprint` value verbatim

#### Scenario: Null connection block produces a null restored List

- **GIVEN** no `elasticsearch_connection` block (null on the model)
- **WHEN** the connection is snapshotted and restored
- **THEN** the restored value SHALL be a null `types.List` of the connection element type

### Requirement: Documentation describes ephemeral envelopes and known properties (REQ-ENVE-008)

The package documentation in `internal/entitycore/doc.go` SHALL describe the ephemeral envelopes alongside the existing resource and data source patterns. The documentation SHALL include:

1. A description of `NewElasticsearchEphemeralResource[T, S]` and `NewKibanaEphemeralResource[T, S]`, including the `T` and `S` type parameter meanings.
2. The rule that `S` must contain only plain Go types (no `terraform-plugin-framework/types`), and that the constructor enforces this at construction time.
3. A note that `Open()` is called by Terraform during `terraform plan` as well as `terraform apply`, and that the user is responsible for noting this in their generated resource documentation when relevant.
4. A note that `Close()` is not guaranteed to run if Terraform is interrupted between Open and Close, and that resource authors must consider this when designing close-time behavior.
5. A minimal example referencing the api_key migration as a reference implementation.

#### Scenario: Documentation content check

- **GIVEN** the contents of `internal/entitycore/doc.go`
- **WHEN** a future ephemeral resource author reads the package documentation
- **THEN** items 1 through 5 SHALL be present and legible

