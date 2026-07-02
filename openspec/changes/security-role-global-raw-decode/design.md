## Context

Elasticsearch 9.5.0-SNAPSHOT changed the security Get Role API response to include `"data_source": []` as a top-level key in the `global` privilege object. The go-elasticsearch typed client's `Role.Global` field is declared as `map[string]map[string]map[string][]string`, which cannot decode an array value. The unmarshal error surfaces on every `GetRole` read for roles that have `global` configured, breaking `TestAccResourceSecurityRole` and production usage.

This is tracked upstream at [elasticsearch-specification#6377](https://github.com/elastic/elasticsearch-specification/issues/6377). The issue is still open; a released fix is not expected in the short term.

Relevant code:
- Read path: `internal/clients/elasticsearch/security_role.go:GetRole` — calls `typedClient.Security.GetRole().Name(rolename).Do(ctx)`, which triggers `Role.UnmarshalJSON` with the mistyped `Global` field.
- Write path: `internal/elasticsearch/security/role/models.go:122` — decodes user-supplied `global` JSON into `map[string]map[string]map[string][]string`.
- Schema field: `internal/elasticsearch/security/role/models.go:46` — `Global` is `customtypes.JSONWithDefaultsValue[map[string]any]` (permissive; no change needed).
- Precedent for raw transport bypass: `internal/clients/elasticsearch/index.go` — uses `typedClient.Transport.Perform` for date-math index names.

## Goals / Non-Goals

**Goals:**
- Fix the `GetRole` read failure on ES 9.5+ without bumping go-elasticsearch.
- Remain forward-compatible when additional heterogeneous `global` categories appear in future ES versions.
- Loosen the write-path type assertion to `map[string]any` so user-supplied `global` JSON with non-standard shapes does not fail on marshal either.
- Preserve the typed API for PutRole and DeleteRole (unchanged, unaffected by the upstream spec bug).

**Non-Goals:**
- Fixing the upstream elasticsearch-specification#6377 bug (tracked separately).
- Upgrading go-elasticsearch (would not fix this until upstream is resolved).
- Changing the provider schema for `global` (already permissive as `map[string]any`).
- Supporting the `data_source` category as a first-class typed attribute (it remains part of the opaque `global` JSON blob).

## Decisions

**Raw transport on read, typed API for write/delete**: The read path bypasses the typed `GetRole` response only for the `global` field. All other fields (applications, cluster, indices, remote_indices, etc.) can still come from the typed response or be decoded from the same raw JSON body. This matches the established precedent in `index.go`.

**Decode strategy for GetRole**: Make a raw `GET /_security/role/<name>` request via `typedClient.Transport.Perform`. Decode the response body into `map[string]json.RawMessage` to extract the per-role entry, then decode each field independently — typed structs for all well-formed fields, `json.RawMessage` for `global` which is then JSON-marshaled back into a string for the `JSONWithDefaultsValue` field in the model. This avoids copying large amounts of typed decoding logic while bypassing the broken `Global` field.

**Alternative considered: error-fallback (D4)**: Catch the specific "cannot unmarshal array" error string from the typed client and fall back to raw GET. Rejected as brittle — error message changes would silently break the fallback.

**Alternative considered: bump go-elasticsearch (D3)**: Not viable; upstream issue still open.

**Write-path loosening (D2)**: The `toAPIModel` decode from `map[string]map[string]map[string][]string` to `map[string]any` is a forward-compat improvement. The typed `PutRole` builder accepts `map[string]json.RawMessage` (already done at line 46–50 in `security_role.go`), so the Go type used in `models.go` is just an intermediate decoding step before marshal; `map[string]any` is safe here.

**No schema change**: The schema field is already `map[string]any`-backed; no Terraform plan-level changes result from this fix.

**Acceptance test coverage**: `TestAccResourceSecurityRole` already exercises the round-trip. On ES 9.5+, it will pass once the read path is fixed. No new test file is required; the existing test validates the fix.

## Risks / Trade-offs

- [Low risk] The raw transport path adds a small amount of manual HTTP error handling (non-2xx status codes, body read). Mitigation: follow the same error-handling pattern as `index.go`.
- [Low risk] Manually extracting fields from the raw JSON may miss future additions in the typed response. Mitigation: only `global` bypasses the typed decode; all other fields remain typed.
- [Low risk] The write-path change from `map[string]map[string]map[string][]string` to `map[string]any` could accept syntactically valid but semantically invalid `global` JSON. Mitigation: the provider has always accepted the `global` field as an opaque JSON string; the decode is a pass-through to the wire, and Elasticsearch will reject malformed payloads at the API level.

## Open questions

None — root cause is well-understood and the fix approach is confirmed by the issue author.
