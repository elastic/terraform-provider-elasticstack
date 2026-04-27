## 1. Update ImportState implementation

- [ ] 1.1 In `internal/fleet/integration_policy/resource.go`, replace
  `resource.ImportStatePassthroughID(ctx, path.Root("policy_id"), req, resp)` with the
  composite-ID aware pattern that mirrors `agentpolicy.ImportState` (see design.md §Changes
  Required for the exact replacement).
- [ ] 1.2 Add `"github.com/elastic/terraform-provider-elasticstack/internal/clients"` to
  the imports in `resource.go`.
- [ ] 1.3 Verify the build: `make build`.

## 2. Update spec

- [ ] 2.1 In `openspec/specs/fleet-integration-policy/spec.md`, replace REQ-006 with the
  corrected import requirement as defined in the delta spec at
  `openspec/changes/fleet-integration-policy-space-import/specs/fleet-integration-policy/spec.md`.

## 3. Acceptance tests

- [ ] 3.1 In `internal/fleet/integration_policy/acc_test.go`, add a test
  `TestAccResourceIntegrationPolicy_importFromSpace` that:
  - Creates an integration policy in a named Kibana space.
  - Removes the resource from Terraform state (`ImportStateVerifyIgnore` or a destroy step
    with `forget` semantics is acceptable; a destroy-and-reimport two-step is the standard
    pattern).
  - Runs `terraform import` with the composite ID `<space_id>/<policy_id>`.
  - Asserts that after the subsequent read, `policy_id` matches the original and `space_ids`
    contains the named space.
- [ ] 3.2 Confirm that the existing `TestAccResourceIntegrationPolicy_*` import test (if
  present) continues to pass against a plain policy ID without `space_ids` set.

## 4. Verification

- [ ] 4.1 `make build` — provider compiles without errors.
- [ ] 4.2 `go test ./internal/fleet/integration_policy/... -v` — unit tests pass.
- [ ] 4.3 Acceptance tests (requires live Kibana with spaces enabled):
  `go test ./internal/fleet/integration_policy/... -v -count=1 -run TestAccResourceIntegrationPolicy_importFromSpace`.
- [ ] 4.4 `make check-openspec` — spec validation passes.
