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

This design adopts the recommendation from the issue's implementation-research comment.

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

**Non-Goals:**
- Supporting `secrets.ssl.key` (the Fleet secret-storage reference variant of the client key) — the
  issue only flags the plain `ssl` object gap, and no existing pattern in this provider handles that
  union type for download sources.
- Changes to `elasticstack_fleet_output` or `elasticstack_fleet_proxy`'s own `ssl`/certificate
  handling.
- Authentication on download sources (the `auth` object and `secrets.auth` variant) — a real,
  spec-grounded gap, but not the one this issue raises; left for a separate change.

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

Mapping follows `internal/fleet/output/models_ssl.go`'s `objectValueToSSL`/`sslToObjectValue` pattern,
adapted into a new `internal/fleet/agentdownloadsource/models_ssl.go`: convert the Terraform object
to/from the kbapi anonymous `Ssl` struct in `toAPICreateModel`/`toAPIUpdateModel`
(`internal/fleet/agentdownloadsource/models.go`), and hydrate it back from `item.Ssl` in
`read.go`'s `readAndHydrateState`.

**Why:** Reuses a proven, tested pattern in the same codebase (`models_ssl_test.go`); matches the
nested shape Kibana actually sends/returns; minimal new code; consistent attribute semantics with
`elasticstack_fleet_output` for users who already use both resources.

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

**Decision:** reject emptiness with validators, extending the pattern `output/schema.go` already uses
for `certificate`, `key`, and `certificate_authorities` (all three already carry nonemptiness
validators there — `certificate`/`key`: `stringvalidator.LengthAtLeast(1)`; `certificate_authorities`:
`listvalidator.SizeAtLeast(1)`). These three are reused directly, unchanged. What's new for this
resource's `ssl`:
- `certificate_authorities`: add `listvalidator.SizeAtMost(10)` (enforcing the spec's `maxItems: 10`)
  plus `listvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1))` so an empty-string element
  can't hide inside an otherwise-nonempty list.
- `ssl` object itself: a new object-level validator requiring at least one of `certificate`,
  `certificate_authorities`, `key` to be set, so `ssl = {}` is rejected outright at plan time.

### Decision 3: Attribute-scoped 9.1.0 version requirement, resource floor unchanged

Elastic docs mark TLS support for agent binary download sources "Generally available since Elastic
Stack 9.1"; the OpenAPI spec itself carries no `x-state`/version annotation on `ssl`. The resource's
existing overall gate stays `8.13.0` (`agentDownloadSourceMinVersion` in `models.go`) for all other
fields; `ssl` gets its own, narrower, attribute-scoped requirement using
`entitycore.NewAttributeVersionRequirement(attr path.Path, minVersion version.Version, errorMessage
string)`, wired via the guarded-closure pattern already used in
`internal/fleet/agentpolicy/version_requirements.go` (append the requirement only when `ssl` is
actually configured).

### Decision 4: PUT-clearing semantics — pre-implementation probe with a branch per outcome

**Probe outcome (Kibana 9.4.0):** A live probe using the generated `kbapi` client
created five independent download sources and verified each PUT with a separate GET.
Omitting `ssl` preserved all three existing fields. Sending `ssl = {}` cleared all
three fields; GET returned a present-but-empty object. In three independent
single-child omission probes, the omitted child cleared and the retained two
persisted. All PUTs returned HTTP 200, and all five probe sources were deleted.

The pre-specified middle branch applies: `toAPIUpdateModel` must send a non-nil,
all-nil-fields `Ssl` struct (explicit `{}`) for whole-block removal. Individual
children can be removed in place by omitting them from a supplied SSL object.
Neither whole-object nor per-child `RequiresReplace` modifiers are needed.
Read hydration must normalize a present-but-empty response object to null.
This behavior was measured on 9.4.0, not separately on the 9.1.0 minimum.

Does omitting `ssl` (or sending `ssl = {}`, which the spec does not declare `nullable: true`, so there
is no JSON `null` option) on `PUT /api/fleet/agent_download_sources/{sourceId}` clear a previously-set
SSL config server-side, or leave it unchanged? The PUT operation description states no
clearing/merge behavior either way, so this is not resolvable from the spec text alone and must be
settled with a live probe before implementation, using the already-generated client:

1. Create a source with `ssl` fully populated (certificate + CA list + key).
2. PUT an update that omits `ssl` from the request body entirely; independently GET the source
   afterward and record whether `ssl` persisted or was cleared.
3. Against a fresh copy of the same starting state, PUT an update sending `ssl = {}`; GET and record
   the result.
4. Against three independent fresh copies of the same fully-populated starting state, PUT three
   separate updates that each omit exactly one child — omit only `certificate`; omit only `key`; omit
   only `certificate_authorities` — and independently GET each afterward to record whether the omitted
   child cleared while the retained two persisted.

Decision branches from the probe's outcome:
- If omitting `ssl` (step 2) clears it: the currently-planned encoding — a nil `Ssl` pointer in
  `toAPIUpdateModel`, which Go's JSON encoding omits by default — is already correct; no extra code is
  needed to support removal.
- If omitting `ssl` does **not** clear it but `ssl = {}` (step 3) does: `toAPIUpdateModel` must send a
  non-nil, all-nil-fields `Ssl` struct (i.e. explicit `{}`) when the user removes the block, instead of
  a nil pointer.
- If neither step 2 nor step 3 clears a previously-set `ssl`: the resource must add a
  `planmodifier.Object` on `ssl` that forces `RequiresReplace` specifically for the transition from a
  known, non-null `ssl` to a null/removed `ssl` (the framework supports conditioning a plan modifier on
  old-vs-new value, the same mechanism already used for `source_id`'s `RequiresReplace` in
  `schema.go`) — removal becomes delete+recreate instead of an in-place PUT the API can't actually
  honor. This same fallback also covers the middle branch above if `ssl = {}` ever turns out not to
  clear reliably either.
- Step 4's three sub-probes determine, per child, which transitions are safe as plain in-place updates
  vs. need the same `RequiresReplace` fallback, applied per-child.

This probe is a concrete pre-implementation task (see tasks.md), not an open-ended research question —
every possible outcome has a specified, Terraform-consistent implementation.

## Risks / Trade-offs

| Risk | Mitigation |
|---|---|
| PUT-clearing semantics are unknown until the probe runs | Every outcome branch is pre-specified (direct encoding or `RequiresReplace` fallback); the probe is a task, not a blocker on design. |
| Reusing `output/models_ssl.go`'s normalization as-is would mask user-configured emptiness as null | Reject emptiness with validators instead (Decision 2), rather than normalizing. |
| `ssl`'s own minimum version (9.1) differs from the resource's overall floor (8.13) | Use the existing attribute-scoped `entitycore.NewAttributeVersionRequirement` mechanism instead of raising the resource floor. |

## Open questions

- The PUT-clearing probe is complete on Kibana 9.4.0; see Decision 4's recorded
  outcome. Use explicit `{}` for whole-block removal and ordinary child omission
  for per-child removal. No replacement fallback is needed for the measured behavior.
- Does the Kibana API enforce any pairing constraint (e.g. `certificate` and `key` must both be set, or
  neither) for client mTLS, or are all three `ssl` sub-fields independently optional as the spec's
  `additionalProperties: false` object with no internal `required` suggests? Not blocking the schema
  shape, but affects whether the provider should add its own validators (e.g.
  `stringvalidator` cross-field checks).
- Should the separate `secrets.ssl.key` field (a Fleet-managed-secret reference variant of `key`,
  distinct from the plain `ssl.key` string) be supported now or deferred? No existing provider code
  handles Fleet secret-reference unions for this resource. Not blocking; see Out of scope.
