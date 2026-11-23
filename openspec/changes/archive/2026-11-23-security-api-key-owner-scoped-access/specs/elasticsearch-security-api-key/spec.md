## ADDED Requirements

### Requirement: Ownership-scoped access is controlled by `restrict_to_owned` (REQ-048)

The resource SHALL expose an optional, computed boolean attribute `restrict_to_owned`, default `false`, stored only in Terraform state. When `false`, read and delete SHALL first send `owner=true` and SHALL retry without the owner restriction when that does not return or invalidate the key. When `true`, read and delete SHALL only send `owner=true` and SHALL NOT retry, so a key owned by a different user is treated as non-existent and never invalidated.

#### Scenario: Key managed with only manage_own_api_key

- GIVEN a resource whose connection user holds `manage_own_api_key` but not `manage_api_key`, and `restrict_to_owned` is unset
- WHEN the key is created, refreshed and destroyed
- THEN every operation SHALL succeed
- AND the key SHALL be invalidated in Elasticsearch after destroy

#### Scenario: Key owned by another user with restrict_to_owned = true

- GIVEN `restrict_to_owned = true` and a key in state that is owned by a different user
- WHEN Terraform refreshes the resource
- THEN the provider SHALL send only an `owner=true` lookup
- AND it SHALL remove the resource from state without invalidating the key

#### Scenario: Existing state without restrict_to_owned

- GIVEN state written at schema version 0, 1 or 2 without a `restrict_to_owned` value
- WHEN the state is upgraded to the current schema version
- THEN `restrict_to_owned` SHALL be `false` in the upgraded state
- AND the next plan for an unchanged configuration SHALL show no changes

## MODIFIED Requirements

### Requirement: Read refreshes state and preserves non-returned sensitive values (REQ-034-REQ-038)

During Terraform `Read` (refresh), the envelope SHALL invoke the package-level read callback (`readAPIKey`) with the resolved API key identifier after standard prelude steps. `readAPIKey` SHALL look the key up with the Get API key API scoped to the current authenticated user (`owner=true`) first. When that lookup returns no key and `restrict_to_owned` is `false`, it SHALL look the key up again without restricting by owner. When the lookups made return no key (HTTP 404 or an empty `api_keys` list), the envelope SHALL remove the resource from Terraform state. When the fallback lookup fails after the owner-scoped lookup returned no key, the error SHALL state that looking up keys owned by other users requires the `manage_api_key` or `read_security` cluster privilege and mention `restrict_to_owned`. When a key is returned, `readAPIKey` SHALL set `key_id`, `name`, `expiration_timestamp`, and `metadata` from the API response. `readAPIKey` SHALL preserve prior-state values for `api_key` and `encoded`, because the Get API does not return them.

During refresh only — after `resp.State.Set` succeeds for a found key on this path — the envelope SHALL invoke `PostRead`, which SHALL save the Elasticsearch cluster version in private state for later planning decisions.

#### Scenario: API key no longer exists on refresh

- GIVEN Terraform `Read` runs for a key that has been removed from Elasticsearch
- WHEN the Get API returns HTTP 404
- THEN the envelope SHALL remove the resource from Terraform state

#### Scenario: Key owned by another user is found by the fallback lookup

- GIVEN `restrict_to_owned` is `false` and the key is owned by a different user
- WHEN Terraform `Read` runs
- THEN the provider SHALL first send an `owner=true` lookup that returns no key
- AND it SHALL then find the key with a lookup that does not restrict by owner and refresh state from it

#### Scenario: Cluster version persisted via PostRead after successful refresh

- GIVEN Terraform `Read` completes, `readAPIKey` succeeds, and the envelope persists refreshed state
- WHEN `PostRead` runs for this resource type
- THEN the current Elasticsearch cluster version SHALL be stored in private state via `saveClusterVersion`

### Requirement: Delete invalidates the API key and removes state (REQ-043)

On delete, the resource SHALL invalidate the API key id parsed from the stored composite `id` by calling the Elasticsearch Invalidate API key API, and SHALL remove the resource from Terraform state after a successful invalidation. The provider SHALL first send the request with `owner=true`. An invalidation SHALL count as successful only when the key id appears in `invalidated_api_keys` or `previously_invalidated_api_keys` of the response. When the `owner=true` request does not invalidate the key (an error, HTTP 404, or a response that does not list the key id) and `restrict_to_owned` is `false`, the provider SHALL retry with `owner=false`. An HTTP 404 on the `owner=false` request SHALL count as success. When `restrict_to_owned` is `true`, the provider SHALL NOT retry: an HTTP 404 SHALL count as success, and a response that does not list the key id SHALL be reported as an error.

#### Scenario: Destroying an API key

- GIVEN Terraform destroys the resource
- WHEN delete runs
- THEN the provider SHALL invalidate the referenced API key and remove the resource from state

#### Scenario: Destroying a key owned by another user

- GIVEN `restrict_to_owned` is `false` and the key is owned by a different user
- WHEN delete runs
- THEN the provider SHALL send an `owner=true` invalidate request that does not list the key id
- AND it SHALL then invalidate the key with an `owner=false` request

#### Scenario: Invalidate response does not list the key

- GIVEN every invalidate request the provider sends returns a response that does not list the key id
- WHEN delete runs
- THEN the provider SHALL report an error and keep the resource in state

### Requirement: State upgrades remain on the concrete type

The concrete `Resource` type SHALL continue to implement `ResourceWithUpgradeState`, providing state upgraders from schema version `0` to `1`, from `1` to `2`, and from `2` to `3`. Each upgrader SHALL set `restrict_to_owned` to `false` when it is absent from prior state, so resources created before the attribute existed plan with no changes; the `2` to `3` upgrader SHALL make no other change.

#### Scenario: State upgrade registration

- **WHEN** the concrete resource type is asserted as `resource.ResourceWithUpgradeState`
- **THEN** the assertion SHALL succeed
- **AND** the upgrade map SHALL contain the three defined upgraders
