## MODIFIED Requirements

### Requirement: Update only changes mutable API key fields (REQ-039-REQ-041)

During update, the resource SHALL call the regular or cross-cluster update API according to `type` from the plan, SHALL identify the target API key by `key_id`, and SHALL omit immutable fields such as `id`, `name`, and `expiration` from the update request payload. Read-after-write SHALL be performed via the envelope's `readFunc` (`readAPIKey`); the envelope's `PostRead` hook (`postReadPersistClusterVersion`) MAY run after update, consistent with how it runs after read.

The resource SHALL implement Update as a `WriteFunc[T]` callback via the entitycore envelope. The resource SHALL NOT override the envelope's `Update` method receiver. The write callback SHALL branch on `req.Plan.Type` to select the appropriate update API.

#### Scenario: Update request payload

- **WHEN** Terraform updates a managed API key in place
- **THEN** the provider SHALL send only mutable fields and SHALL refresh state afterward via the envelope's read-after-write path using `readAPIKey`

#### Scenario: Update branches on key type

- **WHEN** `type = "cross_cluster"` in the plan
- **THEN** the write callback SHALL call the Update cross-cluster API key API

#### Scenario: Update uses regular update API for non-cross-cluster keys

- **WHEN** `type` is `"rest"` or unset in the plan
- **THEN** the write callback SHALL call the Update API key API
