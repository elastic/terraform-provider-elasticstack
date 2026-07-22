# Fix entity-store acceptance-test isolation and add 9.4.2 to CI matrix

## Summary

Four acceptance tests in the `elasticstack_kibana_security_entity_store` family were failing against Elastic Stack 9.4.2 with `entity_types` inconsistent-result errors and HTTP 500s.

Root-cause investigation (see issue #3952) found that the failures are **not** caused by a Kibana 9.4.2 regression or a `flattenStatus` read-reconciliation bug. They are caused by a **test-isolation gap**: all five entity-store acceptance-test packages share the hardcoded Kibana space `"default"`, and the Go test runner executes packages concurrently by default. The Security Entity Store is a **singleton per Kibana space**, so concurrent installs and uninstalls by competing packages race against each other, producing the observed errors.

PR [#4062](https://github.com/elastic/terraform-provider-elasticstack/pull/4062) (merged 2026-07-03) already added HTTP-500 retries and per-test `CleanupEntityStore` cleanup, but it left the shared-space problem intact and documented cross-package concurrency as an explicit non-goal. That non-goal assumption does not hold in practice: three of the five entity-store packages land in shard 0 of the CI 2-shard split, so they run concurrently against the same Kibana instance.

## Proposed fix

Give every entity-store acceptance test its own randomly generated Kibana space. This follows the established convention already used by `internal/kibana/tag`, `internal/kibana/osquery_pack`, and `internal/kibana/synthetics/*`.

Changes required:

1. **Go test files** (5 packages): add `accTestKibanaSpaceIDCharset` const and generate a per-test `spaceID` via `sdkacctest.RandStringFromCharSet(12, accTestKibanaSpaceIDCharset)`. Pass it via `ConfigVariables` in every `TestStep`. Update `t.Cleanup` to use the generated `spaceID` instead of `"default"`. Add `sdkacctest` import.

2. **Terraform `.tf` fixtures** (37 files across 5 packages): add a `variable "space_id"` block, an `elasticstack_kibana_space.test` resource, and wire `space_id = elasticstack_kibana_space.test.space_id` on every store/entity/link/data-source block. Multi-step tests that reuse a directory across steps already receive the same `spaceID` via `ConfigVariables`, so the space is stable throughout the test.

3. **CI matrix** (`.github/workflows/provider.yml`): add `"9.4.2"` between `"9.4.0"` and `"9.5.0-SNAPSHOT"`.

No product code changes are needed.

## Scope

All five packages that touch the entity-store singleton:
- `internal/kibana/security_entity_store`
- `internal/kibana/security_entity_store/entities`
- `internal/kibana/security_entity_store/entity`
- `internal/kibana/security_entity_store_entity_link`
- `internal/kibana/security_entity_store_resolution_group`

## Acceptance criteria (from issue #3952)

- The CI acceptance-test matrix includes `9.4.2`.
- All entity-store family acceptance tests pass against `9.4.2`.
- The race is eliminated: all five packages can be invoked together in a single `go test` command and pass reliably.
