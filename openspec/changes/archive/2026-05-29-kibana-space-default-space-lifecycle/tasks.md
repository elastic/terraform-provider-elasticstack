## 1. Spec

- [x] 1.1 Validate the change with `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate kibana-space-default-space-lifecycle --type change`.
- [x] 1.2 Sync or archive the delta into `openspec/specs/kibana-space/spec.md` after implementation is verified.

## 2. Destroy Guard for Default Space

- [x] 2.1 In `internal/kibana/spaces/delete.go`, add a guard at the top of `deleteSpace` that returns early when `resourceID == "default"`, emitting `tflog.Warn(ctx, "default Kibana space cannot be deleted; removing from Terraform state only")`.
- [x] 2.2 Add `"github.com/hashicorp/terraform-plugin-log/tflog"` to imports in `internal/kibana/spaces/delete.go`.
- [x] 2.3 Verify that `DELETE /api/spaces/space/default` is not called when `space_id = "default"` and `terraform destroy` runs.

## 3. Actionable Create 409 Diagnostic

- [x] 3.1 In `internal/clients/kibanaoapi/spaces.go`, in `CreateSpace`, add an explicit `http.StatusConflict` case after checking the API error and before calling `HandleMutateTypedResponse`. The case SHALL return an error diagnostic of the form: `"Space %q already exists. To manage an existing Kibana space with Terraform, import it first:\n\n    terraform import elasticstack_kibana_space.<NAME> %s"`.
- [x] 3.2 Add unit coverage for the 409 path in `internal/clients/kibanaoapi` (for example, a `CreateSpace` test that serves HTTP 409 and asserts the diagnostic names the space id and includes `terraform import elasticstack_kibana_space.<NAME> <space_id>`).
- [x] 3.3 Confirm `"fmt"` and `"net/http"` are present in imports in `internal/clients/kibanaoapi/spaces.go` (add if missing).

## 4. Acceptance Test

- [x] 4.1 Add `TestAccResourceSpace_DefaultSpace` to `internal/kibana/spaces/acc_test.go`. The test SHALL:
  - Use `acctest.NamedTestCaseDirectory("default_space")` as the `ConfigDirectory`.
  - In step 1: import the default space using `ResourceName: "elasticstack_kibana_space.default"`, `ImportState: true`, and `ImportStateId: "default"`.
  - In step 2: apply the fixture config and check `space_id == "default"` and `name == "Default"`.
  - Use no `CheckDestroy` (the default space is never deleted).
  - Use no `solution` attribute in the fixture (ungated, runs on all stack versions).
- [x] 4.2 Create the test fixture at `internal/kibana/spaces/testdata/TestAccResourceSpace_DefaultSpace/default_space/space.tf` with:
  ```hcl
  provider "elasticstack" {
    kibana {}
  }

  resource "elasticstack_kibana_space" "default" {
    space_id    = "default"
    name        = "Default"
    description = "This is your default space!"
  }
  ```

## 5. Verification

- [x] 5.1 Run `make build` to confirm the provider compiles cleanly after changes.
- [x] 5.2 Run targeted Go unit tests for changed packages: `go test ./internal/kibana/spaces/... ./internal/clients/kibanaoapi/...`.
- [x] 5.3 Run the targeted acceptance test `TestAccResourceSpace_DefaultSpace` against a live stack (requires `TF_ACC=1` and stack environment variables per `dev-docs/high-level/testing.md`).
