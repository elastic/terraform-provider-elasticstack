## Why

The `ElasticsearchResource` envelope's `Create` and `Update` preludes currently delegate all post-write state population to the concrete callbacks. Every callback independently calls `readFunc` and handles the not-found case, duplicating the same boilerplate across every resource that uses the envelope and creating a trap for new resources that forget to do so.

## What Changes

- The envelope's `writeFromPlan` (shared by `Create` and `Update`) will invoke `readFunc` after the concrete callback returns successfully, and append a standard error diagnostic if the resource is not found after the write.
- Concrete create and update callbacks no longer call `readFunc` or handle not-found. They return the written model with enough state for `readFunc` to proceed (composite ID, any create-only field values such as API key secrets, `ElasticsearchConnection`).
- Existing concrete callbacks for `role`, `script`, `role_mapping`, and `system_user` are simplified to remove their inline read and not-found handling.

## Capabilities

### New Capabilities

_None._

### Modified Capabilities

- `entitycore-resource-envelope`: the Create and Update prelude requirements change — read-after-write becomes the envelope's responsibility; the callback contract narrows to "write and return state for readFunc to build on".

## Impact

- `internal/entitycore/resource_envelope.go` — `writeFromPlan` gains the read-after-write step
- `internal/entitycore/resource_envelope_test.go` — new tests for the read-after-write path; existing write happy-path tests updated
- `internal/elasticsearch/security/role/update.go` — remove inline `readRole` call and not-found handling
- `internal/elasticsearch/cluster/script/update.go` — remove inline `readScriptPayload` call, field-carrying, and not-found handling
- `internal/elasticsearch/security/rolemapping/update.go` — remove inline `readRoleMappingResource` call and nil check
- `internal/elasticsearch/security/systemuser/update.go` — remove inline `readSystemUser` call and not-found handling
