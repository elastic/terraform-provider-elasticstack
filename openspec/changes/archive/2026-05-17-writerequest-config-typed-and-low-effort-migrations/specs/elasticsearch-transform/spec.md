## MODIFIED Requirements

### Requirement: Transform state management — enabled (REQ-013–REQ-015)

On create, when `enabled` is `true`, the resource SHALL call the Elasticsearch Start Transform API after a successful Put Transform call. On update, when `enabled` has changed to `true`, the resource SHALL call the Start Transform API. On update, when `enabled` has changed to `false`, the resource SHALL call the Stop Transform API. When `enabled` has not changed during update, the resource SHALL NOT call Start or Stop Transform. On read, the resource SHALL derive the `enabled` value from transform statistics: `enabled` SHALL be `true` when the transform state is `"started"` or `"indexing"`, and `false` otherwise.

The resource SHALL implement Create and Update as `WriteFunc[T]` callbacks via the entitycore envelope. The resource SHALL NOT override the envelope's `Create` or `Update` method receivers. A shared write callback SHALL use `req.Prior == nil` to distinguish Create from Update. On Update, the enabled-state delta SHALL be determined by comparing `req.Plan.Enabled` against `req.Prior.Enabled`.

#### Scenario: Start on create with enabled=true

- **WHEN** `enabled = true` in configuration and create runs successfully
- **THEN** the resource SHALL call Start Transform after the Put Transform API call

#### Scenario: Stop on update with enabled=false

- **WHEN** an enabled transform is updated with `enabled = false`
- **THEN** the resource SHALL call Stop Transform after the Update Transform API call

#### Scenario: No start/stop when enabled unchanged

- **WHEN** `enabled` is unchanged between plan and apply
- **THEN** the resource SHALL NOT call Start or Stop Transform

#### Scenario: Single WriteFunc serves both Create and Update

- **WHEN** the transform resource registers the same `WriteFunc[T]` for both Create and Update
- **THEN** the callback SHALL distinguish Create from Update via `req.Prior == nil`
- **AND** Create behavior SHALL use the Put Transform API; Update behavior SHALL use the Update Transform API

### Requirement: Create and read-after-write (REQ-033)

After a successful Put Transform API call (and optional Start Transform), the resource SHALL call the read function to refresh state via the envelope's read-after-write mechanism, ensuring the stored state reflects the server-side representation of the transform. The resource SHALL NOT perform a manual read-after-write inside the write callback.

#### Scenario: State refreshed after create

- **WHEN** create completes successfully
- **THEN** the envelope SHALL call the read function to populate state from the API response
