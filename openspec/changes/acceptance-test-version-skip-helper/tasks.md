## 1. Core Helper Implementation

- [ ] 1.1 Add `Flavor` enum (`Any`, `Stateful`, `Serverless`) and `String() string` method to `internal/versionutils/testutils.go`
- [ ] 1.2 Add unexported `checkSkip(minVersion, constraints, flavor)` function that creates one ES client and evaluates version + flavor rules
- [ ] 1.3 Add exported `SkipIfUnsupported(t *testing.T, minVersion *version.Version, flavor Flavor)` helper
- [ ] 1.4 Add exported `SkipIfUnsupportedConstraints(t *testing.T, constraints version.Constraints, flavor Flavor)` helper
- [ ] 1.5 Add unit tests for `checkSkip` covering: below-min version, above-min version, serverless bypass, `FlavorStateful` on serverless, `FlavorServerless` on stateful, client-creation error
- [ ] 1.6 Run `make check-lint` and fix any issues

## 2. Representative Package Migration (Dashboard)

- [ ] 2.1 Migrate `internal/kibana/dashboard/acc_test.go` from per-step `SkipFunc` to top-level `SkipIfUnsupported(t, minDashboardAPISupport, FlavorAny)`
- [ ] 2.2 Migrate `internal/kibana/dashboard/acc_gauge_panels_test.go`
- [ ] 2.3 Migrate `internal/kibana/dashboard/acc_heatmap_panels_test.go`
- [ ] 2.4 Migrate `internal/kibana/dashboard/acc_esql_control_panels_test.go`
- [ ] 2.5 Migrate remaining `internal/kibana/dashboard/acc_*_panels_test.go` files with identical `minDashboardAPISupport` skips
- [ ] 2.6 Run `make check-lint` and targeted dashboard acceptance tests to confirm no regressions

## 3. Elasticsearch Package Migrations

- [ ] 3.1 Migrate `internal/elasticsearch/index/datastreamlifecycle/acc_test.go`
- [ ] 3.2 Migrate `internal/elasticsearch/index/ilm/acc_test.go` cold-allocate, hot-actions, warm-downsample tests
- [ ] 3.3 Migrate `internal/elasticsearch/index/template/acc_test.go` and `acc_from_sdk_test.go` template tests
- [ ] 3.4 Migrate `internal/elasticsearch/security/api_key/acc_test.go` single-version tests
- [ ] 3.5 Migrate `internal/elasticsearch/transform/transform_test.go`
- [ ] 3.6 Run `make check-lint` and targeted ES acceptance tests

## 4. Fleet Package Migrations

- [ ] 4.1 Migrate `internal/fleet/agentpolicy/acc_test.go` tests with uniform `minVersionAgentPolicy` skips
- [ ] 4.2 Migrate `internal/fleet/integration/acc_test.go` uniform-skip tests
- [ ] 4.3 Migrate `internal/fleet/integration_policy/acc_test.go` uniform-skip tests
- [ ] 4.4 Migrate `internal/fleet/output/acc_test.go`
- [ ] 4.5 Migrate `internal/fleet/proxy/acc_test.go`
- [ ] 4.6 Migrate `internal/fleet/agentdownloadsource/acc_test.go`
- [ ] 4.7 Migrate `internal/fleet/elastic_defend_integration_policy/acc_test.go`
- [ ] 4.8 Migrate `internal/fleet/enrollmenttokens/acc_test.go`
- [ ] 4.9 Migrate `internal/fleet/serverhost/acc_test.go`
- [ ] 4.10 Run `make check-lint` and targeted Fleet acceptance tests

## 5. Kibana Package Migrations

- [ ] 5.1 Migrate `internal/kibana/agentbuildertool/acc_test.go`
- [ ] 5.2 Migrate `internal/kibana/agentbuilderworkflow/acc_test.go`
- [ ] 5.3 Migrate `internal/kibana/agentbuilderagent/acc_test.go`
- [ ] 5.4 Migrate `internal/kibana/dataview/acc_test.go` uniform-skip tests
- [ ] 5.5 Migrate `internal/kibana/defaultdataview/acc_test.go`
- [ ] 5.6 Migrate `internal/kibana/maintenance_window/acc_test.go`
- [ ] 5.7 Migrate `internal/kibana/security_detection_rule/acc_test.go` uniform-skip tests
- [ ] 5.8 Migrate `internal/kibana/slo/acc_test.go` uniform-skip tests
- [ ] 5.9 Migrate `internal/kibana/synthetics/monitor/acc_test.go` uniform-skip tests
- [ ] 5.10 Migrate `internal/kibana/synthetics/parameter/acc_test.go`
- [ ] 5.11 Migrate `internal/kibana/synthetics/privatelocation/acc_test.go`
- [ ] 5.12 Migrate `internal/kibana/streams/acc_test.go`
- [ ] 5.13 Migrate `internal/kibana/security_exception_item/acc_test.go` uniform-skip tests
- [ ] 5.14 Run `make check-lint` and targeted Kibana acceptance tests

## 6. Provider and Remaining Migrations

- [ ] 6.1 Migrate `provider/provider_test.go` — `TestElasticsearchAPIKeyConnection`, `TestFleetBearerTokenConfiguration`, `TestFleetConfiguration`
- [ ] 6.2 Migrate any remaining test files with uniform per-step `SkipFunc` not covered in prior batches
- [ ] 6.3 Verify no new lint violations introduced across the full migration
- [ ] 6.4 Run a representative subset of acceptance tests across all domains to confirm no regressions

## 7. Final Verification and Archive

- [ ] 7.1 Confirm all 284 target test cases now use the top-level helper
- [ ] 7.2 Confirm the 36 non-uniform `SkipFunc` cases remain untouched (progressive gating, partial skipping, custom skips)
- [ ] 7.3 Run `make check-openspec` and `make check-lint` pass cleanly
- [ ] 7.4 Archive the change with `openspec archive acceptance-test-version-skip-helper`
