## ADDED Requirements

### Requirement: Typed client implementation for cluster settings
The resource SHALL use the go-elasticsearch Typed API for cluster settings operations. `GetSettings` SHALL use `Cluster.GetSettings().Do(ctx)` with flat settings enabled. `PutSettings` SHALL use `Cluster.PutSettings().Do(ctx)`. Manual JSON decoding into `map[string]any` from raw response bodies SHALL be replaced with typed API response handling.

#### Scenario: Typed API read with flat settings
- GIVEN a successful Cluster Get Settings API call
- WHEN the provider processes the response
- THEN the typed API `getsettings.Response` SHALL provide `Persistent`, `Transient`, and `Defaults` as `map[string]json.RawMessage`
- AND the provider SHALL unmarshal each `RawMessage` value to `any` to maintain the existing `map[string]any` contract with callers

#### Scenario: Typed API write sends settings
- GIVEN cluster settings to update
- WHEN the provider calls the Cluster Put Settings API
- THEN the request SHALL be built using typed API request builders
- AND manual `json.Marshal` of a `map[string]any` into a raw request body SHALL NOT occur
