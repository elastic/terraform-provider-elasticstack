## 1. Fleet client

- [x] 1.1 Add generic typed-format `GetPackagePolicyTyped` and `UpdatePackagePolicyTyped` to `internal/clients/fleet/package_policy.go`, and make `GetDefendPackagePolicy` / `UpdateDefendPackagePolicy` delegate to them; verify with `go build ./...`

## 2. Snapshot and restore

- [x] 2.1 Add `internal/fleet/integration_policy/input_config.go` with `snapshotInputConfig` (inputs keyed by `type` + `policy_template`, streams additionally by `data_stream.dataset`, ambiguous identities skipped), `deepFill` (restore absent or null keys, recurse into objects, current values win), and `restoreInputConfig` (merge into the post-update typed inputs, strip `compiled_input` / `compiled_stream`, report whether anything changed)
- [x] 2.2 Add `readInputConfigSnapshot` and `restorePreservedInputConfig`, the latter issuing a typed update with only `version` and `inputs`, and only when something was restored
- [x] 2.3 Wire both into `Update` in `update.go`: snapshot before the simplified update, restore after it; return errors from either step

## 3. Tests

- [x] 3.1 Unit tests in `input_config_test.go` for snapshot (input, stream, empty, ambiguous, nil), deep fill (missing, null, nested, list, idempotence) and restore (APM keys restored while values written by the update are kept, no-op when nothing was lost, whole config and stream config restored, removed inputs ignored, empty snapshot)
- [x] 3.2 Acceptance test `TestAccResourceIntegrationPolicyPreservesInputConfig`: write input and stream config out of band through the typed API, update the policy from Terraform, assert the config is still present and the update was applied
- [x] 3.3 Acceptance test `TestAccResourceIntegrationPolicyAPMPreservesAPIKeys`: create an APM integration policy, assert Kibana stored both API keys, update an input var from Terraform, assert both API keys are still present
- [x] 3.4 Confirm both acceptance tests fail with the restore step disabled and pass with it enabled
- [x] 3.5 Run the existing `integration_policy` and `elastic_defend_integration_policy` acceptance suites and confirm no regressions

## 4. Documentation and validation

- [x] 4.1 Document the preserved `config` behaviour in `internal/fleet/integration_policy/resource-description.md` and regenerate docs with `make docs-generate`
- [x] 4.2 Run `make check-openspec`, unit tests, and golangci-lint and confirm they pass
