## ADDED Requirements

### Requirement: Typed client implementation for SLM policy CRUD
The resource SHALL use the go-elasticsearch Typed API for all SLM operations. `GetSlm` SHALL use `Slm.GetLifecycle().Do(ctx)`, `PutSlm` SHALL use `Slm.PutLifecycle().Do(ctx)`, and `DeleteSlm` SHALL use `Slm.DeleteLifecycle().Do(ctx)`. The typed API response type `types.SLMPolicy` SHALL replace the custom `models.SnapshotPolicy` type.

#### Scenario: Typed API read maps SLM policy
- GIVEN a successful Get Lifecycle API response
- WHEN the provider extracts the policy
- THEN the typed API request SHALL be issued via the typed client
- AND because `types.SLMPolicy.Retention` uses value-typed `int` fields that cannot distinguish "unset" from "zero", the response MAY be parsed from raw JSON to preserve nullability for `retention` sub-fields
- AND fields (`name`, `repository`, `schedule`, `config`, `retention`) SHALL map to the resource state accurately

#### Scenario: Typed API write sends SLM policy
- GIVEN an SLM policy to create or update
- WHEN the provider calls the Put Lifecycle API
- THEN the request SHALL be issued via the typed client
- AND because the typed API does not expose `expand_wildcards` on `types.SLMConfiguration` and `types.Retention` uses plain `int` without `omitempty`, a raw-body wrapper MAY be used so the provider controls exact field presence
- AND manual JSON marshaling into a legacy `models.SnapshotPolicy` type SHALL NOT occur
