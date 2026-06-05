# entitycore-resource-envelope Specification

## Purpose

The entitycore resource envelope centralizes common Terraform Plugin Framework behavior for Elasticsearch-backed resources that share the standard connection block, lifecycle preludes, model contract, composite ID handling, and opt-in import convention.
## Requirements
### Requirement: Envelope constructor produces shared Elasticsearch resource behavior

The system SHALL provide a generic constructor `NewElasticsearchResource[T]()` that returns an envelope owning shared Elasticsearch resource behavior. The Elasticsearch envelope constructor SHALL take the Terraform type-name suffix and an options struct, not a component enum and a positional callback list. The envelope SHALL provide Metadata, Schema, Create, Read, Update, Delete, and Configure behavior, and SHALL satisfy `resource.Resource`. Concrete resources SHALL embed the envelope and may choose to implement additional Plugin Framework interfaces such as ImportState or state upgrade support.

#### Scenario: Constructor returns complete resource envelope

- **WHEN** `NewElasticsearchResource[T]("security_user", opts)` is called with non-nil required callbacks in `opts`
- **THEN** the returned value SHALL provide the shared Metadata, Schema, Create, Read, Update, Delete, and Configure methods
- **AND** the returned value SHALL satisfy `resource.Resource`
- **AND** concrete resources embedding the returned value SHALL NOT need thin Create or Update wrappers when their behavior fits the callback contract

### Requirement: Envelope owns Configure and Metadata

The system SHALL implement `Configure` and `Metadata` on the Elasticsearch envelope using the same provider-data conversion and type-name composition rules as `ResourceBase`, with the Elasticsearch namespace segment implied by the envelope rather than passed by the caller. The configured `*clients.ProviderClientFactory` SHALL be reachable from the envelope's Read, Create, Update, and Delete preludes.

#### Scenario: Metadata builds the Terraform type name

- **WHEN** an envelope is constructed via `NewElasticsearchResource[T]("security_user", opts)`
- **THEN** its `Metadata` SHALL set the type name to `<provider_type_name>_elasticsearch_security_user`

### Requirement: Envelope injects elasticsearch_connection block into schema

The system SHALL inject the `elasticsearch_connection` block into the schema returned by the concrete schema factory before exposing it via the `Schema` method.

#### Scenario: Schema includes injected connection block

- **WHEN** an envelope is constructed with a schema factory that returns a schema lacking an `elasticsearch_connection` block
- **THEN** calling `Schema` on the envelope SHALL return a schema that includes the `elasticsearch_connection` block produced by the canonical provider helper
- **AND** the concrete schema attributes and other blocks SHALL remain unchanged

### Requirement: Envelope owns the Read prelude

The `ElasticsearchResource` envelope SHALL resolve the read identity using a lenient three-step fallback when the decoded state model does not implement `WithReadResourceID` (or `GetReadResourceID` returns empty) and no write fallback is supplied: (1) attempt to parse `model.GetID().ValueString()` as a composite ID with `clients.CompositeIDFromStr`; if successful (non-nil result), use `compID.ResourceID`; (2) otherwise fall back to `model.GetResourceID().ValueString()` if it is non-empty; (3) otherwise use the raw `model.GetID().ValueString()` string as the resource ID. The envelope SHALL NOT return an error diagnostic solely because the state `id` is not in composite format. If all three fallback steps resolve to an empty string the envelope SHALL return an error diagnostic with summary "Invalid resource identifier" and SHALL NOT invoke the concrete read function.

This replaces the "Read falls back to composite ID resource segment" scenario in the canonical spec, which previously stated that a non-composite ID always fails.

#### Scenario: Read succeeds with plain (non-composite) ID — GetResourceID available

- GIVEN a state `id` attribute that is not in `<cluster_uuid>/<resource_id>` format (e.g. `my-job-id`)
- AND `model.GetResourceID()` returns a non-empty string (e.g. `my-job-id`)
- WHEN `Read` runs
- THEN the envelope SHALL NOT return an error diagnostic for the ID format
- AND the concrete `readFunc` SHALL be invoked with `my-job-id` as the `resourceID` argument

#### Scenario: Read succeeds with plain (non-composite) ID — GetResourceID empty

- GIVEN a state `id` attribute that is not in `<cluster_uuid>/<resource_id>` format (e.g. `my-job-id`)
- AND `model.GetResourceID()` returns an empty string
- WHEN `Read` runs
- THEN the envelope SHALL NOT return an error diagnostic for the ID format
- AND the concrete `readFunc` SHALL be invoked with `my-job-id` (the raw ID string) as the `resourceID` argument

#### Scenario: Read still fails when all fallbacks produce empty string

- GIVEN a state `id` attribute that is empty
- AND `model.GetResourceID()` also returns an empty string
- WHEN `Read` runs
- THEN the envelope SHALL return an error diagnostic with summary "Invalid resource identifier"
- AND the concrete `readFunc` SHALL NOT be invoked

#### Scenario: Read with composite ID continues to work

- GIVEN a state `id` attribute in `<cluster_uuid>/<resource_id>` format (e.g. `abc123/my-job-id`)
- WHEN `Read` runs
- THEN the envelope SHALL parse the composite ID and invoke `readFunc` with `my-job-id` as the `resourceID` argument
- AND behavior SHALL be unchanged from the prior implementation

### Requirement: Envelope owns the Create and Update preludes

The system SHALL implement `Create` and `Update` on `NewElasticsearchResource[T]` by deserializing the relevant framework inputs, deriving the write resource ID from the model, resolving the scoped Elasticsearch client from the model's connection block via `GetElasticsearchClient`, enforcing any optional version requirements declared by the planned model, invoking the corresponding concrete callback with a structured request object, and then invoking `readFunc` with the model returned by the callback. State SHALL be set from the model returned by `readFunc`, not directly from the concrete callback.

Create and update callbacks SHALL share the type `WriteFunc[T]` and receive a `WriteRequest[T]` containing `Plan`, `Prior`, `Config`, and `WriteID`. `Prior` SHALL be a `*T`: `nil` for create invocations and a non-nil pointer to the decoded prior state model for update invocations. Callbacks that distinguish create from update SHALL inspect `req.Prior == nil`. `Config` SHALL be the Terraform configuration decoded into `T` by the envelope before the callback is invoked, in the same manner as `Plan` and `Prior`.

Create and update callbacks SHALL return `WriteResult[T]` carrying the written model used for read-after-write identity resolution.

#### Scenario: Create callback receives nil Prior and decoded config

- **WHEN** `Create` runs for a resource whose callback fits the envelope contract
- **THEN** the callback SHALL receive `WriteRequest[T]` with `Prior == nil`
- **AND** the callback SHALL receive the planned model in `Plan` and the Terraform configuration decoded into `T` in `Config`

#### Scenario: Update callback receives prior state and decoded config

- **WHEN** `Update` runs for a resource whose callback fits the envelope contract
- **THEN** the callback SHALL receive `WriteRequest[T]` with `Prior` pointing at the decoded prior-state model
- **AND** the callback SHALL receive both the planned model in `Plan` and the Terraform configuration decoded into `T` in `Config`

#### Scenario: Write-only attributes accessible via decoded Config

- **WHEN** a schema attribute is declared `WriteOnly: true`
- **THEN** the decoded `Config T` value in `WriteRequest[T]` SHALL carry the practitioner-supplied value for that attribute
- **AND** the decoded `Plan T` value SHALL NOT carry the practitioner-supplied value for that attribute, consistent with framework write-only semantics

### Requirement: Envelope supports post-read side effects
The system SHALL allow Elasticsearch envelope users to provide an optional post-read hook. The system SHALL define `ElasticsearchPostReadRequest[T]` as a struct with fields `Client *clients.ElasticsearchScopedClient`, `Prior T`, `State T`, and `Private any`. The `PostReadFunc[T]` type SHALL be `func(ctx context.Context, req ElasticsearchPostReadRequest[T]) (T, diag.Diagnostics)`.

When `PostRead` is configured, the envelope SHALL invoke it after the read callback completes successfully, pass the read-callback result in `req.State`, populate `req.Prior` with the plan model (Create/Update) or the state model decoded before the read (plain Read), and commit the model returned by PostRead to state. If PostRead returns error diagnostics, the envelope SHALL append those diagnostics and SHALL NOT call `resp.State.Set`. The hook SHALL NOT run when the entity is not found or when `readFunc` returns error diagnostics.

#### Scenario: Post-read hook on write path receives plan as Prior and read-callback result as State
- **WHEN** `ElasticsearchResourceOptions.PostRead` is set and the envelope completes a Create or Update operation
- **THEN** PostRead SHALL be invoked with `req.Prior` equal to the plan model and `req.State` equal to the model returned by the read callback
- **AND** the model returned by PostRead SHALL be passed to `resp.State.Set`
- **AND** `resp.State.Set` SHALL NOT be called with the read-callback model directly

#### Scenario: Post-read hook on plain Read path receives prior state as Prior and refreshed model as State
- **WHEN** `ElasticsearchResourceOptions.PostRead` is set and the envelope executes a plain Read operation
- **THEN** PostRead SHALL be invoked with `req.Prior` equal to the state model decoded from Terraform state before the read callback ran and `req.State` equal to the model returned by the read callback
- **AND** the model returned by PostRead SHALL be passed to `resp.State.Set`

#### Scenario: Post-read error diagnostics prevent state set
- **WHEN** PostRead returns diagnostics containing at least one error
- **THEN** the envelope SHALL append those diagnostics and SHALL NOT call `resp.State.Set`

#### Scenario: Post-read hook not set — state set directly from read callback result
- **WHEN** `ElasticsearchResourceOptions.PostRead` is nil
- **THEN** the model returned by the read callback SHALL be passed directly to `resp.State.Set`
- **AND** no PostRead invocation SHALL occur

### Requirement: Shared version requirement type is envelope-neutral

The system SHALL define the shared requirement type as `VersionRequirement`, not `DataSourceVersionRequirement`, and the optional `WithVersionRequirements` interface SHALL return `[]VersionRequirement`.

#### Scenario: Resource and data source envelopes use the same requirement type

- **WHEN** a model implements `WithVersionRequirements`
- **THEN** both Kibana and Elasticsearch envelopes SHALL accept the same `VersionRequirement` return type

### Requirement: Envelope owns the Delete prelude

The `ElasticsearchResource` envelope SHALL implement `Delete` by deserializing the prior state into the generic model `T`, resolving the resource ID using the same lenient three-step fallback as the updated Read prelude (composite parse → `GetResourceID()` → raw ID), resolving the scoped Elasticsearch client from the model's connection block via `GetElasticsearchClient`, and invoking the concrete delete function with `(context, *clients.ElasticsearchScopedClient, resourceID string, T)`. The envelope SHALL NOT return an error diagnostic solely because the state `id` is not in composite format. If all three fallback steps produce an empty string the envelope SHALL return an error diagnostic and SHALL NOT invoke the concrete delete function.

This replaces the "Composite ID parse failure short-circuits delete" scenario in the canonical spec, which previously stated that a non-composite ID always fails.

#### Scenario: Delete succeeds with plain (non-composite) ID — GetResourceID available

- GIVEN a state `id` attribute that is not in `<cluster_uuid>/<resource_id>` format (e.g. `my-job-id`)
- AND `model.GetResourceID()` returns a non-empty string (e.g. `my-job-id`)
- WHEN `Delete` runs
- THEN the envelope SHALL NOT return an error diagnostic for the ID format
- AND the concrete `deleteFunc` SHALL be invoked with `my-job-id` as the `resourceID` argument

#### Scenario: Delete succeeds with plain (non-composite) ID — GetResourceID empty

- GIVEN a state `id` attribute that is not in `<cluster_uuid>/<resource_id>` format (e.g. `my-job-id`)
- AND `model.GetResourceID()` returns an empty string
- WHEN `Delete` runs
- THEN the envelope SHALL NOT return an error diagnostic for the ID format
- AND the concrete `deleteFunc` SHALL be invoked with `my-job-id` (the raw ID string) as the `resourceID` argument

#### Scenario: Delete with composite ID continues to work

- GIVEN a state `id` attribute in `<cluster_uuid>/<resource_id>` format (e.g. `abc123/my-job-id`)
- WHEN `Delete` runs
- THEN the envelope SHALL parse the composite ID and invoke `deleteFunc` with `my-job-id` as the `resourceID` argument
- AND behavior SHALL be unchanged from the prior implementation

#### Scenario: Client resolution failure still short-circuits delete

- GIVEN `GetElasticsearchClient` returns error diagnostics
- WHEN `Delete` runs
- THEN the diagnostics SHALL be appended to `resp.Diagnostics`
- AND the concrete delete function SHALL NOT be invoked

### Requirement: Envelope requires a non-nil delete callback

The system SHALL require concrete resources to supply a non-nil delete callback. Resources whose API has no delete operation SHALL pass a callback that returns `nil` diagnostics. The envelope SHALL NOT special-case a nil callback to mean "no API call".

#### Scenario: Constructor accepts a no-op delete callback

- **WHEN** the delete callback is supplied as a function that returns `nil` diagnostics without performing an API call
- **THEN** the envelope SHALL invoke that callback during `Delete` and proceed normally, removing the resource from state

### Requirement: Envelope keeps ImportState opt-in

The system SHALL NOT implement `ImportState` on the envelope. Import support SHALL remain opt-in for concrete resources, matching the provider-wide convention. Concrete resources that support import SHALL implement `ImportState` themselves, typically as a passthrough on the `id` attribute.

#### Scenario: Envelope resource without ImportState does not satisfy ResourceWithImportState

- **WHEN** an envelope is constructed via `NewElasticsearchResource[T]`
- **THEN** the returned resource SHALL satisfy `resource.Resource`
- **AND** SHALL satisfy `resource.ResourceWithConfigure`
- **AND** SHALL NOT satisfy `resource.ResourceWithImportState`

#### Scenario: Concrete resource adds ImportState passthrough

- **WHEN** a concrete resource that embeds `*ElasticsearchResource[T]` defines its own `ImportState` method using `resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)`
- **THEN** that resource SHALL satisfy `resource.ResourceWithImportState`
- **AND** the `id` attribute SHALL be set to the supplied import identifier

### Requirement: Model type constraint exposes ID and connection block

The system SHALL define a type constraint `ElasticsearchResourceModel` requiring `GetID() types.String`, `GetResourceID() types.String`, and `GetElasticsearchConnection() types.List`. Concrete `Data` types SHALL satisfy this constraint by declaring value-receiver getter methods over their existing `ID`, natural write identity, and `ElasticsearchConnection` fields.

#### Scenario: Concrete model satisfies the constraint

- **WHEN** a concrete `Data` struct declares `GetID() types.String` returning `d.ID`, `GetResourceID() types.String` returning its plan-safe write identity field, and `GetElasticsearchConnection() types.List` returning `d.ElasticsearchConnection`
- **THEN** that struct SHALL satisfy `ElasticsearchResourceModel`
- **AND** SHALL be usable as the type parameter to `NewElasticsearchResource[T]`

### Requirement: Envelope preserves resource type names exactly

The system SHALL allow concrete migrations to preserve their existing Terraform type name. Each migrated security resource SHALL initialize the envelope with the literal name suffix that produces its current type name.

#### Scenario: Migrated user resource preserves its type name

- **WHEN** the user resource is migrated to the envelope using `NewElasticsearchResource[Data]("security_user", opts)`
- **THEN** its Terraform type name SHALL remain `<provider_type_name>_elasticsearch_security_user`

#### Scenario: Migrated system_user resource preserves its type name

- **WHEN** the system user resource is migrated to the envelope using name suffix `security_system_user`
- **THEN** its Terraform type name SHALL remain `<provider_type_name>_elasticsearch_security_system_user`

#### Scenario: Migrated role resource preserves its type name

- **WHEN** the role resource is migrated to the envelope using name suffix `security_role`
- **THEN** its Terraform type name SHALL remain `<provider_type_name>_elasticsearch_security_role`

#### Scenario: Migrated role_mapping resource preserves its type name

- **WHEN** the role mapping resource is migrated to the envelope using name suffix `security_role_mapping`
- **THEN** its Terraform type name SHALL remain `<provider_type_name>_elasticsearch_security_role_mapping`

### Requirement: Envelope coexists with ResourceBase-only entities

The system SHALL NOT change the behavior of resources that embed `*ResourceBase` directly. Resources that do not migrate to the envelope SHALL continue to operate with their existing Configure, Metadata, and Client wiring.

#### Scenario: ResourceBase-only resource continues to work

- **WHEN** a resource that embeds `*ResourceBase` (without the envelope) is configured and used
- **THEN** its Configure, Metadata, and Client behavior SHALL remain unchanged

### Requirement: Acceptance test for plain-ID import compatibility (REQ-ES-ENV-001)

The acceptance test suite for `elasticstack_elasticsearch_ml_anomaly_detection_job` SHALL include a test named `TestAccResourceAnomalyDetectionJobFrom0_12_2` that simulates a job originally stored with a plain `job_id` as the state `id` (as produced by provider ≤ 0.12.2) and verifies that the current provider can successfully refresh and apply using that state without returning a "Wrong resource ID" diagnostic.

#### Scenario: Plain-ID state from old provider can be refreshed by current provider

- GIVEN an existing anomaly detection job whose Terraform state stores `id = "<job_id>"` (plain, non-composite)
- AND the current provider implements the lenient-ID fallback
- WHEN `terraform plan` or `terraform apply` is run with the current provider
- THEN the provider SHALL successfully read the job from Elasticsearch using the plain `job_id` as the resource identifier
- AND the plan SHALL NOT include an error diagnostic about wrong resource ID format

### Requirement: Envelope auto-injects `timeouts` attribute into resource schema
The system SHALL inject a `timeouts` attribute produced by `terraform-plugin-framework-timeouts/resource/timeouts.AttributesAll(ctx)` into every Elasticsearch envelope resource schema. The system SHALL copy the factory-supplied schema attributes first and then assign the envelope's `timeouts` attribute, silently overwriting any factory-supplied attribute with the same key. Concrete resource schema factories MUST NOT include a `timeouts` attribute in their output.

#### Scenario: Schema gains `timeouts` attribute with create/read/update/delete sub-attributes
- **WHEN** a Plugin Framework client requests the schema of any Elasticsearch envelope resource
- **THEN** the response schema SHALL contain an attribute named `timeouts` of the `timeouts.AttributesAll` shape
- **AND** the `timeouts` attribute SHALL expose optional `create`, `read`, `update`, and `delete` string sub-attributes

#### Scenario: Factory-supplied `timeouts` attribute is silently replaced
- **WHEN** a concrete resource schema factory returns a schema whose `Attributes` map already contains a `timeouts` key
- **THEN** the envelope SHALL replace the factory's value with `timeouts.AttributesAll(ctx)`
- **AND** the envelope SHALL NOT panic or surface a diagnostic for the collision

### Requirement: Envelope wraps operation context with the configured timeout
The system SHALL wrap the context for every Create, Read, Update, and Delete operation with `context.WithTimeout` after the model is decoded from `req.Plan`/`req.State`/`req.Config` and **before** scoped client resolution, version-requirement enforcement, the resource callback, read-after-write, and post-read run. All API-touching steps SHALL execute under the resolved deadline. The system SHALL defer cancellation so the wrapped context is released when the operation returns. The system SHALL append diagnostics returned by `timeouts.Value.Create/Read/Update/Delete` and SHALL NOT continue past the wrap if those diagnostics contain an error.

The system SHALL define `ResourceTimeouts` as a struct with fields `Create, Read, Update, Delete time.Duration`. The system SHALL accept a `Timeouts ResourceTimeouts` field on `ElasticsearchResourceOptions[T]`. For each operation the system SHALL use the corresponding `ResourceTimeouts` field as the default if non-zero, falling back to the package-level constants `DefaultResourceCreateTimeout = 20 * time.Minute`, `DefaultResourceReadTimeout = 5 * time.Minute`, `DefaultResourceUpdateTimeout = 20 * time.Minute`, and `DefaultResourceDeleteTimeout = 20 * time.Minute`.

#### Scenario: Create wraps context using plan `timeouts.create` or per-op default
- **WHEN** the envelope dispatches a Create operation
- **THEN** the envelope SHALL call `planModel.GetTimeouts().Create(ctx, defaultCreate)` where `defaultCreate` is `opts.Timeouts.Create` if non-zero, otherwise `DefaultResourceCreateTimeout`
- **AND** the envelope SHALL wrap the operation context with `context.WithTimeout` using the resolved duration
- **AND** the envelope SHALL defer `cancel()` so the context is released when Create returns

#### Scenario: Read, Update, and Delete apply per-operation timeouts
- **WHEN** the envelope dispatches a Read, Update, or Delete operation
- **THEN** the envelope SHALL apply the same ctx-wrap pattern as Create using the corresponding `timeouts.Value` method and the matching default constant
- **AND** Update SHALL derive the timeout from the plan model
- **AND** Read and Delete SHALL derive the timeout from the state model

#### Scenario: Version-requirement enforcement runs under the operation timeout
- **WHEN** the envelope dispatches any of Create, Read, Update, or Delete
- **THEN** `EnforceVersionRequirements` and any other API-issuing envelope plumbing SHALL execute with the wrapped (timeout-bound) context
- **AND** a slow `EnforceVersionRequirements` call SHALL be subject to the same deadline as the resource callback

#### Scenario: Null or unknown stored `timeouts` falls back to the default
- **WHEN** a Read, Update, or Delete operation runs against a state whose `timeouts` value is null, unknown, or has no entry for the current operation (for example after upgrading from a provider version that did not expose `timeouts`)
- **THEN** `model.GetTimeouts().<Op>(ctx, default)` SHALL return the envelope-supplied default with no diagnostics
- **AND** the operation SHALL proceed under that default-bound context

#### Scenario: Configured per-op default overrides the envelope default
- **WHEN** `ElasticsearchResourceOptions.Timeouts.Create` is set to a non-zero duration
- **AND** the practitioner did not specify `timeouts.create` in configuration
- **THEN** the envelope SHALL use the configured per-op default in place of `DefaultResourceCreateTimeout`

#### Scenario: Practitioner-supplied timeout overrides per-op default
- **WHEN** the practitioner specifies `timeouts = { create = "30m" }` in configuration
- **THEN** the envelope SHALL wrap Create with `30m` regardless of `Options.Timeouts.Create` or `DefaultResourceCreateTimeout`

#### Scenario: Timeouts diagnostic prevents callback invocation
- **WHEN** `model.GetTimeouts().Create(ctx, default)` returns diagnostics containing at least one error
- **THEN** the envelope SHALL append those diagnostics
- **AND** the envelope SHALL NOT invoke the Create callback

#### Scenario: Envelope persists `timeouts` to state independent of callback-returned model
- **WHEN** a Read operation completes and the resource `readFunc` returns a model whose `timeouts` field is a zero value (for example because the callback reconstructed the model without copying `ResourceTimeoutsField`)
- **THEN** the envelope SHALL write the prior-state `timeouts` value back into `resp.State` after `resp.State.Set`
- **AND** the operation SHALL succeed without a `timeouts` value-conversion diagnostic
- **WHEN** a Create or Update operation completes and read-after-write returns a model whose `timeouts` field is a zero value
- **THEN** the envelope SHALL write the plan model's `timeouts` value into `outState` after `outState.Set`
- **AND** the operation SHALL succeed without a `timeouts` value-conversion diagnostic

### Requirement: Envelope may skip read-after-write when write result is authoritative
The system SHALL accept `SkipReadAfterWrite` on `ElasticsearchResourceOptions`. When `SkipReadAfterWrite` is true, after a successful Create or Update write callback the envelope SHALL persist `WriteResult.Model` to state directly without invoking the read callback. When `SkipReadAfterWrite` is false or unset, the envelope SHALL retain the existing read-after-write behavior. The read callback SHALL still be used for Read operations regardless of this option.

#### Scenario: SkipReadAfterWrite persists write callback model without re-read
- **WHEN** `ElasticsearchResourceOptions.SkipReadAfterWrite` is true
- **AND** the envelope completes a Create or Update write callback successfully
- **THEN** the envelope SHALL set state from `WriteResult.Model` without calling the read callback
- **AND** the read callback SHALL still be available for Read/refresh operations

#### Scenario: Default Create/Update still read-after-write
- **WHEN** `SkipReadAfterWrite` is false or unset
- **THEN** the envelope SHALL invoke the read callback after Create/Update to refresh state before persisting

### Requirement: Resource models embed `WithResourceTimeouts`
The system SHALL define `WithResourceTimeouts` as an interface with a single method `GetTimeouts() timeouts.Value`. The system SHALL define `ResourceTimeoutsField` as an embeddable struct with field `Timeouts timeouts.Value `tfsdk:"timeouts"`` and a value-receiver `GetTimeouts()` method that returns the field. The system SHALL require `ElasticsearchResourceModel` to embed `WithResourceTimeouts`. Concrete resource models satisfy the constraint by embedding `ResourceTimeoutsField` (or by declaring an equivalent field plus method).

#### Scenario: Type constraint enforces `WithResourceTimeouts`
- **WHEN** a developer attempts to instantiate `ElasticsearchResource[T]` with a model type that does not satisfy `WithResourceTimeouts`
- **THEN** the Go compiler SHALL reject the instantiation with a type-constraint violation

