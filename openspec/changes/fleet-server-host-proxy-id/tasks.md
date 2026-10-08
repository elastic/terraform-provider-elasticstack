## 1. Schema and model

- [ ] 1.1 Add `proxy_id` (`schema.StringAttribute{Optional: true}`, no Computed/plan-modifiers) to `internal/fleet/serverhost/schema.go`, matching `agentdownloadsource/schema.go`'s `proxy_id` attribute and description style
- [ ] 1.2 Add `ProxyID types.String \`tfsdk:"proxy_id"\`` to `serverHostModel` in `internal/fleet/serverhost/models.go`
- [ ] 1.3 Populate `m.ProxyID = types.StringPointerValue(data.ProxyId)` in `populateFromAPI`

## 2. Create/update body mapping

- [ ] 2.1 Set `body.ProxyId = m.ProxyID.ValueStringPointer()` in `toAPICreateModel`
- [ ] 2.2 Extract or duplicate `agentdownloadsource.proxyIDForUpdate(plan, prior types.String) *string` (with its doc comment on the `omitempty` interaction) for use by `serverhost`; prefer extracting to a shared location (e.g. `internal/fleet`) over duplicating if it does not disturb `agentdownloadsource`'s existing call sites or tests
- [ ] 2.3 Grow `serverHostModel.toAPIUpdateModel`'s signature to `toAPIUpdateModel(ctx context.Context, prior serverHostModel)` and set `body.ProxyId` via the helper from 2.2, called with `(m.ProxyID, prior.ProxyID)`
- [ ] 2.4 Update `updateServerHost` in `internal/fleet/serverhost/update.go` to pass `req.Prior` through to `req.Plan.toAPIUpdateModel(ctx, ...)`
- [ ] 2.5 Update/add unit test coverage in `internal/fleet/serverhost/models_test.go` for: create with `proxy_id` set; update with `proxy_id` set→set (unchanged), set→different value, set→unset (expect `""` sent), never-set→still-unset (expect `nil` sent)

## 3. Documentation and examples

- [ ] 3.1 Update `examples/resources/elasticstack_fleet_server_host/resource.tf` to add an `elasticstack_fleet_proxy` resource and wire its `proxy_id` into `elasticstack_fleet_server_host.proxy_id`, matching the issue's HCL snippet
- [ ] 3.2 Regenerate provider docs (`docs/resources/fleet_server_host.md`) via the existing `make` docs target so the new attribute description renders
- [ ] 3.3 Add a CHANGELOG entry following the repo's existing format

## 4. Acceptance tests

- [ ] 4.1 Add `TestAccResourceFleetServerHost_ProxyID` to `internal/fleet/serverhost/acc_test.go`, mirroring `TestAccResourceFleetAgentDownloadSource_ProxyID`: a `with_proxy` step asserting `proxy_id` is set and matches the wired `elasticstack_fleet_proxy.test.proxy_id`, followed by a `without_proxy` step asserting `proxy_id` is unset (`TestCheckNoResourceAttr`)
- [ ] 4.2 Add an additional step (or extend 4.1) that updates `proxy_id` from one real proxy to a second real proxy, to exercise the set→different-value update path
- [ ] 4.3 Add `testdata/TestAccResourceFleetServerHost_ProxyID/with_proxy/main.tf` and `.../without_proxy/main.tf` fixtures, following the layout of `internal/fleet/agentdownloadsource/testdata/TestAccResourceFleetAgentDownloadSource_ProxyID/`
- [ ] 4.4 During implementation, run this test against the stack version matrix available in CI/local dev to resolve the open question on whether `PUT /fleet/fleet_server_hosts/{itemId}` treats an omitted `proxy_id` as "leave unchanged" the same way `agent_download_sources` does; adjust Decision 4 in design.md if the live behavior differs
- [ ] 4.5 During implementation, confirm whether `proxy_id` requires a minimum stack version newer than the resource's existing `8.6.0` floor; add a dedicated `entitycore.VersionRequirement` gate only if the acceptance run against the minimum-supported stack shows it's needed

## 5. Validation and cleanup

- [ ] 5.1 Run `make build` and `make check-lint` — fix any issues
- [ ] 5.2 Run `make check-openspec` — confirm this change validates
- [ ] 5.3 Run the full unit test suite for `internal/fleet/serverhost`
- [ ] 5.4 Run the new and existing `internal/fleet/serverhost` acceptance tests against a real Kibana/Fleet stack matching the minimum version (per `dev-docs/high-level/testing.md`)
- [ ] 5.5 Self-review with the `requirements-verification` skill against this change's delta spec
