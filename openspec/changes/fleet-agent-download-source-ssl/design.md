## Context

`elasticstack_fleet_agent_download_source` never sets or reads the Kibana Fleet `ssl` request/response
field on `POST /api/fleet/agent_download_sources` and `PUT /api/fleet/agent_download_sources/{sourceId}`.
Users who point an agent binary download source at a host with a self-signed or privately-issued
certificate (a common air-gapped/on-prem pattern) have no way to supply a trusted CA bundle or a
client certificate/key for mTLS through this resource today, and no opaque/passthrough attribute
covers it. The gap is purely additive: `ssl` is optional on both request bodies, so existing
configurations are unaffected.

The generated `kbapi` client already exposes `Ssl *struct{ Certificate *string; CertificateAuthorities
*[]string; Key *string }` on `PostFleetAgentDownloadSourcesJSONBody`,
`PutFleetAgentDownloadSourcesSourceidJSONBody`, and the read-response item — confirmed directly in
`generated/kbapi/kibana.gen.go`. No `generated/kbapi` regeneration is required. The field has no
`verification_mode` sibling, unlike `elasticstack_fleet_output`'s `ssl` block.

The same request/response bodies also already expose a sibling `Secrets.Ssl.Key` union field (e.g.
`generated/kbapi/kibana.gen.go:84028-84041` for the POST body, `:84175-84196` for PUT, and the
analogous `Secrets.Ssl.Key` on both GET response shapes), alongside `Secrets.Auth.{ApiKey,Password}`.
This design originally deferred using it entirely (see Decision 1a below for why it is now used for
`key`).

This design originally adopted the recommendation from issue #5243's implementation-research comment
(a plain, validator-rejects-empty `ssl` object, deferring `secrets.ssl.key`). It has since been
updated to resolve issue #5242 (the `secrets` gap for this resource) for the `ssl.key` sub-attribute,
per the deviation recorded below.

## Deviation from the #5242 implementation-research recommendation

Issue #5242's implementation-research comment recommended **Approach A**: per-field write-only
attributes (`ssl.key_wo`, `auth.api_key_wo`, `auth.password_wo`) with automatic bcrypt-hash-based
change detection, private-state threading, a `secrets_generation` marker attribute, and a new
`ModifyPlan` hook on the Kibana resource envelope — modeled on
`elasticstack_elasticsearch_connector`'s `secret_value` pattern. That research reached only a
`research-needs-human` outcome (score 84/100, round 5/5, "Grounded" check still failing) and listed
several blocking open questions (bcrypt input-length handling, version-gate confirmation, and how
`GetVersionRequirements` would inspect write-only config values with no existing mechanism to do so).

A subsequent human comment on #5242 from the issue reporter (`@tobio`) explicitly contradicts that
recommendation, grounded in (a) the `auth` attribute having since shipped in #5263 as plain
`Sensitive` strings with no write-only/hash machinery at all, and (b) live probes against a Kibana
9.4.0 stack. It decides: **plain `Sensitive` attributes, not write-only-plus-hash**, for `ssl.key` too,
so it stays consistent with the shipped `auth.password`/`auth.api_key` shape; Approach A (and Approach
B, the JSON-blob alternative) are both dropped. This design adopts that human decision instead of the
research's Approach A recommendation — the explicit contradiction rule takes precedence over the
default "adopt the research comment's recommendation" rule. The `secrets_generation` schema sketch,
the private-state/`ModifyPlan` adaptation items, and the bcrypt-length concern from the research are
therefore superseded and not part of this design; they do not apply to a plain `Sensitive` string with
no change-detection machinery to build (a plan diff on a `Sensitive` attribute is already detected by
ordinary Terraform config diffing, same as `auth.password` today).

The research's other open questions are resolved by the human comment's probes and carried into
Decisions 1, 2, and 4 below:

| Research open question | Resolution |
|---|---|
| Does Kibana echo plaintext `secrets.*`? | No — `auth.password`/`auth.api_key` are moved into Fleet secrets and read back as `secrets.auth.*: {id}`, never plaintext. |
| Does Kibana echo a plain `ssl.key`? | Yes, verbatim, if sent via the plain `ssl.key` field. Only `secrets.ssl.key` returns `{id}` and suppresses the plaintext field from the response — this is why `key` must be wired through `secrets.ssl.key`, not `ssl.key`, to get secret-storage behavior at all. |
| What minimum version gates `secrets.*`/`ssl`/`auth`? | Unchanged from this design's existing Decision 3: `ssl` at 9.1.0 (docs-based), `auth` at 9.4.0 (shipped); only 9.4.0 was available to test against, so the 9.1.0 gate remains unverified by live probe. |
| Can `auth.api_key` and basic auth coexist? | No (confirmed by live probe) — already enforced by the shipped `auth` validators (`ConflictsWith`/`AlsoRequires` in `schema.go`); no change needed here. |
| `ssl` clearing / PUT-omission semantics | Resolved directly into Decision 4 below: omitting `ssl` leaves it unchanged; `ssl: {}` clears it; omitting one child clears only that child. No `RequiresReplace` needed anywhere. |
| Pairing constraint between `certificate` and `key`? | None confirmed — no change to the independent-optionality of `ssl`'s three children. |
| Does `fleet-agent-download-source`'s spec forbid `auth`/`secrets`? | Yes, the "Auth and secrets not in v1 schema" requirement bans both; the delta spec below amends it to allow `ssl` (including the `secrets.ssl.key` wiring), consistent with this change's existing "Modified Capabilities" scope. |

The research's `auth.api_key_wo`/`auth.password_wo` write-only redesign and Fleet-Server-version
question remain genuinely out of scope for this change: `auth` already shipped in #5263 using plain
`Sensitive` strings and is not being redesigned here.

## Goals / Non-Goals

**Goals:**
- Let practitioners supply `certificate_authorities`, `certificate`, and `key` for
  `elasticstack_fleet_agent_download_source`, matching the actual Kibana API shape (an object with a
  CA list, not a single string).
- Reuse the existing, tested `elasticstack_fleet_output` `ssl` mapping pattern rather than inventing a
  new one.
- Keep the "ssl block" concept consistent across Fleet resources that already expose one.
- Gate `ssl` with its own narrower, attribute-scoped minimum-version requirement (9.1.0) without
  raising the resource's existing `8.13.0` floor for configurations that don't use `ssl`.
- Resolve the empty-value and PUT-clearing ambiguities explicitly (see Decisions) instead of silently
  inheriting `elasticstack_fleet_output`'s unaddressed edge cases.
- Wire `ssl.key` through the Fleet `secrets.ssl.key` managed-secret field so Kibana stores it as a
  Fleet secret rather than echoing it back in plaintext, matching the shipped `auth.password`/
  `auth.api_key` behavior and closing the `secrets.ssl.key` part of #5242 for this resource.

**Non-Goals:**
- Exposing the `{id}`-reference form of `secrets.ssl.key`/`secrets.auth.*` (pointing at a
  pre-existing Fleet secret store entry by ID, rather than submitting a raw value for Fleet to store)
  — no existing provider pattern does this for any resource, and no demonstrated user need.
- Changes to `elasticstack_fleet_output` or `elasticstack_fleet_proxy`'s own `ssl`/certificate
  handling.
- Redesigning authentication on download sources (the `auth` object) — already shipped in #5263 as
  plain `Sensitive` `username`/`password`/`api_key`/`headers` attributes, with its own secret-storage
  behavior server-side; not revisited here. The `secrets.auth.*` part of #5242 is therefore already
  closed functionally and needs no further schema change.
- Write-only attributes, hash-based change detection, private state, or a `secrets_generation` marker
  for `ssl.key` — see "Deviation from the #5242 implementation-research recommendation" above; this
  design uses a plain `Sensitive` string instead, consistent with `auth.password`/`auth.api_key`.

## Decisions

### Decision 1: Nested `ssl` single-object attribute mirroring `elasticstack_fleet_output`

Add a `ssl` attribute of type `schema.SingleNestedAttribute` (Optional) to
`internal/fleet/agentdownloadsource/schema.go`, with:
- `certificate_authorities` — `types.List` of strings (Optional)
- `certificate` — `types.String` (Optional)
- `key` — `types.String` (Optional, **Sensitive: true**)

This is the same nested shape `internal/fleet/output/schema.go`'s `ssl` block already uses, including
marking `key` sensitive. There is no `verification_mode` attribute, since the download-source API has
none.

Mapping follows `internal/fleet/output/models_ssl.go`'s `objectValueToSSL`/`sslToObjectValue` pattern
for `certificate`/`certificate_authorities` only, adapted into a new
`internal/fleet/agentdownloadsource/models_ssl.go`: convert the Terraform object to/from the kbapi
anonymous `Ssl` struct in `toAPICreateModel`/`toAPIUpdateModel`
(`internal/fleet/agentdownloadsource/models.go`), and hydrate those two fields back from `item.Ssl` in
`read.go`'s `readAndHydrateState`.

`key` is handled differently from `certificate`/`certificate_authorities`, per Decision 1a below.

**Why:** Reuses a proven, tested pattern in the same codebase (`models_ssl_test.go`); matches the
nested shape Kibana actually sends/returns; minimal new code; consistent attribute semantics with
`elasticstack_fleet_output` for users who already use both resources.

### Decision 1a: `key` is wired through `secrets.ssl.key`, not the plain `ssl.key` field

Both `PostFleetAgentDownloadSourcesJSONBody`/`PutFleetAgentDownloadSourcesSourceidJSONBody` expose a
`Secrets.Ssl.Key` field alongside the plain `Ssl.Key` field (`generated/kbapi/kibana.gen.go`, e.g.
lines ~84028-84041 for POST, ~84175-84196 for PUT). It is a two-arm union
(`...SecretsSslKey0{Id string}` / `...SecretsSslKey1 = string`) with generated constructors
`FromPostFleetAgentDownloadSourcesJSONBodySecretsSslKey1(v string) error` (and the PUT equivalent) for
the raw-string arm. The same union shape exists on both GET response items' `Secrets.Ssl.Key`.

A live-Kibana probe (recorded in the human comment resolving #5242's open questions) found that
sending the plain `ssl.key` field causes Kibana to echo it back verbatim in plaintext on every
subsequent GET — i.e. plaintext storage in Terraform state would be unavoidable if `key` mapped to
`ssl.key`. Sending the same raw string through `secrets.ssl.key` instead causes Fleet to store it as a
managed secret (the same outcome `auth.password`/`auth.api_key` already get by being sent as plain
`auth.password`/`auth.api_key` and moved server-side into Fleet secrets), and `secrets.ssl.key` always
reads back as `{id}`, never the plaintext value.

**Decision:** `toAPICreateModel`/`toAPIUpdateModel` set `body.Secrets.Ssl.Key` (via the string-arm
constructor) whenever `ssl.key` is configured in Terraform, and never set the plain `body.Ssl.Key`
field. On read, `item.Ssl.Key` (the plain field) is therefore expected to always be absent once a key
has been set through `secrets.ssl.key`, and `item.Secrets.Ssl.Key` only ever carries an opaque `{id}`
— so `key` is populated by carrying the prior Terraform state value forward (the same
`preserveRedactedSecret` helper `read.go` already uses for `auth.password`/`auth.api_key`), never by
reading either `item.Ssl.Key` or `item.Secrets.Ssl.Key` directly. This means `key` behaves like
`auth.password`/`auth.api_key` for state purposes: the provider never persists a value it read back
from the API into that attribute, only the value the practitioner last configured.

**Why:** Matches the only mechanism confirmed, by live probe, to make Fleet treat the key as a managed
secret rather than plaintext; reuses the existing `preserveRedactedSecret` read-side pattern instead
of inventing a new one; keeps the Terraform-facing shape identical to the originally-planned plain
`ssl.key` string (practitioners write `ssl = { key = var.key }` exactly as before — only the internal
wire routing changed).

**Rejected alternative — flat top-level attributes mirroring `elasticstack_fleet_proxy`**
(`ssl_certificate`, `ssl_certificate_authorities` as a single string, `ssl_key`): rejected because
`certificate_authorities` is an array (up to 10 entries) in the actual API, so a flat string loses the
ability to supply more than one CA — a functional regression, not just a style choice — and it would
diverge from the `elasticstack_fleet_output` convention for the same concept.

### Decision 2: Reject ambiguous-empty `ssl` input with validators, not `Computed`

Directly reusing `output/models_ssl.go`'s `sslToObjectValue` as-is would turn a user-configured,
non-null `ssl = {}` (or any populated `ssl` object whose `certificate`/`key` is an empty string, or
whose `certificate_authorities` is `[]`) into `types.ObjectNull`/`types.ListNull` on the next read — a
plan-after-apply inconsistency for a plain `Optional` attribute. Marking `ssl` `Computed` does not fix
this: `Computed` only lets the provider supply a value when the practitioner's config has no value at
all; it does not let the provider override a non-null, explicitly-configured value during planning, so
a known `ssl = {}` in config would still conflict with a null refresh.

**Decision (superseded from the original plan — see below):** this change originally planned to
reject emptiness with validators, extending `output/schema.go`'s nonemptiness validators
(`certificate`/`key`: `stringvalidator.LengthAtLeast(1)`; `certificate_authorities`:
`listvalidator.SizeAtLeast(1)`) plus a new object-level "at least one of" validator rejecting
`ssl = {}` outright.

**Revised decision (per the human comment resolving #5242):** accept `ssl = {}` and empty/omitted
children, matching how `auth = {}` and an all-null `auth` are already accepted
(`internal/fleet/agentdownloadsource/schema.go`'s `auth` has no such validators at all). Drop:
- the object-level "at least one of `certificate`/`certificate_authorities`/`key`" validator;
- `certificate`'s and `key`'s `stringvalidator.LengthAtLeast(1)` (an explicitly-configured empty
  string is now accepted for both);
- `certificate_authorities`'s `listvalidator.SizeAtLeast(1)` (an explicitly-configured empty list,
  `[]`, is now accepted).

Keep, unchanged, only the two validators that guard the API's own structural constraints rather than
ambiguous emptiness:
- `certificate_authorities`: `listvalidator.SizeAtMost(10)` (the spec's `maxItems: 10`);
- `certificate_authorities`: `listvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1))` (an
  empty-string element can't hide inside an otherwise-nonempty list — this is about a malformed
  individual entry, not an ambiguous-empty collection, so it stays).

**Why:** Consistency with `auth`'s existing, already-shipped permissiveness is more valuable than the
originally-planned strictness, and the strictness was never load-bearing for a real API constraint —
the Fleet API itself has no `minProperties`/non-empty requirement on `ssl` or its children (confirmed:
no such constraint in `oas.yaml:38142-38235`/`:38712-38805`). Diverging from `auth`'s precedent for no
functional reason would be a worse user experience, not a safer one.

### Decision 3: Attribute-scoped 9.1.0 version requirement, resource floor unchanged

Elastic docs mark TLS support for agent binary download sources "Generally available since Elastic
Stack 9.1"; the OpenAPI spec itself carries no `x-state`/version annotation on `ssl`. The resource's
existing overall gate stays `8.13.0` (`agentDownloadSourceMinVersion` in `models.go`) for all other
fields; `ssl` gets its own, narrower, attribute-scoped requirement using
`entitycore.NewAttributeVersionRequirement(attr path.Path, minVersion version.Version, errorMessage
string)`, wired via the guarded-closure pattern already used in
`internal/fleet/agentpolicy/version_requirements.go` (append the requirement only when `ssl` is
actually configured).

### Decision 4: PUT-clearing semantics — resolved by live probe, no `RequiresReplace` needed

This was originally planned as a pre-implementation probe with a branch per outcome (see tasks.md
history). The probe has since been run against a live Kibana 9.4.0 stack (recorded in the human
comment resolving #5242's open questions) and resolved unconditionally:

- Omitting `ssl` entirely from a `PUT` leaves the previously-set value unchanged server-side (Kibana
  treats omission as a merge, same as `auth`).
- Sending `ssl: {}` clears it; a subsequent `GET` then returns `"ssl": {}`.
- Omitting a single child (`certificate`, `key`, or `certificate_authorities`) while sending the other
  two clears only the omitted child and leaves the other two untouched.

**Decision:** every `ssl` transition — full removal, per-child removal, or any combination — is
encoded as an in-place `PUT` with the appropriate fields present or omitted; none of them requires
resource replacement. `toAPIUpdateModel` therefore:
- sends the full `Ssl`/`Secrets.Ssl` object whenever `ssl` is configured (known, non-null), with each
  child field present if set in config and omitted (nil pointer) if not — Go's
  `json:"...,omitempty"` tags already produce this encoding from a straightforward per-child mapping,
  with no special-casing needed;
- sends an explicit, non-nil, all-nil-fields `Ssl{}` (i.e. `ssl: {}` on the wire) when `ssl` was
  previously set and is now removed from config — mirroring `authForUpdate`'s existing treatment of
  removing a previously-configured `auth` block in `models.go` — rather than a nil `Ssl` pointer,
  since a nil pointer would omit the field and leave the prior value unchanged instead of clearing it;
- sends nothing (nil `Ssl`/`Secrets.Ssl`) when `ssl` was never configured and still isn't.

No `planmodifier.Object`/`RequiresReplace` fallback is added for `ssl`, unlike what the original
pre-probe plan anticipated — every transition Terraform can produce has a direct in-place encoding.

**Why:** This is the simplest possible implementation consistent with the confirmed API behavior, and
it reuses the `auth` removal pattern (`authForUpdate` in `models.go`) directly instead of introducing
a second removal mechanism.

## Risks / Trade-offs

| Risk | Mitigation |
|---|---|
| Reusing `output/models_ssl.go`'s normalization as-is would mask user-configured emptiness as null | Normalize to null only when the API returns nothing for `certificate`/`certificate_authorities`, same as `output`; emptiness is now accepted rather than rejected (Decision 2), so there is no rejection-vs-normalization conflict left to resolve. |
| `ssl`'s own minimum version (9.1) differs from the resource's overall floor (8.13) | Use the existing attribute-scoped `entitycore.NewAttributeVersionRequirement` mechanism instead of raising the resource floor. |
| Sending `key` through `secrets.ssl.key` is a different wire path from `certificate`/`certificate_authorities`, so a future reader of `models_ssl.go` could assume all three round-trip the same way | Decision 1a documents the asymmetry explicitly and co-locates the `key`-specific `preserveRedactedSecret` handling next to the `auth.password`/`auth.api_key` precedent it mirrors, rather than hiding it inside the shared `certificate_authorities`/`certificate` mapping helper. |
| The 9.1.0 version gate for `ssl` (and by extension `secrets.ssl.key`) remains docs-based, not independently verified against a live pre-9.1 stack | Unchanged risk from Decision 3; only 9.4.0 was available to test in the probe that resolved #5242's other open questions. |

## Open questions

- Does the Kibana API enforce any pairing constraint (e.g. `certificate` and `key` must both be set, or
  neither) for client mTLS, or are all three `ssl` sub-fields independently optional as the spec's
  `additionalProperties: false` object with no internal `required` suggests? The live probe did not
  test this specifically; not blocking the schema shape, but affects whether the provider should add
  its own validators (e.g. `stringvalidator` cross-field checks).
- Is the per-child clearing behavior confirmed by the live probe (omitting one child leaves the other
  two untouched) also true specifically for the `secrets.ssl.key`/plain-`ssl` combination this design
  uses, as opposed to the all-plain-`ssl` shape the probe's steps describe? The probe's answers appear
  to post-date the decision to route `key` through `secrets.ssl.key` (the human comment re-grounds
  explicitly against that shape), but tasks.md includes a verification step to confirm this directly
  against a live Kibana before considering the implementation complete.
