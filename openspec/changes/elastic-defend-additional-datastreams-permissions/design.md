## Context

`elasticstack_fleet_elastic_defend_integration_policy` (`internal/fleet/elastic_defend_integration_policy/`) never maps `additional_datastreams_permissions` on the Defend bootstrap create (`buildBootstrapRequest`, `common.go:411`) or finalize/update (`buildFinalizeRequest`, `common.go:468`) requests, nor reads it back in `populateModelFromAPI` (`common.go:45`). The generated client already exposes `AdditionalDatastreamsPermissions *[]string` on `kbapi.PackagePolicyCreateRequestTypedInputs`, `kbapi.PackagePolicyUpdateRequestTypedInputs`, and the `kbapi.PackagePolicy` response type, so the gap is entirely in this resource's schema and request-mapping code.

`elasticstack_fleet_integration_policy` already implements this field (`internal/fleet/integration_policy/schema.go:115-122`, `models.go:53,73-79,177-183,290,354-370`, `resource.go:45-49`, `capabilities.go:27-49`), gated at Kibana 9.1.0 via `MinVersionAdditionalDatastreamsPermissions`, resolved through a per-resource `integrationPolicyFeatures` capability struct threaded into `toAPIModel`. Its revoke-on-unset behavior sends an empty array when the attribute is null and the server supports the field, because neither its `Create` nor `Update` passes prior state into the request builder, so "unset" must mean "revoke."

The Defend resource's `models.go:51-54` currently has:

```go
func (model elasticDefendIntegrationPolicyModel) GetVersionRequirements(_ context.Context) ([]entitycore.VersionRequirement, diag.Diagnostics) {
    if !typeutils.IsKnown(model.AgentPolicyIDs) {
        return nil, nil
    }
    return []entitycore.VersionRequirement{...agent_policy_ids requirement...}, nil
}
```

This early-return means any second, independent version requirement added after it would be skipped whenever `agent_policy_ids` happens to be unknown. This must be restructured to accumulate requirements before adding a second attribute-scoped requirement for `additional_datastreams_permissions`.

## Goals / Non-Goals

**Goals**
- Add `additional_datastreams_permissions` to the Defend resource, mapped on both `buildBootstrapRequest` and `buildFinalizeRequest`, and read back in `populateModelFromAPI`.
- Gate the attribute at Kibana >= 9.1.0 via an attribute-scoped `entitycore.VersionRequirement`, evaluated only when the value is known and non-null.
- Restructure `GetVersionRequirements` so the `agent_policy_ids` and `additional_datastreams_permissions` requirements are both evaluated regardless of which attribute(s) are configured.
- Make the attribute `Optional` + `Computed` with no revoke-on-unset: an unset attribute never sends anything, and an explicit `[]` is the supported way to clear permissions.

**Non-Goals**
- Any change to `elasticstack_fleet_integration_policy` or `elasticstack_fleet_managed_integration`, which already implement this field (with different, unchanged semantics).
- Regenerating `generated/kbapi` — the field already exists in the generated client.
- A shared cross-resource package-policy mapping helper (considered and rejected by the research as Approach B, given the three resources' differing lifecycle shapes; not revisited here).
- Client-side validation of the `<type>-<suffix>` datastream-prefix grammar; the server already rejects malformed values with a 400, and that regex has changed across Kibana patch versions (see Kibana PR #282404 / the 9.5.1 breaking-change notes), so duplicating it client-side would drift.
- Implementation itself — this change produces only the OpenSpec proposal/design/tasks/delta-spec artifacts.

## Decisions

### Decision 1 — Optional + Computed, no revoke-on-unset

Unlike `integration_policy`'s plain `Optional` attribute with `SizeAtLeast(1)` (which treats "remove from config" as the only way to signal intent, and revokes via an explicit `[]` when unset against a supporting server), the Defend attribute is `Optional` + `Computed` with a `UseStateForUnknown` plan modifier, and allows an explicit empty list:

- Known value (including `[]`): send `*[]string` of that value in both `buildBootstrapRequest` and `buildFinalizeRequest`.
- Null or unknown value: omit the field entirely (`nil` pointer) — regardless of server version. There is no "send empty array to revoke" branch and therefore no `elasticDefendFeatures`/capability-resolution struct is needed for this field, unlike `integration_policy`'s `resolveIntegrationPolicyFeatures`.
- A practitioner who wants to clear previously granted permissions sets `additional_datastreams_permissions = []` explicitly; removing the attribute from configuration leaves server-side permissions untouched.

### Decision 2 — Read-back mapping

In `populateModelFromAPI`:
- API returns a non-empty list: map to a known list value via `types.ListValueFrom`.
- API returns an absent or empty list: map to the modeled attribute's existing value when state/plan already holds a known, non-null (possibly empty) list — i.e. an explicit `[]` round-trips as `[]`, not null. Otherwise (the attribute was never configured) map to `types.ListNull(types.StringType)`, matching the `Computed` plan-modifier contract (no manufactured non-null default when nothing was ever configured or returned).

### Decision 3 — Version gate restructuring

Add `MinVersionAdditionalDatastreamsPermissions = version.Must(version.NewVersion("9.1.0"))` to `resource.go` alongside the existing `MinVersionPolicyIDs`. Restructure `elasticDefendIntegrationPolicyModel.GetVersionRequirements` to build a `[]entitycore.VersionRequirement` slice incrementally (appending the `agent_policy_ids` requirement when that attribute is known, and the `additional_datastreams_permissions` requirement when that attribute is known and non-null) instead of returning early, so neither requirement is silently skipped because of the other's state. This is the `entitycore.AppendVersionRequirementIf`-style accumulation pattern already available in `internal/entitycore/version_requirements.go`.

9.1.0 is reused directly from `integration_policy.MinVersionAdditionalDatastreamsPermissions` rather than independently re-derived, per the human decision below.

### Decision 4 — No client-side grammar validation

Rely on the server to reject malformed datastream-prefix values (e.g. `foo`, `logs*`). The server's validation regex has changed across Kibana patch versions (9.5.1 backport), so a client-side `stringvalidator.RegexMatches` would need to track that drift. This matches `integration_policy`'s existing behavior (no grammar validator beyond `SizeAtLeast(1)`, which this resource does not even adopt per Decision 1).

## Deviation from research recommendation

The automated implementation-research comment (round 3/5, outcome `research-needs-human`, score 96/100) recommended "Approach A": mirror `integration_policy` field-for-field, including its revoke-on-unset semantics (send `[]` when the attribute is unset against a 9.1.0+ server) and an `Optional` (non-`Computed`) attribute rejecting an explicit `[]` via `listvalidator.SizeAtLeast(1)`.

A subsequent human review comment (2026-10-16T05:31:16Z, "Human review: open questions resolved and implementation decision") resolves the research's open questions and **explicitly supersedes that one piece of the recommendation**: *"Decision: proceed, with the attribute Optional + Computed (no revoke-on-unset). This supersedes Approach A's revoke-on-unset semantics; the rest of the research stands."* This design follows that decision: Decision 1 above replaces the research's revoke-on-unset request-building and its `SizeAtLeast(1)` schema validator with an Optional+Computed attribute that allows an explicit empty list as the clearing mechanism. The remainder of the research's recommendation — reusing the 9.1.0 version floor, relying on server-side grammar validation, and the general shape of mapping the field into `buildBootstrapRequest`/`buildFinalizeRequest`/`populateModelFromAPI` — is retained.

The human comment also resolves the research's three open questions directly:

## Open questions

- **`min-version` (blocking) — resolved: 9.1.0 is correct.** `additional_datastreams_permissions` first appears in Kibana's `common/types/models/package_policy.ts` on the `9.1` branch, absent on `9.0` and `8.18`. It is on the shared package-policy schema, so the Defend path is covered. `package_policies_to_agent_permissions.ts` applies it to every package including `endpoint`. Reuse the 9.1.0 floor from `integration_policy` (source-history evidence; no live 9.0 stack was probed).
- **`client-side-grammar-validation` — resolved: no, rely on the server.** The server accepts `logs-*`, `traces-x-*`, `synthetics-x-*`, `profiles-x`; it rejects `foo`, `foo-bar`, `logs*`, `*-x`, `profiling-x`. Kibana PR #282404 (backported to 9.3/9.4/9.5) changed the regex, so enforcement differs by patch version; duplicating it client-side would drift. A server 400 is clear enough. Same as `integration_policy`.
- **`migration-communication` — resolved: largely moot with the Optional+Computed decision.** Probed update semantics against a live Kibana 9.5.5 stack: omitted -> value retained; `[]` -> stored `[]`; `null` -> cleared. With Optional+Computed and no revoke-on-unset, an unset attribute never sends anything, so existing Defend policies keep their out-of-band permissions on the first post-upgrade apply — no migration hazard. A short CHANGELOG entry (new attribute, 9.1.0+) is sufficient.

## Risks / Trade-offs

- **Divergence from `integration_policy`'s pattern.** The Defend resource intentionally does not mirror `integration_policy`'s revoke-on-unset / reject-empty-list behavior for this one field. This is a deliberate, human-confirmed choice (see Deviation section) to avoid a migration hazard specific to Defend policies that may already carry out-of-band permissions; it does mean the two sibling resources now have different semantics for an identically-named attribute, which should be called out in docs for each resource.
- **No capability/feature-resolution struct needed for this field**, unlike `integration_policy`. This keeps the Defend implementation smaller, but means a future fourth consumer of this field with yet another unset-semantics requirement cannot reuse either existing pattern verbatim — acceptable, since no such consumer is in scope.
- **Version floor reused without independent re-confirmation against the Endpoint/Defend-specific code path** (the research's original open question). The human decision accepts the source-history evidence as sufficient and does not require a live pre-9.1.0 probe before shipping; the acceptance test plan below still exercises the version gate directly.

## Test plan (adjusted from the research, per the human decision)

- **Acceptance**, mirroring `TestAccResourceIntegrationPolicyAdditionalDatastreamsPermissions`, with a server-side check via `fleetclient.GetDefendPackagePolicy`: create with one value, import + `ImportStateVerify`, update to two values, **remove the attribute from config and assert the server-side value is retained** (not cleared), then set `additional_datastreams_permissions = []` and assert it is cleared server-side.
- **Version gate**: a separate acceptance test against a stack below 9.1.0 asserting the attribute-scoped error when the attribute is set, and that omitting it still allows create/update to succeed.
- **Unit**: request-body cases (populated list, explicit empty list, null/omitted) for both `buildBootstrapRequest` and `buildFinalizeRequest`; `populateModelFromAPI` round-trips including the state-dependent empty-list-vs-null rule (Decision 2); `GetVersionRequirements` cases covering the attribute alone, `agent_policy_ids` alone, both configured, and neither configured (regression coverage for the `models.go:51-54` early-return restructuring in Decision 3).
- Drop the research's "explicit empty list rejected" test — `[]` is valid here, unlike in `integration_policy`.
