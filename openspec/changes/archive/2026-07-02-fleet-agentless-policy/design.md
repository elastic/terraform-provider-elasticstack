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

**Spike results (Task 3.3, empirically confirmed against a live Kibana 9.4.3 Cloud Hosted deployment; full findings in a comment block at the top of `internal/fleet/agentlesspolicy/update.go`):** all seven in-place-updatable candidates (`description`, `vars_json`, `inputs` vars, `global_data_tags`, `additional_datastreams_permissions`, `var_group_selections`, `package.title`) are accepted and persisted by `PUT /api/fleet/package_policies/{id}`, confirming that half of the allowlist. One caveat: toggling an input's `enabled` sub-field was accepted (200) but NOT persisted for the `cloud_security_posture` package tested (it silently reverted to its prior value) — this may be package-specific business logic (CSPM enforces "only one enabled input" on create) rather than a general PUT limitation; Task 5's implementation should verify enabled-state changes against the actual response body rather than assuming a 200 means the change took effect.

**Important contradiction — flagged for orchestrator review:** the RequiresReplace premise ("not fully documented" / "not confirmed empirically") is now resolved, and the result contradicts the assumption that `name`, `namespace`, and `package.version` are API-enforced immutable: the same PUT endpoint accepted and persisted changes to all three (including setting `package.version` to a version string that isn't installed or in the registry — PUT performs no registry validation at all, unlike Create). `package.name` was not a clean pass-through, but its rejections were about registry/agentless-capability ("package not installed", "package does not support agentless deployment mode"), not about the field being immutable. Despite this, **the RequiresReplace partitioning above is kept unchanged for Tasks 4/5**: forcing replacement for `name`, `namespace`, `package.version`, and `package.name` is now understood to be a deliberate Terraform-side safety choice (avoiding an in-place path that bypasses Fleet's package-install/upgrade lifecycle) rather than an API constraint. `cloud_connector.*` and `policy_template` were not probed by the spike (out of scope per the task's field list) and remain RequiresReplace by design-only inference.

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

**Escape hatch for a positive-but-wrong self-managed classification (`skip_topology_check`, see Open Question 6):** the implemented heuristic's positive classification of "self-managed" is not actually "inconclusive" from the fail-open fallback's point of view -- it is a confident (if occasionally wrong) verdict, so the fail-open path above does not cover it. The `skip_topology_check` Optional bool schema attribute closes that gap: it lets a user bypass `checkDeploymentTopology` entirely for a deployment they know is genuinely Elastic Cloud Hosted or Serverless but that the heuristic mis-classifies (e.g. PrivateLink routing that never surfaces the Elastic Cloud edge-proxy headers this heuristic depends on).

### Decision 8: `vars_json`, not typed vars blocks

Integration-level and input-level vars are encoded as JSON strings (`vars_json` attributes), not as typed Terraform map attributes.

**Why:** Reuses the existing `VarsJsonType` modeling from `integration_policy`, which is already proven and tested. The variable schema is controlled by the integration package, not the provider — trying to model it as typed attributes would require schema introspection at plan time and would produce a worse user experience for custom integrations. Users who need strongly-typed vars can use `jsonencode()` in their HCL.

## Open questions

1. **Which fields does `PUT /api/fleet/package_policies/{id}` honor for hidden agentless-created policies? -- resolved.** Empirically confirmed via the Task 3.3 spike against a live Kibana 9.4.3 Cloud Hosted deployment (full findings in `internal/fleet/agentlesspolicy/update.go`). The PUT endpoint does not reject updates for agentless-created policies; all seven candidate in-place-updatable fields from Decision 3 are accepted and persisted (with a caveat on `inputs[*].enabled` toggling, see Decision 3). **Unexpectedly, the endpoint also accepts and persists changes to fields Decision 3 marks RequiresReplace** (`name`, `namespace`, `package.version` -- including to an uninstalled/nonexistent version string, since PUT performs no registry validation) **and does not reject `package.name` changes for immutability reasons** (only for registry/agentless-capability reasons). This means the RequiresReplace partitioning for identity/structural fields is a **Terraform-side safety policy**, not an API-enforced constraint -- the API would technically allow an in-place rename, namespace move, or version bump, but doing so would bypass Fleet's package-install/upgrade lifecycle (index templates, ingest pipelines, and Kibana saved objects for a new package version are not provisioned by PUT). The design keeps RequiresReplace for these fields deliberately, on safety grounds, and this rationale is flagged prominently for orchestrator review given it changes the "why" from an inferred API constraint to an explicit product decision.

2. **Acceptance test matrix: which integrations work in CI? -- resolved (descoped).** CSPM (GA, `cloud_security_posture` package) is the golden-path test actually implemented (Task 8: `TestAccResourceAgentlessPolicy` and its siblings in `internal/fleet/agentlesspolicy/acc_test.go`), and it alone was sufficient to exercise every requirement in specs/fleet-agentless-policy/spec.md end-to-end, including in-place updates, import, `force_delete`, cloud connectors, and RequiresReplace semantics. A second, beta-integration package for secondary coverage was considered but not added: no gap in requirement coverage was identified that a second package would close, and CSPM's own credential-vars/CloudFormation-URL quirks (documented throughout acc_test.go) already exercised the trickiest edge cases (server-injected computed vars, per-credential-type var shapes). Left for a future change if a specific regression ever motivates it.

3. **Final name for the shared modeling package -- resolved.** `internal/fleet/policyshape/` was verified against existing repo conventions during Task 1 and kept as the final name (see `internal/fleet/policyshape/doc.go`'s package-decision comment, called for by tasks.md 1.2). No rename occurred.

4. **`condition` and `deprecated` on inputs/streams — resolved.** `condition` SHALL be promoted to a first-class Optional string attribute on inputs and streams in `policyshape` during Phase 1, and `integration_policy` SHALL gain the same attribute (additive, non-breaking — no state upgrader required; existing configs and state are unaffected). This closes an existing functionality gap: the Fleet package policy API exposes `condition` at integration, input, and stream levels (top-level on `KibanaHTTPAPIsSimplifiedCreatePackagePolicyRequest`, and on `PackagePolicyRequestMappedInput` / `PackagePolicyRequestMappedInputStream`), and it round-trips through the read response (`PackagePolicyMappedInputs`), but `integration_policy` currently drops it entirely. `deprecated` is API-returned deprecation metadata (not user-configurable) and SHALL NOT be modeled in v1.

   **Follow-up, empirically confirmed:** `condition` is not yet a released feature as of this change. Verified directly against a live Kibana 9.5.0-SNAPSHOT (via an isolated, since-torn-down `docker compose` stack) that `condition` on package-policy inputs/streams round-trips correctly through create/read using this provider's code, but both currently-released 9.4.0 and 9.4.3 reject it with an "Additional properties are not allowed" HTTP 400. `condition` is therefore gated behind `MinVersionCondition` (9.5.0) using the same soft, attribute-scoped `EnforceMinVersion` pattern as `SupportsPolicyIDs`/`SupportsOutputID` (see `internal/fleet/integration_policy/capabilities.go` and `resource.go`): the gate only produces an error when `condition` is actually set on an input or stream and the connected Kibana is older than 9.5.0, so it does not affect any configuration that leaves `condition` unset. See `openspec/changes/fleet-agentless-policy/specs/fleet-integration-policy/spec.md` ("Version gating for `condition`") for the corresponding requirement and scenarios.

   **Correction (post-archive, PR #4034 review):** this resolution's gating description only covered `integration_policy`, but `agentlesspolicy` also surfaces `condition` on its own inputs/streams via the same shared `policyshape.InputType`/`StreamType`, and Task 5/6's implementation shipped without a matching gate on that resource -- a user on `agentlesspolicy`'s own 9.3.0 floor who set `condition` got the same raw, unhelpful Kibana 400 this gate exists to prevent. Fixed by (a) relocating `MinVersionCondition` from `integration_policy`'s `resource.go` into `policyshape` itself (`internal/fleet/policyshape/version.go`), since the constant is a property of the shared `condition` attribute both resources expose, not of `integration_policy` specifically -- `integration_policy` now keeps a package-level alias pointing at `policyshape.MinVersionCondition` so its existing call sites are unaffected -- and (b) adding the identical `resolveAgentlessPolicyFeatures`/`validateInputConditionSupport` gating pair to `agentlesspolicy` (`internal/fleet/agentlesspolicy/capabilities.go` and `models_convert.go`), wired into both `toCreateBody` and `buildUpdateBody`. See `internal/fleet/agentlesspolicy/condition_test.go` for the corresponding test coverage (mirroring `integration_policy/models_test.go`'s `TestConditionHandling`).

5. **Deployment topology check implementation -- resolved.** Implemented in Task 6 as a `GET /api/status` heuristic (`internal/fleet/agentlesspolicy/topology.go`, `checkDeploymentTopology`): Kibana's `build_flavor` combined with the presence of Elastic Cloud's `X-Found-Handling-Cluster`/`X-Found-Handling-Instance` edge-proxy response headers. None of the three originally-considered approaches were used -- (a) and (c) were not viable (no such cluster setting or dry-run parameter was found to exist), and (b) (`buildSha`/`uuid` pattern-matching) was superseded by the proxy-header signal, which is a more direct and reliable indicator of Elastic Cloud routing. See Open Question 6 for the one known gap in this heuristic (non-standard network routing that never surfaces the proxy headers) and its `skip_topology_check` escape hatch.

6. **Topology opt-in override -- resolved.** Decision 7 specifies a **fail-open** fallback when detection is inconclusive (proceed rather than block), but the detection heuristic implemented for Decision 7 (Kibana's `build_flavor` plus the `X-Found-Handling-Cluster`/`X-Found-Handling-Instance` Elastic Cloud edge-proxy headers on `GET /api/status` -- see `internal/fleet/agentlesspolicy/topology.go`) has no "inconclusive" bucket for a well-formed 200 response: absence of both signals is treated as a *positive* self-managed classification with no escape hatch. This is a real gap for a narrow case -- a genuine Elastic Cloud Hosted or Serverless deployment with non-standard network routing (e.g. a PrivateLink setup that does not route through Elastic's public edge proxy) that never emits the proxy headers would be permanently blocked with no Terraform-native workaround. (Elastic Cloud Enterprise, by contrast, does NOT need this override: agentless integrations run on Elastic's own SaaS agentless-compute plane, which does not exist in a self-hosted ECE install, so ECE is genuinely unsupported and is correctly blocked by the same fail-closed path as on-prem self-managed.)

   The resource now exposes an explicit `skip_topology_check` Optional bool schema attribute (default `false`) as this opt-in, exactly as anticipated above: when `true`, `createAgentlessPolicy` (`internal/fleet/agentlesspolicy/create.go`) skips calling `checkDeploymentTopology` entirely -- it is not merely ignored, the live `GET /api/status` probe it would make is not issued at all. It is documented as escape-hatch only ("Use only if you are certain this is running against a supported Elastic Cloud Hosted or Serverless deployment and the automatic detection is producing a false positive") and does not weaken the fail-closed path when unset, nor does it affect version gating (Kibana 9.3.0+ is enforced by a separate, unconditional check). It is client-side only: never sent to any API and never read back from state (same treatment as `force`/`force_delete`). See specs/fleet-agentless-policy/spec.md's "Deployment topology preflight check" requirement (`skip_topology_check` scenario) and "Schema attributes" requirement, and `internal/fleet/agentlesspolicy/create_test.go`'s `TestCreateAgentlessPolicy_topologyGatesFleetCall` for the test coverage (self-managed-shaped fake Kibana with `skip_topology_check=true` proceeds to the fleet POST; the same shape with the flag unset/false does not; a confirmed cloud-hosted shape proceeds regardless of the flag).
