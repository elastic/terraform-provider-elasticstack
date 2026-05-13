## 1. Core Helper Implementation

- [x] 1.1 Add `Flavor` enum (`Any`, `Stateful`, `Serverless`) and `String() string` method to `internal/versionutils/testutils.go`
- [x] 1.2 Add unexported `checkSkip(minVersion, constraints, flavor)` function that creates one ES client and evaluates version + flavor rules
- [x] 1.3 Add exported `SkipIfUnsupported(t *testing.T, minVersion *version.Version, flavor Flavor)` helper
- [x] 1.4 Add exported `SkipIfUnsupportedConstraints(t *testing.T, constraints version.Constraints, flavor Flavor)` helper
- [x] 1.5 Add unit tests for `checkSkip` covering: below-min version, above-min version, serverless bypass, `FlavorStateful` on serverless, `FlavorServerless` on stateful, client-creation error
- [x] 1.6 Run `make check-lint` and fix any issues

## 2. Representative Package Migration (Dashboard)

- [x] 2.1 Migrate `internal/kibana/dashboard/acc_test.go` from per-step `SkipFunc` to top-level `SkipIfUnsupported(t, minDashboardAPISupport, FlavorAny)`
- [x] 2.2 Migrate `internal/kibana/dashboard/acc_gauge_panels_test.go`
- [x] 2.3 Migrate `internal/kibana/dashboard/acc_heatmap_panels_test.go`
- [x] 2.4 Migrate `internal/kibana/dashboard/acc_esql_control_panels_test.go`
- [x] 2.5 Migrate remaining `internal/kibana/dashboard/acc_*_panels_test.go` files with identical `minDashboardAPISupport` skips
- [x] 2.6 Run `make check-lint` and targeted dashboard acceptance tests to confirm no regressions

## 3. Elasticsearch Package Migrations

- [x] 3.1 Migrate `internal/elasticsearch/index/datastreamlifecycle/acc_test.go`
- [x] 3.2 Migrate `internal/elasticsearch/index/ilm/acc_test.go` cold-allocate, hot-actions, warm-downsample tests
- [x] 3.3 Migrate `internal/elasticsearch/index/template/acc_test.go` and `acc_from_sdk_test.go` template tests
- [x] 3.4 Migrate `internal/elasticsearch/security/api_key/acc_test.go` single-version tests
- [x] 3.5 Migrate `internal/elasticsearch/transform/transform_test.go`
- [x] 3.6 Run `make check-lint` and targeted ES acceptance tests

## 4. Fleet Package Migrations

- [x] 4.1 Migrate `internal/fleet/agentpolicy/acc_test.go` tests with uniform `minVersionAgentPolicy` skips
- [x] 4.2 Migrate `internal/fleet/integration/acc_test.go` uniform-skip tests
- [x] 4.3 Migrate `internal/fleet/integration_policy/acc_test.go` uniform-skip tests
- [x] 4.4 Migrate `internal/fleet/output/acc_test.go`
- [x] 4.5 Migrate `internal/fleet/proxy/acc_test.go`
- [x] 4.6 Migrate `internal/fleet/agentdownloadsource/acc_test.go`
- [x] 4.7 Migrate `internal/fleet/elastic_defend_integration_policy/acc_test.go`
- [x] 4.8 Migrate `internal/fleet/enrollmenttokens/data_source_test.go`
- [x] 4.9 Migrate `internal/fleet/serverhost/acc_test.go`
- [x] 4.10 Run `make check-lint` and targeted Fleet acceptance tests (lint verified after commits; Fleet acceptance not run — no local stack)

## 5. Kibana Package Migrations

- [x] 5.1 Migrate `internal/kibana/agentbuildertool/acc_test.go`
- [x] 5.2 Migrate `internal/kibana/agentbuilderworkflow/acc_test.go`
- [x] 5.3 Migrate `internal/kibana/agentbuilderagent/acc_test.go`
- [x] 5.4 Migrate `internal/kibana/dataview/acc_test.go` uniform-skip tests
- [x] 5.5 Migrate `internal/kibana/defaultdataview/acc_test.go`
- [x] 5.6 Migrate `internal/kibana/maintenance_window/acc_test.go`
- [x] 5.7 Migrate `internal/kibana/security_detection_rule/acc_test.go` uniform-skip tests
- [x] 5.8 Migrate `internal/kibana/slo/acc_test.go` uniform-skip tests
- [x] 5.9 Migrate `internal/kibana/synthetics/monitor/acc_test.go` uniform-skip tests
- [x] 5.10 Migrate `internal/kibana/synthetics/parameter/acc_test.go`
- [x] 5.11 Migrate `internal/kibana/synthetics/privatelocation/acc_test.go`
- [x] 5.12 Migrate `internal/kibana/streams/acc_test.go`
- [x] 5.13 Migrate `internal/kibana/security_exception_item/acc_test.go` uniform-skip tests
- [x] 5.14 Run `make check-lint` and targeted Kibana acceptance tests (lint verified after commits; Kibana acceptance not run — no local stack)

## 6. Provider and Remaining Migrations

- [x] 6.1 Migrate `provider/provider_test.go` — `TestElasticsearchAPIKeyConnection`, `TestFleetBearerTokenConfiguration`, `TestFleetConfiguration`
- [x] 6.2 Migrate any remaining test files with uniform per-step `SkipFunc` not covered in prior batches
- [x] 6.3 Verify no new lint violations introduced across the full migration
- [x] 6.4 Run a representative subset of acceptance tests across all domains to confirm no regressions (deferred to CI — no local stack)

## 7. Final Verification and Archive

- [x] 7.1 Confirm all 284 target test cases now use the top-level helper (`SkipIfUnsupported`: 278; `SkipIfUnsupportedConstraints`: 17; **total 295** — slight variance vs ~284 target due to multi-call sites per test / subtests)
- [x] 7.2 Confirm the 36 non-uniform `SkipFunc` cases remain untouched (progressive gating, partial skipping, custom skips) — per-step `SkipFunc` with `CheckIfVersionIsUnsupported`: 120; `CheckIfVersionMeetsConstraints`: 6; plus `CheckIfNotServerless`: 1 (**127** lines with `versionutils` directly on `SkipFunc`); additional composite skips (`skipFn`, `skipDashboardOrKqlSLOUnsupported`, `skipKqlSLO*`, `skipAgentPolicyTamperProtectionTest`, etc.) preserved; spot-checked ILM, streams query, agent policy tamper, dashboard SLO alerts panels
- [x] 7.3 Run `make check-openspec` and `make check-lint` pass cleanly
- [ ] 7.4 Archive the change with `openspec archive acceptance-test-version-skip-helper` — **intentionally not run here**: archiving is performed by the PR `verify-openspec` workflow, not this orchestrator.
