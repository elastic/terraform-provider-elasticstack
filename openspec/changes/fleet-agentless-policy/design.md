## Context

Fleet agentless policies differ fundamentally from standard Fleet agent policies. Rather than deploying Elastic Agent to a target host, Kibana provisions the agent runtime itself in the cloud. The `POST /api/fleet/agentless_policies` endpoint is a **bundled create**: in one call, Kibana provisions a hidden managed agent policy *and* a package policy on top of it. The response's `id` is the **package policy** ID — this is the identifier used for all subsequent read, update, and delete operations.

The agentless API surface is sparse by design:

| Verb | Path | Notes |
|---|---|---|
| POST | `/api/fleet/agentless_policies` | Create (bundled) |
| DELETE | `/api/fleet/agentless_policies/{policyId}` | Delete |
| — | *(no dedicated GET)* | Read via package_policies fallback |
| — | *(no dedicated PUT)* | Update via package_policies fallback |
| GET | `/api/fleet/package_policies/{id}` | Read — works on agentless-created policies |
| PUT | `/api/fleet/package_policies/{id}` | Update — allowlist applies (see Decision 3) |

The API is experimental and was added in Kibana **9.3.0**. It is only supported on **Elastic Cloud Hosted** and **Serverless** (Security or Observability) deployments.

The existing `elasticstack_fleet_integration_policy` resource already models the same `inputs`, `streams`, and `vars` shape. Rather than duplicating that logic, this change extracts it into a shared package that both resources import.

The in-flight `fleet-cloud-connector` change (`openspec/changes/fleet-cloud-connector`) adds full CRUD for cloud connectors. The agentless resource references a connector by raw `cloud_connector_id` string — no hard dependency on that change landing first.

## Goals / Non-Goals

**Goals:**

- Full lifecycle (Create, Read, Update, Delete) for agentless policies with import support.
- Phase 1 extracts shared inputs/streams/vars modeling from `integration_policy` into `internal/fleet/policyshape/` so both resources share the implementation without code duplication.
- Phase 2 implements `elasticstack_fleet_agentless_policy` using the shared package.
- Hybrid update semantics: in-place updates for mutable fields, replacement for structural/identity fields.
- Deployment-topology preflight check rejects self-managed stacks with a clear, actionable error.
- Version gating at Kibana 9.3.0 via the existing `EnforceMinVersion` pattern.
- Cloud connector referenced by raw `cloud_connector_id` string; no hard dependency on the `fleet-cloud-connector` change.
- Import via `<space_id>/<policy_id>` composite form (and bare `<policy_id>` for default space).

**Non-Goals:**

- A dedicated data source for listing agentless policies (no dedicated list endpoint exists; the package_policies list endpoint does not filter by agentless origin).
- Surfacing the hidden managed agent policy ID or its details.
- Supporting self-managed Kibana deployments (not supported by the API).
- Providing typed cloud provider blocks (like the `aws {}` / `azure {}` blocks in the `fleet-cloud-connector` resource) — users supply `vars_json` directly because the integration package controls the variable schema, not the provider.
- Retrofitting write-only secret hashing from `fleet-cloud-connector` to this resource in this change.
- Support for Kibana versions older than 9.3.0.

## Decisions

### Decision 1: Dedicated resource, not an extension of `elasticstack_fleet_integration_policy`

Implement `elasticstack_fleet_agentless_policy` as a separate resource rather than adding an `agentless = true` flag to `integration_policy`.

**Why:** The agentless API is structurally different — a separate endpoint, a different lifecycle (no agent policy ID required, bundled create), and a different set of fields that force replacement vs. update in-place. Merging the two resources would require significant schema branching inside `integration_policy` and would make the resource harder to understand. The existing `supports_agentless` flag on `agent_policy` is a different concept (making a *regular* agent policy able to host agentless package policies via cloud connectors). A dedicated resource is unambiguous.

### Decision 2: Phase 1 extracts shared modeling into `internal/fleet/policyshape/`

Before implementing the agentless resource, extract `InputType`, `InputsType`, `VarsJsonType`, defaults merging, canonical JSON normalization, and secret helpers from `internal/fleet/integration_policy/` into a new package `internal/fleet/policyshape/`. The `integration_policy` resource is then migrated to import from this package.

**Why:** The agentless policy body has nearly identical inputs/streams/vars structure to an integration policy. Duplicating the Plugin Framework custom types and the normalization logic would create maintenance debt immediately. Extracting first (Phase 1) and migrating `integration_policy` to use the shared package is a behaviour-preserving refactor that can be validated independently before Phase 2 begins. Phase 1 is behaviour-preserving with one explicit, additive exception: the `condition` field is promoted to a first-class Optional attribute on inputs and streams (see Open Question 4 resolution). This is non-breaking (no state upgrader required) and closes an existing functionality gap — the Fleet package policy API exposes `condition` at integration, input, and stream levels, but `integration_policy` currently drops it. The name `policyshape` is a working name; implementers should validate it against existing naming conventions in the repo.

### Decision 3: Hybrid update semantics

Split the resource attributes into two groups:

- **In-place updatable** (via `PUT /api/fleet/package_policies/{id}`): `description`, `vars_json`, `inputs`, `global_data_tags`, `additional_datastreams_permissions`, `var_group_selections`, `package.title`.
- **Replace-on-change** (`RequiresReplace`): `name`, `package.name`, `package.version`, `namespace`, `cloud_connector.*`, `policy_template`, `policy_id`, `space_ids`.
- **Create-only** (sent on create only; not read back; not on PUT; not `RequiresReplace`): `force`, `force_delete`, `create_dataset_templates`. Post-create changes to these are no-ops.

**Why:** The agentless endpoint has no PATCH semantics; the package_policies PUT endpoint is the only update path. However, the Kibana API behavior for hidden agentless-created policies under PUT is not fully documented. An explicit spike task (see Tasks §3) must probe which fields Kibana actually honors to validate or refine the in-place allowlist before shipping. The current partition is the expected behavior based on API design, not confirmed empirically.

### Decision 4: Read via package_policies fallback

On Read, call `GET /api/fleet/package_policies/{id}` (space-aware) with the stored `policy_id`. HTTP 404 means the policy was deleted out of band — remove from state without error. The agentless-specific metadata (cloud_connector configuration, policy_template) may not all round-trip through this endpoint; implementation must document which fields are reconstructed from state vs. read from the API.

**Why:** There is no `GET /api/fleet/agentless_policies/{id}` endpoint. The `POST` response provides a full `KibanaHTTPAPIsAgentlessPolicy` object. For subsequent reads, the package policy endpoint is the documented fallback per Kibana's own API guide.

### Decision 5: `force_delete` attribute for managed-policy conflicts

Add a `force_delete` boolean attribute (default `false`) that passes `?force=true` to the DELETE endpoint. When `force_delete = false` and the API returns a conflict, surface a helpful error.

**Why:** The DELETE endpoint accepts a `force` query parameter (`DeleteFleetAgentlessPoliciesPolicyidParams.Force`). Without surfacing this, users who encounter a managed-policy conflict have no Terraform-native resolution path. The pattern mirrors `fleet_cloud_connector`'s `force_delete` attribute.

### Decision 6: Composite ID and SpaceImporter

The resource `id` is set to the composite string `"<space_id>/<policy_id>"`. `policy_id` is Optional+Computed with `UseStateForUnknown`; when omitted from config, the API-assigned ID is used. Import accepts both `"<space_id>/<policy_id>"` and bare `"<policy_id>"`.

**Why:** Matches the existing space-aware Fleet import pattern used across the provider (via the `SpaceImporter` helper in `internal/fleet/space_importer.go`).

### Decision 7: Deployment topology preflight check

Add a preflight check that queries the Kibana status or cluster info to detect self-managed stacks and refuse agentless policy operations with a clear error message directing users to Elastic Cloud Hosted or Serverless.

**Why:** The agentless API silently accepts requests on self-managed stacks (the endpoint exists) but the policy will never activate because Elastic's cloud infrastructure is not present. Failing fast with a clear error is better than a policy that appears created but never collects data. The check can use a heuristic similar to what Fleet itself uses (e.g., checking for the presence of the cloud metadata plugin or a Fleet server setup mode indicator).

**Fallback when detection is inconclusive (fail-open):** If the preflight heuristic cannot reliably determine the topology (neither confidently cloud-hosted nor confidently self-managed), the resource SHALL **fail open** — it SHALL proceed with the operation rather than block potentially-legitimate cloud-hosted setups that the heuristic mis-classifies. This trades the risk of silently creating a non-functional policy on a mis-classified self-managed stack for fewer false-negatives on legitimate cloud setups. When the heuristic positively identifies a self-managed stack, the resource still fails closed per the primary rule. If a later API call fails for topology reasons, the surfaced error diagnostic SHALL still guide the user toward the cloud-hosted requirement.

### Decision 8: `vars_json`, not typed vars blocks

Integration-level and input-level vars are encoded as JSON strings (`vars_json` attributes), not as typed Terraform map attributes.

**Why:** Reuses the existing `VarsJsonType` modeling from `integration_policy`, which is already proven and tested. The variable schema is controlled by the integration package, not the provider — trying to model it as typed attributes would require schema introspection at plan time and would produce a worse user experience for custom integrations. Users who need strongly-typed vars can use `jsonencode()` in their HCL.

## Open questions

1. **Which fields does `PUT /api/fleet/package_policies/{id}` honor for hidden agentless-created policies?** The in-place update allowlist in Decision 3 is based on API design inference, not confirmed empirical testing. A spike task (Tasks §3, task 3.3) must probe this before Phase 2 acceptance tests are written. If the PUT endpoint rejects updates entirely for agentless-created policies, the design shifts to replace-on-all-changes.

2. **Acceptance test matrix: which integrations work in CI?** CSPM (GA, `cloud_security_posture` package) is the intended golden-path test. A representative beta integration for secondary coverage should be identified during implementation. Acceptance tests are gated on Kibana ≥ 9.3 and require a cloud-hosted topology.

3. **Final name for the shared modeling package.** `internal/fleet/policyshape/` is a working name. The implementer should verify naming against existing conventions in the repo (e.g., `internal/fleet/` already has `integration_policy`, `agentpolicy`, etc.) and adjust if needed. The OpenSpec spec refers to the behavior, not the package name.

4. **`condition` and `deprecated` on inputs/streams — resolved.** `condition` SHALL be promoted to a first-class Optional string attribute on inputs and streams in `policyshape` during Phase 1, and `integration_policy` SHALL gain the same attribute (additive, non-breaking — no state upgrader required; existing configs and state are unaffected). This closes an existing functionality gap: the Fleet package policy API exposes `condition` at integration, input, and stream levels (top-level on `KibanaHTTPAPIsSimplifiedCreatePackagePolicyRequest`, and on `PackagePolicyRequestMappedInput` / `PackagePolicyRequestMappedInputStream`), and it round-trips through the read response (`PackagePolicyMappedInputs`), but `integration_policy` currently drops it entirely. `deprecated` is API-returned deprecation metadata (not user-configurable) and SHALL NOT be modeled in v1.

5. **Deployment topology check implementation.** The precise mechanism for detecting cloud-hosted vs. self-managed is TBD. Possible approaches: (a) check for the `xpack.fleet.agentless.enabled` cluster setting via the ES cluster settings API; (b) check the Kibana status `buildSha` / `uuid` against known cloud patterns; (c) attempt a preflight `POST` with `dry_run: true` if such a parameter exists. This should be resolved early in Phase 2 implementation (Tasks §6).

6. **Topology opt-in override.** Decision 7 specifies a **fail-open** fallback when detection is inconclusive (proceed rather than block). Should the resource additionally offer an explicit provider/resource-level opt-in (e.g. a `force_cloud_topology` / `skip_topology_check` flag) for users who run legitimately cloud-hosted stacks that the heuristic mis-classifies? If added, it must be documented as escape-hatch only and must not weaken the positive self-managed detection (fail-closed path). Deferred to Phase 2 implementation once the detection heuristic is known.
