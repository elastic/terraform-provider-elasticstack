## Why

`elasticstack_kibana_space` fails when targeting the default Kibana space (`space_id = "default"`):

1. **Create (409 Conflict)**: `POST /api/spaces/space` is rejected because the default space already exists on every Kibana deployment. The current error message is an opaque "409 Conflict" with no guidance on how to recover.
2. **Destroy (400 Bad Request)**: `DELETE /api/spaces/space/default` is always rejected by Kibana because the default space is protected and cannot be deleted. This causes `terraform destroy` to fail unconditionally.

The inability to delete the default Kibana space is a hard platform invariant — no version of Kibana will ever allow it. Practitioners managing the default space need a way to configure it via Terraform without hitting these errors.

## What Changes

- **Destroy**: Add a guard in `deleteSpace` (`internal/kibana/spaces/delete.go`) that detects `resourceID == "default"` and returns without calling the API, emitting a `tflog.Warn` to surface the skip to operators.
- **Create 409**: Add an explicit `http.StatusConflict` handler in `kibanaoapi.CreateSpace` (`internal/clients/kibanaoapi/spaces.go`) that returns an actionable diagnostic pointing the user to `terraform import` instead of the opaque HTTP error.
- **Acceptance test**: Add `TestAccResourceSpace_DefaultSpace` in `internal/kibana/spaces/acc_test.go` that imports the default space, verifies an update succeeds, and confirms `terraform destroy` completes without error. The test uses no `solution` attribute so it runs on all supported stack versions.

## Capabilities

### New Capabilities

- _(none)_

### Modified Capabilities

- `kibana-space`: extend destroy and create-time behavior so the default Kibana space can be managed by Terraform via import, updated without errors, and removed from state cleanly on `terraform destroy`.

## Impact

- Specs: delta spec under `openspec/changes/kibana-space-default-space-lifecycle/specs/kibana-space/spec.md`.
- Provider behavior: `internal/kibana/spaces/delete.go` and `internal/clients/kibanaoapi/spaces.go`.
- Acceptance tests: `internal/kibana/spaces/acc_test.go` and a new test fixture at `internal/kibana/spaces/testdata/TestAccResourceSpace_DefaultSpace/default_space/main.tf`.
- No schema surface changes.
