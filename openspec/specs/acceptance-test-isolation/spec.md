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

### Requirement: Elastic Stack 9.4.2 included in acceptance-test CI matrix (REQ-ACC-002)

The `version` list in `.github/workflows/provider.yml` SHALL include `"9.4.2"` inserted between `"9.4.0"` and `"9.5.0-SNAPSHOT"`.

#### Scenario: Entity-store tests pass in CI against 9.4.2

- GIVEN `"9.4.2"` added to the CI matrix and per-test space isolation applied to all five packages
- WHEN the acceptance-test jobs run for the `9.4.2` matrix entry
- THEN all entity-store family tests pass
- AND the issue #3952 closure gate ("tests run successfully against that version") is satisfied

#### Scenario: Existing matrix entries are unaffected

- GIVEN the insertion of `"9.4.2"` between `"9.4.0"` and `"9.5.0-SNAPSHOT"`
- WHEN CI runs the full matrix
- THEN all other version entries continue to behave as before

