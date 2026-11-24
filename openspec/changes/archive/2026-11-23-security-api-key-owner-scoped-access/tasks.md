## 1. Owner-scoped read and delete

- [x] 1.1 `GetAPIKey` looks the key up with `owner=true` first and, unless `restrict_to_owned` is set, retries without the owner filter when the key is not returned; a failed fallback after an owner-scoped miss reports the missing privilege and the `restrict_to_owned` option
- [x] 1.2 `DeleteAPIKey` invalidates with `owner=true` first and, unless `restrict_to_owned` is set, retries with `owner=false` when the key is not reported as invalidated; the response is checked for the key id in `invalidated_api_keys` or `previously_invalidated_api_keys`
- [x] 1.3 Unit tests cover the owner-scoped success, the fallback on 404 and on empty responses, the fallback error, and the no-fallback behavior of `restrict_to_owned = true`

## 2. restrict_to_owned attribute

- [x] 2.1 Add `restrict_to_owned` (Optional + Computed, default `false`) to the resource schema and model, and pass it to the read and delete callbacks
- [x] 2.2 Bump the schema version to 3 and backfill `restrict_to_owned = false` in the 0→1, 1→2 and new 2→3 state upgraders, with unit tests
- [x] 2.3 The ephemeral resource invalidates on close with `owner=true` and no fallback
- [x] 2.4 Regenerate the resource documentation

## 3. Acceptance coverage

- [x] 3.1 `TestAccResourceSecurityAPIKeyManageOwnAPIKeyOnly` creates, refreshes, re-plans and destroys keys through a connection whose user holds only `manage_own_api_key` and `monitor`, with `restrict_to_owned` unset and set to `true`
- [x] 3.2 The acceptance destroy check looks keys up regardless of owner, so keys created by other users are checked for invalidation
