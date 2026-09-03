## Why

`elasticstack_kibana_security_role` rejects configurations that use `dynamic`
blocks inside a `kibana {}` block with the error:

> Either one of the `feature` or `base` privileges must be set for kibana role!

The error is emitted during Terraform's plan phase (`ValidateResourceConfig`)
before dynamic-block `for_each` expressions are fully resolved. At that point
the `feature` attribute inside each `kibana` set element is an **unknown** set.
The resource's `configValidator.ValidateResource` calls `kibanaPrivilegeCounts`,
which treats an unknown `feature` set as length 0. Because `base` is also absent
the check fires — even though at apply time the `feature` blocks will be populated
correctly.

This regression was introduced with the Plugin Framework migration in v0.15.0 (PR
#3071). The old SDKv2 implementation validated inside the create/update handlers
(after all values were resolved), so unknown values never reached the check.

## What Changes

- In `ValidateResource` (`internal/kibana/security_role/validators.go`), add an
  early-continue inside the `kibana` element loop when the element object itself,
  or its `feature` or `base` attribute, is unknown. The constraint is still
  enforced at apply time by the existing `validateKibanaPrivileges` call inside
  `expandKibana` (`expand.go`), so no invariant is lost.
- Add an acceptance test configuration that uses a `dynamic "feature"` block to
  confirm the plan/apply cycle succeeds where it previously errored.

No schema changes, no API changes, and no changes to the API-facing validation
(which runs only after values are known) are required.

## Capabilities

### New Capabilities

- _(none)_

### Modified Capabilities

- `elasticstack_kibana_security_role`: Fix `ValidateResourceConfig` to skip
  per-element privilege validation when `feature` or `base` is unknown (deferred
  from an unresolved `dynamic` block), matching the pattern established by
  `ExactlyOneOfNestedAttrsValidator` in this repo.

## Impact

- **Specs**: Delta under
  `openspec/changes/kibana-security-role-dynamic-block-validation/specs/elasticstack-kibana-security-role/spec.md`
  capturing the new unknown-deferral requirement.
- **Implementation**: `internal/kibana/security_role/validators.go` (~6 lines in
  the element loop); an acceptance test config exercising a `dynamic "feature"` block.
