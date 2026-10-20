## Context

`elasticstack_fleet_server_host` (`internal/fleet/serverhost`) manages Fleet server hosts via `POST/GET/PUT/DELETE /api/fleet/fleet_server_hosts`. The Kibana OpenAPI schema for the create/update request bodies includes an `ssl` object (`generated/kbapi/kibana.gen.go:86223-86234` for POST, `:86294-86305` for PUT) that the resource never sends or reads. The object has ten optional sub-fields, `additionalProperties: false`, `nullable: true`, and no schema-level `description`, `default`, or required-together constraint on any child:

- `certificate` / `key` — the TLS certificate Fleet Server presents to agents, and its key.
- `certificate_authorities` — CAs used to validate agent client certificates (max 10 items).
- `client_auth` — enum `none | optional | required`.
- `agent_certificate` / `agent_certificate_authorities` / `agent_key` — client cert/key/CA Elastic Agent uses when connecting *to* Fleet Server.
- `es_certificate` / `es_certificate_authorities` / `es_key` — client cert/key/CA Fleet Server uses when connecting *to* Elasticsearch.

Kibana's Fleet Server host update replaces the stored `ssl` object rather than merging its children. The generated API model can represent the deferred fields, but the Terraform schema intentionally cannot. Therefore, a PUT that would replace `ssl` must be blocked when the current host has non-null deferred fields, unless the implementation can preserve them without exposing them in Terraform.

The issue's triage reason ("ssl carries certificate/key/client_auth configuration") and this repo's existing structural precedent (`internal/fleet/agentdownloadsource`'s `ssl` block, which maps exactly `certificate`/`certificate_authorities`/`key` plus its own write-only sibling) both point at the first four fields — `certificate`, `key`, `certificate_authorities`, `client_auth` — as the right first slice. The `agent_*`/`es_*` fields and the parallel `secrets.ssl` managed-secret object (added later, Kibana 9.3.0) are a materially larger surface with their own unverified minimum-version and redaction questions; they are deferred to a follow-up change.

## Goals / Non-Goals

**Goals:**
- Expose `certificate`, `key` (sensitive), `certificate_authorities`, and `client_auth` as an optional `ssl` nested attribute on `elasticstack_fleet_server_host`, mapped to/from the Kibana `ssl` object on create, update, and read.
- Reuse `internal/fleet/sslcore.Values`/`Extract`/`Hydrate` for the certificate/key/CA conversion, following `internal/fleet/output/models_ssl.go`'s pattern of embedding `sslcore.Values` and adding a resource-specific sibling field (there: `verification_mode`; here: `client_auth`).
- Reject plan-time configurations where `ssl` is present but all children are null, following `agentdownloadsource/ssl_validator.go`.
- Gate `ssl` behind its own attribute-scoped minimum version requirement (9.1.0), independent of the resource's overall version floor (the resource currently has none in production code), so existing `ssl`-free configurations are unaffected on any stack version.
- Resolve, via acceptance testing during implementation rather than by assumption, three behaviors the OAS does not document for this specific endpoint: whether `PUT` with an omitted `ssl`, an empty object, or explicit JSON `null` actually clears a previously-set value (each tested independently); whether `ssl.key` is redacted on read the way `agentdownloadsource`'s secret-backed key is; and the precise minimum Kibana version.

**Non-Goals:**
- `agent_certificate`, `agent_certificate_authorities`, `agent_key`, `es_certificate`, `es_certificate_authorities`, `es_key`. Deferred to a follow-up change.
- `secrets.ssl` Fleet-managed-secret storage (storing `key` as a `{id}` reference instead of a literal). Deferred to a follow-up change; its own minimum Fleet Server version is not established in available docs.
- A provider-enforced default for `client_auth`, or a provider-enforced "certificate and key must be set together" constraint. The Kibana API declares neither at the schema level (no default, no required-together rule), and the defaults cited in Elastic's standalone-Agent runtime-TLS documentation describe `server.ssl.*` config, not this API contract — inventing a constraint the API doesn't declare would reject configurations Kibana itself accepts.
- Advanced TLS settings (`supported_protocols`, `cipher_suites`) — not exposed on this API field at all.

## Decisions

### Decision 1: Mirror `agentdownloadsource`'s `ssl` block structure, add `client_auth`

Add `internal/fleet/serverhost/models_ssl.go` (object↔API conversion, following `internal/fleet/output/models_ssl.go`'s `outputSslModel` embedding `sslcore.Values` plus a sibling field) and `internal/fleet/serverhost/ssl_validator.go` (all-nil-subattribute validator, following `agentdownloadsource/ssl_validator.go`, extended to include `client_auth` in the checked sub-attribute list). The schema attribute in `internal/fleet/serverhost/schema.go` is:

```hcl
ssl = {
  certificate             = <optional, string>
  key                     = <optional, string, sensitive>
  certificate_authorities = <optional, list(string)>  # 1-10 non-empty entries
  client_auth             = <optional, string>          # "none" | "optional" | "required"
}
```

**Why:** `agentdownloadsource` is the most recently shipped, most battle-tested version of this exact pattern in the repo (shared `sslcore` helpers, same validator shape). Reusing it minimizes review risk and keeps the Fleet domain internally consistent. `client_auth` is modeled as a sibling field alongside the embedded `sslcore.Values`, exactly as `output`'s `verification_mode` is — both are resource-specific additions to the shared certificate/key/CA core.

Considered: a fully bespoke schema for `fleet_server_host`'s `ssl` ignoring the shared core. Rejected — `sslcore` exists specifically to keep the null-normalizing and pointer-conversion logic correct in one place across resources that share this exact certificate/CA/key shape.

### Decision 2: `client_auth` has no provider-side default; omitted fields sent as absent

Kibana returns no `client_auth` default (probe 1.9), so `client_auth` is Optional-only and hydrated as null when absent. An omitted value is omitted from the request.

**Why:** The OAS declares no default for `client_auth` on this endpoint and the probe confirmed Kibana applies none. Elastic's cited defaulting behavior ("required if CAs configured else none") comes from standalone Elastic Agent's `server.ssl.*` runtime documentation, not from this Kibana API's contract.

### Decision 3: Attribute-scoped version gate at 9.1.0, added via a new `GetVersionRequirements` on `serverHostModel`

`serverHostModel` does not currently implement `GetVersionRequirements` (unlike `agentdownloadsource`'s `model`, which gates its own `ssl` and `auth` attributes this way). This change adds the method, scoped so only an `ssl`-configured plan incurs the gate:

```go
func (m serverHostModel) GetVersionRequirements(_ context.Context) ([]entitycore.VersionRequirement, diag.Diagnostics) {
    var reqs []entitycore.VersionRequirement
    if typeutils.IsKnown(m.Ssl) {
        reqs = append(reqs, entitycore.NewAttributeVersionRequirement(
            path.Root("ssl"),
            *serverHostSslMinVersion,
            fmt.Sprintf("ssl is only supported in Elastic Stack %s and above", serverHostSslMinVersion),
        ))
    }
    return reqs, nil
}
```

**Why:** Matches `agentdownloadsource/models.go`'s existing pattern exactly (`GetVersionRequirements` returning per-attribute requirements only when that attribute is configured). Because `serverHostModel` has no resource-wide version floor today, this is a pure addition with no risk of affecting existing `ssl`-free configurations on any stack version — the new method returns an empty slice when `ssl` is unset.

The 9.1.0 floor comes from the Kibana release notes entry "Adds SSL options to Fleet Server host settings" (kibana#208091). It is unverified: only a 9.5.5 stack was probed, so below-floor behavior is covered by unit tests of `GetVersionRequirements` only.

### Decision 4: Removal semantics, `key` read-back, and deferred fields (resolved by probes)

1. **Removal:** a PUT omitting `ssl` leaves the stored object unchanged, while `ssl: {}` and explicit `ssl: null` both clear it, and a PUT replaces the whole object. Removing `ssl` therefore sends `ssl: {}` in place (no replacement, no explicit-null request path); dropping one child sends the remaining children in place.
2. **`key` read-back:** Kibana echoes `ssl.key` verbatim on create, update, and read, so `populateFromAPI` hydrates it from the response with no preservation logic.
3. **`client_auth`:** no server default, so it is Optional-only (Decision 2).
4. **Deferred fields:** hosts can store non-null `agent_*`/`es_*` fields and a PUT drops them, so an update that sends `ssl` is rejected before mutation when the host stores any (Decision 5).

See "Probe results" for the evidence.

## Risks / Trade-offs

| Risk | Mitigation |
|---|---|
| Removing `ssl` from config doesn't clear it server-side | Resolved: Kibana ignores an omitted `ssl` but clears on `ssl: {}`, which the provider sends on removal; covered by `_SSL` and `_SSLChildRemoval` acceptance tests. |
| `ssl.key` is redacted on read, stomping a configured key | Resolved: Kibana echoes the key verbatim, so it is hydrated from the response; round-trip and import are asserted in acceptance tests. |
| Kibana returns a `client_auth` default that differs from the omitted plan value | Resolved: no default is returned; `client_auth` is Optional-only and `_SSLClientAuthOmitted` asserts no diff. |
| Updating the supported SSL fields erases existing deferred `agent_*`/`es_*` values | Inspect the current stored object before a replacing PUT and return an actionable diagnostic without mutation when deferred values are present; verify with a seeded-host acceptance test. |
| 9.1.0 floor is wrong (not independently confirmed against this repo's pinned OAS changelog) | Explicit version-boundary acceptance tests (below-floor error, at-floor success, `ssl`-omitted-below-floor success) catch a wrong floor during implementation rather than after release. |
| Changing TLS settings on an enrolled Fleet Server host can disconnect connected Elastic Agents in real deployments | Documented in the resource's generated docs; out of scope for the Terraform acceptance test itself, which only asserts API/state, not live agent connectivity. |

### Decision 5: Raw PUT plus follow-up GET, and a pre-update guard read

Kibana returns `ssl` in a PUT response as a JSON-encoded string, which the generated typed response cannot decode. `UpdateFleetServerHost` therefore issues the raw PUT, ignores the body, and reads the host back (a missing host after the PUT is an error); `PutFleetServerHost` is the exported PUT-only form used by delete when it clears the default flag. Because a PUT replaces the whole stored `ssl` object and Kibana ignores an omitted `ssl` (probe 1.2), `ssl` is sent only when the planned value differs from prior state, and the deferred-field guard (a GET before the PUT) runs only in that case, including removals. The guard and the PUT are not atomic, so a concurrent writer adding `agent_*`/`es_*` fields between them could still be overwritten; this race is accepted.

## Open questions

All open questions are resolved by the "Probe results" section: `ssl` removal is sent as `ssl: {}` in place, `client_auth` is Optional-only with no server default, and `ssl.key` is hydrated verbatim. The `agent_*`/`es_*` fields are out of scope (tracked as a follow-up). The only unresolved item is the 9.1.0 minimum version, which is unverified (only 9.5.5 was probed; below-floor behavior is covered by unit tests only).

## Probe results (Task 1, live Kibana 9.5.5)

Run with `curl` against `POST/PUT/GET/DELETE /api/fleet/fleet_server_hosts` on a Kibana 9.5.5 stack. All probe hosts were deleted afterwards.

| Probe | Result |
|---|---|
| 1.1 create with all four `ssl` fields | Accepted; create response and subsequent GET return all four fields. |
| 1.2 PUT omitting `ssl` | `ssl` persists unchanged (omission does NOT clear). |
| 1.3 PUT `"ssl": {}` | Clears: GET returns `"ssl": {}` (all supported fields absent; hydrates to null). |
| 1.4 PUT `"ssl": null` | Clears: GET no longer returns `ssl`. |
| 1.5 PUT replacing `ssl` with one child omitted (each of `certificate`, `key`, `certificate_authorities`, `client_auth`) | Each omitted child is cleared in place; retained children persist. PUT replaces the whole `ssl` object. |
| 1.6 `key` on create/GET | Echoed verbatim on create and GET (no redaction, no `secrets.ssl` on this stack). |
| 1.7 minimum version | 9.5.5 accepts non-null `ssl`. Lower bound cannot be exercised (only 9.5.5 is available) and the pinned OAS carries no changelog; 9.1.0 (kibana#208091) is retained as the unverified floor. |
| 1.9 `client_auth` omitted | No server default: create response and GET omit `client_auth`. |
| extra: deferred fields | `agent_*`/`es_*` fields can be seeded via the API and are returned on GET. A PUT with only supported fields replaces the object and silently drops them, confirming the protection requirement. |
| extra: PUT response shape | The PUT response returns `ssl` as a JSON-encoded string (e.g. `"ssl":"{}"`) instead of an object, so the generated client cannot decode the update response when `ssl` is present. GET and POST return an object. |

### Decisions driven by the probes

- **Decision 3/4 removal**: in-place. Because `{}` and explicit `null` both clear, the provider sends an empty `ssl` object (`{}`, serializable through the generated struct pointer) when `ssl` is removed from configuration and the prior state had `ssl`. Explicit JSON null serialization (Task 2.5) and `RequiresReplace` plan modifiers (Task 3.4) are not needed. Child removal is in place because the PUT replaces the entire object. When neither plan nor prior state has `ssl`, `ssl` is omitted from the PUT.
- **Key hydration**: the key is echoed verbatim, so `populateFromAPI` hydrates `ssl.key` from the response with no preservation logic.
- **client_auth**: no server default, so `client_auth` is Optional-only and omitted from requests when null.
- **Version floor**: 9.1.0 retained; below-floor behavior is covered by unit tests of `GetVersionRequirements` only.
- **Deferred-field protection**: before any PUT that sends `ssl`, the provider reads the host and returns an error diagnostic when it holds non-null `agent_*`/`es_*` fields.
- **Update response**: the update client path must tolerate the string-encoded `ssl` in the PUT response; the resource state is built from a follow-up read.
