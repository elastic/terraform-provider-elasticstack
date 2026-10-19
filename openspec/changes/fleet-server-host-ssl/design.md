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

Probe create with `client_auth` omitted and inspect the response. If Kibana does not return a default, keep `client_auth` Optional-only and hydrate it as null. If Kibana returns a default, model `client_auth` as Optional+Computed so the response value can be represented without Terraform's inconsistent-result error. In either branch, an omitted configured value must be omitted from the initial request; for the defaulting branch, unknown/computed plan values must not be sent as an explicit client-chosen value.

**Why:** The OAS declares no default for `client_auth` on this endpoint, so the provider must establish the actual response behavior before selecting the schema/hydration branch. Elastic's cited defaulting behavior ("required if CAs configured else none") comes from standalone Elastic Agent's `server.ssl.*` runtime documentation, not from this Kibana API's documented contract.

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

The 9.1.0 floor comes from the Kibana release notes entry "Adds SSL options to Fleet Server host settings" (kibana#208091) cited in the research driving this change; it is not independently re-verified against `generated/kbapi`'s pinned OAS changelog in this proposal. **Confirming or adjusting this exact version is an implementation-time task**, resolved by the version-boundary acceptance tests in tasks.md, not a design-time blocker.

### Decision 4: Removal/clear semantics and `key` read-back are resolved by acceptance testing, not assumed

Two behaviors are genuinely unknown for this specific endpoint until exercised against a live Kibana:

1. **Does `PUT .../fleet_server_hosts/{itemId}` clear a previously-set `ssl` when sent an omitted field, an explicit empty object, or explicit JSON `null`?** `agentdownloadsource`'s own spec (`openspec/specs/fleet-agent-download-source/spec.md`, "ssl removal consistency with PUT semantics") documents that its answer for *its* endpoint came from a pre-implementation probe, and explicitly calls out that the behavior does not generalize. Probe each payload independently against a fresh populated host. If explicit JSON `null` is the only clearing form that works, use a request path that serializes `"ssl": null` explicitly; a nil generated `Ssl` pointer is insufficient because its `omitempty` tag omits the field. Assert the serialized payload. Use a clearing payload shown to work, or force replacement only if all three forms fail.
2. **Is `ssl.key` redacted on `GET`, like `agentdownloadsource`'s secret-backed key?** If so, `internal/entitycore/generic_resource_write.go`'s read-after-write lifecycle would otherwise overwrite a practitioner-configured `key` with a redacted value on every create/update/refresh, not just on import. If redaction is observed, `models.go`'s `populateFromAPI` must preserve the configured key from prior state/plan instead of hydrating it from the (redacted) response — the same `preserveRedactedSecret`-style precedent `agentdownloadsource` uses at `read.go:235`.
3. **Does Kibana return a `client_auth` default when omitted?** Resolve with the create probe before choosing Optional-only versus Optional+Computed. The apply/refresh acceptance test must prove the chosen branch is plan-consistent and produces no subsequent diff when configuration continues to omit `client_auth`.
4. **Can a host contain non-null deferred `agent_*`/`es_*` fields?** Seed such a host and verify an update that would replace `ssl` is rejected before mutation with an actionable diagnostic, leaving those values intact. This protects existing configurations without expanding the Terraform schema.

**Why defer to testing rather than design-time assumption:** The OAS has no description field documenting either behavior for this path, and the one documented precedent in this repo (`agentdownloadsource`) is explicit that its answer is endpoint-specific, not a general Fleet API behavior. Choosing wrong here would either silently fail to clear TLS config server-side while Terraform shows it removed, or silently corrupt a practitioner's configured key on every apply. Both are concretely testable and must be resolved before this ships, per the acceptance tests in tasks.md.

## Risks / Trade-offs

| Risk | Mitigation |
|---|---|
| Removing `ssl` from config doesn't actually clear it server-side (PUT semantics unknown) | Acceptance tests independently probe omission, `{}`, and explicit JSON `null` before the clearing strategy is finalized; forced replacement is the fallback if all fail (Decision 4). |
| `ssl.key` is redacted on read, silently stomping a configured key on every create/update/refresh | Acceptance test asserts round-trip stability of `key` across create→refresh and create→update; hydration logic preserves the configured value if redaction is observed (Decision 4). |
| Kibana returns a `client_auth` default that differs from the omitted plan value | Probe the create response; use Optional+Computed for the returned value and verify apply/refresh has no inconsistent result or subsequent diff (Decision 2). |
| Updating the supported SSL fields erases existing deferred `agent_*`/`es_*` values | Inspect the current stored object before a replacing PUT and return an actionable diagnostic without mutation when deferred values are present; verify with a seeded-host acceptance test. |
| 9.1.0 floor is wrong (not independently confirmed against this repo's pinned OAS changelog) | Explicit version-boundary acceptance tests (below-floor error, at-floor success, `ssl`-omitted-below-floor success) catch a wrong floor during implementation rather than after release. |
| Changing TLS settings on an enrolled Fleet Server host can disconnect connected Elastic Agents in real deployments | Documented in the resource's generated docs; out of scope for the Terraform acceptance test itself, which only asserts API/state, not live agent connectivity. |

## Open questions

- Does Kibana's `fleet_server_hosts` PUT actually clear a previously-set `ssl` value when sent an omitted field, an explicit empty object, or explicit JSON `null`? The OAS doesn't document this for this path, and `agentdownloadsource`'s precedent is for a different endpoint. Non-blocking — each form is tested against a fresh populated host; implementation uses a form that clears or a forced-replacement plan modifier if none do.
- Does Kibana enforce a `client_auth` default or a `certificate`+`key` pairing requirement server-side for this specific endpoint (the OAS declares neither, and the cited defaults come from standalone-Agent runtime docs, not this API)? Non-blocking — the provider does not invent a constraint the API doesn't declare; relevant only to user-facing documentation/examples.
- Does Kibana return a `client_auth` default when it is omitted from create? Non-blocking — the pre-implementation probe selects the plan-consistent schema/hydration branch in Decision 2.
- What is the minimum Kibana/Fleet Server version to gate the new `ssl` attribute on — 9.1.0 is the best available evidence (when the field was added per Kibana release notes) but not independently confirmed against the `kibana-openapi-spec` changelog pinned in this repo. Non-blocking — implement with 9.1.0 as the `GetVersionRequirements` floor and confirm/adjust via the explicit version-boundary acceptance tests in tasks.md.
- Is `ssl.key` returned on GET, or redacted like `agentdownloadsource`'s secret-backed key? This affects create/update/refresh hydration, not just import — a redacted key would need a conditional preservation policy in `models.go`, not only an import-time fix. Non-blocking — resolved by the acceptance test's import/refresh assertions during implementation.
- Are the `agent_*`/`es_*` sub-fields in scope for this change or a separate follow-up issue? Resolved for this change: out of scope (see Non-Goals); tracked as a follow-up.
