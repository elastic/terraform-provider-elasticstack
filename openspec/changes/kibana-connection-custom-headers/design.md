## Context

The Kibana connection block (`kbConnectionBlockSpec()` in `internal/schema/kb_connection_shared.go`)
is the single source of truth from which the provider-level `kibana` block, every entity-local
`kibana_connection` block, the ephemeral-resource variant, and the action-resource variant are all
generated. It currently has no attribute analogous to the Elasticsearch connection block's `headers`
map (`attrHeaders`, `internal/schema/connection_constants.go`; wired into `esConnectionBlockSpec()` in
`internal/schema/es_connection_shared.go`), so there is no way to attach a custom header (e.g. an
identity-aware proxy token) to outgoing Kibana API requests.

The Kibana client's HTTP transport (`internal/clients/kibanaoapi/client.go`) already has a single
injection point for request headers: a custom `http.RoundTripper` (`transport.RoundTrip`) that today
unconditionally handles two cases:
- `kbn-xsrf: true` is added (via `req.Header.Add`, not `Set`) for every non-GET/HEAD request.
- `Authorization` is set only when a provider auth method (bearer token, API key, or basic auth) is
  configured.

Both writes are conditional, so a user-supplied `headers["kbn-xsrf"]` or `headers["Authorization"]`
could reach the wire unmodified in some request/auth combinations even if the custom-headers loop runs
strictly before the existing writes. Per the human direction on the issue, this change proceeds with
the research recommendation and its post-critique fixes: no Stack version gate, acceptance coverage run
across the provider's supported Stack-version matrix, and explicit handling of `KIBANA_ENDPOINT`
precedence in the new acceptance test.

The client currently uses Go's default redirect behavior. Because the custom transport injects
configured headers on every round trip, an unconstrained redirect could forward a proxy credential to
a different origin. The redirect boundary therefore forms part of the credential-delivery design.

## Goals / Non-Goals

**Goals:**
- Let users attach one or more supported custom HTTP headers to every outgoing Kibana API request, configurable
  at the provider level (`kibana { headers = {...} }`) and per-entity (`kibana_connection { headers =
  {...} }`), via the existing shared-spec mechanism so no per-entity schema changes are needed.
- Prevent a configured custom header from silently colliding with provider-managed headers
  (`kbn-xsrf`, `Authorization`) or being ignored by Go's special handling of `Host`, by rejecting those
  names outright.
- Prevent configured custom headers from being forwarded when Kibana or an intermediary redirects to
  a different origin, including an HTTPS-to-HTTP destination.
- Keep Fleet's client configuration unaffected by Kibana-only custom headers, since Fleet's endpoint
  can diverge from Kibana's.
- Ship as a normal additive feature with no stack-version gate, and prove the behavior holds across the
  provider's supported Stack-version test matrix.

**Non-Goals:**
- Adding a `headers` attribute to the Fleet connection block (tracked as a non-blocking follow-up).
- Changing the existing Elasticsearch `headers` attribute's behavior.
- Adding new redaction logic to the debug-log transport (`internal/debugutils/http_log.go`);
  `Authorization` is already masked by the existing debug formatter, while arbitrary custom
  proxy-token headers are not generally redacted. This change only documents that limitation.
- Extending the ephemeral Kibana connection snapshot (`internal/entitycore/ephemeral_connection_snapshot.go`)
  — no Kibana ephemeral resource is registered in `provider/plugin_framework.go` today, so there is no
  reachable ephemeral `kibana_connection` for a `headers` field to serialize.
- Any change to generated Kibana client code (`generated/kbapi/kibana.gen.go`); the needed extension
  points (`WithHTTPClient`) already exist and are already used.

## Decisions

### 1. Add `headers` to the shared `kbConnectionBlockSpec()`, not to individual entities

Add a `headers` entry (`kind: connAttrMap, sensitive: true, writeOnly: true`) to the `attrs` slice
returned by `kbConnectionBlockSpec()`, matching the shape of the Elasticsearch `headers` attribute in
`esConnectionBlockSpec()`. Reuse the existing `attrHeaders` constant from
`internal/schema/connection_constants.go` and add a new Kibana-specific description constant (the
existing `descHeaders` text says "...to Elasticsearch", so Kibana needs its own wording).

Those flags are shared-spec inputs, not a promise that every generated schema marks the attribute
write-only. The existing factories apply `Sensitive` to provider, managed-resource, data-source, and
ephemeral attributes, while only the action factory applies `WriteOnly`. Consequently, entity-local
managed-resource and data-source header values remain in Terraform state, marked sensitive; action
header values are write-only; and ephemeral values follow Terraform's ephemeral-state semantics.

Why:
- `kbConnectionBlockSpec()` is the sole generator for the managed-resource, ephemeral-resource, and
  action-resource variants of the Kibana connection block (via `connAttrFactory[T]` in
  `internal/schema/connection_spec.go`), so one attribute addition here reaches every call site
  (`GetKbFWConnectionBlock()`, the entity-local block, etc.) without per-entity changes.
- `connAttrMap` is already a supported `connAttrKind` with working `mapAttr()` implementations for all
  three entity-kind factories (`fwConnAttrFactory`, `ephemeralConnAttrFactory`, `actionConnAttrFactory`
  in `connection_spec.go`), because the Elasticsearch `headers` attribute already exercises that path.
  No new generator code is required.

Alternatives considered:
- A Kibana-only environment variable (e.g. `KIBANA_HEADERS`) with no HCL attribute: rejected because it
  has no precedent anywhere in this provider (not even for the Elasticsearch `headers` attribute),
  cannot vary per connection block, and requires an awkward delimited-string encoding for header names
  and values that may themselves contain `=` or `,`.

### 2. Validate the reserved header names at config-build time, not as a schema-level map validator

The `connAttrSpec` struct (`internal/schema/connection_spec.go`) only carries `validators
[]validator.String` — there is no map-level validator plumbing today, and the existing Elasticsearch
`headers` attribute has no validators at all. Rather than extending `connAttrSpec`/`connAttrFactory[T]`
with a new `[]validator.Map` field (which would ripple into all three entity-kind factories for a
single attribute), reject the reserved names where `headers` is actually decoded into
`kibanaoapi.Config`: inside `buildKibanaOapiConfigFromFramework` in
`internal/clients/config/kibana_oapi.go`, immediately after `kibConfig.Headers.ElementsAs(...)`.
Case-insensitively comparing each configured header key against `kbn-xsrf`, `authorization`, and
`host` and returning a diagnostic error that names the offending key blocks use of that connection
before it sends its first request. `Host` must be rejected because Go sends the request authority from
`req.Host` or the request URL rather than `req.Header["Host"]`; accepting it into the ordinary header
map would not put the configured value on the wire. Provider-level configuration is built during
provider configuration, while a new entity-local scoped connection may not be built until Create;
this design therefore does not promise that every reserved-name error appears during planning.

Why:
- Matches the precedent already used for endpoint/CA decoding errors in the same function (returning
  `fwdiags.Diagnostics` on invalid input).
- Avoids a second place (a schema-level validator function) that would need to know about the three
  reserved names, keeping the reserved-name policy in a single function.
- Guarantees rejection before the connection can send an HTTP request, which is the security boundary
  this validation protects, without expanding the shared schema-generator validator surface.

Alternatives considered:
- Reject at the point `transport.RoundTrip` consumes `Config.Headers`: rejected because it defers the
  error until request execution rather than scoped-client construction.
- Extend `connAttrSpec` with a `[]validator.Map` field: this would provide uniform plan-time validation,
  including for new entity-local resources, but is rejected as unnecessary generator surface area for
  a single attribute when rejection before the first request satisfies the required safety property.

### 3. Inject custom headers in `transport.RoundTrip`, after the reserved-name check removes any collision risk

In `internal/clients/kibanaoapi/client.go`, add `Headers map[string]string` to `Config`, and in
`transport.RoundTrip` iterate `t.Headers` and `req.Header.Set(name, value)` for each entry. Because
reserved names are already rejected at config-build time (Decision 2), ordering relative to the
existing `kbn-xsrf`/`Authorization` writes is no longer safety-critical, but the custom-headers loop
should still run before those writes, as defense in depth, so a future change to the reserved-name list
fails safe (the provider's own header always wins) rather than fails open.

Why:
- Reuses the single existing injection point that already handles per-request header writes and is
  already covered by `transport_test.go`.
- `req.Header.Set` (rather than `Add`) matches how `Authorization` is already written, and is correct
  for accepted single-valued custom headers after the transport-special `Host` name is rejected.

### 4. Reject cross-origin redirects when custom headers are configured

Configure the Kibana HTTP client with a redirect policy whenever `Config.Headers` is non-empty.
Compare every redirect destination with the initial request's origin, where origin consists of the
scheme, canonical host, and effective port. Allow same-origin redirects, but return an error before
sending a redirected request when any origin component differs. This rejects cross-host and
cross-port redirects as well as HTTPS-to-HTTP downgrades.

Why:
- Redirected requests pass through `transport.RoundTrip` again, so clearing headers only on the
  redirect request would not help: the transport would re-inject them.
- Rejecting the redirect gives the practitioner a visible failure instead of silently following an
  unexpected destination without the configured proxy credential.
- Applying the restriction only when custom headers are configured preserves the client's existing
  redirect behavior for unaffected configurations.

Alternatives considered:
- Strip custom headers on cross-origin redirects and continue: rejected because the custom transport
  would re-inject them unless additional request-scoped suppression were introduced, and continuing
  without a required proxy credential would usually fail later with a less useful error.
- Disable all redirects: rejected because ordinary same-origin redirects do not cross the credential
  boundary and may be required by a Kibana proxy.

### 5. Explicitly drop `Headers` in `toFleetConfig()`

The Kibana and Fleet client configurations share an underlying configuration shape, and Fleet is
currently derived by converting the effective Kibana configuration. Once custom headers are added,
that conversion would carry them into the Fleet client unchanged. Since Fleet's endpoint can be
configured independently of Kibana's (through the Fleet block or environment), forwarding a
Kibana-proxy header by default risks leaking that header's value to a different host. The Fleet
derivation boundary therefore explicitly clears Kibana custom headers.

Why:
- Keeps the default behavior safe (no implicit credential forwarding to a potentially different host)
  without requiring a decision about Fleet's own `headers` attribute in this change.
- Fleet gaining its own `headers` attribute is left as a clearly-scoped, non-blocking follow-up (see
  Open Questions).

### 6. No Stack/Kibana minimum-version gate; validate via the supported-version test matrix instead

This feature only shapes outgoing HTTP requests before they reach any Kibana/Fleet API operation, so it
has no API-surface dependency on a specific Kibana version. No source consulted during research
confirms or rules out a minimum version for this kind of provider-side request-shaping capability, so
this is an absence-of-evidence inference, not a verified fact (tracked as a non-blocking open question
below) — per the human direction on the issue, this is accepted as non-blocking and the change proceeds
without a `VersionRequirement` check. To guard against an unknown real version dependency surfacing
later, the acceptance test added for this change runs against the provider's standard
supported-Stack-version test matrix (not a single pinned version), so a real dependency would still be
caught by CI rather than assumed away silently.

### 7. Prove headers reach the wire with a local recording reverse-proxy acceptance test, accounting for `KIBANA_ENDPOINT` precedence

There is no real Cloudflare Access or Google IAP available in CI, so state-shape acceptance tests alone
(following `TestAccDataSourceEnrichPolicyConnectionAPIKey` /
`TestAccDataSourceEnrichPolicyConnectionAPIKeyEmptyHeaders`,
`internal/elasticsearch/enrich/acc_test.go`) cannot prove headers are actually transmitted — they only
prove the value is stored in state, and the provider-level `kibana { headers = {...} }` block is
provider configuration, not resource state, so it has no state address to assert against at all. Add an
acceptance test that points at a local `httptest.Server` acting as a minimal recording reverse-proxy:
it records received headers and forwards the request to the real Kibana test instance, for both the
provider-level `kibana` block and a resource-level `kibana_connection` override.

Because `KIBANA_ENDPOINT` currently overrides both provider- and resource-level configured endpoints by
default (`internal/clients/config/kibana_oapi.go`, `internal/clients/config/base.go`), the test must
not let a CI-configured `KIBANA_ENDPOINT` silently bypass the local proxy. The test sets
`TF_ELASTICSTACK_PREFER_CONFIGURED_KIBANA_ENDPOINT=true` for its own test run (or points
`KIBANA_ENDPOINT` itself at the proxy) and asserts that the proxy actually recorded at least one
request with the expected custom header — not just that `terraform apply` succeeded — so a silent
endpoint bypass cannot pass undetected.

Why:
- Matches the research's resolved testability plan and the human direction's explicit instruction to
  "account for `KIBANA_ENDPOINT` precedence."
- Asserting on recorded requests (not just apply success) is the only way to catch a silent bypass
  where the real Kibana endpoint was hit instead of the proxy.

If `internal/acctest` does not already have a reusable header-recording proxy helper, this change adds
one as part of the acceptance-test boundary.

### 8. Test each security boundary at the narrowest observable layer

Use config-level tests for header decoding, reserved-name rejection, and Fleet exclusion; transport
tests for ordinary request injection; HTTP client tests with controlled redirect endpoints for
same-origin success and cross-origin rejection; and acceptance tests with a recording proxy for
provider-level and entity-local delivery. Redirect tests assert that rejected destinations receive no
request at all, including host, effective-port, and scheme changes. Acceptance coverage explicitly
controls endpoint environment precedence and runs through the supported Stack-version matrix.

Why:
- Config and client tests isolate failure boundaries and can prove that rejected inputs or redirects
  do not produce network requests.
- Acceptance tests remain focused on the user-visible schema-to-wire path that unit tests cannot prove.

## Risks / Trade-offs

- **Sensitive managed-resource and data-source header values remain in Terraform state** → Mark them
  sensitive so Terraform masks them in normal CLI output, document the persistence explicitly, and
  reserve true state exclusion for the action variant's write-only attribute and Terraform's
  ephemeral-value semantics.
- **A reserved-name error for a new entity-local connection may surface during apply rather than
  planning** → Build and validate the scoped configuration before its first request, so invalid input
  cannot reach Kibana even though the diagnostic timing varies by context.
- **A cross-origin redirect that previously succeeded will fail when custom headers are configured** →
  Allow same-origin redirects, reject the redirect before the destination receives a request, and
  return an error that identifies the credential-protection boundary.
- **Debug logging can expose custom header values** → Document that the existing debug transport logs
  request headers, that existing formatting masks `Authorization` but not arbitrary custom proxy-token
  headers, and do not claim that `Sensitive` redacts provider debug logs.
- **A recording proxy adds acceptance-test complexity and can be bypassed by endpoint environment
  precedence** → Force configured-endpoint preference (or point the environment endpoint at the proxy)
  and assert that the proxy actually observed the expected header.
- **Kibana headers are deliberately unavailable to Fleet** → Drop them at the Fleet derivation
  boundary; a separate Fleet `headers` attribute requires its own user requirement and follow-up.
