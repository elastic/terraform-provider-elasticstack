# acceptance-test-isolation Specification

## Purpose
TBD - created by archiving change fix-entity-store-test-isolation. Update Purpose after archive.
## Requirements
### Requirement: Per-test randomised Kibana space for entity-store acceptance tests (REQ-ACC-001)

Every acceptance test in the five entity-store packages (`security_entity_store`, `security_entity_store/entities`, `security_entity_store/entity`, `security_entity_store_entity_link`, `security_entity_store_resolution_group`) SHALL generate a unique, randomly generated Kibana `space_id` per test function and create an `elasticstack_kibana_space.test` resource in the Terraform fixture.

The random `space_id` SHALL be generated with `sdkacctest.RandStringFromCharSet(12, accTestKibanaSpaceIDCharset)` where `accTestKibanaSpaceIDCharset = "abcdefghijklmnopqrstuvwxyz0123456789_-"`. The `space_id` SHALL be passed to every `TestStep` in the test via `ConfigVariables` so the space is stable throughout multi-step tests.

All store, entity, entity-link, and data-source resource blocks in the fixture SHALL reference `space_id = elasticstack_kibana_space.test.space_id` rather than the hardcoded string `"default"`. The `t.Cleanup` call to `acctest.CleanupEntityStore` SHALL use the generated `spaceID`, not `"default"`.

#### Scenario: Concurrent test packages do not collide via the singleton

- GIVEN five entity-store acceptance-test packages running concurrently against one Kibana instance
- WHEN each test generates a distinct random `space_id` and creates its own `elasticstack_kibana_space.test`
- THEN no two tests share the same entity-store singleton
- AND there are no `entity_types inconsistent result` or HTTP 500 errors from concurrent install/uninstall operations

#### Scenario: Multi-step test does not recreate its space mid-run

- GIVEN an acceptance test with two or more `TestStep`s that reuse the same `ConfigDirectory`
- WHEN the same `spaceID` value is passed via `ConfigVariables` to every step
- THEN Terraform detects no change to `space_id` between steps
- AND neither the space nor the entity store is recreated mid-test

#### Scenario: Cleanup belt-and-suspenders targets the correct space

- GIVEN an acceptance test that registers `t.Cleanup(func() { acctest.CleanupEntityStore(t, spaceID) })`
- WHEN the test fails before reaching Terraform's own destroy sequence
- THEN the cleanup function calls `CleanupEntityStore` with the generated `spaceID`, not `"default"`
- AND the entity store in the correct space is successfully uninstalled

#### Scenario: Destroy ordering prevents destroy-time 500s

- GIVEN an `elasticstack_kibana_security_entity_store` resource that references `space_id = elasticstack_kibana_space.test.space_id`
- WHEN `terraform destroy` runs
- THEN Terraform's dependency graph destroys the entity store (and waits for `not_installed`) before destroying the space
- AND no HTTP 500 errors occur due to concurrent singleton operations against a destroyed space

### Requirement: Exactly one 9.4.x GA entry in the acceptance-test CI matrix (REQ-ACC-002)

The stack-version list pinned at `.github/versions/acceptance-test-matrix.json` (consumed by `.github/workflows/provider.yml`'s acceptance-test matrix) SHALL include exactly one `9.4.x` GA entry — the latest pullable `9.4` patch — and SHALL NOT include more than one `9.4.x` patch at a time. The latest pullable patch is the latest published `9.4` patch whose Elasticsearch, Kibana, and applicable Agent image manifests all resolve. When those manifests are unavailable, the generator's image-probe fallback retains the previously pinned `9.4` patch without failing the run; that retained pin still satisfies this requirement. This requirement is a version-range invariant, not a pin to a specific patch string, so an automated `9.4` patch bump satisfies it without a spec change.

#### Scenario: Pinned list has exactly one latest 9.4.x GA entry

- **GIVEN** the pinned versions artifact after a successful version-matrix computation
- **WHEN** the `9.4.x` entries are inspected
- **THEN** there SHALL be exactly one `9.4.x` GA version
- **AND** that version SHALL be the latest pullable `9.4` patch from the generator's GA catalog (the latest published patch whose compose-stack manifests resolve, or the previously pinned `9.4` patch when image-probe fallback retains it)

#### Scenario: Automated patch bump keeps exactly one 9.4.x entry

- **GIVEN** the pinned versions artifact contains a `9.4.x` GA entry
- **WHEN** the version-matrix generator (`ci-version-matrix-generation` capability) updates that entry to a newer published patch of the same minor
- **THEN** the pinned versions artifact SHALL still contain exactly one `9.4.x` entry
- **AND** all other version entries SHALL continue to behave as before

#### Scenario: Two 9.4.x patches cannot both be present

- **GIVEN** the pinned versions artifact is inspected at any point in time
- **WHEN** more than one `9.4.x` version string is present
- **THEN** this requirement is violated, regardless of which specific patch numbers are involved

