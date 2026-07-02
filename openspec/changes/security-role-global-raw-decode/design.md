## Context

Elasticsearch 9.5.0-SNAPSHOT changed the security Get Role API response to include `"data_source": []` as a top-level key in the `global` privilege object. The go-elasticsearch typed client's `Role.Global` field is declared as `map[string]map[string]map[string][]string`, which cannot decode an array value. The unmarshal error surfaces on every `GetRole` read for roles that have `global` configured, breaking `TestAccResourceSecurityRole` and production usage.

This is tracked upstream at [elasticsearch-specification#6377](https://github.com/elastic/elasticsearch-specification/issues/6377). The issue is still open; a released fix is not expected in the short term.

Relevant code:

- Read path: `internal/clients/elasticsearch/security_role.go:GetRole` — calls `typedClient.Security.GetRole().Name(rolename).Do(ctx)`, which triggers `Role.UnmarshalJSON` with the mistyped `Global` field.
- Read consumers: `internal/elasticsearch/security/role/read.go` (resource) and `internal/elasticsearch/security/role/data_source.go` — both call `GetRole` then `fromAPIModel`, which reads `role.Global` and marshals it into the schema's `JSONWithDefaultsValue` field.
- Default normalizer: `internal/elasticsearch/security/role/global_defaults.go:populateGlobalPrivilegesDefaults` — strips server-injected empty `global` defaults; currently strips only `role: {}`, not the new `data_source: []`.
- Write path: `internal/clients/elasticsearch/security_role.go:PutRole` — marshals `role.Global` to `map[string]json.RawMessage` and passes it to `req.Global(...)`. Unaffected by the upstream bug for the shapes users configure today.
- Schema field: `internal/elasticsearch/security/role/models.go:46` — `Global` is `customtypes.JSONWithDefaultsValue[map[string]any]` (permissive; no change needed).
- Precedent for raw transport bypass: `internal/clients/elasticsearch/index.go` — uses `typedClient.Transport.Perform` for date-math index names.

## Goals / Non-Goals

**Goals:**

- Fix the `GetRole` read failure on ES 9.5+ without bumping go-elasticsearch.
- Remain forward-compatible when additional heterogeneous `global` categories appear in future ES versions.
- Strip the new `data_source: []` server default from state so `TestAccResourceSecurityRole`'s exact-match assertion holds and users don't see perpetual diffs.
- Preserve the typed API for PutRole and DeleteRole (unchanged, unaffected by the upstream spec bug).

**Non-Goals:**

- Fixing the upstream elasticsearch-specification#6377 bug (tracked separately).
- Upgrading go-elasticsearch (would not fix this until upstream is resolved).
- Changing the provider schema for `global` (already permissive as `map[string]any`).
- Supporting the `data_source` category as a first-class typed attribute (it remains part of the opaque `global` JSON blob).
- Loosening the write-path `global` decode (`toAPIModel`) — the write path is unchanged in this change. Write-side forward-compat for user-configured array-typed `global` categories (e.g. a user writing `data_source: []` in HCL) is a possible follow-up but is not required to fix the failing test and is out of scope here.

## Decisions

**Raw transport on read; carry `global` as raw JSON out-of-band (not via `types.Role.Global`).** The read path bypasses the typed `GetRole` response entirely for the `global` field. A raw `GET /_security/role/<name>` request is made via `typedClient.Transport.Perform`; the response body is decoded as `map[string]json.RawMessage` to locate the per-role entry, then the non-`global` fields are decoded into the typed `types.Role` struct, while `global` is kept as `json.RawMessage` and returned **alongside** the typed role (not assigned to `types.Role.Global`, which is typed `map[string]map[string]map[string][]string` and cannot represent array-typed categories like `data_source: []`). Both `fromAPIModel` variants (resource and data source) consume the raw `global` JSON directly into `customtypes.JSONWithDefaultsValue`. This matches the established precedent in `index.go` and avoids reintroducing the decode failure.

**Why not reassemble into `*types.Role.Global`**: `types.Role.Global`'s type cannot hold `data_source: []`. Any attempt to marshal the raw blob back into a `map[string]map[string]map[string][]string`-compatible alias reintroduces the exact `cannot unmarshal array` error. Carrying `global` out-of-band is the only sound approach.

**Default-stripping extension.** The existing `populateGlobalPrivilegesDefaults` strips the empty `role: {}` server default so state matches user intent. ES 9.5 adds `data_source: []` as another server-injected empty default; without stripping it, state would contain `{"application":{},"profile":{...},"data_source":[]}` while the user configured `{"application":{},"profile":{...}}`, causing `TestAccResourceSecurityRole`'s exact-match assertion to fail and producing perpetual diffs. The normalizer is extended to strip `data_source` when it is an empty array, generalized to "strip server-injected empty `global` defaults."

**Alternative considered: error-fallback (D4)**: Catch the specific "cannot unmarshal array" error string from the typed client and fall back to raw GET. Rejected as brittle — error message changes would silently break the fallback.

**Alternative considered: bump go-elasticsearch (D3)**: Not viable; upstream issue still open.

**Alternative considered: loosen write-path decode to `map[string]any` (D2)**: Considered but **out of scope** for this change. As written it does not compile (`role.Global` is `map[string]map[string]map[string][]string`; assigning `map[string]any` is a type error), and the failing test does not require write-side changes. If write-side forward-compat for user-configured array categories is desired, it should be a separate change decoding `global` straight to `map[string]json.RawMessage` and passing it directly to `req.Global()`.

**No schema change**: The schema field is already `map[string]any`-backed; no Terraform plan-level changes result from this fix.

**Acceptance test coverage**: `TestAccResourceSecurityRole` already exercises the round-trip. On ES 9.5+, it will pass once the read path is fixed and `data_source: []` is stripped from state. A unit test for the extended default-stripping is added.

## Risks / Trade-offs

- [Low risk] The raw transport path adds a small amount of manual HTTP error handling (non-2xx status codes, body read). Mitigation: follow the same error-handling pattern as `index.go`.
- [Low risk] `GetRole`'s return contract changes (carries raw `global` alongside the typed role), so both `fromAPIModel` variants must be updated in lockstep. Mitigation: both call sites are in the same package and are updated together.
- [Low risk] Manually extracting fields from the raw JSON may miss future additions in the typed response. Mitigation: only `global` bypasses the typed decode; all other fields remain typed.
- [Low risk] Generalizing the default-stripper could over-strip a user who explicitly configures `data_source: []`. Mitigation: the strip only fires on the *empty* array (`[]`), matching the existing `role: {}` empty-object behavior; an explicitly-empty `data_source` is semantically identical to omitted.

## Open questions

None — root cause is well-understood and the fix approach is confirmed by the issue author.
