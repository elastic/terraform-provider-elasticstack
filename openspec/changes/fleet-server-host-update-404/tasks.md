## 1. Schema fix

- [ ] 1.1 In `internal/fleet/serverhost/schema.go`, add `PlanModifiers` to the `host_id` `schema.StringAttribute`:
  ```go
  PlanModifiers: []planmodifier.String{
      stringplanmodifier.UseStateForUnknown(),
      stringplanmodifier.RequiresReplace(),
  },
  ```
  Ensure the required imports (`"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"` and `"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"`) are added.
- [ ] 1.2 Verify that no other code change is needed: the `Update` handler at `internal/fleet/serverhost/update.go` reads `planModel.HostID.ValueString()` — with `UseStateForUnknown()` in place the plan value will now carry the prior-state UUID correctly.

## 2. Acceptance test extension

- [ ] 2.1 Locate the acceptance test file for `elasticstack_fleet_server_host` (likely `internal/fleet/serverhost/acc_test.go` or similar) and find `TestAccResourceFleetServerHost_computedID`.
- [ ] 2.2 Append an UPDATE `resource.TestStep` to the test that changes `name` or `hosts` while omitting `host_id` from config. Assert that the apply succeeds and that `host_id` is still set to a non-empty value in state after the update.
- [ ] 2.3 Confirm the new step's config does not set `host_id`, so it exercises the `UseStateForUnknown()` path specifically.

## 3. Validation and cleanup

- [ ] 3.1 Run `make build` — confirm the provider compiles with no errors.
- [ ] 3.2 Run `make check-lint` — fix any linter findings.
- [ ] 3.3 Run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate fleet-server-host-update-404 --type change` — confirm the change validates.
- [ ] 3.4 Run the targeted acceptance test (`go test -run TestAccResourceFleetServerHost_computedID ./internal/fleet/serverhost/... -v`) against a real Fleet-enabled Kibana per `dev-docs/high-level/testing.md`.
- [ ] 3.5 Add a CHANGELOG entry following the repo's existing format, noting the `host_id` update fix and the `RequiresReplace` behavior change for explicit `host_id` edits.
