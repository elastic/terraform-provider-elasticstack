## ADDED Requirements

### Requirement: Envelope auto-injects `timeouts` attribute into resource schema
The system SHALL inject a `timeouts` attribute produced by `terraform-plugin-framework-timeouts/resource/timeouts.AttributesAll(ctx)` into every Kibana envelope resource schema. The system SHALL copy the factory-supplied schema attributes first and then assign the envelope's `timeouts` attribute, silently overwriting any factory-supplied attribute with the same key. Concrete resource schema factories MUST NOT include a `timeouts` attribute in their output.

#### Scenario: Schema gains `timeouts` attribute with create/read/update/delete sub-attributes
- **WHEN** a Plugin Framework client requests the schema of any Kibana envelope resource
- **THEN** the response schema SHALL contain an attribute named `timeouts` of the `timeouts.AttributesAll` shape
- **AND** the `timeouts` attribute SHALL expose optional `create`, `read`, `update`, and `delete` string sub-attributes

#### Scenario: Factory-supplied `timeouts` attribute is silently replaced
- **WHEN** a concrete resource schema factory returns a schema whose `Attributes` map already contains a `timeouts` key
- **THEN** the envelope SHALL replace the factory's value with `timeouts.AttributesAll(ctx)`
- **AND** the envelope SHALL NOT panic or surface a diagnostic for the collision

### Requirement: Envelope wraps operation context with the configured timeout
The system SHALL wrap the context for every Create, Read, Update, and Delete operation with `context.WithTimeout` after the model is decoded from `req.Plan`/`req.State`/`req.Config` and **before** scoped client resolution, version-requirement enforcement, space-ID validation, the resource callback, read-after-write, and post-read run. All API-touching steps SHALL execute under the resolved deadline. The system SHALL defer cancellation so the wrapped context is released when the operation returns. The system SHALL append diagnostics returned by `timeouts.Value.Create/Read/Update/Delete` and SHALL NOT continue past the wrap if those diagnostics contain an error.

The system SHALL accept a `Timeouts ResourceTimeouts` field on `KibanaResourceOptions[T]`. For each operation the system SHALL use the corresponding `ResourceTimeouts` field as the default if non-zero, falling back to the package-level constants `DefaultResourceCreateTimeout`, `DefaultResourceReadTimeout`, `DefaultResourceUpdateTimeout`, and `DefaultResourceDeleteTimeout` defined alongside the Elasticsearch envelope.

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

#### Scenario: Version-requirement enforcement and space-ID validation run under the operation timeout
- **WHEN** the envelope dispatches any of Create, Read, Update, or Delete
- **THEN** `EnforceVersionRequirements`, space-ID validation, and any other API-issuing envelope plumbing SHALL execute with the wrapped (timeout-bound) context
- **AND** a slow `EnforceVersionRequirements` call SHALL be subject to the same deadline as the resource callback

#### Scenario: Null or unknown stored `timeouts` falls back to the default
- **WHEN** a Read, Update, or Delete operation runs against a state whose `timeouts` value is null, unknown, or has no entry for the current operation (for example after upgrading from a provider version that did not expose `timeouts`)
- **THEN** `model.GetTimeouts().<Op>(ctx, default)` SHALL return the envelope-supplied default with no diagnostics
- **AND** the operation SHALL proceed under that default-bound context

#### Scenario: Configured per-op default overrides framework default
- **WHEN** `KibanaResourceOptions.Timeouts.Create` is set to a non-zero duration
- **AND** the practitioner did not specify `timeouts.create` in configuration
- **THEN** the envelope SHALL use the configured per-op default in place of `DefaultResourceCreateTimeout`

#### Scenario: Practitioner-supplied timeout overrides per-op default
- **WHEN** the practitioner specifies `timeouts = { create = "30m" }` in configuration
- **THEN** the envelope SHALL wrap Create with `30m` regardless of `Options.Timeouts.Create` or the framework default

#### Scenario: Timeouts diagnostic prevents callback invocation
- **WHEN** `model.GetTimeouts().Create(ctx, default)` returns diagnostics containing at least one error
- **THEN** the envelope SHALL append those diagnostics
- **AND** the envelope SHALL NOT invoke the Create callback

### Requirement: Resource models embed `WithResourceTimeouts`
The system SHALL require `KibanaResourceModel` to embed `WithResourceTimeouts`. Concrete Kibana resource models satisfy the constraint by embedding `ResourceTimeoutsField`. The `WithResourceTimeouts` interface, `ResourceTimeoutsField` struct, and per-op default constants are shared with the Elasticsearch envelope and defined once in the entitycore package.

#### Scenario: Type constraint enforces `WithResourceTimeouts`
- **WHEN** a developer attempts to instantiate `KibanaResource[T]` with a model type that does not satisfy `WithResourceTimeouts`
- **THEN** the Go compiler SHALL reject the instantiation with a type-constraint violation
