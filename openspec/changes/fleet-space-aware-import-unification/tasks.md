## 1. Shared SpaceImporter

- [ ] 1.1 Create `internal/fleet/space_importer.go`: define `SpaceImporter` struct with `idFields []path.Path`, `NewSpaceImporter(fields ...path.Path) *SpaceImporter`, and `ImportState` method that parses composite IDs via `clients.CompositeIDFromStrFw`, sets each `idField` to the resource ID, and sets `space_ids` when a space prefix is present
- [ ] 1.2 Add unit tests for `SpaceImporter.ImportState` in `internal/fleet/space_importer_test.go`: composite ID sets resource ID and space_ids; plain ID sets resource ID and leaves space_ids nil; multiple idFields all receive the resource ID

## 2. Bug Fix — fleet_output

- [ ] 2.1 Embed `*SpaceImporter` in `outputResource`, wire `NewSpaceImporter(path.Root("output_id"))` in `newOutputResource()`, remove the explicit `ImportState` method
- [ ] 2.2 Add acceptance test `TestAccResourceFleetOutput_importFromSpace` in `internal/fleet/output/acc_test.go` mirroring the pattern from `TestAccResourceIntegrationPolicy_importFromSpace`: create a Kibana space, deploy an output into it, import using composite ID, verify `space_ids` is populated

## 3. Bug Fix — fleet_server_host

- [ ] 3.1 Embed `*SpaceImporter` in `serverHostResource`, wire `NewSpaceImporter(path.Root("host_id"))` in `newServerHostResource()`, remove the explicit `ImportState` method
- [ ] 3.2 Add acceptance test `TestAccResourceFleetServerHost_importFromSpace` in `internal/fleet/serverhost/acc_test.go` mirroring the same pattern

## 4. Migration — existing resources

- [ ] 4.1 Migrate `fleet_agent_policy`: embed `*SpaceImporter` wired to `path.Root("policy_id")`, remove the bespoke `ImportState` method from `internal/fleet/agentpolicy/resource.go`
- [ ] 4.2 Migrate `fleet_integration_policy`: embed `*SpaceImporter` wired to `path.Root("policy_id")`, remove the bespoke `ImportState` method from `internal/fleet/integration_policy/resource.go` (removes strict empty-segment validation, aligning with standard behavior)
- [ ] 4.3 Migrate `fleet_elastic_defend_integration_policy`: embed `*SpaceImporter` wired to `path.Root("policy_id")`, remove the bespoke `ImportState` method from `internal/fleet/elastic_defend_integration_policy/resource.go` (also removes the extra `ImportStatePassthroughID` call for `id`)
- [ ] 4.4 Migrate `fleet_agent_binary_download_source`: embed `*SpaceImporter` wired to `path.Root("source_id")`, remove the bespoke `ImportState` method and `setImportStateAttributes` helper from `internal/fleet/agentdownloadsource/resource.go` (removes hardcoded `"default"` fallback and `strings.SplitN` usage)

## 5. Spec Updates

- [ ] 5.1 Verify `make check-openspec` passes with the four delta specs in this change

## 6. Verification

- [ ] 6.1 `make build` passes
- [ ] 6.2 `make lint` passes (check for unused imports left behind after `ImportState` method removals)
- [ ] 6.3 Existing import acceptance tests for `fleet_agent_policy`, `fleet_integration_policy`, `fleet_elastic_defend_integration_policy`, and `fleet_agent_binary_download_source` continue to pass
