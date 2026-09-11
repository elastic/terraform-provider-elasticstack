## Context

See `proposal.md` — Why. Requirements are in `specs/fleet-managed-integration/spec.md`.

Three mechanics shape the approach:

- `Provider.Resources()` (`provider/plugin_framework.go`) returns `p.resources(ctx)` and **appends** `p.experimentalResources(ctx)` when `p.version == AccTestVersion` or `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL=true`. The two lists are concatenated, not merged, so an entry present in both would yield a duplicate `elasticstack_fleet_managed_integration` type name.
- The `docs-generate` Makefile target runs `tfplugindocs` with `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL=false`. That is the reason no `docs/resources/fleet_managed_integration.md` exists today, and the reason one appears as soon as the resource is stable.
- `provider/plugin_framework_experimental_test.go` exists primarily to pin the current experimental placement of this resource; its five tests also each assert `elasticstack_fleet_agentless_policy` stays absent. `elasticstack_kibana_tag` is the only other entry in `experimentalResources()` and has no gate coverage of its own, so the env-var gate mechanism is currently tested only through the resource this change is promoting.

## Goals / Non-Goals

**Goals:**

- Promote the resource with no behavioural change beyond availability and diagnostic wording.
- Keep the `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL` gate itself under test after the resource leaves it.

**Non-Goals:**

- Refactoring the experimental-gate mechanism, the `AccTestVersion` sentinel, or `entitycore` version-requirement plumbing.
- Touching the Kibana 9.5.0 floor, `policyshape.MinVersionCondition`, or the topology preflight.
- Covering the `DataSources()` experimental gate (which gates `kibanatag.NewDataSource` through a separate code path). It is untested today and stays untested here; extending gate coverage to data sources is a candidate follow-up, not part of this promotion.

## Decisions

**Move, not copy.** The registration entry is deleted from `experimentalResources()` and added to `resources()` in the same commit. The alternative — adding to `resources()` first and cleaning up later — would register the type twice for anyone with the opt-in exported, which the Plugin Framework surfaces as a provider-level duplicate-type error rather than a silent no-op. The spec pins this with an explicit single-registration scenario.

**Express the spec delta as REMOVED + ADDED rather than MODIFIED.** The old requirement's "Resource registered as experimental" scenario has to go, and a MODIFIED block replaces a requirement wholesale — OpenSpec validate and archive both reject a MODIFIED block that drops a scenario the current spec still has. The requirement is also renamed, which only REMOVED + ADDED expresses.

**Drop the `elasticstack_fleet_agentless_policy` clause from the replacement requirement.** The old registration requirement also asserted the agentless-policy resource stays absent, duplicating the dedicated **elasticstack_fleet_agentless_policy resource — REMOVED** requirement the main spec already carries. Since this change rewrites the registration requirement anyway, the replacement leaves that behaviour to its dedicated owner, so a future agentless-policy edit touches one requirement instead of two.

**Retarget the gate tests to `elasticstack_kibana_tag` rather than deleting them.** Once the resource is stable, `TestProvider_managedIntegrationNotRegisteredInDefaultProvider` fails outright, and `TestProvider_managedIntegrationRegisteredWhenExperimentalEnvEnabled` becomes vacuous — its `require.Contains` keeps passing because the stable list now carries the resource, so it no longer exercises the gate. The behaviour they cover — that the env var and the `acctest` sentinel actually gate the experimental slice — is still live for `elasticstack_kibana_tag`. Deleting them would leave the gate untested; retargeting preserves the coverage at no cost. `TestProvider_experimentalResourcesIncludesManagedIntegration` and `TestProvider_stableResourcesExcludeManagedIntegrationTypes` invert into a single stable-registration assertion. One accepted gap: the retargeted tests pin behaviour no spec requirement owns (`openspec/specs/kibana-tag/spec.md` has no registration requirement); adding a kibana-tag delta for it is out of scope for this promotion.

**Rename the test file to `plugin_framework_registration_test.go`.** After retargeting, the file covers both stable and experimental registration, so `_experimental_` is no longer an accurate name. Use `git mv` so the diff reads as a rename. The `registeredResourceTypeNames` / `stableResourceTypeNames` helpers stay as-is; they are already generic and are used only in this file.

**Leave the version-gate requirement alone.** The spec requires only that practitioner-facing errors name the 9.5.0 release. Dropping `(experimental API)` from the message string keeps that true, so this is an implementation-level edit with no delta. Rewriting the requirement would invite a reviewer to think the floor moved — and the floor is exactly what should not move, since 9.5.0 is both the API's minimum version and its GA release.

**Generate the docs page rather than hand-writing it.** No `templates/resources/fleet_managed_integration.md.tmpl` is needed; the resource has no example-heavy behaviour that the default template mishandles, and the schema descriptions already carry the deployment constraints. Add a template only if review of the generated page shows it reads poorly.

## Risks / Trade-offs

- **Users on Kibana < 9.5.0 now see the resource in the schema and hit the version error only at plan/apply time** → Accepted and unchanged in kind: this is how every version-gated resource in the provider behaves. The diagnostic already names the required version. Note this is not a preview-exposure risk: the feature is GA from 9.5 and the gate floor is 9.5.0, so the stable resource never reaches a stack where the API is in preview (see `proposal.md` — Upstream GA status).
- **The generated docs page lands with wording that was written for an experimental resource** → Review the generated Markdown as part of the change rather than trusting the generator; the schema-description edit is what the page renders from.
- **Stale references to the change directory in code comments** → comments across the package — the `GetVersionRequirements` and model doc comments in `models.go`, the schema factory comment in `schema.go`, the package comment in `resource.go`, plus `acc_test.go` and `entitycore_contract_test.go` — point at `openspec/changes/fleet-managed-integration/...`, a path that now lives at `openspec/changes/archive/2026-07-22-fleet-managed-integration/`. Fix the comments while editing the package; leaving them deepens the drift.

## Migration Plan

No state migration and no user action. The changelog is regenerated from merged PR titles, so the PR title and description must state the promotion and that `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL` is no longer needed for this resource. Practitioners who set the variable solely for this resource can drop it, and leaving it set stays harmless. Rollback is a revert of the single commit.
