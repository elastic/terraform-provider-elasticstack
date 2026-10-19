## 1. Prep

- [ ] 1.1 Probe a live 9.4.x Kibana cluster (if available) to confirm whether the main `GET /api/agent_builder/agents/{id}` returns `access_control`; record the result in design.md "Open questions" and revisit the 9.5.0 gate decision if it reliably does
- [ ] 1.2 Confirm against a live cluster what `access_mode` a newly created agent gets when `access_control` is omitted entirely; record the actual value for use in acceptance test assertions (see design.md default-mode discrepancy)

## 2. Version constants

- [ ] 2.1 Add `MinAccessControlAPIVersion = semver.Must(semver.NewVersion("9.5.0"))` to `internal/kibana/agentbuilder/versions.go`, extending the existing rollout-timeline doc comment with a `9.5.0` entry for access control

## 3. Models

- [ ] 3.1 Add `AccessControl *AgentAccessControl` (`json:"access_control,omitempty"`) to `models.Agent` in `internal/models/agent_builder.go`
- [ ] 3.2 Add `type AgentAccessControl struct { AccessMode string \`json:"access_mode"\` }` to the same file
- [ ] 3.3 Add a model type for the dedicated sub-resource `GET` response if its JSON shape differs from `AgentAccessControl` (confirm against `kbapi.GetAgentBuilderAgentsIdAccessControlResponse`); reuse `AgentAccessControl` if the shape matches

## 4. Client wrapper

- [ ] 4.1 Add `GetAgentAccessControl(ctx context.Context, client *Client, spaceID, agentID string) (*models.AgentAccessControl, diag.Diagnostics)` to `internal/clients/kibanaoapi/agentbuilder_agent.go`, wrapping `client.API.GetAgentBuilderAgentsIdAccessControlWithResponse` with `kibanautil.SpaceAwarePathRequestEditor(spaceID)`, following `GetAgent`'s structure exactly (including its `HandleGetRawResponse[...]` 404 handling)

## 5. Schema

- [ ] 5.1 Add `attrAccessControl = "access_control"` (or inline literal, matching this file's existing `attrX` constant convention) to `internal/kibana/agentbuilderagent/schema.go`
- [ ] 5.2 Add the `access_control` `schema.SingleNestedAttribute` to `getResourceSchema`: `Optional: true, Computed: true`, `PlanModifiers: []planmodifier.Object{objectplanmodifier.UseStateForUnknown()}`, nested `Attributes: map[string]schema.Attribute{"access_mode": schema.StringAttribute{Required: true, Validators: []validator.String{stringvalidator.OneOf("public", "shared", "private")}}}`
- [ ] 5.3 Add a `MarkdownDescription` documenting the enum values, the 9.5.0 minimum version, and the relinquish-not-reset removal semantics

## 6. Model struct + population

- [ ] 6.1 Add an `AccessControl types.Object \`tfsdk:"access_control"\`` field (or a typed nested struct per this package's existing conventions — check whether other nested attributes in this resource use `types.Object` directly or a generated nested model type) to `agentModel` (resource-only; not `agentBaseModel`, since this change is resource-only per design.md Non-Goals)
- [ ] 6.2 Implement a small nested model type (e.g. `agentAccessControlModel{ AccessMode types.String }`) and conversion helpers `accessControlToObjectValue(ctx, *models.AgentAccessControl) (types.Object, diag.Diagnostics)` and the reverse accessor, following the pattern already used for other nested/object fields in this codebase (see `internal/kibana/dashboard/models_access_control.go` for the closest shape, adapting for the required-field and resource-only-attribute differences noted in design.md Decision 1)
- [ ] 6.3 Implement `populateAccessControlFromAPI` on `agentModel` taking the sub-resource response (or nil) and setting `model.AccessControl` to null when nil/unsupported, or to the converted object when present

## 7. GetVersionRequirements override

- [ ] 7.1 Add a `GetVersionRequirements` method directly on `agentModel` (shadowing the embedded `agentVersionGate` one) in `internal/kibana/agentbuilderagent/models.go`: always include the existing 9.3.0 `agentbuilder.MinCoreAPIVersion` requirement, then conditionally append the 9.5.0 `agentbuilder.MinAccessControlAPIVersion` requirement when `typeutils.IsKnown(model.AccessControl) && !model.AccessControl.IsNull()`, following `internal/kibana/alertingrule/models.go`'s conditional-append pattern
- [ ] 7.2 Verify `agentDataSourceModel`'s `GetVersionRequirements` (still satisfied by the embedded `agentVersionGate`) is unaffected, since the data source does not gain `access_control` in this change

## 8. Create path

- [ ] 8.1 In `toAPICreateModel` (`internal/kibana/agentbuilderagent/models.go`), when `typeutils.IsKnown(model.AccessControl) && !model.AccessControl.IsNull()`, decode `access_mode` from the nested object and set `body.AccessControl = &struct{ AccessMode ... }{AccessMode: kbapi.PostAgentBuilderAgentsJSONBodyAccessControlAccessMode(mode)}`; omit otherwise

## 9. Update path

- [ ] 9.1 Change `toAPIUpdateModel`'s signature to accept a `sendAccessControl bool` (alongside the existing `supportsSkillIDs bool`), following design.md Decision 3
- [ ] 9.2 In `updateAgent` (`internal/kibana/agentbuilderagent/update.go`), compute `sendAccessControl := typeutils.IsKnown(req.Config.AccessControl) && !req.Config.AccessControl.IsNull()` from `req.Config`, and thread it into the `toAPIUpdateModel` closure the same way `supportsSkillIDs` is threaded today
- [ ] 9.3 In `toAPIUpdateModel`, when `sendAccessControl` is true, decode `access_mode` from the plan's `AccessControl` and set `body.AccessControl`; when false, leave `body.AccessControl` nil regardless of what the plan carries forward

## 10. Read path (refresh + post-write)

- [ ] 10.1 Identify every call site of `kibanaoapi.GetAgent` in `internal/kibana/agentbuilderagent` (the refresh/`read.go` path, and the read-after-write step invoked by `create.go`/`update.go` via the shared envelope) and add the conditional second call after each successful (non-404) main `GET`
- [ ] 10.2 Add a shared helper (e.g. `populateAccessControl(ctx, client, spaceID, agentID string, model *agentModel) diag.Diagnostics`) that: calls `client.EnforceMinVersion(ctx, agentbuilder.MinAccessControlAPIVersion)`; on unsupported-version, sets `model.AccessControl` to a null object and returns no error; on supported, calls the new `kibanaoapi.GetAgentAccessControl` wrapper and populates the field; on API error, surfaces diagnostics
- [ ] 10.3 Wire the helper into each call site identified in 10.1, ensuring the existing 404-removes-from-state handling for the main GET runs and returns before the helper is ever invoked

## 11. Acceptance tests

- [ ] 11.1 Add `minKibanaAgentBuilderAccessControlAPIVersion = agentbuilder.MinAccessControlAPIVersion` to the version-constant block at the top of `internal/kibana/agentbuilderagent/acc_test.go`, alongside `minKibanaAgentBuilderAPIVersion`/`minKibanaAgentBuilderWorkflowAPIVersion`
- [ ] 11.2 Add `TestAccResourceAgentBuilderAgentAccessControl`, gated with `versionutils.SkipIfUnsupported(t, minKibanaAgentBuilderAccessControlAPIVersion, versionutils.FlavorAny)`, covering: (a) create with `access_control` omitted — assert the live cluster's actual default (per task 1.2, not a hardcoded literal) and an empty second `terraform plan`; (b) create with `access_mode = "private"`, then update to `"shared"`, then to `"public"`; (c) `ImportState` + `ImportStateVerify: true`; (d) update removing `access_control` from config — assert state keeps reporting the last-known server mode and a following empty-plan check; (e) `ExpectError` for an invalid enum value and for `access_control = {}` (missing required `access_mode`)
- [ ] 11.3 Add a drift-detection step: after create, change the mode out-of-band via a direct API call in the test, then assert `terraform plan` reports a non-empty diff
- [ ] 11.4 Add a version-gating negative test restricted to a stack below 9.5.0 (skip-gated to not run on 9.5.0+) asserting that configuring `access_control` surfaces a clear version-requirement error and that no request to the access-control sub-resource is made; add a separate always-run compatibility check that agents still apply successfully with `access_control` omitted regardless of stack version
- [ ] 11.5 Add the corresponding `testdata`/`ConfigDirectory` fixtures (`create_access_control`, `update_access_control`, `update_no_access_control`, etc.) following this test file's existing `acctest.NamedTestCaseDirectory(...)` convention

## 12. Documentation and examples

- [ ] 12.1 Regenerate `docs/resources/kibana_agentbuilder_agent.md` via the repo's existing docs-generation `make` target once the schema `MarkdownDescription`s are final
- [ ] 12.2 Add an `access_control` example to `examples/resources/elasticstack_kibana_agentbuilder_agent/resource.tf`
- [ ] 12.3 Add a CHANGELOG entry following the repo's existing format

## 13. Validation

- [ ] 13.1 Run `make build` and `make check-lint`
- [ ] 13.2 Run `make check-openspec`
- [ ] 13.3 Run the new/updated acceptance tests from task 11 against a Kibana ≥ 9.5.0 (per `dev-docs/high-level/testing.md`), plus the version-gating negative test against a < 9.5.0 stack if one is available
- [ ] 13.4 Self-review with the `requirements-verification` skill against this change's delta spec
