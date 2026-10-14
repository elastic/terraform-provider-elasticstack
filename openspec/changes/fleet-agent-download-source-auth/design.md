## Context

`elasticstack_fleet_agent_download_source` manages Fleet's "agent binary download source" settings via the generated Kibana client, but the resource schema never maps Kibana's request-body `auth` object on either `POST /api/fleet/agent_download_sources` or `PUT /api/fleet/agent_download_sources/{sourceId}`. Per Kibana's Fleet settings docs, `auth` lets Elastic Agents authenticate to a private/self-hosted artifact registry when downloading binaries, via one of: no auth (default), HTTP Basic (`username`+`password`), or API key (`api_key`), plus optional custom HTTP headers sent on every request.

The OpenAPI schema for `auth` (identical on POST and PUT) is a nullable, non-required object with four optional, flat fields: `api_key` (string), `username` (string), `password` (string), `headers` (array, up to 100 `{key, value}` pairs). There is no discriminator; which "mode" is active is implied by which fields are set. The generated Go client (`generated/kbapi/kibana.gen.go`) already exposes `Auth` on `PostFleetAgentDownloadSourcesJSONBody`, `PutFleetAgentDownloadSourcesSourceidJSONBody`, and the response item type — no `kbapi` regeneration is needed, only provider-side schema/model work.

Separately, the spec also defines `secrets.auth` (`api_key`/`password` as a `{id}` secret-store reference) on both request and response schemas, with no `writeOnly`/`readOnly` annotation (confirmed by `grep -n writeOnly generated/kbapi/oas.yaml` returning zero matches). This is a Kibana secret-store-reference mechanism, not a Terraform-style write-only attribute, and is out of scope for this change (see "Out of scope"). There is also a top-level `ssl` object (unrelated TLS settings), also out of scope.

This document's recommendation and structure come from the issue's [implementation research comment](https://github.com/elastic/terraform-provider-elasticstack/issues/5241#issuecomment-6159099656) (Approach A). That research left two questions blocking before implementation could start; both were subsequently resolved by live-Kibana verification posted as a [follow-up human comment](https://github.com/elastic/terraform-provider-elasticstack/issues/5241#issuecomment-6159427095) from `@tobio` against a local Kibana 9.4.0 stack. See "Open questions" below for the verbatim research questions and "Resolution of open questions" for the empirical answers that unblock implementation.

## Goals / Non-Goals

**Goals:**
- Expose `auth.username`, `auth.password`, `auth.api_key`, and `auth.headers` as an optional nested Terraform attribute, matching the OAS shape 1:1.
- Enforce the same credential-mode exclusivity Kibana's server already enforces (`username`+`password` paired; `password`/`api_key` mutually exclusive), without blocking a headers-only `auth` block.
- Correctly preserve just-written secrets across the mandatory read-after-write (`finalizeWrite`), since Kibana never echoes `password`/`api_key` back.
- Correctly clear `auth` on update when the block is removed from configuration, given Kibana's merge-on-omit / replace-on-`{}` PUT semantics.
- Gate `auth` behind the Kibana version where dedicated auth modes are GA (9.4.0), without raising the resource's existing 8.13.0 baseline for configurations that omit `auth`.
- Give import a correct, explicit story for the two fields the API will never echo back (`password`, `api_key`).

**Non-Goals:**
- `secrets.auth` secret-store `{id}` reference support (Approach B from the research comment) — deferred as a possible follow-up. This change only supports plaintext values in Terraform config/state (mitigated by `Sensitive: true`, which hides the value from CLI output and plans but does not encrypt it in state — the same tradeoff as every other `Sensitive` password/token field already in this provider, e.g. `output.kafka.password`).
- The top-level `ssl` object (CA, client cert, client key) and `secrets.ssl.key` — unrelated TLS configuration, not flagged in the issue.
- Any change to `generated/kbapi` — the client already exposes everything needed.

## Decisions

### Decision 1: Plain nested `auth` attribute mirroring the `output` resource's `kafka` block

Add an optional `auth` `SingleNestedAttribute` to `internal/fleet/agentdownloadsource/schema.go`:

```
auth = optional, single nested object {
  username = optional string
  password = optional string, Sensitive
  api_key  = optional string, Sensitive
  headers  = optional list of { key = required string, value = required string }, maxItems 100
}
```

This mirrors `internal/fleet/output/schema.go`'s existing `kafka.username`/`kafka.password`/`kafka.headers` shape (plain `username`, `Sensitive` `password`, `ListNestedAttribute` `headers` with required `key`/`value`). `toAPICreateModel`/`toAPIUpdateModel` (`internal/fleet/agentdownloadsource/models.go`) set `kbapi...JSONRequestBody.Auth` directly from these fields when the block is present, and leave it nil when absent — matching how `host`/`name` are already built today.

**Why:** Smallest, most idiomatic change; reuses an established repo pattern (kafka's nested shape and `Sensitive` convention) with no new abstractions, and matches the OAS shape 1:1. Considered and rejected: a `secrets.auth`-aware union (Approach B in the research comment) — the docs and OAS give no indication typical users need secret-store references rather than literal credentials; that is a natural v2 extension, not something blocking this gap fix.

### Decision 2: Credential-mode validators reuse the relative-path pattern from `es_connection_shared.go`

Apply `stringvalidator.AlsoRequires` between `auth.username` and `auth.password` (each requires the other, relative to the `auth` object), and `stringvalidator.ConflictsWith` between `auth.password` and `auth.api_key`, matching `internal/schema/es_connection_shared.go:51,59,68,78`. `headers` carries no validator tying it to the other three fields, so a headers-only `auth` block (no credentials at all) remains valid, per the docs and per live-Kibana confirmation (see "Resolution of open questions").

**Why:** `ExactlyOneOf` was considered and rejected: it would wrongly reject both a headers-only `auth` block and `auth = {}`, which the docs and the live-Kibana probe both confirm are valid.

### Decision 3: Update sends `auth: {}` on removal, never bare omission

`toAPIUpdateModel` SHALL send an explicit empty `Auth` struct (`"auth":{}`) when the planned `auth` block is null/absent but the prior state had `auth` configured. When the planned `auth` block is non-null, the full `Auth` struct is populated as usual. When `auth` was already absent in both prior state and plan, the field is left nil (omitted) — there is nothing to clear and omission avoids an unnecessary wire write.

**Why:** Confirmed by live probing (see "Resolution of open questions"): omitting the `auth` key on PUT is a **merge** — Kibana leaves existing credentials untouched — while sending `auth: {}` is a **replace** that clears them. A nil pointer would therefore silently leave stale credentials in place when a practitioner removes the `auth` block from configuration. The generated typed struct's `omitempty` pointer sub-fields marshal an empty struct to `{}` correctly, so the raw-body (`WithBody`) escape hatch considered in the research comment is not needed.

### Decision 4: Read hydration decodes `username`/`headers` verbatim; preserves `password`/`api_key` via the `kafka.password` nil-pointer pattern

Extend `readAndHydrateState` (`internal/fleet/agentdownloadsource/read.go`) to:
- Decode `auth.username` directly from the response (`types.StringPointerValue`).
- Decode `auth.headers` rebuilt verbatim from the response list, null if the API returns no headers, matching `fromAPIKafkaModel`'s headers handling (`internal/fleet/output/models_kafka.go:584-598`) — and keep an explicit empty list (`headers = []`) distinct from a null/absent list, since the live probe confirmed `headers: []` round-trips as `[]`, not as omitted.
- Apply `kafka.password`'s exact nil-pointer preservation condition (`internal/fleet/output/models_kafka.go:529-544,571-581`) to `auth.password` and `auth.api_key` independently: when the API response's pointer for that field is `nil` (omitted — confirmed to always be the case on every read, since Kibana moves the value into its secret store and returns no placeholder), fall back to a caller-supplied prior value; when the API returns a non-nil value, use it.

This requires a signature change to `readAndHydrateState` and its two callers, since there is no existing "fallback auth" plumbing for this resource (unlike kafka's output path, which already threads `prior`):
- The write path (`finalizeWrite` in `write.go`, called after both Create and Update) passes the **planned** `auth` as fallback, so a just-rotated password or a brand-new credential set is not immediately nulled out by the mandatory post-write read.
- The plain refresh/`Read` path (`readAgentDownloadSource` in `read.go`, which already threads `prior`) passes the **prior-state** `auth` as fallback, exactly mirroring kafka's `configuredPassword` capture.

**Why:** Confirmed by live probing: Kibana omits `password`/`api_key` from every read response (POST, GET, list, PUT), always, with no masked placeholder — so the `kafka.password` nil-pointer check applies unmodified, no placeholder-detection branch is needed. `username` and `headers` come back verbatim and need no preservation logic.

### Decision 5: Import leaves `password`/`api_key` null, verified with a scoped `ImportStateVerifyIgnore` + custom `ImportStateCheck`

`SpaceImporter.ImportState`/`KibanaSpaceImporter.SeedState` seed only identity and `space_ids` into state on import — there is no prior-state secret to fall back to on the first post-import read, for any imported resource, regardless of whether it was originally created through Terraform or the Kibana UI. So `auth.password`/`auth.api_key` hydrate to null on import whenever the API redacts them (which the live probe confirms is always). Acceptance tests covering import of a source with server-side-only credentials use `ImportStateVerifyIgnore` scoped to exactly `auth.password` and `auth.api_key` (not the whole `auth` block, so `headers`/`username` are still verified), paired with a custom `ImportStateCheck` asserting those two fields are explicitly null — following the precedent in `internal/apm/source_map/acc_test.go:159-167,210` (scoped ignore + custom check for fields the API won't return) and `internal/elasticsearch/security/user/acc_test.go:293-304` (`TestAccImportedUserDoesNotResetPassword`, asserting a secret field is absent post-import via direct `*terraform.InstanceState` inspection).

**Why:** A plain, unscoped `ImportStateVerify` would fail by comparing imported state (secrets null) against Terraform-created state (secrets populated). Scoping the ignore to just the two never-echoed fields keeps ordinary create/update/rotation assertions strict everywhere else.

### Decision 6: Per-field 9.4.0 version gate, resource baseline unchanged

Add a version requirement in `internal/fleet/agentdownloadsource/models.go`'s `GetVersionRequirements` that fires only when `auth` is non-null in the plan, pinned to Kibana 9.4.0 (dedicated Basic/API-key auth modes are GA since Elastic Stack 9.4 per Fleet settings docs), following the `addMinimum`/`NewAttributeVersionRequirement` pattern in `internal/fleet/agentpolicy/version_requirements.go:32-125`. `headers`, part of the same `auth` object in the OAS, shares the same 9.4.0 gate for simplicity (no earlier GA date specific to headers was found). The resource's existing 8.13.0 baseline (`agentDownloadSourceMinVersion`) is untouched for configurations that omit `auth` entirely.

**Why:** Keeps the version-gating granular and consistent with how other optional, newer fields are gated elsewhere in the Fleet domain (`agentpolicy`), rather than raising the whole resource's minimum version for users who don't need `auth`.

## Open questions

Copied verbatim from the issue's implementation research comment (now resolved — see "Resolution of open questions" immediately below):

- Does Kibana echo `auth.password` / `auth.api_key` back in GET/POST/PUT responses with the real value, omit the field, or return a masked placeholder? The OAS response schema is structurally identical to the request and has no redaction marker on these bare `string` fields (confirmed by inspecting the spec directly), so this cannot be resolved from the spec alone — it requires an empirical check against a running Kibana (see Testable section). **Blocking** for whether the `kafka.password`-style `nil`-pointer preservation check applies as-is, needs placeholder detection instead, and for `ImportStateVerify` expectations.
- What are PUT's merge-vs-replace semantics for `auth`? The generated client can only *omit* the `auth` key (via a nil pointer, `json:"auth,omitempty"`) or send `auth: {}` — it cannot send an explicit JSON `null` even though the OAS marks `auth` nullable. Confirming via a live-API probe whether omission clears a previously-set `auth` (replace semantics) or leaves it untouched (merge semantics), and whether `auth: {}` behaves differently from omission, is **blocking** for `toAPIUpdateModel` and for the `auth_remove` acceptance test.
- Confirmed decided (no longer open): `username`+`password` use `AlsoRequires` and `password`/`api_key` use `ConflictsWith`, matching `es_connection_shared.go:51,59,68,78` — `headers` remains independent of both, so a headers-only `auth` block stays valid.

## Resolution of open questions

Both blocking open questions above were resolved by a [follow-up comment](https://github.com/elastic/terraform-provider-elasticstack/issues/5241#issuecomment-6159427095) from `@tobio`, who probed a local Kibana 9.4.0 stack directly (`/api/fleet/agent_download_sources`, POST/GET/list/PUT/DELETE) after the research comment was posted. These findings are empirical, not spec-derived, and are the basis for Decisions 3 and 4 above:

- **Redaction on read:** Kibana omits `auth.password`/`auth.api_key` on every read (POST/GET/PUT/list responses), moving the secret into Fleet's internal secret store and returning a reference under `secrets.auth.{password,api_key}.id` instead. There is no masked placeholder. `username` and `headers` come back verbatim. The secret `id` changes on every PUT that sends a credential, so `secrets` is not usable for drift detection and stays unmapped in this resource. The `kafka.password`-style nil-pointer preservation pattern applies as proposed, with no placeholder-detection branch needed.
- **PUT merge vs. replace:** omitting the `auth` key leaves existing auth unchanged (merge); sending `auth: {}` or `auth: null` clears it (replace); sending a new credential (e.g. `api_key` after previously-set basic auth) replaces the full credential set, not just the changed sub-field; a headers-only `auth` body replaces everything, leaving only headers. `toAPIUpdateModel` must therefore send `"auth":{}` — not a nil pointer — when the Terraform `auth` block is removed.
- **Validators match Kibana's server-side rules**, confirming Decision 2 directly against the live API: `username`+`password`+`api_key` together returns 400 (`Cannot specify both username/password and api_key authentication`); `username` alone returns 400 (`Username and password must be provided together`); a headers-only `auth` block is accepted.
- **Headers:** `headers: []` is accepted and round-trips as `[]` (not omitted), so null and empty list must be kept distinct in the read-mapping (Decision 4). Up to 100 headers are accepted; 101 is rejected with HTTP 400 (`array size is [101], but cannot be greater than [100]`), matching the OAS `maxItems: 100` and giving the acceptance-test boundary pair (`auth_headers_max` / `auth_headers_over_max`) concrete expected behavior.

## Out of scope

- The top-level `ssl` object (CA, client cert, client key) and the `secrets.ssl.key` reference — separate TLS configuration, unrelated to this `auth` gap, and not flagged in this issue.
- `secrets.auth` secret-store `{id}` reference support (Approach B in the research comment) — deferred as a possible follow-up.
- Any change to `generated/kbapi` — the client already exposes `Auth` on both request bodies and the response item; no regeneration is required.
