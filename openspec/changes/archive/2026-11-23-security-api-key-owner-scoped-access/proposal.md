## Why

`elasticstack_elasticsearch_security_api_key` cannot be managed by a user that holds only the `manage_own_api_key` cluster privilege (issue #4351). Elasticsearch authorizes id-scoped Get API key and Invalidate API key requests under `manage_own_api_key` only when the request sets `owner=true`, sets `username` and `realm_name` to the caller's identity, or comes from the API key itself. The provider sent `ids` alone, so refresh and destroy failed with HTTP 403 unless the user also held `manage_api_key` or `manage_security`, even for keys the user owns.

Always sending `owner=true` is not an option. It narrows the request to keys owned by the current user. A Terraform workspace shared by several privileged users can hold keys created by another user, and those keys would be silently dropped from state on refresh or left valid after destroy.

## What Changes

- Read looks the key up with `owner=true` first. When that does not return the key, the provider retries without the owner filter. A missing key is reported as not found only when neither lookup returns it.
- Delete invalidates with `owner=true` first. When the response does not list the key as invalidated or previously invalidated, the provider retries with `owner=false`. An error is reported only when that attempt also fails or does not list the key.
- The Invalidate API key response is now checked: the key id must appear in `invalidated_api_keys` or `previously_invalidated_api_keys`. Before, any 2xx response counted as success.
- New optional attribute `restrict_to_owned` (bool, default `false`). When `true`, read and delete only use `owner=true` and never fall back. A key owned by a different user is treated as non-existent and is never invalidated.
- Schema version 2 → 3, with a state upgrader that backfills `restrict_to_owned = false`. The v0 and v1 upgraders backfill it too, so existing resources plan with no diff and trigger no Update API key call.
- The ephemeral `elasticstack_elasticsearch_security_api_key` invalidates on close with `owner=true` and no fallback. The connection that opens the key always owns it.

Non-goals:

- No change to create or update. Both already work under `manage_own_api_key` for keys the caller owns.
- No ownership enforcement by default. Shared workspaces keep today's behavior.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `elasticsearch-security-api-key`: REQ-034–REQ-038 (read), REQ-043 (delete) and the state upgrade registration describe the owner-scoped lookup with fallback. A new requirement covers `restrict_to_owned` and its state backfill.
- `elasticsearch-security-api-key-ephemeral`: REQ-EPH-003 states that close invalidates with `owner=true`.

## Impact

- `internal/clients/elasticsearch/security_api_key.go`: `GetAPIKey` and `DeleteAPIKey` take a `restrictToOwned` flag and implement the fallback.
- `internal/elasticsearch/security/apikey/`: schema attribute, model field, state upgraders, read/delete callbacks, ephemeral close.
- Unit tests for the client fallback and the state upgraders. A new acceptance test manages keys with a user holding only `manage_own_api_key` and `monitor`.
- Users: no migration. Users holding `manage_api_key` see no behavior change. Users holding only `manage_own_api_key` can now manage their own keys.
