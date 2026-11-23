## Context

`ManageOwnApiKeyClusterPrivilege` in Elasticsearch authorizes `GetApiKeyRequest` and `InvalidateApiKeyRequest` only when `owner=true` is set, when `username` and `realm_name` match the caller, or when the caller is the API key named in the request. The provider's read (`GetAPIKey`) and delete (`DeleteAPIKey`) sent only the key id. With `manage_own_api_key` alone, every refresh after create and every destroy failed with HTTP 403.

When an `owner=true` request matches no key, Elasticsearch does not fail it. Get API key returns an empty `api_keys` list (HTTP 404 at the REST layer when an id was given). Invalidate API key returns HTTP 200 with empty `invalidated_api_keys` and `previously_invalidated_api_keys`. This holds from 8.0 to current releases.

## Goals / Non-Goals

**Goals:** manage own keys with `manage_own_api_key`; keep today's behavior for shared workspaces; give users who want strict ownership a way to opt in.

**Non-Goals:** enforcing ownership by default; changing create or update.

## Decisions

**Owner-first with fallback, for both read and delete.** Each operation first sends `owner=true`. That succeeds under `manage_own_api_key` and covers the common case in one request. When the key is not returned or not invalidated, the provider retries without the owner filter. That retry needs `manage_api_key` (or `read_security` for reads) and reaches keys owned by other users, which is what happens today.

- A 404 or empty result on the owner-scoped attempt only means no key owned by the caller matched. It falls through to the fallback and is not treated as "gone" or "deleted".
- When the fallback lookup fails after an owner-scoped miss, the error names the missing privilege and points to `restrict_to_owned`. A user holding only `manage_own_api_key` whose key was removed outside Terraform sees that error rather than a bare 403.

Alternative considered: always set `owner=true` (the first version of the PR). Rejected: on shared workspaces, keys owned by another user would drop out of state while staying valid.

Alternative considered: unscoped first, `owner=true` on 403. Rejected: the common `manage_own_api_key` case would always make a failing request first.

**Validate the invalidate response.** The key id must appear in `invalidated_api_keys` or `previously_invalidated_api_keys`. Otherwise an owner-scoped invalidate that matched nothing would look successful.

**`restrict_to_owned` (Optional + Computed, default `false`).** It turns off the fallback. Read then treats a key owned by another user as not found and drops it from state. Delete never invalidates such a key. With `restrict_to_owned = true`, an owner-scoped invalidate that does not list the key fails with an error that names the attribute. A 404 is treated as already deleted. The attribute is stored in state only and is not sent to Elasticsearch.

**State upgrade to schema version 3.** States written before this change have no `restrict_to_owned`. Without a backfill, the first plan would move it from null to `false`, which is an in-place update that calls the Update API key API. That call fails on clusters older than 8.4. Upgraders 0→1 and 1→2 backfill `false`, and a new 2→3 upgrader does only that.

**Ephemeral resource.** The key is opened and closed through the same connection, so the caller always owns it. Close invalidates with `owner=true` and no fallback.

## Risks / Trade-offs

- [Up to two requests on read and delete for keys owned by other users] → Only shared-workspace keys take the second request. Own keys take one.
- [Fallback errors for `manage_own_api_key`-only users when a key disappears outside Terraform] → The error explains the cause and the `restrict_to_owned` opt-in. Invalidated keys stay readable until the retention period removes them, so this only happens after that.
