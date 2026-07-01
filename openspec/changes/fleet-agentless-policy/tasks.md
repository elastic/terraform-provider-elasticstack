## 1. Phase 1 preparation: shared modeling package

- [x] 1.1 Identify all types and functions in `internal/fleet/integration_policy/` that are candidates for extraction into the shared package: `InputType`, `InputsType`, `VarsJsonType`, and associated value types, defaults merging (`models_defaults.go`), canonical JSON normalization, and secret helpers (`secrets.go`)
- [x] 1.2 Decide final package name for the shared package (working name: `internal/fleet/policyshape/`); document the decision as a comment in `resource.go` of the new package
- [x] 1.3 Create the shared package directory and move (not copy) each candidate type/function, updating all import paths in `integration_policy/`
- [x] 1.4 Verify `go build ./...` passes with no import cycles after the extraction
- [x] 1.5 Update all unit tests in `integration_policy/` that previously tested the now-shared code to import from the new package
- [x] 1.6 Add or migrate unit tests for `VarsJsonType` normalization: semantically equivalent JSON → no diff; changed JSON → diff
- [x] 1.7 Add or migrate unit tests for defaults merging: user value overrides default; missing user value uses default
- [x] 1.8 Add or migrate unit tests for secret helpers: secret reference preserved on update; raw value does not appear in state
- [x] 1.9 **Additive schema change:** add an Optional `condition` string attribute to input and stream elements in the shared `InputType`, surfaced in both `integration_policy` and `agentless_policy`. Wire it to the API `condition` field on create/update and read it back. Non-breaking (no state upgrader); verify existing resources plan without diff.
- [x] 1.10 Run `make check-lint` and fix any linting issues from the extraction
- [x] 1.11 Run integration_policy acceptance tests to confirm Phase 1 parity: `go test -v -run TestAcc ./internal/fleet/integration_policy/ -timeout 30m`

## 2. kbapi client wrappers

- [x] 2.1 Create `internal/clients/fleet/agentless_policy.go` with thin wrappers mirroring the style of `proxy.go`:
  - `CreateAgentlessPolicy(ctx, client, spaceID, body) (*kbapi.KibanaHTTPAPIsAgentlessPolicy, diag.Diagnostics)` wrapping `PostFleetAgentlessPolicies`
  - `ReadAgentlessPolicyViaPackagePolicy(ctx, client, spaceID, policyID) (*kbapi.PackagePolicy, diag.Diagnostics)` wrapping `GetFleetPackagePoliciesPackagepolicyid` (returns nil on HTTP 404) — delegates to the existing `GetPackagePolicy` in `package_policy.go` to avoid duplicating the same generated call
  - `UpdateAgentlessPolicyViaPackagePolicy(ctx, client, spaceID, policyID, body) (*kbapi.PackagePolicy, diag.Diagnostics)` wrapping `PutFleetPackagePoliciesPackagepolicyid` — delegates to the existing `UpdatePackagePolicy`
  - `DeleteAgentlessPolicy(ctx, client, spaceID, policyID, force bool) diag.Diagnostics` wrapping `DeleteFleetAgentlessPoliciesPolicyid` (no-op on HTTP 404)
  - Note: signatures return `diag.Diagnostics` (not `error`) to match every existing Fleet client wrapper (`proxy.go`, `agent_policy.go`, `package_policy.go`); the task's pseudocode used `error` but the codebase convention is Plugin Framework diagnostics end-to-end.
- [x] 2.2 Map kbapi non-2xx responses into provider diagnostics consistently with other Fleet clients — reused `kibanaoapi.HandleMutateTypedResponse` for Create, and the package-private `handleDeleteResponse` (`internal/clients/fleet/responses.go`, backed by `diagutil.HandleStatusResponse`) for Delete, same as `proxy.go`/`agent_policy.go`.
- [x] 2.3 Add unit tests for each wrapper's error-handling paths (404 → nil/no-op, non-2xx → error) — see `internal/clients/fleet/agentless_policy_test.go`.

## 3. Resource: skeleton, model, and spike

- [x] 3.1 Create `internal/fleet/agentlesspolicy/` directory mirroring `internal/fleet/proxy/` in structure: `resource.go`, `models.go`, `schema.go`, `create.go`, `read.go`, `update.go`, `delete.go`
- [x] 3.2 Implement `models.go` with `agentlessPolicyModel`, `GetID`, `GetResourceID`, `GetSpaceID`, `GetKibanaConnection`, `GetVersionRequirements` (MinVersion: 9.3.0)
- [x] 3.3 **Spike task:** Apply a test agentless policy in a cloud-hosted test environment, then call `PUT /api/fleet/package_policies/{id}` with each candidate in-place-updatable field (`description`, `vars_json`, `inputs`, `global_data_tags`, `additional_datastreams_permissions`, `var_group_selections`, `package.title`). Document which fields Kibana actually honors and which return errors. Record findings in a comment block in `update.go`. Adjust the RequiresReplace list if the spike contradicts the design. (`create_dataset_templates` is create-only and is intentionally excluded from this PUT probe.)
  - Done against a live Kibana 9.4.3 Cloud Hosted deployment; full findings recorded in `internal/fleet/agentlesspolicy/update.go`. All 7 in-place-updatable candidates confirmed accept+persist (with a caveat on `inputs[*].enabled`, see update.go). **Contradiction found and flagged:** `name`, `namespace`, and `package.version` (and, conditionally, `package.name`) are also accepted and persisted by PUT — not API-enforced immutable. RequiresReplace partitioning was kept unchanged (see design.md Decision 3 / Open Question 1) as a deliberate Terraform-side safety choice; flagged prominently for orchestrator review.

## 4. Resource: schema

- [ ] 4.1 Implement `getSchema` in `schema.go` covering all identity attributes (`id`, `policy_id`, `name`, `description`, `namespace`, `space_ids`) with correct Optional/Computed/Required, `UseStateForUnknown`, and `RequiresReplace` plan modifiers per the spec
- [ ] 4.2 Add schema for `package` (Required object: `name` (Required, RequiresReplace), `version` (Required, RequiresReplace), `title` (Optional+Computed, in-place updatable, not RequiresReplace))
- [ ] 4.3 Add schema for `policy_template` (Optional string, `RequiresReplace`)
- [ ] 4.4 Add schema for top-level `vars_json` (via shared `VarsJsonType`), `var_group_selections`, and `inputs` (reusing `InputsType`/`InputType` from the shared package); input/stream-level vars use the `vars` attribute key (matching `integration_policy`); `inputs` is Optional+Computed with `UseStateForUnknown`
- [ ] 4.5 Add schema for `cloud_connector` (Optional object: `enabled`, `cloud_connector_id`, `name`, `target_csp`) with `RequiresReplace` on all sub-fields
- [ ] 4.6 Add schema for `global_data_tags` and `additional_datastreams_permissions` (both Optional, updatable in-place); add `create_dataset_templates` (Optional, create-only — not read back, not sent on Update, not RequiresReplace)
- [ ] 4.7 Add schema for operation flags `force` (Optional bool, create-only, not read back from API) and `force_delete` (Optional bool, not read back from API)
- [ ] 4.8 Add schema for computed fields `created_at` and `updated_at`
- [ ] 4.9 Add `kibana_connection` block following the pattern of other Fleet resources

## 5. Resource: CRUD + import

- [ ] 5.1 Implement `create.go`: compile config model to `PostFleetAgentlessPoliciesJSONRequestBody`, call `CreateAgentlessPolicy`, decode response into model, set state
- [ ] 5.2 Implement `read.go`: call `ReadAgentlessPolicyViaPackagePolicy`, handle nil (404 → remove from state), decode response into model, set state; preserve `force`, `force_delete`, and `create_dataset_templates` from plan (not returned by API)
- [ ] 5.3 Implement `update.go` (based on spike findings from 3.3): call `UpdateAgentlessPolicyViaPackagePolicy` with the in-place-updatable allowlist fields only; decode response into model; set state
- [ ] 5.4 Implement `delete.go`: call `DeleteAgentlessPolicy` with `force = force_delete`; handle 404 as no-op; surface helpful error for conflict when `force_delete = false`
- [ ] 5.5 Implement `ImportState` accepting composite `"<space_id>/<policy_id>"` form and plain `"<policy_id>"` form using `SpaceImporter`

## 6. Resource: version gating and deployment check

- [ ] 6.1 Wire `GetVersionRequirements` in `models.go` to enforce Kibana ≥ 9.3.0 using the existing `EnforceMinVersion` pattern; add a test asserting the version check fires before any API call
- [ ] 6.2 Implement the deployment topology preflight check in `create.go` that detects self-managed stacks and returns a clear diagnostic; investigate the available detection mechanism (e.g., checking `xpack.fleet.agentless.enabled` via the Fleet settings endpoint or a stack capability check) and document the chosen approach in a code comment
- [ ] 6.3 Add the "experimental" notice to the resource description in `schema.go`

## 7. Resource: registration

- [ ] 7.1 Register `agentlesspolicy.NewResource()` in the Plugin Framework provider's resource list
- [ ] 7.2 Run `make build` and confirm the resource appears in `terraform providers schema -json` output

## 8. Acceptance tests

- [ ] 8.1 Add `acc_test.go` for the full agentless policy lifecycle using the CSPM package (`cloud_security_posture`): create → read → update description → update vars → destroy; gate on Kibana ≥ 9.3 and cloud-hosted topology
- [ ] 8.2 Add coverage for import via composite `<space_id>/<policy_id>` ID
- [ ] 8.3 Add coverage for `force_delete = true`: create a policy, simulate a managed-policy conflict, verify delete with `force_delete = true` succeeds
- [ ] 8.4 Add coverage for version-skip gating: when Kibana < 9.3.0, the test is skipped (not failed)
- [ ] 8.5 Add coverage for cloud connector reference: create policy with `cloud_connector.cloud_connector_id` set to an existing connector ID, verify it round-trips
- [ ] 8.6 Add coverage for RequiresReplace fields: change `name` in config and verify the plan shows destroy+create (not in-place update)
- [ ] 8.7 Add coverage for `inputs` update in-place: change an input's `vars` and verify the plan shows an update (not destroy+create)

## 9. Documentation and examples

- [ ] 9.1 Add `examples/resources/elasticstack_fleet_agentless_policy/resource.tf` with a complete CSPM AWS example including `inputs`, `vars_json`, and `cloud_connector`
- [ ] 9.2 Add `examples/resources/elasticstack_fleet_agentless_policy/import.sh` showing both composite and plain import forms
- [ ] 9.3 Generate provider docs (`docs/resources/fleet_agentless_policy.md`) via the existing `make` target; verify the rendered description includes the experimental notice and a note about cloud-hosted/serverless requirement
- [ ] 9.4 Add a CHANGELOG entry following the repo's existing format (new resource: `elasticstack_fleet_agentless_policy`)

## 10. Validation and cleanup

- [ ] 10.1 Run `make build` — fix any compilation errors
- [ ] 10.2 Run `make check-lint` — fix any lint issues
- [ ] 10.3 Run `make check-openspec` — confirm the change validates cleanly
- [ ] 10.4 Run the full unit test suite for the new packages: `go test ./internal/fleet/agentlesspolicy/... ./internal/clients/fleet/...` plus the shared package
- [ ] 10.5 Run Phase 1 integration_policy acceptance tests one final time to confirm parity was maintained end-to-end: `go test -v -run TestAcc ./internal/fleet/integration_policy/ -timeout 30m`
- [ ] 10.6 Self-review with the `requirements-verification` skill against this change's specs
