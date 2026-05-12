## Context

Kibana's Fleet package-policy create/update API
(`POST|PUT /api/fleet/package_policies`) accepts a `package_policy_request` body.
Inspecting the generated client at
`generated/kbapi/kibana.gen.go` (`PackagePolicyRequestMappedInputs`) confirms the
request schema has no top-level `enabled` field — only inputs and streams have an
`enabled` flag (`PackagePolicyRequestMappedInput.Enabled`,
`PackagePolicyRequestMappedInputStream.Enabled`).

The response type `PackagePolicy` *does* include `Enabled bool`, but Kibana
hardcodes it to `true` for created policies and there is no documented or
implemented mechanism to disable a package policy by writing to that field. In the
Kibana UI, package policies are not toggled on/off as a unit — they are removed
from agent policies, or their inputs are disabled.

The Terraform resource exposed `enabled` as `Optional + Computed` with a default
of `true`. `populateFromAPI` mapped the response field into state, but
`toAPIModel` never wrote it back to the request. Because the API always returns
`true`, any user setting `enabled = false` triggered Terraform's
"produced an unexpected new value: .enabled: was cty.False, but now cty.True"
consistency error.

## Goals / Non-Goals

**Goals:**

- Remove the no-op `enabled` attribute so users can no longer configure it
  ineffectively or trigger the inconsistency error.
- Provide an automatic state migration so existing v2 state files (which have
  `enabled = true` baked in) load cleanly under the new schema.
- Keep the per-input and per-stream `enabled` toggles, which are the actual
  Kibana-supported way to disable telemetry from an integration policy.
- Keep the OpenSpec functional spec in sync with the implementation.

**Non-Goals:**

- Add a server-side or client-side mechanism to "disable" a package policy as a
  whole. Kibana doesn't support this, and emulating it client-side (e.g.
  disabling every input) would silently mutate user-managed inputs.
- Deprecate or rename per-input or per-stream `enabled` — those are correct.
- Refactor the v0 / v1 / v2 model structs themselves; only the conversion
  targets and the live model change.

## Decisions

### Decision 1: Remove the attribute outright (no deprecation period)

**Rationale:** The attribute has never worked. Any existing config that sets
`enabled = true` is a no-op on the wire and matches the API's always-true
return; removing it changes nothing functionally for those users. Configs that
set `enabled = false` are already broken (they error on apply on Stack 9.x as
the new acceptance test demonstrates), so deprecating-then-removing only
extends a window of broken-looking config without buying compatibility.

**Alternative considered:** Mark `enabled` as `Computed`-only (drop `Optional`).
Rejected because it still leaves a useless read-only field whose value never
changes (`true`), and it doesn't fix the underlying confusion that Kibana cannot
disable a package policy this way.

**Alternative considered:** Deprecate for one minor release, remove next.
Rejected because the deprecation message would have to say "this never worked"
and the only correct user response is to remove the field — which we can do for
them via state upgrade now.

### Decision 2: Bump schema version to 3 with a v2 → v3 state upgrader

**Rationale:** State files written by prior provider versions contain `enabled`
in the v2 schema. Without an upgrader, Terraform would refuse to load existing
state ("schema version 2 is greater than known version" or attribute mismatch).
A v2 → v3 upgrader using the prior v2 schema as `PriorSchema` is the standard
plugin-framework pattern. The upgrader is a pure structural drop; no API call is
needed.

The existing v0 → v2 and v1 → v2 upgraders are retargeted to v3 (renamed
`toV3` / `upgradeV0ToV3` / `upgradeV1ToV3`) so all three migration paths land on
the live schema. They already construct the live model directly, so they
naturally stop populating `enabled` once the field is removed from the model.

### Decision 3: Keep the v2 schema/model in tree under `schema_v2.go`

**Rationale:** Terraform plugin-framework state upgraders need a `PriorSchema`
that exactly matches the old schema; otherwise `req.State.Get` won't decode the
prior state. We capture this as `getSchemaV2()` and `integrationPolicyModelV2`
in a dedicated `schema_v2.go`, mirroring how `schema_v0.go` and `schema_v1.go`
already preserve their respective prior versions.

### Decision 4: Spec captures the upgrade behaviour

**Rationale:** The OpenSpec functional spec
(`openspec/specs/fleet-integration-policy/spec.md`) documents schema and
behaviour the provider must preserve. Removing `enabled` is a schema-visible
change; the spec is updated in lock-step (schema block, REQ-022 mapping,
REQ-024/025 carry-over lists) and a new requirement covers the v2 → v3 upgrade
contract.

## Risks / Trade-offs

- **[Risk]** Users with `enabled = true` (the schema default) explicitly set in
  HCL will see a Terraform plan-time error after the upgrade ("Unsupported
  argument").
  - **Mitigation**: Document in the CHANGELOG breaking-change section that the
    field must be removed from configuration. The fix is a one-line deletion.
- **[Risk]** Users with `enabled = false` were already broken; after the
  upgrade they get a plan-time "Unsupported argument" instead of an apply-time
  inconsistency error.
  - **Mitigation**: Same CHANGELOG entry. The new error message is clearer and
    earlier in the workflow.
- **[Risk]** External tooling that reads provider state JSON might rely on the
  `enabled` attribute being present.
  - **Mitigation**: It always read `true`. Tools should treat its absence as
    equivalent to `true`. Called out in the CHANGELOG.

## Open Questions

- None. The Kibana Fleet team has not signalled any plan to add a top-level
  `enabled` write field; if that ever changes, the attribute can be re-added in
  a future schema version with a real implementation.
