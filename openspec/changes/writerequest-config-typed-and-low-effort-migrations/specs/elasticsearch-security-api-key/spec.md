## MODIFIED Requirements

### Requirement: Update only changes mutable API key fields (REQ-039-REQ-041)

During update, the resource SHALL call the regular or cross-cluster update API according to `type` from the plan, SHALL identify the target API key by `key_id`, and SHALL omit immutable fields such as `id`, `name`, and `expiration` from the update request payload. After a successful update API call, the resource SHALL call `readAPIKey` and persist refreshed state via `resp.State.Set`. This update-time follow-up SHALL NOT invoke the envelope `PostRead` hook or persist cluster version to private state.

The resource SHALL implement Update as a `WriteFunc[T]` callback via the entitycore envelope. The resource SHALL NOT override the envelope's `Update` method receiver. The write callback SHALL branch on `req.Plan.Type` to select the appropriate update API.

#### Scenario: Update request payload

- **WHEN** Terraform updates a managed API key in place
- **THEN** the provider SHALL send only mutable fields and SHALL refresh state afterward via `readAPIKey` from the write callback (without `PostRead`)

#### Scenario: Update branches on key type

- **WHEN** `type = "cross_cluster"` in the plan
- **THEN** the write callback SHALL call the Update cross-cluster API key API

#### Scenario: Update uses regular update API for non-cross-cluster keys

- **WHEN** `type` is `"rest"` or unset in the plan
- **THEN** the write callback SHALL call the Update API key API
