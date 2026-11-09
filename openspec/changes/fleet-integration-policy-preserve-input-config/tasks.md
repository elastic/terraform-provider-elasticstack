## 1. Fleet client

- [x] 1.1 Add generic typed-format `GetPackagePolicyTyped` and `UpdatePackagePolicyTyped` to `internal/clients/fleet/package_policy.go`, and make `GetDefendPackagePolicy` / `UpdateDefendPackagePolicy` delegate to them; verify with `go build ./...`

## 2. Snapshot and restore

- [x] 2.1 Add `internal/fleet/integration_policy/input_config.go` with `snapshotInputConfig` (inputs keyed by `type` + `policy_template`, streams additionally by `data_stream.dataset`, ambiguous identities skipped), `deepFill` (restore absent keys, including null-valued ones, and keys null now, recurse into objects, current values win), and `restoreInputConfig` (merge into the post-update typed inputs, strip `compiled_input` / `compiled_stream`, report whether anything changed)
- [x] 2.2 Add `readInputConfigSnapshot` and `restorePreservedInputConfig`, the latter issuing a typed update with only `version`, `policy_id`, `policy_ids` and `inputs`, and only when something was restored
- [x] 2.3 Wire both into `Update` in `update.go`: snapshot before the simplified update, restore after it; return errors from either step
- [x] 2.4 Count identities over all inputs and streams (with or without config) before and after the update and only snapshot/restore identities unique on both sides; treat a policy not found by the restoring read as an error

## 3. Tests

- [x] 3.1 Unit tests in `input_config_test.go` for snapshot (input, stream, empty, ambiguous, nil), deep fill (missing, null, nested, list, idempotence) and restore (APM keys restored while values written by the update are kept, no-op when nothing was lost, whole config and stream config restored, removed inputs ignored, empty snapshot)
- [x] 3.1a Unit tests for configured/unconfigured duplicate inputs and streams, identities that become ambiguous after the update, a mocked 404 on the restoring read (error, no update sent), and an empty snapshot sending no request
- [x] 3.1b Unit test that the restoring request body contains exactly `version`, `policy_id`, `policy_ids` and `inputs`
- [x] 3.1c Unit tests that nested keys with null values removed under a kept parent are restored (input and stream config), and that no restoring update is sent when a restored null cannot be carried by the typed request
- [x] 3.2 Acceptance test `TestAccResourceIntegrationPolicyPreservesInputConfig`: write input and stream config out of band through the typed API, update the policy from Terraform, assert the config is still present and the update was applied; failure messages never include config values
- [x] 3.3 Acceptance test `TestAccResourceIntegrationPolicyAPMPreservesAPIKeys`: create an APM integration policy (latest GA package version, or latest prerelease when the registry lists no GA version), assert Kibana stored both API keys, update an input var from Terraform, assert both API keys are still present
- [x] 3.4 Confirm both acceptance tests fail with the restore step disabled and pass with it enabled
- [x] 3.5 Run the existing `integration_policy` and `elastic_defend_integration_policy` acceptance suites and confirm no regressions

## 4. Documentation and validation

- [x] 4.1 Document the preserved `config` behaviour in `internal/fleet/integration_policy/resource-description.md` and regenerate docs with `make docs-generate`
- [x] 4.2 Run `make check-openspec`, unit tests, and golangci-lint and confirm they pass
