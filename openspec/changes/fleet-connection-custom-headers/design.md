## Context

The provider's `fleet {}` connection block is hand-written in `GetFleetFWConnectionBlock()`
(`internal/schema/connection.go:72-129`), a `fwschema.ListNestedBlock` capped at one entry via
`listvalidator.SizeAtMost(1)`. Unlike `kibana {}` / `kibana_connection {}`, it is not generated from the
shared `connectionBlockSpec` mechanism (`internal/schema/connection_spec.go`) that `kbConnectionBlockSpec()`
and `esConnectionBlockSpec()` use, and there is no entity-local `fleet_connection` block anywhere in the
provider — `GetFleetFWConnectionBlock()` has exactly one call site, the top-level provider schema
(`provider/plugin_framework.go:180`).

`fleet.Config` is a type alias of `kibanaoapi.Config` (`internal/clients/fleet/client.go:27`), and
`fleet.NewClient` is a thin wrapper around `kibanaoapi.NewClientWithLabel`
(`internal/clients/fleet/client.go:31-33`). The shared `transport.RoundTrip`
(`internal/clients/kibanaoapi/client.go:149-171`) already injects whatever `Config.Headers` it is given on
every request, for both the Kibana and Fleet clients, and the cross-origin-redirect guard added for Kibana
headers (`client.go:77-105`) activates whenever `Config.Headers` is non-empty — it is not Kibana-specific.
No transport-layer change is needed to deliver Fleet headers once `Config.Headers` is populated for the
Fleet client.

Today, `toFleetConfig()` (`internal/clients/config/kibana_oapi.go:228-232`) derives the Fleet config from
the effective Kibana config and explicitly sets `Headers = nil`. `newFleetConfigFromFramework`
(`internal/clients/config/fleet.go:33-74`) then applies the `fleet {}` block's own auth/endpoint/CA/
insecure overrides on top, but has no equivalent step for headers today. The reserved-header-name and
case-insensitive-duplicate validation that Kibana headers already go through lives entirely inline inside
`buildKibanaOapiConfigFromFramework` (`internal/clients/config/kibana_oapi.go:106-128`), keyed off the
package-level `reservedKibanaHeaderNames` map (`kibana_oapi.go:35-42`). There is no reusable function today
— only an inline loop coupled to the Kibana decoding path.

Per the human direction on the issue, this change proceeds with the research's recommended approach
(Approach A: a dedicated `fleet { headers }` attribute) and two explicit constraints beyond the research's
own recommendation: Fleet's header validation must reuse the exact same reserved-name list as Kibana's
(not a Fleet-specific list), and no same-host auto-forward convenience default (Approach B) is implemented,
even as a secondary behavior layered on top of Approach A. This is also authored as a fresh capability spec
rather than a delta to the existing `provider-kibana-connection` spec.

## Goals / Non-Goals

**Goals:**
- Let users attach custom HTTP headers to every outgoing Fleet API request via `fleet { headers = {...}
  } }`, independent of `kibana { headers }`.
- Reject Fleet's configured header names using the exact same reserved-name list and case-insensitive
  duplicate rule already enforced for Kibana headers, via one shared implementation rather than two
  independently-maintained lists.
- Keep `kibana.headers` and `fleet.headers` fully independent in both directions: configuring one never
  implicitly populates, suppresses, or overrides the other.
- Ship as a normal additive feature with no Kibana/Fleet/Stack minimum-version gate, validated by running
  the new acceptance coverage across the provider's standard supported-Stack-version test matrix.

**Non-Goals:**
- Any automatic forwarding of `kibana.headers` to Fleet, including a same-host/same-endpoint convenience
  default (research Approach B). Explicitly rejected by human direction, not just deferred.
- Any change to `kibana.headers`' own behavior, validation, or redirect protection, or to the existing
  `provider-kibana-connection` spec text (including its "Fleet does not gain its own `headers` attribute as
  part of this requirement" sentence — see "Relationship to the existing `provider-kibana-connection` spec"
  below).
- An entity-local `fleet_connection` block. None exists today; `fleet {}` is provider-level only.
- Any change to generated Fleet/Kibana client code or to `transport.RoundTrip`'s header-injection logic —
  it already handles an arbitrary `Config.Headers` map for both clients.

## Decisions

### 1. Add `headers` directly to `GetFleetFWConnectionBlock()`, not via `connectionBlockSpec`

Add a `headers` entry to the `Attributes` map inside `GetFleetFWConnectionBlock()`
(`internal/schema/connection.go:81-123`):

```go
attrHeaders: fwschema.MapAttribute{
    MarkdownDescription: "Custom HTTP headers to send on every request to Fleet.",
    Optional:            true,
    Sensitive:           true,
    ElementType:         types.StringType,
},
```

reusing the existing `attrHeaders` constant (`internal/schema/connection_constants.go`) for the attribute
name, matching the exact attribute contract `fwConnAttrFactory.mapAttr` already produces for `kibana.headers`
(`connection_spec.go:202-209`): `Optional`, `Sensitive`, `types.StringType` elements, no validators.

Why:
- `fleet {}` is hand-written and has exactly one call site (the top-level provider schema). Routing it
  through `connectionBlockSpec`/`connAttrFactory[T]` would pull in ephemeral- and action-resource variants
  that Fleet does not have and gain nothing, since there is no entity-local or ephemeral Fleet connection to
  generate for.
- Matching the Kibana attribute's exact contract (rather than inventing a new one) is what the human
  direction's "reuse the exact same reserved-name list" implies for the schema layer too: the same kind of
  attribute, validated the same way.

Alternatives considered:
- Lift `fleet {}` onto `connectionBlockSpec` as part of this change, then add `headers` once: rejected as
  disproportionate schema-generator churn for a single attribute, and out of scope for a change whose
  purpose is the `headers` attribute itself, not a Fleet schema-generation refactor.

### 2. Extract a shared reserved-header-name/case-collision validator; reuse it unchanged for Fleet

Factor the validation loop in `buildKibanaOapiConfigFromFramework`
(`internal/clients/config/kibana_oapi.go:106-128`) into a standalone function in the `config` package, e.g.:

```go
// validateConnectionHeaders rejects headers whose name matches a reserved,
// provider- or transport-controlled header (case-insensitively) and rejects
// multiple configured keys that differ only by case. entityLabel identifies
// the connection in diagnostics ("Kibana" or "Fleet"). It returns a
// whitespace-trimmed copy of headers with trimmed keys/values.
func validateConnectionHeaders(headers map[string]string, entityLabel string) (map[string]string, fwdiags.Diagnostics)
```

sharing the existing `reservedKibanaHeaderNames` map as-is (rename to a neutral name, e.g.
`reservedConnectionHeaderNames`, since it is no longer Kibana-specific) rather than introducing a second,
Fleet-specific list. `buildKibanaOapiConfigFromFramework` calls this helper for `kibana.headers` exactly as
it does today (same diagnostics, same `"Kibana"` label, same resulting behavior — this is a refactor, not a
behavior change for Kibana). `newFleetConfigFromFramework` (Decision 3) calls the same helper for
`fleet.headers` with `entityLabel: "Fleet"`, so a header named `kbn-xsrf`, `authorization`, `host`,
`content-length`, `transfer-encoding`, or `trailer` is rejected on `fleet.headers` for the same reason it is
rejected on `kibana.headers` today: Fleet requests already set `kbn-xsrf` and `Authorization` through the
same transport as Kibana requests, and the other four are Go-transport-controlled `http.Request` fields, not
ordinary header map entries, on both clients identically.

Why:
- Directly satisfies the human direction that Fleet reuse "the exact same reserved-name list" — a second,
  independently maintained list for Fleet would risk drifting from Kibana's over time, which the shared
  helper prevents structurally.
- Keeps the accepted-header policy in one function, matching the precedent set by the Kibana change's own
  design (Decision 2 of `kibana-connection-custom-headers/design.md`), now shared instead of duplicated.

Alternatives considered:
- Keep two separate inline loops (one for Kibana, a new copy for Fleet): rejected because it directly
  contradicts the human direction to reuse the exact same list, and because any future change to the
  reserved-name list would need to be applied in two places instead of one.

### 3. Decode and validate `fleet.headers` in `newFleetConfigFromFramework`, after `toFleetConfig()`

In `internal/clients/config/fleet.go`, inside `newFleetConfigFromFramework`, after the existing
`config := kibanaCfg.toFleetConfig()` call (which leaves `config.Headers` `nil`) and within the existing
`if len(cfg.Fleet) > 0 { fleetCfg := cfg.Fleet[0]; ... }` block, decode `fleetCfg.Headers` the same way
`kibConfig.Headers` is decoded today (`typeutils.IsKnown` guard, then `ElementsAs`), pass the result through
`validateConnectionHeaders(headers, "Fleet")`, and assign `config.Headers = validated` when non-empty.
Return early with the validator's diagnostics on error, matching how `buildKibanaOapiConfigFromFramework`
already returns early on invalid Kibana headers.

Because this runs entirely inside the `cfg.Fleet` block and only ever reads `fleetCfg.Headers`, it cannot be
influenced by, or influence, `kibana.headers` — `toFleetConfig()`'s existing clear of the inherited Kibana
`Headers` value happens first and is never consulted afterward for this field.

Why:
- Mirrors the existing decode-then-validate-then-assign shape already used for Kibana headers in
  `buildKibanaOapiConfigFromFramework`, so there is one pattern for "how a connection's `headers` map
  becomes `Config.Headers`" across both entry points.
- Placing it after `toFleetConfig()` and inside the `cfg.Fleet` block, rather than inside `toFleetConfig()`
  itself, keeps `toFleetConfig()` as a pure "derive Fleet defaults from Kibana, excluding Kibana's own
  headers" step, and keeps the new Fleet-specific decoding next to the rest of the `fleet {}` block's own
  overrides (auth, endpoint, CA certs, insecure) it already applies in the same function.

Alternatives considered:
- Decode `fleet.headers` earlier, inside `toFleetConfig()`: rejected because `toFleetConfig()` is a method
  on `kibanaOapiConfig` and has no access to the Fleet-specific `cfg.Fleet` block; threading it through would
  change that method's signature for a concern (Fleet's own schema attribute) that belongs with the rest of
  the Fleet-block decoding in `fleet.go`.

### 4. No same-host auto-forward (Approach B), as an explicit non-goal rather than a silent omission

The research comment raised, as an open question (see "Open questions" below), whether a same-host
auto-forward convenience default should be layered on top of the explicit `fleet.headers` attribute. The
human direction resolves this explicitly: "Do NOT do approach B as a convenience default." No endpoint-
comparison logic is added anywhere in the config pipeline. `kibana.headers` and `fleet.headers` remain two
independently configured attributes; a user who wants the same header value in both must set it in both
blocks (optionally by referencing the same Terraform variable or local in each).

Why:
- Matches the explicit human direction over the research's own tentative framing of Approach B as "worth
  asking about," not just its recommendation (which already favored Approach A alone).
- Avoids introducing new implicit, string-comparison-based security-relevant behavior (as the research's own
  "Cons" analysis of Approach B noted) for a convenience that was not requested.

### 5. Relationship to the existing `provider-kibana-connection` spec

The existing requirement "Fleet client excludes Kibana-only headers"
(`openspec/specs/provider-kibana-connection/spec.md`) states: "The Fleet client SHALL exclude custom headers
configured on the Kibana connection... Fleet does not gain its own `headers` attribute as part of this
requirement." That requirement's behavior — Kibana-only headers never reaching Fleet — is unchanged by this
change (Decision 4: no auto-forward). Its closing sentence scopes what *that specific requirement* grants:
it does not itself grant Fleet a `headers` attribute. This change adds Fleet's `headers` attribute through a
new, separate requirement in a new capability (`provider-fleet-connection`), not through that requirement,
so the sentence remains accurate in its own scope and is not amended. Per the human direction ("this should
be a fresh spec"), this change does not touch `openspec/specs/provider-kibana-connection/spec.md` or its
change-time delta at `openspec/changes/kibana-connection-custom-headers/`.

## Open questions

Copied verbatim from the issue's implementation-research comment, since that comment's `### Open questions`
section is the authoritative source for this section. The human direction on the issue resolves the first
three explicitly; see the inline resolution notes.

- Should Fleet's `headers` validation reuse the exact same reserved-name list (`authorization`,
  `content-length`, `host`, `kbn-xsrf`, `trailer`, `transfer-encoding`) as Kibana's, given Fleet calls also
  set `kbn-xsrf` internally? (Likely yes, but not yet confirmed against any Fleet-specific reserved header
  needs.)
  - **Resolved by human direction**: yes — Fleet reuses the exact same reserved-name list as Kibana (see
    Decision 2).
- Is there any appetite for *also* doing a non-blocking, same-host auto-forward (Approach B) as a
  convenience default, with the explicit `fleet.headers` attribute taking precedence? This wasn't scoped by
  the original design doc and would need its own proposal.
  - **Resolved by human direction**: no — Approach B is explicitly out of scope for this change (see
    Decision 4).
- Should this be delivered as a new OpenSpec change amending
  `openspec/changes/kibana-connection-custom-headers/specs/provider-kibana-connection/spec.md:149-160`
  (which currently mandates the exclusion), or as a fresh spec for Fleet connection headers?
  - **Resolved by human direction**: a fresh spec — this change introduces the new `provider-fleet-
    connection` capability instead of amending `provider-kibana-connection` (see Decision 5).
- Can `/api/fleet/*` be fronted by a hostname/proxy different from the rest of Kibana in any supported
  deployment? Elastic docs only ever show Fleet calls going to "the Kibana host URL" and never address a
  split-hostname topology either way — this is unverified in both directions and affects how strongly
  Approach A's "duplication" con should be weighted.
  - **Still open.** Not addressed by the human direction and not required to resolve this change: Approach
    A's `fleet.headers` attribute works whether or not Fleet and Kibana share a hostname, since it is
    configured independently of `kibana.headers` either way.

## Risks / Trade-offs

- **Users whose Fleet and Kibana sit behind the same proxy must configure the same header twice** (once in
  `kibana { headers }`, once in `fleet { headers }`) → Accepted per human direction, which explicitly rejects
  an automatic same-host default; users can reference the same Terraform variable/local in both blocks to
  avoid literal duplication.
- **A second, Fleet-facing `Sensitive: true` map attribute still leaves header values in Terraform state** →
  Same trade-off already accepted for `kibana.headers`; `fleet {}` has no action-resource or ephemeral
  variant, so there is no write-only or ephemeral alternative to offer here, unlike Kibana's connection
  block.
- **Extracting the reserved-header validator changes a function Kibana's header validation depends on** →
  Mitigate by keeping the extracted function's behavior for the Kibana call site identical (same diagnostics,
  same reserved-name map, same case-collision rule) and covering it with the existing Kibana header tests
  continuing to pass, not just new Fleet tests.
