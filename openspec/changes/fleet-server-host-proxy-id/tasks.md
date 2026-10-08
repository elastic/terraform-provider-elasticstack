## 1. Schema and model

- [x] 1.1 Add `proxy_id` (`schema.StringAttribute{Optional: true}`, no Computed/plan-modifiers) to `internal/fleet/serverhost/schema.go`, matching `agentdownloadsource/schema.go`'s `proxy_id` attribute and description style
- [x] 1.2 Add `ProxyID types.String \`tfsdk:"proxy_id"\`` to `serverHostModel` in `internal/fleet/serverhost/models.go`
- [x] 1.3 Map API `proxy_id` into state so a non-empty value is preserved and both nil and empty API values become Terraform null, matching REQ-012 and the sibling resource's hydration behavior

## 2. Create/update body mapping

- [x] 2.1 Set `body.ProxyId = m.ProxyID.ValueStringPointer()` in `toAPICreateModel`
- [x] 2.2 Extract or duplicate `agentdownloadsource.proxyIDForUpdate(plan, prior types.String) *string` (with its doc comment on the `omitempty` interaction) for use by `serverhost`; prefer extracting to a shared location (e.g. `internal/fleet`) over duplicating if it does not disturb `agentdownloadsource`'s existing call sites or tests
- [x] 2.3 Grow `serverHostModel.toAPIUpdateModel`'s signature to `toAPIUpdateModel(ctx context.Context, prior serverHostModel)` and set `body.ProxyId` via the helper from 2.2, called with `(m.ProxyID, prior.ProxyID)`
- [x] 2.4 Update `updateServerHost` in `internal/fleet/serverhost/update.go` to pass `req.Prior` through to `req.Plan.toAPIUpdateModel(ctx, ...)`
- [x] 2.5 Update/add unit test coverage in `internal/fleet/serverhost/models_test.go` for: create with `proxy_id` set; update with `proxy_id` set→set (unchanged), set→different value, set→unset (expect `""` sent), never-set→still-unset (expect `nil` sent)

## 3. Documentation and examples

- [x] 3.1 Update `examples/resources/elasticstack_fleet_server_host/resource.tf` to add an `elasticstack_fleet_proxy` resource and wire its `proxy_id` into `elasticstack_fleet_server_host.proxy_id`, matching the issue's HCL snippet
- [x] 3.2 Regenerate provider docs (`docs/resources/fleet_server_host.md`) via the existing `make` docs target so the new attribute description renders
- [x] 3.3 Add a CHANGELOG entry following the repo's existing format

## 4. Acceptance tests

- [x] 4.1 Add `TestAccResourceFleetServerHost_ProxyID` to `internal/fleet/serverhost/acc_test.go`, mirroring `TestAccResourceFleetAgentDownloadSource_ProxyID`: a `with_proxy` step asserting `proxy_id` is set and matches the wired `elasticstack_fleet_proxy.test.proxy_id`, followed by a `without_proxy` step asserting `proxy_id` is unset (`TestCheckNoResourceAttr`)
- [x] 4.2 Add an additional step (or extend 4.1) that updates `proxy_id` from one real proxy to a second real proxy, to exercise the set→different-value update path
- [x] 4.3 Add `testdata/TestAccResourceFleetServerHost_ProxyID/with_proxy/main.tf` and `.../without_proxy/main.tf` fixtures, following the layout of `internal/fleet/agentdownloadsource/testdata/TestAccResourceFleetAgentDownloadSource_ProxyID/`
- [x] 4.4 During implementation, run this test against the stack version matrix available in CI/local dev to resolve the open question on `PUT /fleet/fleet_server_hosts/{itemId}` omitted and empty-string `proxy_id` semantics; update Decision 4 and REQ-014/REQ-017 if live behavior differs
- [x] 4.5 During implementation, confirm whether `proxy_id` requires a minimum stack version newer than the resource's existing `8.6.0` floor; if so, document the compatibility requirement in the delta spec and add the corresponding `entitycore.VersionRequirement` gate

## 5. Validation and cleanup

- [x] 5.1 Run `make build` and `make check-lint` — fix any issues
- [x] 5.2 Run `make check-openspec` — confirm this change validates
- [x] 5.3 Run the full unit test suite for `internal/fleet/serverhost`
- [x] 5.4 Run the new and existing `internal/fleet/serverhost` acceptance tests against a real Kibana/Fleet stack matching the minimum version (per `dev-docs/high-level/testing.md`)
- [x] 5.5 Self-review with the `requirements-verification` skill against this change's delta spec
