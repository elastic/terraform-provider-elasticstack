## Why

`elasticstack_fleet_managed_integration` was shipped behind the provider's experimental gate to match the tech-preview status of the Kibana Fleet `managed_integrations` API. That feature is now generally available (see Upstream GA status below), so the gate costs users more than it protects them. The resource is absent from the provider schema unless `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL=true` is exported, it has no generated documentation page (`make docs-generate` runs with the flag set to `false`), and its schema and version-gate diagnostics tell practitioners the behaviour may still change.

### Upstream GA status

Elastic Managed integrations are **generally available since Elastic Stack 9.5** — technical preview from 9.0 to 9.4 — and generally available on Elastic Cloud Serverless ([Elastic Managed integrations](https://www.elastic.co/docs/manage-data/ingest/managed-integrations/managed-integrations)).

That boundary lines up exactly with the resource's existing `9.5.0` minimum-version gate, which is what makes this promotion safe rather than merely plausible: the resource already refuses to run against any stack older than 9.5.0, so every Kibana version it can reach is one where the feature is GA. There is no version window in which a stable resource would be talking to a tech-preview API. It also confirms the version gate itself needs no change — see Out of Scope.

## What Changes

- Register `elasticstack_fleet_managed_integration` in the stable `resources()` list in `provider/plugin_framework.go` and remove it from `experimentalResources()`. It must appear in exactly one list — `Resources()` appends the experimental slice to the stable one, so an entry in both would register the same type name twice.
- Rewrite the `**This resource is experimental**: ...` sentence in the resource schema's `MarkdownDescription` in `internal/fleet/managedintegration/schema.go`. The Kibana 9.5.0 minimum-version statement currently lives inside that sentence, so it cannot be deleted outright; the replacement wording is "The underlying Fleet managed integrations API requires Kibana 9.5.0 or later." The Elastic Cloud Hosted / Serverless-only deployment constraint that follows stays unchanged.
- Remove the `(experimental API)` suffix from the version-gate diagnostic in `internal/fleet/managedintegration/models.go` (`GetVersionRequirements`).
- Remove the instruction to export `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL=true` from `examples/resources/elasticstack_fleet_managed_integration/README.md`; it becomes wrong once the resource is stable.
- Generate the previously missing `docs/resources/fleet_managed_integration.md` page, which `make docs-generate` now produces because the resource is in the default schema.
- Update the `fleet-managed-integration` spec requirement **Resource type and registration** so it mandates stable registration instead of `experimentalResources()`.

Not a breaking change: the resource becomes available to configurations that previously needed the opt-in env var, and existing opted-in configurations keep working unchanged. The `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL` mechanism itself stays in place for `elasticstack_kibana_tag`, which remains experimental.

## Out of Scope

- The Kibana 9.5.0 minimum-version gate stays exactly as it is: promotion out of experimental is about provider-level stability guarantees, not about widening version support (see Upstream GA status). Practitioners on older stacks continue to get a version diagnostic, now without the `(experimental API)` wording.
- No schema attributes, CRUD behaviour, or API endpoints change.

## Capabilities

### New Capabilities

<!-- None — this change promotes an existing capability's registration; no new capability is introduced. -->

### Modified Capabilities

- `fleet-managed-integration`: the **Resource type and registration** requirement is removed and replaced by **Resource type and stable registration**. The new requirement mandates registration in the stable `resources()` list and forbids the resource description from calling the resource experimental. It adds scenarios for availability without the opt-in, single registration when the opt-in is set, and documentation generation. The discoverability scenario carries over with a retitle, minus its `elasticstack_fleet_agentless_policy` absence assertion — the dedicated **elasticstack_fleet_agentless_policy resource — REMOVED** requirement already owns that behaviour. The delta is expressed as REMOVED + ADDED rather than MODIFIED, and the **Version gate for managed_integrations endpoint** requirement is deliberately left untouched; see design.md for both rationales.

## Impact

- `provider/plugin_framework.go` — resource moves between the two registration lists.
- `provider/plugin_framework_experimental_test.go` — five tests pin the current experimental placement; they are retargeted or replaced within the same file, which is renamed to `plugin_framework_registration_test.go` (see design.md).
- `internal/fleet/managedintegration/schema.go` — `MarkdownDescription` wording.
- `internal/fleet/managedintegration/models.go` — `GetVersionRequirements` error message and the stale comment above it.
- `internal/fleet/managedintegration/entitycore_contract_test.go` — two assertions pin the exact error-message string.
- `internal/fleet/managedintegration/` — stale change-path comments across the package (`schema.go`, `models.go`, `resource.go`, `acc_test.go`, `entitycore_contract_test.go`) point at the pre-archive change directory.
- `examples/resources/elasticstack_fleet_managed_integration/README.md` — drops the experimental env-var instruction.
- `docs/resources/fleet_managed_integration.md` — new generated page.
- `openspec/specs/fleet-managed-integration/spec.md` — updated when this change is archived.
- No dependency, API-client, or `kbapi` changes. `CHANGELOG.md` needs no manual edit; the `## [Unreleased]` section is regenerated from merged PR history, which is why the Migration Plan in design.md requires the PR title to state the promotion.
