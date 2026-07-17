## 1. Fix unknown-deferral in ValidateResource

- [ ] 1.1 In `internal/kibana/security_role/validators.go`, inside the
  `ValidateResource` method, replace the current loop body (lines 62-72) with the
  unknown-aware version:
  - After the `obj, ok := elem.(types.Object)` type assertion, add a guard that
    calls `continue` when `obj.IsUnknown()`.
  - After the guard, read `featureAttr` and `baseAttr` from `obj.Attributes()`.
  - If `featureAttr.IsUnknown()` or `baseAttr.IsUnknown()`, call `continue`.
  - Only then call `kibanaPrivilegeCounts(obj)` and
    `validateKibanaPrivileges(baseLen, featureLen)`.

  The resulting loop should match the implementation sketch in `design.md`.

## 2. Unit tests

- [ ] 2.1 In `internal/kibana/security_role/` (create `validators_test.go` if it
  does not exist), add unit tests for `ValidateResource` covering:
  - A `kibana` element with a known `feature` set (non-empty) → no error.
  - A `kibana` element with a known `base` set (non-empty) → no error.
  - A `kibana` element with both `base` and `feature` absent (known empty sets) → error.
  - A `kibana` element where `feature` is an unknown set → no error (validation
    deferred).
  - A `kibana` element where `base` is an unknown set → no error (validation
    deferred).
  - A fully-unknown `types.Object` element → no error (validation deferred).

## 3. Acceptance test

- [ ] 3.1 In the existing acceptance test package for `elasticstack_kibana_security_role`
  (`internal/kibana/security_role/` or the acceptance test directory under
  `internal/kibana/`), add a test case `TestAccResourceKibanaSecurityRoleDynamicFeature`
  that uses a `dynamic "feature"` block:

  ```hcl
  locals { features = ["discover"] }

  resource "elasticstack_kibana_security_role" "dynamic_feature" {
    name = "test-acc-dynamic-feature-%[1]s"
    elasticsearch {}
    kibana {
      spaces = ["*"]
      dynamic "feature" {
        for_each = local.features
        content {
          name       = feature.value
          privileges = ["read"]
        }
      }
    }
  }
  ```

  The test must exercise plan and apply. Prior to the fix, this configuration
  errors on plan; after the fix it must plan and apply successfully.

## 4. Spec validation

- [ ] 4.1 Run
  `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate kibana-security-role-dynamic-block-validation --type change`
  and resolve any reported problems.
- [ ] 4.2 When implementation is complete, sync the delta spec or archive the
  change per the project workflow.
