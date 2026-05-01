## ADDED Requirements

### Requirement: Typed client implementation for SLM policy CRUD
The resource SHALL use the go-elasticsearch Typed API for all SLM operations. `GetSlm` SHALL use `Slm.GetLifecycle().Do(ctx)`, `PutSlm` SHALL use `Slm.PutLifecycle().Do(ctx)`, and `DeleteSlm` SHALL use `Slm.DeleteLifecycle().Do(ctx)`. The typed API response type `types.SLMPolicy` SHALL replace the custom `models.SnapshotPolicy` type.

#### Scenario: Typed API read maps SLM policy
- GIVEN a successful Get Lifecycle API response
- WHEN the provider extracts the policy
- THEN the typed API `types.SLMPolicy` SHALL be used directly
- AND fields (`name`, `repository`, `schedule`, `config`, `retention`) SHALL map to the resource state without intermediate model conversion

#### Scenario: Typed API write sends SLM policy
- GIVEN an SLM policy to create or update
- WHEN the provider calls the Put Lifecycle API
- THEN the request SHALL be built using typed API request builders accepting `*types.SLMPolicy`
- AND manual JSON marshaling into `models.SnapshotPolicy` SHALL NOT occur
