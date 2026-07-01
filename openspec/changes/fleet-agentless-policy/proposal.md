## Why

Fleet agentless policies enable cloud-hosted data collection without deploying Elastic Agent on the target host — Kibana provisions a hidden managed agent policy and a package policy atomically, then hosts the agent runtime itself. As agentless adoption grows (Cloud Security Posture Management, SaaS integrations, Cloud Asset Discovery), teams need to provision agentless policies alongside their standard agent policies in Terraform. Without this resource, agentless policies become a manual step in otherwise automated provisioning workflows.

This change closes the Fleet agentless policy gap called out in [`elastic/kibana#260388`](https://github.com/elastic/kibana/issues/260388) and resolves [`#2121`](https://github.com/elastic/terraform-provider-elasticstack/issues/2121).

## What Changes

Two-phase delivery in a single OpenSpec change:

**Phase 1 — Refactor (behaviour-preserving):** Extract the shared inputs/streams/vars typed modeling from `internal/fleet/integration_policy/` into a new shared package (`internal/fleet/policyshape/`). This includes `InputType`, `InputsType`, `VarsJsonType`, defaults merging, canonical JSON normalization, and secret helpers. The existing `elasticstack_fleet_integration_policy` resource is migrated to import from the shared package with no user-visible schema change. Full acceptance-test parity gates entry to Phase 2.

**Phase 2 — New resource:** Add `elasticstack_fleet_agentless_policy` with full lifecycle management:

- **Create** → `POST /api/fleet/agentless_policies` (bundled: hidden agent policy + package policy; response `id` is the package policy ID)
- **Read** → `GET /api/fleet/package_policies/{id}` (no dedicated agentless GET endpoint; the underlying package policy is always readable)
- **Update** → Hybrid: `PUT /api/fleet/package_policies/{id}` for an allowlist of mutable fields (`description`, `vars`, `inputs`, `global_data_tags`, `additional_datastreams_permissions`, `var_group_selections`, `package.title`); `RequiresReplace` on identity/structural fields (`name`, `package.name`, `package.version`, `namespace`, `cloud_connector.*`, `policy_template`). `force`, `force_delete`, and `create_dataset_templates` are create-only (sent on create, not read back, not on PUT, not `RequiresReplace`).
- **Delete** → `DELETE /api/fleet/agentless_policies/{id}` with optional `force=true`

Additional capabilities in this change:

- Resource gated behind minimum Kibana 9.3.0 (experimental API).
- Preflight check refuses self-managed stacks with a clear error directing users to Elastic Cloud Hosted or Serverless.
- Import via the composite `<space_id>/<policy_id>` form (or bare `<policy_id>` for the default space).
- Cloud connector referenced by raw ID string — no hard dependency on the in-flight `fleet-cloud-connector` change.

## Capabilities

### New Capabilities

- `fleet-agentless-policy`: Defines the schema and runtime behaviour of the `elasticstack_fleet_agentless_policy` resource, including the hybrid CRUD lifecycle (POST agentless endpoint for create, package_policy endpoints for read/update, agentless endpoint for delete), RequiresReplace/in-place-update field partitioning, version gating, and deployment-topology preflight check.
- `fleet-policyshape`: A reusable provider package (`internal/fleet/policyshape/`) containing the shared typed modeling for Fleet package policy inputs, streams, and vars — extracted from `integration_policy` and structured for adoption by both `integration_policy` and `agentless_policy`.

### Modified Capabilities

- `fleet-integration-policy`: Phase 1 refactor migrates the existing `elasticstack_fleet_integration_policy` resource to import `InputType`, `InputsType`, `VarsJsonType`, defaults merging, and secret helpers from `internal/fleet/policyshape/` instead of defining them inline. No user-visible schema change; acceptance tests must maintain full parity.

## Impact

- **New code**: `internal/fleet/policyshape/` (shared package), `internal/fleet/agentlesspolicy/` (resource), `internal/clients/fleet/agentless_policy.go` (thin client wrappers).
- **Modified code**: `internal/fleet/integration_policy/` migrated to import from `policyshape/`; provider entrypoint wired for new resource.
- **New docs/examples**: `docs/resources/fleet_agentless_policy.md`, `examples/resources/elasticstack_fleet_agentless_policy/`.
- **Generated clients**: `generated/kbapi` already includes `PostFleetAgentlessPolicies` and `DeleteFleetAgentlessPoliciesPolicyid`. The package_policy read/update endpoints (`GetFleetPackagePoliciesPackagepolicyid`, `PutFleetPackagePoliciesPackagepolicyid`) are also present. No regeneration required for Phase 2.
- **Acceptance test infra**: requires Kibana ≥ 9.3 on Elastic Cloud Hosted or Serverless; skip-gating uses the existing version-check pattern.
- **Backward compatibility**: additive only — no breaking changes to existing resources or data sources. Phase 1 refactor is behaviour-preserving.
