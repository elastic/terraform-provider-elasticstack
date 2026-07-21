## 1. Pre-implementation spike

- [ ] 1.1 Confirm the exact Kibana version that introduced `/api/fleet/managed_integrations` by checking kibana#276925 and the shipped OAS changelog; record the version in `design.md` "Open Questions" answer and in `capabilities.go` constant
- [ ] 1.2 Review the new `KibanaHTTPAPIsManagedIntegration` response type in `generated/kbapi/oas.yaml` to confirm field mapping against the existing schema (document any discrepancies)
- [ ] 1.3 Confirm `onlyCreateOnlyFlagsChanged` short-circuit behaviour under full-replace semantics; decide whether to retain, simplify, or remove

## 2. New client — `internal/clients/fleet/managed_integration.go`

- [ ] 2.1 Create `internal/clients/fleet/managed_integration.go` with wrappers: `CreateManagedIntegration`, `ReadManagedIntegration`, `UpdateManagedIntegration`, `DeleteManagedIntegration` — all targeting `managed_integrations` endpoints
- [ ] 2.2 Wire `SpaceAwarePathRequestEditor(spaceID)` into all four wrappers (mirroring the existing agentless policy client)
- [ ] 2.3 Retain `ConflictRetry` wrapping and `?force=` mapping on Delete; map `404` to nil/no-op for Read and Delete
- [ ] 2.4 Remove the `ReadAgentlessPolicyViaPackagePolicy` and `UpdateAgentlessPolicyViaPackagePolicy` fallback wrappers (do not port them)
- [ ] 2.5 Add unit tests for the new client wrappers, including the 404 sentinel behaviour
- [ ] 2.6 Delete `internal/clients/fleet/agentless_policy.go` and its test file

## 3. Move and rename the resource package

- [ ] 3.1 Copy `internal/fleet/agentlesspolicy/` → `internal/fleet/managedintegration/`; update `package` declaration in every file
- [ ] 3.2 Rename resource `Metadata.TypeName` from `elasticstack_fleet_agentless_policy` to `elasticstack_fleet_managed_integration` in `resource.go`
- [ ] 3.3 Update all internal cross-references: `internal/fleet/integration_policy`, `internal/fleet/policyshape` doc comments, and `provider/plugin_framework.go`
- [ ] 3.4 In `provider/plugin_framework.go`: remove `agentlesspolicy.NewResource` from `experimentalResources()`; add `managedintegration.NewResource` there
- [ ] 3.5 Delete the old `internal/fleet/agentlesspolicy/` package directory

## 4. `capabilities.go` — version gate update

- [ ] 4.1 Replace the 9.3.0 `EnforceMinVersion` floor constant with the Kibana version confirmed in task 1.1
- [ ] 4.2 Keep the 9.5.0 `condition` gate unchanged
- [ ] 4.3 Update any comments referencing "agentless_policies" to reference "managed_integrations"

## 5. `schema.go` — schema changes

- [ ] 5.1 Drop `RequiresReplace` from `name` attribute (now updatable in-place)
- [ ] 5.2 Drop `RequiresReplace` from `package.version` attribute (now updatable in-place)
- [ ] 5.3 Keep `RequiresReplace` on `package.name` (immutable upstream)
- [ ] 5.4 Rewrite `global_data_tags` from `ListNestedAttribute{name, value:string}` to `MapNestedAttribute` keyed by tag name, item `{string_value: StringAttribute, number_value: Float32Attribute}`, with `stringvalidator.ConflictsWith`+`AtLeastOneOf` — mirror `internal/fleet/agentpolicy/schema.go`
- [ ] 5.5 Keep `cloud_connector` as `SingleNestedAttribute` with all sub-fields `RequiresReplace` and `name`/`target_csp` retained
- [ ] 5.6 Update schema description text: replace "agentless" → "managed integration"; keep the experimental notice and ECH/Serverless-only note
- [ ] 5.7 Update attr-types map in `models.go` to reflect the new `global_data_tags` shape

## 6. `models_convert.go` — simplification

- [ ] 6.1 Change `toCreateBody` return type to `kbapi.PostFleetManagedIntegrationsJSONRequestBody` (alias for `KibanaHTTPAPIsCreateManagedIntegrationRequest`); verify the existing construction is correct
- [ ] 6.2 Replace `populateFromPackagePolicy(*kbapi.PackagePolicy)` with `populateFromManagedIntegration(*kbapi.KibanaHTTPAPIsManagedIntegration)` for the Read path; mirror `populateFromCreateResponse` since response types are now identical
- [ ] 6.3 Delete the `PackagePolicy`-leakage normalizers: `decodeMappedInputs`, `mappedInputWire`/`mappedStreamWire`, `globalDataTagValueToString`, and the dual-shape decode branches
- [ ] 6.4 Rewrite `globalDataTagsToModel` / `globalDataTagsRawFromModel` for the new `MapNestedAttribute{name → {string_value|number_value}}` shape, using `internal/fleet/agentpolicy` conversion as reference
- [ ] 6.5 Keep `validateInputConditionSupport` and the 9.5.0 condition gate wiring unchanged
- [ ] 6.6 Keep the `mappedInputKey("<policy_template>-<input_type>")` keying logic (request/response inputs map is keyed the same way)
- [ ] 6.7 Update `models_convert_test.go` for the clean `KibanaHTTPAPIsManagedIntegration` response type and the new `global_data_tags` shape; add a number-value round-trip test case

## 7. `update.go` — full-replace simplification

- [ ] 7.1 Rewrite `buildUpdateBody` to accept only the plan (no `*kbapi.PackagePolicy` "current" parameter); build `KibanaHTTPAPIsCreateManagedIntegrationRequest` directly from plan
- [ ] 7.2 Derive `cloud_connector` `{enabled, cloud_connector_id}` from state (not plan) and always re-send it when a connector is associated; never send `name`/`target_csp` on PUT
- [ ] 7.3 Include `name` and `package.version` from plan (now updatable)
- [ ] 7.4 Replace `buildUpdateInputs`/`overlayInputFromPlan` with the `decodeInputs`+`applyCreateInputs` helpers from the create path
- [ ] 7.5 Evaluate and remove `mergeVarsInto` if full-replace makes it unnecessary
- [ ] 7.6 Re-evaluate `onlyCreateOnlyFlagsChanged` short-circuit per task 1.3 decision
- [ ] 7.7 Update `update_test.go` to cover: in-place `name` change, in-place `package.version` change, cloud_connector re-sent from state on update, full-replace body content

## 8. `create.go` / `read.go` / `delete.go`

- [ ] 8.1 `create.go`: swap the single client call to `CreateManagedIntegration`; response handling targets `KibanaHTTPAPIsManagedIntegration` (no change to response handling logic)
- [ ] 8.2 `read.go`: swap the single client call to `ReadManagedIntegration`; drop the `package_policies` fallback entirely; preserve create-only flags (`force`, `create_dataset_templates`, `skip_topology_check`) and `cloud_connector.name`/`target_csp` from prior state as today
- [ ] 8.3 `delete.go`: swap the single client call to `DeleteManagedIntegration`; `force_delete` → `?force=` mapping unchanged

## 9. Tests — unit and model conversion

- [ ] 9.1 Update `kbapi_roundtrip_test.go` response-shape assertions for the clean `KibanaHTTPAPIsManagedIntegration` type (no leaked PackagePolicy fields)
- [ ] 9.2 Rewrite `global_data_tags` unit tests for the `Map + string_value/number_value` shape
- [ ] 9.3 Add update unit tests: in-place `name` and `package.version` changes; cloud_connector re-sent from state; full-replace PUT body structure
- [ ] 9.4 Ensure `create_test.go` and `delete_test.go` use the new client wrappers
- [ ] 9.5 Update `enabled_convergence_test.go` and `condition_test.go` for the renamed package (mechanical)

## 10. Examples and documentation

- [ ] 10.1 Rename `examples/resources/elasticstack_fleet_agentless_policy/` → `examples/resources/elasticstack_fleet_managed_integration/`; update the resource type name in all `.tf` files
- [ ] 10.2 Add an example demonstrating in-place `package.version` update
- [ ] 10.3 Run the provider documentation generator (`make generate`) to produce `docs/resources/fleet_managed_integration.md`; delete `docs/resources/fleet_agentless_policy.md`
- [ ] 10.4 Verify the generated docs include the experimental notice and the ECH/Serverless-only note

## 11. Acceptance tests

- [ ] 11.1 Move and rename the acceptance test files from `agentlesspolicy` to `managedintegration`; update composite-ID and resource-type strings
- [ ] 11.2 Rename `testdata/TestAccResourceAgentlessPolicy*` directories to `TestAccResourceManagedIntegration*`; update fixture HCL to use the new resource type
- [ ] 11.3 Add an acceptance test for in-place `package.version` bump (the headline new capability)
- [ ] 11.4 Add an acceptance test for in-place `name` change (no replacement)
- [ ] 11.5 Update `global_data_tags` acceptance test to the new `Map + string_value/number_value` shape; add a number-value case
- [ ] 11.6 Skip-gate acceptance tests against the new version floor established in task 1.1
- [ ] 11.7 Add a test for update with cloud_connector — assert `cloud_connector` is re-sent from state

## 12. CHANGELOG and validation

- [ ] 12.1 Add a CHANGELOG entry documenting the rename, the schema changes, and the migration path for existing users
- [ ] 12.2 Run `make build` and fix any compilation errors
- [ ] 12.3 Run `make check-lint` and fix any lint issues
- [ ] 12.4 Run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate fleet-managed-integration --type change` and resolve any reported problems
- [ ] 12.5 Run targeted unit tests for `internal/fleet/managedintegration` and `internal/clients/fleet`
