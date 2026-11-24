## MODIFIED Requirements

### Requirement: Close invalidates the key when `invalidate_on_close = true` (REQ-EPH-003)

`Close()` SHALL be called by Terraform after the apply completes. The provider SHALL:
- When `invalidate_on_close = false` (default): perform no action. The API key persists in Elasticsearch until it expires or is manually invalidated.
- When `invalidate_on_close = true`: call the Elasticsearch Invalidate API key API (`POST /_security/api_key/invalidate`) using the `key_id` from the `Open()` result, with `owner=true` and without retrying with `owner=false`. The key is always owned by the connection that opened it, so the `manage_own_api_key` cluster privilege is sufficient. After invalidation, the key SHALL be immediately unusable.

#### Scenario: Close with invalidate_on_close = false

- GIVEN `invalidate_on_close = false` (or unset) in the ephemeral resource configuration
- WHEN Terraform calls Close after the apply completes
- THEN the provider SHALL NOT call the Invalidate API key API
- AND the API key SHALL remain valid in Elasticsearch

#### Scenario: Close with invalidate_on_close = true

- GIVEN `invalidate_on_close = true` in the ephemeral resource configuration
- WHEN Terraform calls Close after the apply completes
- THEN the provider SHALL call `POST /_security/api_key/invalidate` with the `key_id` from the result and `owner=true`
- AND the API key SHALL be marked as invalidated in Elasticsearch

#### Scenario: Run interrupted before Close is called

- GIVEN `invalidate_on_close = true`
- WHEN Terraform is killed before `Close()` is invoked
- THEN the API key SHALL remain alive until it expires naturally (or indefinitely if no expiration was set)
- AND the provider SHALL document this as a known limitation
