## 1. ES client helpers

- [ ] 1.1 Add `GetIndicesWithILMPolicy` to `internal/clients/elasticsearch/index.go` — queries `GET /_all/_settings/index.lifecycle.name?flat_settings=true`, parses response and returns index names whose `index.lifecycle.name` matches the given policy.
- [ ] 1.2 Add `ClearILMPolicyFromIndices` to `internal/clients/elasticsearch/index.go` — issues `PUT /{indices}/_settings` with `{"index.lifecycle.name": null}`.
- [ ] 1.3 Add unit tests for both helpers or verify via acceptance test coverage.

## 2. ILM resource Delete handler

- [ ] 2.1 Update `internal/elasticsearch/index/ilm/delete.go` to call `GetIndicesWithILMPolicy` before `DeleteIlm`.
- [ ] 2.2 If indices match, call `ClearILMPolicyFromIndices`; surface any diagnostic and short-circuit if the clear fails.
- [ ] 2.3 Then proceed with existing `DeleteIlm` call.
- [ ] 2.4 Verify `go build ./internal/elasticsearch/index/ilm/...` and `go vet` pass.

## 3. Acceptance test

- [ ] 3.1 Create/update `internal/fleet/integration/testdata/TestAccResourceIntegration_destroyWithILMCrossDependency/create/main.tf` (already exists from repro test).
- [ ] 3.2 Update `internal/fleet/integration/acc_test.go` `TestAccResourceIntegration_destroyWithILMCrossDependency` to expect **success** on ILM policy destroy instead of `ExpectError`, confirming the fix.
- [ ] 3.3 Add targeted acceptance test for ILM resource itself that creates a policy, an index with the policy reference, and asserts the ILM resource deletes successfully.
- [ ] 3.4 Run the test against the local stack (`make docker-fleet`) and confirm it passes.

## 4. Existing tests

- [ ] 4.1 Run `go test ./internal/elasticsearch/index/ilm/...` — all unit tests pass.
- [ ] 4.2 Run `go test ./internal/clients/elasticsearch/...` — all unit tests pass.
- [ ] 4.3 Run fleet integration acceptance tests (`TestAccResourceIntegration`, `TestAccResourceIndexTemplateIlmAttachment_fleet`) to confirm no regressions.

## 5. Sync and archive

- [ ] 5.1 Sync the delta spec changes into `openspec/specs/elasticsearch-index-lifecycle/spec.md` using the OpenSpec sync workflow.
- [ ] 5.2 Archive the change with `openspec archive change fix-ilm-delete-in-use-indices`.
