## 1. Prep and discovery

- [x] 1.1 Confirm the minimum Kibana version that ships `/api/fleet/cloud_connectors` (resource + data source); record in design.md "Open Questions" answer
- [x] 1.2 Verify `golang.org/x/crypto/bcrypt` is already a (transitive) provider dependency; if not, add it via `go get`
- [x] 1.3 Verify Plugin Framework version supports `WriteOnly` string attributes (PF ≥ 1.11); bump in `go.mod` if required
- [x] 1.4 Confirm `entitycore.KibanaResource` exposes a hook for private-state read/write during `ModifyPlan`; if not, plan the smallest extension needed (separate task group)

## 2. `internal/utils/writeonlyhash` helper

- [x] 2.1 Create package `internal/utils/writeonlyhash` with `Hasher` struct, `New(resourceTypeName string) *Hasher`, `Compute(value string) ([]byte, error)`, `Matches(value string, storedHash []byte) bool`, `PrivateStateKey(attributePath string) string`
- [x] 2.2 Implement bcrypt-backed hashing with per-resource-type salt derived deterministically from the resource type string
- [x] 2.3 Implement diagnostic-safe error returns that never include the input value
- [x] 2.4 Add unit tests covering: constructor produces per-type salt, roundtrip Matches=true, different value Matches=false, different type salts produce different hashes, errors do not leak input
- [x] 2.5 Add usage docs as Go doc comments on the exported symbols

## 3. Generated client wrappers

- [x] 3.1 Add `internal/clients/fleet/cloud_connector.go` with thin wrappers `CreateCloudConnector`, `GetCloudConnector` (read), `UpdateCloudConnector`, `DeleteCloudConnector` (force-aware), `ListCloudConnectors` — each mirroring the structure of `proxy.go` and using `spaceAwarePathRequestEditor`
- [x] 3.2 Map kbapi non-2xx responses into provider diagnostics consistently with other Fleet clients
- [x] 3.3 Translate HTTP 404 to a sentinel/nil result for Read and a no-op for Delete

## 4. Resource: skeleton + model

- [x] 4.1 Create `internal/fleet/cloudconnector/` directory mirroring `internal/fleet/proxy/`
- [x] 4.2 Implement `models.go` with `cloudConnectorModel`, `GetID`, `GetResourceID`, `GetSpaceID`, `GetKibanaConnection`, `GetVersionRequirements`
- [x] 4.3 Implement `cloudConnectorVarsElement` model covering all four union arms and the computed `secret_ref` field (the design's `secret_value_wo_version` companion is intentionally omitted — drift detection uses bcrypt hashes in private state per Decision 5)
- [x] 4.4 Implement `awsBlockModel` and `azureBlockModel` for the typed sugar
- [x] 4.5 Implement `populateFromAPI` performing the dual representation: raw `vars` map always; matching typed block populated only when (a) the modelled keys for that block are exactly the set of keys present in the API response (no extras), AND (b) the API's `cloud_provider` matches the typed block

## 5. Resource: schema

- [x] 5.1 Implement `getSchema` in `schema.go` covering identity (`id`, `cloud_connector_id`, `space_id`, `name`, `cloud_provider`, `account_type`), `force_delete`, the typed `aws` and `azure` blocks (Optional+Computed), and the `vars` map (Optional+Computed) using the exact arm-mapped shape from design.md
- [x] 5.2 Add `RequiresReplace` plan modifiers on `cloud_provider`, `cloud_connector_id`, `space_id`
- [x] 5.3 Add `UseStateForUnknown` plan modifiers where dual-population could produce informational diffs (typed block, `vars`, computed read-only fields)
- [x] 5.4 Add the per-element `ConfigValidator` for `vars` enforcing the arm-exclusivity rules and rejecting computed-only fields in config
- [x] 5.5 Add the resource-level `ConfigValidator` enforcing `ExactlyOneOf(aws, azure, vars)` and provider-block-matches-cloud_provider
- [x] 5.6 Mark `secret_value` and `aws.external_id` as `WriteOnly` and `Sensitive`

## 6. Resource: CRUD + import

- [x] 6.1 Implement `create.go` calling `POST /api/fleet/cloud_connectors`, populating dual representation from the response
- [x] 6.2 Implement `read.go` calling `GET /api/fleet/cloud_connectors/{id}`, treating 404 as removed-from-state
- [x] 6.3 Implement `update.go` calling `PUT /api/fleet/cloud_connectors/{id}` (omitting `cloudProvider`), preserving existing secret refs when secret values are not re-supplied in config
- [x] 6.4 Implement `delete.go` calling `DELETE /api/fleet/cloud_connectors/{id}` with `?force=` driven by `force_delete`, surfacing a helpful error mentioning `package_policy_count` on in-use conflicts
- [x] 6.5 Implement `ImportState` accepting the composite `"<space_id>/<cloud_connector_id>"` form
- [x] 6.6 Implement compilers: `compileAWS`, `compileAzure`, `compileVars` (config drives branch selection; plan supplies field values; prior state used to preserve secret refs on update)

## 7. Resource: write-only drift detection

- [x] 7.1 Wire `internal/utils/writeonlyhash` into the resource for each write-only attribute (`vars[*].secret_value`, `aws.external_id`, future expansion-ready)
- [x] 7.2 Implement `ModifyPlan` (or `PlanModifier`s on the relevant attributes) that reads config write-only values, looks up stored hashes from private state, compares, and marks the resource as needing update on mismatch
- [x] 7.3 Emit a plan-time warning diagnostic naming each changed write-only attribute (no values)
- [x] 7.4 On successful Create/Update, write fresh hashes to private state for every write-only attribute that is set
- [x] 7.5 Handle the imported-resource case: absence-of-hash is "no comparison possible"; first apply baselines the hash

## 8. Resource: registration

- [x] 8.1 Register `cloudconnector.NewResource()` in the Plugin Framework provider's resource list
- [x] 8.2 Confirm the resource is discoverable via `make build` and a smoke `terraform providers schema -json`

## 9. Data source

- [x] 9.1 Create `internal/fleet/cloudconnector/datasource/` (or co-located, matching existing fleet datasource layout)
- [x] 9.2 Implement schema: `space_id`, `kuery`, `page`, `per_page`, `kibana_connection`, and computed `cloud_connectors` list (excluding `vars`)
- [x] 9.3 Implement Read calling `GET /api/fleet/cloud_connectors` with the configured query parameters
- [x] 9.4 Map API items to the data source model, omitting `vars`
- [x] 9.5 Register the data source in the provider entrypoint

## 10. Acceptance tests

- [x] 10.1 Add `acc_test.go` covering full lifecycle with typed `aws` block (create → read → update name → update role_arn → destroy)
- [x] 10.2 Add coverage for typed `azure` block (create → read → update → destroy)
- [x] 10.3 Add coverage for generic `vars` block with each of the four union arms exercised at least once
- [x] 10.4 Add coverage for the dual-state-population rule: input via typed block → state has typed block + `vars`; input via `vars` matching known keys → state has `vars` + typed block; input via `vars` with an unknown extra key → state has only `vars` (typed block null)
- [x] 10.5 Add coverage for import via composite ID
- [x] 10.6 Add coverage for `force_delete`: in-use without force produces clear error; with force succeeds
- [x] 10.7 Add coverage for write-only secret drift: change `aws.external_id` in config between two applies and assert plan detects the change with warning diagnostic
- [x] 10.8 Add coverage for version gating against a too-old Kibana (skip-gated; document expected error)
- [x] 10.9 Add data source acceptance test: create N connectors, read with no kuery → all returned; read with `kuery` filtering to a single provider → only that subset returned; read with no connectors → empty list

## 11. Documentation and examples

- [x] 11.1 Add `examples/resources/elasticstack_fleet_cloud_connector/resource.tf` covering AWS, Azure, and generic `vars` usage with at least one write-only secret
- [x] 11.2 Add `examples/resources/elasticstack_fleet_cloud_connector/import.sh` showing the composite ID import form
- [x] 11.3 Add `examples/data-sources/elasticstack_fleet_cloud_connectors/data-source.tf` covering filtered and unfiltered reads
- [x] 11.4 Generate provider docs (`docs/resources/fleet_cloud_connector.md`, `docs/data-sources/fleet_cloud_connectors.md`) via the existing `make` target; verify rendered descriptions explain typed-vs-vars representation and write-only secret handling with the bcrypt-hash drift contract
- [x] 11.5 Add a CHANGELOG entry following the repo's existing format
- [x] 11.6 Mark the resource as preview/experimental in the resource docs while cloud connectors remain preview in Kibana

## 12. Validation and cleanup

- [ ] 12.1 Run `make build` and `make check-lint` — fix any issues
- [ ] 12.2 Run `make check-openspec` — confirm the change validates
- [ ] 12.3 Run targeted acceptance tests against a real Kibana matching the minimum version (per `dev-docs/high-level/testing.md`)
- [ ] 12.4 Run the full unit test suite for `internal/utils/writeonlyhash` and `internal/fleet/cloudconnector`
- [ ] 12.5 Verify generated docs render correctly and links to upstream Elastic docs resolve
- [ ] 12.6 Self-review with the `requirements-verification` skill against this change's specs
