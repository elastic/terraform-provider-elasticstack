## MODIFIED Requirements

### Requirement: Envelope owns Create and Update preludes

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
