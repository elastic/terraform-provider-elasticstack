## Context

`elasticstack_fleet_agent_download_source` (`internal/fleet/agentdownloadsource/`) currently maps `ssl.key`, `auth.password`, and `auth.api_key` only to their plain Fleet API wire fields (`ssl.key`, `auth.password`, `auth.api_key`). It never touches the parallel `secrets.ssl.key` / `secrets.auth.password` / `secrets.auth.api_key` fields on `POST /api/fleet/agent_download_sources` or `PUT /api/fleet/agent_download_sources/{sourceId}`, confirmed by direct inspection of `internal/fleet/agentdownloadsource/schema.go`, `models.go`, and `models_ssl.go` as of this change, and corroborated by the triggering kibana-api-gap scan (issue #5242).

An archived OpenSpec design doc (`openspec/changes/archive/2026-10-14-fleet-agent-download-source-ssl/design.md`, Decision 1a) claims `ssl.key` was already wired through `secrets.ssl.key`. That claim does not match the current code — `models_ssl.go`'s `kbapiSsl` struct still maps `Key` to the plain `ssl.key` wire field, and `schema.go`'s `ssl` validators are still strict (reject `ssl = {}`, empty strings). This proposal treats the archived doc as stale for grounding purposes and uses the current code plus today's fresh scan and the human decisions below as ground truth.

The generated client (`generated/kbapi/kibana.gen.go`) already exposes the fields needed: `PostFleetAgentDownloadSourcesJSONBody.Secrets.{Ssl.Key, Auth.{Password,ApiKey}}` and the matching `PutFleetAgentDownloadSourcesSourceidJSONBody.Secrets...`, each a two-arm `anyOf`: an `{id: string}` object (`...SecretsSslKey0`/`...SecretsAuthPassword0`/`...SecretsAuthApiKey0`) or a raw string (`...SecretsSslKey1`/`...SecretsAuthPassword1`/`...SecretsAuthApiKey1`). No client regeneration is required.

This resource also does not yet implement `resource.ResourceWithModifyPlan` (confirmed via `internal/fleet/agentdownloadsource/resource.go`'s interface assertions), and its Create/Update path cannot currently reach Terraform private state at all (see Decision 6) — both are required for the repo-mandated `writeonlyhash` drift-detection pattern this change adopts.

## Goals / Non-Goals

**Goals**
- Wire `ssl.key` to `secrets.ssl.key`, and `auth.password`/`auth.api_key` to `secrets.auth.password`/`secrets.auth.api_key`, via new write-only sibling attributes (`ssl.key_wo`, `auth.password_wo`, `auth.api_key_wo`).
- Detect drift on all three write-only attributes using `internal/utils/writeonlyhash`, per `coding-standards.md:61-63`, with no `_wo_version` companion attributes.
- Fix the `writeonlyhash` helper's bcrypt 72-byte input limit so it can safely hash PEM-sized values.
- Correctly avoid clearing a managed secret on updates that don't intend to touch it (GET-before-PUT id-reference logic), given the human-confirmed finding that omitting `secrets.*` while sending `ssl`/`auth` is unsafe.
- Keep existing version floors (`ssl` 9.1.0, `auth` 9.4.0) unchanged; add no new Terraform-side version gate for `secrets.ssl.key`.
- Consolidate `auth` credential-mode validation (plain and write-only forms) into one object-level validator.

**Non-Goals**
- Deprecating or removing the plain `ssl.key`/`auth.password`/`auth.api_key` attributes (research's rejected Approach B).
- Changing storage-gating behavior itself — Fleet-Server-version-based runtime gating of whether Kibana can store a secret at all (9.3.0 SSL / 9.4.0 auth) is server-side behavior this resource must tolerate (literal-resend fallback), not something the Terraform schema controls.
- Migrating the two other existing `_wo` resources (`security/user`, `elasticsearch_connector`... connectors already uses `writeonlyhash`) to any new pattern.
- Implementation itself — this change produces only the OpenSpec proposal/design/tasks/delta-spec artifacts.

## Decisions

### Decision 1 — `ssl.key` wiring target
Send the existing plain `ssl.key` value under `secrets.ssl.key` as a literal, rather than under plain `ssl.key`; `ssl.key_wo` also targets `secrets.ssl.key` and uses Decision 3's literal/id-reference logic. The two Terraform attributes are mutually exclusive because they map to one API field. This changes the wire mapping for existing `ssl.key` configurations, even though the schema change is additive. Because Fleet Server's ability to store the secret is gated at runtime (Fleet-Server-version-based, not stack version) and Kibana silently merges `secrets.ssl.key` into plaintext `ssl.key` when storage is unavailable, read hydration must distinguish configured plain values from write-only values and must never persist an echoed write-only literal in `ssl.key`.

### Decision 2 — Write-only mechanism: `writeonlyhash`, no `_wo_version`
Add `ssl.key_wo`, `auth.password_wo`, `auth.api_key_wo` as `WriteOnly` + `Sensitive` string attributes with no accompanying `_wo_version` attributes. Drift is detected by hashing the configured value with `internal/utils/writeonlyhash` and comparing against the hash stored in private state from the last successful apply, inside a new `ModifyPlan` method on `*agentdownloadsource.Resource`. **This supersedes the `_wo_version` design in the automated implementation-research comment** (see "Deviation from research recommendation" below).

### Decision 3 — Secret id-reference vs. literal resend
Before each PUT, GET the current resource by `source_id` and inspect `secrets.ssl`/`secrets.auth` on the response:
- If a write-only field's hash is unchanged since the last successful apply (per `writeonlyhash.Matches` against the stored private-state hash) **and** the GET returned a secret id for that field, send `{id: <existing-id>}` for that field in the PUT, preserving the managed secret without resending the literal.
- If the hash differs (first apply, or the configured value changed), send the literal value from configuration. A resent literal mints a new secret id server-side; Kibana does not validate caller-supplied ids in the `{id}` arm, so there is no risk of referencing the wrong secret as long as the id read in this same request cycle is used.
- If the GET returns no secret id for a field whose write-only attribute is configured (Fleet Server secret storage unavailable, so Kibana stored the value merged into plaintext), resend the literal from configuration rather than attempting an `{id}` reference, since there is no stored secret to reference.
- Never omit `secrets.*` for a field that has a configured write-only value while still sending the surrounding `ssl`/`auth` object — per the human-confirmed probe, that clears the existing secret. The literal-vs-`{id}` branching above is the only way to avoid an unwanted clear.
- Read write-only values from `req.Config`, not `req.Plan`, because the framework represents write-only values as null in the plan. If the API echoes a `key_wo` literal as plain `ssl.key`, do not copy it into Terraform state's stateful `ssl.key` sibling. Conversely, when a configured plain `ssl.key` is redacted by the API, preserve its configured value during post-write read and refresh. Preserve the `ssl` parent object for a configuration that contains only `key_wo`.
- When a configured write-only value is removed, a stored hash with a now-null configuration signals an intentional clear and forces an update, even when all ordinary attributes are unchanged. Keep the hash until the clear succeeds, then remove it; a later re-add has no prior hash and sends the literal.

### Decision 4 — Consolidated `auth` validator
Replace the current per-attribute `AlsoRequires`/`ConflictsWith` validators on `auth.username`/`password`/`api_key` with one object-level validator on `auth` enforcing: `username` and one of `password` or `password_wo` must either both be configured or both be absent; `password`/`password_wo` are mutually exclusive; `api_key`/`api_key_wo` are mutually exclusive; any configured member of the password family conflicts with any configured member of the api_key family. This replaces the per-attribute validators referenced in `schema.go:121-143` rather than layering a second validator set alongside them.

### Decision 5 — bcrypt 72-byte fix
Fix `internal/utils/writeonlyhash/hasher.go`'s `passwordBytes` (currently `salt + ':' + value`, `hasher.go:97-102`): when the salted value exceeds 72 bytes, compute SHA-256 over it and base64-encode the digest before handing it to bcrypt; values at or under 72 bytes are passed through unchanged, so existing stored hashes for values ≤72 bytes (e.g. `elasticsearch_connector`'s existing secrets) remain valid without rehashing. Add a unit test hashing a PEM-sized (>72 byte) value and asserting `Compute`/`Matches` round-trip correctly (today, `Compute` errors outright on such input via bcrypt's `ErrPasswordTooLong`, rather than silently truncating).

### Decision 6 — Threading private state through `KibanaResource[T]`
`internal/entitycore/kibana_resource_envelope.go`'s `KibanaWriteRequest[T]` has no `Private PrivateStateStorage` field (unlike `entitycore.WriteRequest[T]` in `internal/entitycore/resource_envelope.go:87-95`, used by `ElasticsearchResource[T]`, which already carries `Private`), and the envelope's write-adapter wiring discards the real `PrivateStateStorage` value via `_`. Additionally, this resource's `finalizeWrite` (`write.go:60`) always returns `SkipReadAfterWrite: true`, and the envelope's `postRead` hook — which does receive `Private` — is wired with `postReadRunsOnSkip: false`, so it never fires here. As currently structured, this resource's Create/Update path has no way to persist a `writeonlyhash` hash after a successful write.

Fix: extend `entitycore.KibanaWriteRequest[T]` with a `Private PrivateStateStorage` field and stop discarding it in the envelope's write-adapter wiring, mirroring the existing `ElasticsearchResource[T]`/`WriteRequest[T]` pattern exactly. This is a small, additive, backward-compatible change to the shared envelope (existing `KibanaWriteFunc[T]` callbacks that ignore the new field are unaffected). `writeAgentDownloadSource` (`write.go:31-42`) reads write-only literals from `req.Config` (not the null write-only fields in `req.Plan`) and then stores each configured value's hash via `req.Private.SetKey` after a successful create/update. For a removal, it retains the old hash while the clearing request is pending and clears the key only after success. `ModifyPlan` itself is a new plain method on `*agentdownloadsource.Resource` (not an envelope feature) that reads/compares hashes via the framework-provided `req.Private`/`resp.Private` directly, mirroring `internal/elasticsearch/connector/resource/modify_plan.go`'s `(*contentConnectorResource).ModifyPlan`. `Delete` is overridden on `*agentdownloadsource.Resource` to call the embedded `KibanaResource[T].Delete` and then clear the three hash keys from `resp.Private`, mirroring `internal/elasticsearch/connector/resource/resource.go:54-61`'s `(*contentConnectorResource).Delete` override. Raw bcrypt hash bytes are JSON-wrapped before `SetKey`/after `GetKey` (containing `$`, not valid raw JSON), following `internal/elasticsearch/connector/resource/secret.go`'s `encodeSecretHashForPrivateState`/`decodeSecretHashFromPrivateState` pattern.

### Decision 7 — Version floors unchanged
No new Terraform-side minimum-version gate for `secrets.ssl.key`/`secrets.auth.*`. Keep the existing `ssl` (9.1.0) and `auth` (9.4.0) gates as the only Terraform-side floors; storage-gating itself is a Fleet-Server-version-based runtime behavior this resource tolerates via the literal-resend fallback (Decision 3), not something a Terraform version requirement can usefully express.

### Decision 8 — Terraform CLI 1.11 minimum
Document the Terraform CLI 1.11 minimum required for write-only arguments, reusing the existing note from the provider's other `_wo` resources, in both `design.md`'s own impact note and the eventual resource documentation (tracked as a follow-up task, not part of this proposal-only change's deliverable).

## Deviation from research recommendation

The automated implementation-research comment (round 5/5, outcome `research-needs-human`) recommended "Approach A": `ssl.key` → `secrets.ssl.key` wiring plus `_wo`/`_wo_version` sibling attributes modeled on the `security/user` and `elasticsearch_connector` precedents. Its own round-5 critic finding, under "Open questions," flagged that this `_wo_version` design conflicts with the mandatory `writeonlyhash` standard in `coding-standards.md:61-63` and called this a blocking, unresolved question.

A subsequent human decision (issue comment, 2026-10-15T22:25:28Z, "Decisions on the open questions from the implementation research") resolves this explicitly in favor of `writeonlyhash`: **"This supersedes the `_wo_version` design in the research."** Concretely: no `_wo_version` attributes are added; drift detection uses `writeonlyhash` hash-in-private-state (Decision 2); updates reference the existing secret by `{id}` when the hash is unchanged and resend the literal otherwise (Decision 3); the 72-byte bcrypt limit is fixed rather than worked around (Decision 5). This proposal follows the human decision, not the research's original `_wo_version` recommendation, for this specific mechanism. The broader shape of Approach A that the human decision does not contradict — wiring `ssl.key` to `secrets.ssl.key`, adding `_wo` siblings at all (vs. the rejected Approach B), and a consolidated auth validator — is retained from the research as-is.

The human decision also corrects two factual claims the research had left as open probes: a PUT that omits `secrets.*` while still sending `ssl`/`auth` clears the existing secret (invalidating the research's originally-sketched conditional-omission write path at acceptance-test-outline point 5), and bcrypt errors rather than silently truncates inputs over 72 bytes (sharpening the research's drift-detection concern into a must-fix blocker, not a latent risk).

## Risks / Trade-offs

- **GET-before-PUT adds a request per write.** Accepted: correctness (avoiding an accidental secret clear) outweighs one extra GET per Create/Update.
- **Shared envelope change (`KibanaWriteRequest[T]`) touches code used by other Kibana-backed resources.** Mitigated by making the new field purely additive (zero value `nil` for existing callbacks that don't read it) and mirroring an already-proven pattern from `ElasticsearchResource[T]`.
- **bcrypt helper change affects the one existing consumer (`elasticsearch_connector`).** Mitigated by the ≤72-byte passthrough guarantee in Decision 5, verified by a unit test, so no existing stored hash is invalidated.
- **Resending a literal mints a new secret id every apply where the hash last differed.** This is existing, human-confirmed Kibana/Fleet behavior, not something this resource can avoid; it is only a concern if something external depends on secret-id stability, which is out of scope for this resource's contract.

## Open questions

- Does Fleet clear only the managed `secrets.ssl.key` secret when a PUT omits `key`, and does the
  same hold for the write-only path when `key_wo`/`key_wo_version` are removed entirely? Open in the
  archived design (`design.md:282-287`), not re-probed this round. *(blocking)*
- Does omitting `secrets.auth.password`/`secrets.auth.api_key`/`secrets.ssl.key` from a PUT (while
  still sending `username`/`headers`/plain `password`/`api_key`/`certificate`/
  `certificate_authorities` unconditionally) leave Fleet's existing managed credential untouched, or
  does Fleet treat the omission as "clear"? This must be answered with a documented live-stack probe
  of the exact PUT payload before the conditional-omission write path (Approach A, point 4) ships —
  if the probe shows the secret is cleared, the design needs an alternative (e.g. resending the
  secret unconditionally, like connectors does, rather than gating on `_wo_version`). New to the
  `_wo` design, not previously probed. *(blocking)*
- Is there a confirmed minimum Kibana/Fleet stack version for `secrets.ssl.key` specifically,
  narrower than the existing 9.1.0 `ssl`-object gate? Neither the OAS nor Elastic's documentation
  settles this; acceptance test 11 is the proposed way to confirm it. *(blocking)*
- How should the `auth` nested validators be restructured so `username` accepts either `password` or
  `password_wo`, and the two secret pairs stay mutually exclusive across both forms? No existing
  repo validator expresses "requires one of two alternatives"; needs a small custom validator.
  *(blocking)*
- What is the minimum Terraform CLI version required for write-only arguments? Could not be
  verified against HashiCorp's docs from this environment; an already-accepted constraint for the
  two existing `_wo` resources regardless. *(non-blocking)*
- **This research's `_wo_version`-sibling design conflicts with a mandatory, newer repo-wide
  standard and must be reworked before implementation.**
  `dev-docs/high-level/coding-standards.md:61-63` states: *"Resources exposing `WriteOnly` +
  `Sensitive` attributes (passwords, tokens, API secrets) MUST detect drift using the shared
  `internal/utils/writeonlyhash` helper. The helper stores bcrypt hashes of applied values in
  resource private state, surfacing in-config edits as plan-time updates without revealing the
  value."* `dev-docs/high-level/writeonly-secret-hashing.md:5-9` explains the rationale is to avoid
  depending on "user discipline" with a practitioner-supplied version string. Only
  `internal/elasticsearch/connector/resource/secret.go` currently uses this helper
  (`internal/utils/writeonlyhash/hasher.go:54,64,75,86`); neither of this draft's two reference
  resources (`security/user`'s `password_wo_version`, `connectors`' `secrets_wo_version`) has been
  migrated to it, and the writeonly-secret-hashing doc only discusses the connectors case — it does
  not even mention security/user. Point 4 of Approach A (and the entire `_wo_version` schema/write
  design in points 4-5) needs to be redesigned against the hasher-in-private-state contract instead
  of a companion `_wo_version` state attribute: `ssl.key_wo`/`auth.password_wo`/`auth.api_key_wo`
  would each need a `Compute`/`Matches` round-trip against resource private state rather than a
  stored `_wo_version` string, and the schema would drop the three `_wo_version` attributes
  entirely. This is a substantial redesign that this round did not have time to work through (the
  mechanics of wiring private state through this resource's read-derived `write.go`/`read.go`
  lifecycle, parallel to the existing `_wo_version` carry-forward problem already identified in
  Problem framing, are unresearched). *(blocking — supersedes the `_wo_version` design in its
  current form)*
- `internal/utils/writeonlyhash/hasher.go:97-102`'s `passwordBytes` helper concatenates
  `salt + ':' + value` and passes the result straight to `bcrypt.GenerateFromPassword`
  (`:64-65`), which silently truncates its input at 72 bytes. A PEM-encoded private key (the
  expected shape of `ssl.key_wo`) is almost always far longer than 72 bytes, so two different keys
  sharing the same first ~72 bytes after the salt prefix would hash identically and falsely report
  "no drift." This needs to be resolved (e.g. pre-hashing the value with SHA-256 before handing it
  to the bcrypt-based helper) before `writeonlyhash` can be safely reused for `ssl.key_wo`; it may
  be a pre-existing latent issue in the helper itself, independent of this feature. *(blocking)*

Note: the second-to-last and last items above are addressed by Decisions 2 and 5 respectively. The remaining items (managed-secret-clearing, auth-secrets-omission-safety, secrets-ssl-key-version-floor, auth-wo-validator-design, and the non-blocking Terraform-CLI-version item) are addressed by Decisions 3, 4, 7, and 8 respectively, per the human decision comment, which states "All seven [open questions] are resolved by the above."
