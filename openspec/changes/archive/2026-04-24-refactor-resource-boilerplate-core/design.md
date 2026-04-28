## Context

Many Plugin Framework resources in this repository repeat the same three pieces of wiring in `resource.go`: a `*clients.ProviderClientFactory` field, a `Configure` method that converts provider data into that factory, and a `Metadata` method that constructs the Terraform type name from a hard-coded suffix. The duplication is spread across Elasticsearch, Kibana, Fleet, and APM packages, even though the pattern is substantially the same.

The refactor needs to stay narrow. `ImportState` is not uniform: some resources use passthrough import, some parse composite IDs, and some do not implement import at all. The refactor also needs to respect existing type names exactly during rollout, including legacy suffix spellings such as `agentbuilder_tool`. Although APM resources use Kibana-backed APIs, their Terraform type names live under the `apm_*` namespace, so naming concerns and client-resolution concerns cannot be treated as the same axis.

## Goals / Non-Goals

**Goals:**
- Define a shared embedded core for Plugin Framework resources that centralizes canonical `Configure`, client storage, and `Metadata`.
- Introduce a typed component namespace for type-name generation, including `elasticsearch`, `kibana`, `fleet`, and `apm`.
- Standardize the default `Configure` behavior on early return after appended diagnostics.
- Trial the embedded core on four representative resources:
  - `internal/elasticsearch/ml/jobstate`
  - `internal/kibana/agentbuildertool`
  - `internal/fleet/integration`
  - `internal/apm/agent_configuration`
- Add guardrails that detect accidental interface drift when methods are promoted through embedding.

**Non-Goals:**
- Unifying `ImportState` behind the embedded core.
- Renaming existing Terraform resource type names as part of this change.
- Extending the abstraction to data sources in the initial rollout.
- Moving CRUD, schema, validation, or state-upgrade logic into the shared core.

## Decisions

### Introduce a provider-wide `internal/resourcecore` package

The shared logic will live in a provider-wide package dedicated to Plugin Framework resource wiring, tentatively `internal/resourcecore`. This keeps the code grouped by logical concept, matches the repository guidance for provider-wide shared helpers, and avoids adding another generic `utils` surface.

**Alternative considered:** keep the logic as free functions only. Rejected for this change because the primary goal is to reduce repeated receiver boilerplate in `resource.go`, and an embedded core gives a cleaner end state for the pilot resources.

### The core will own only client wiring and type-name construction

The core will carry:
- a typed `Component` value
- a literal `resourceName` suffix segment
- the configured `*clients.ProviderClientFactory`

The core will expose:
- `Configure(...)`
- `Metadata(...)`
- `Client() *clients.ProviderClientFactory`

The `Client()` accessor keeps concrete resources from depending on a mutable exported field while still allowing embedded use across packages.

**Alternative considered:** export the client field directly. Rejected because it weakens encapsulation and makes later evolution of the core more difficult.

### Type names will be constructed from literal namespace parts, without normalization

`Metadata` will build the Terraform type name as:

`<provider_type_name>_<component>_<resource_name>`

The `component` value will come from well-known typed constants such as `ComponentElasticsearch`, `ComponentKibana`, `ComponentFleet`, and `ComponentAPM`. The `resourceName` value will be passed in as the final suffix segment to preserve current compatibility, rather than being derived from package names or transformed automatically.

This allows future resources to use names like `agent_builder_tool` if desired, while allowing the initial pilot to preserve current spellings such as `agentbuilder_tool`.

**Alternative considered:** derive the resource suffix from the Go package name or auto-convert identifiers to snake case. Rejected because current resource type names are not uniform enough for safe derivation.

### `ImportState` remains explicit on concrete resources

The embedded core will not define `ImportState`. Resources that already implement import will continue to do so explicitly, and resources that do not support import will remain non-importable. This avoids accidental satisfaction of `resource.ResourceWithImportState` through promoted methods.

**Alternative considered:** add a default passthrough import implementation to the core. Rejected because it would blur intentional differences between passthrough, composite-ID, and no-import resources.

### The pilot rollout will be staged to maximize shape coverage

The initial rollout sequence will be:
1. `internal/elasticsearch/ml/jobstate` — simple Elasticsearch resource with passthrough import.
2. `internal/kibana/agentbuildertool` — Kibana resource with passthrough import and a legacy type-name suffix spelling.
3. `internal/fleet/integration` — resource with canonical `Configure` and `Metadata`, but no import and additional upgrade-state behavior.
4. `internal/apm/agent_configuration` — resource that uses Kibana-backed APIs while needing the new `apm` naming component.

This order exercises the core across the main namespace and behavior combinations before any broader adoption.

### Add a conformance harness for promoted-method safety

The change will include a small conformance test surface that asserts the intended interface shape of representative resource forms. At minimum, the harness will verify that:
- an embedded-core resource satisfies `resource.ResourceWithConfigure`
- a no-import resource does not accidentally satisfy `resource.ResourceWithImportState`
- a custom-import resource continues to own its explicit import behavior

**Alternative considered:** rely only on existing concrete resource interface assertions. Rejected because those assertions do not protect against architectural drift in the abstraction itself.

## Risks / Trade-offs

- [Promoted methods make behavior ownership less obvious in review] → Keep the core small, preserve explicit interface assertions on each concrete resource, and add the conformance harness.
- [Canonical early-return `Configure` could subtly change resources that currently assign `client` despite diagnostics] → Limit the pilot to resources already matching or compatible with the canonical behavior, and audit outliers separately.
- [Component naming could be confused with client-resolution kind] → Document `Component` as a Terraform type-name namespace, not as a client selector.
- [The embedded pattern may still feel less readable than helper-only wrappers] → Use the four-resource pilot to compare review clarity before wider rollout.

## Migration Plan

1. Add `internal/resourcecore` with the typed component constants, constructor, `Configure`, `Metadata`, and `Client()` accessor.
2. Add unit and conformance tests for type-name generation and promoted-method/interface behavior.
3. Convert the four pilot resources one at a time, preserving existing type names and import behavior.
4. Run targeted tests for the shared package and each converted resource package after each step.
5. If the pilot shows reduced readability or unexpected interface coupling, stop rollout and retain the package only as a helper surface for later reconsideration.

Rollback is straightforward because the change is internal to resource wiring: each pilot resource can be reverted to explicit `client`, `Configure`, and `Metadata` methods without user state migration.

## Pilot verification (task 4)

**4.1 Targeted tests.** `go test` was run for `./internal/resourcecore/...` and the four pilot packages (`./internal/elasticsearch/ml/jobstate/...`, `.../kibana/agentbuildertool/...`, `.../fleet/integration/...`, `.../apm/agent_configuration/...`) using a plain `go` binary (unwrapped). `resourcecore` reported passing unit tests; pilot packages compile and, where acceptance tests are not enabled, exit successfully with any `TestAcc*` tests skipped as usual. `make build` also succeeded. Full `TF_ACC=1` runs require a local Elastic stack and were exercised in prior implementation passes on this change.

**4.2 Type names and import boundaries.** Unchanged and consistent with the pre-refactor contract:

| Resource | `Metadata` / type-name segments | Import support |
| --- | --- | --- |
| ML job state | `elasticsearch` + `ml_job_state` | Passthrough `ImportState` (explicit) |
| Agent Builder tool | `kibana` + `agentbuilder_tool` (legacy spelling) | Passthrough `ImportState` (explicit) |
| Fleet integration | `fleet` + `integration` | None (`ResourceWithImportState` not implemented) |
| APM agent configuration | `apm` + `agent_configuration` | Passthrough `ImportState` (explicit) |

**4.3 Readability decision.** **Continue the embedded core rollout** beyond the pilot, subject to the usual per-resource review. Rationale: the four pilots cover the intended matrix (ES, Kibana with legacy suffix, no-import plus upgrade state, APM+Kibana API); concrete `resource.go` files stay small; explicit interface assertions and the conformance tests keep import behavior and method promotion visible. Stopping at the pilot or reverting to helper-only is unnecessary given no material readability regression, while helper-only would forfeit the embed goal without new evidence.

## Open Questions

- Whether a sibling abstraction for data sources should be proposed later, after the resource pilot has proven readable.
- Whether `internal/resourcecore` is the best final package name, or whether another provider-wide logical name would better match existing conventions.
