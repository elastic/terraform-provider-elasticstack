## Why

The `elasticstack_kibana_security_enable_rule` resource partially adopted the `KibanaResource` envelope but bypasses it for Create and Update by defining wrapper-level `Create` and `Update` methods that shadow the envelope's promoted methods. This means those operations do not flow through the envelope's client resolution, version-enforcement, or read-after-write paths. Completing the migration aligns the resource with the codebase's standard envelope pattern and eliminates the bespoke override code.

## What Changes

- Remove the bespoke `Create(ctx, req, resp)` method from `create.go` in `internal/kibana/security_enable_rule`.
- Remove the bespoke `Update(ctx, req, resp)` and `upsert` methods from `update.go` in `internal/kibana/security_enable_rule`.
- Implement a `writeSecurityEnableRule` callback with the `KibanaWriteFunc[enableRuleModel]` signature (handling both Create and Update).
- Replace `PlaceholderKibanaWriteCallback` in `resource.go` with `writeSecurityEnableRule` for both Create and Update.
- Remove the redundant inline `EnforceVersionRequirements` call from the delete callback (`delete.go`), which duplicates the envelope's own version-requirement enforcement.

## Capabilities

### New Capabilities

<!-- none -->

### Modified Capabilities

- `kibana-kibana-resource-envelope`: extend the "Kibana resources fully implement envelope CRUD callbacks" requirement to cover `security_enable_rule`; extend the `PlaceholderKibanaWriteCallback` SHALL NOT be used list and the lifecycle-dispatch scenario to include `security_enable_rule`.

## Impact

- **Files changed**: `internal/kibana/security_enable_rule/create.go`, `internal/kibana/security_enable_rule/update.go`, `internal/kibana/security_enable_rule/delete.go`, `internal/kibana/security_enable_rule/resource.go`.
- **No user-visible behavior change**: the resource's Terraform schema and API semantics are unchanged; only the internal dispatch path moves from wrapper override to envelope callback.
- **Existing acceptance and unit tests** remain valid without schema or behavior adjustments.
