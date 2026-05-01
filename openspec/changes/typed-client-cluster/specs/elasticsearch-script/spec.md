## ADDED Requirements

### Requirement: Typed client implementation for stored script CRUD
The resource SHALL use the go-elasticsearch Typed API for stored script operations. `GetScript` SHALL use `Core.GetScript().Do(ctx)`, `PutScript` SHALL use `Core.PutScript().Do(ctx)`, and `DeleteScript` SHALL use `Core.DeleteScript().Do(ctx)`. The typed API response type `types.StoredScript` SHALL replace the custom `models.Script` type for API fields `lang` and `source`.

#### Scenario: Typed API read maps stored script
- GIVEN a successful Get Stored Script API response
- WHEN the provider processes the response
- THEN `types.StoredScript` SHALL provide `Lang` and `Source` directly
- AND the provider SHALL preserve `params` from prior state when the typed API response does not include params

#### Scenario: Typed API write sends stored script
- GIVEN a stored script to create or update
- WHEN the provider calls the Put Stored Script API
- THEN the request SHALL be built using typed API request builders
- AND `types.StoredScript` SHALL be used for the script body fields (`Lang`, `Source`)
- AND manual JSON marshaling into `models.Script` SHALL NOT occur

#### Scenario: Context parameter preserved
- GIVEN a script resource with `context` configured
- WHEN create or update runs via the typed API
- THEN the `context` value SHALL be passed as a query parameter to the Put Script API
- AND `context` SHALL continue to be preserved from state on read because it is not returned by the Get Script API
