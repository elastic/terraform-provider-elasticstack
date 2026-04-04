## Why

Synthetics private locations are scoped to a Kibana space via the API path (`/s/<space_id>/api/synthetics/...`). The Terraform resource currently issues all Private Location API calls as if the target were the default space, so practitioners cannot manage locations in non-default spaces. Aligning with `elasticstack_kibana_synthetics_monitor` (which already exposes `space_id`) removes that gap and matches how monitors and locations are used together per space.

## What Changes

- Add an optional Terraform attribute `space_id` on `elasticstack_kibana_synthetics_private_location` that selects the Kibana space for create, read, and delete.
- When `space_id` is unset or empty, behavior remains the default space (same as today).
- Changing `space_id` after create SHALL require replacement (consistent with other identity-related attributes on this resource).
- Extend the legacy Kibana client (`kbapi`) private location helpers to accept a space argument for URL construction, mirroring the monitor API pattern.
- Update acceptance tests and generated documentation as needed.

## Capabilities

### New Capabilities

- (none)

### Modified Capabilities

- `kibana-synthetics-private-location`: Add requirements for `space_id`, space-scoped API usage, and replacement/import behavior involving the chosen space.

## Impact

- **Code**: `internal/kibana/synthetics/privatelocation/` (schema, CRUD), `libs/go-kibana-rest/kbapi/api.kibana_synthetics.go` (private location function signatures and `basePath` usage), possibly `internal/kibana/synthetics/privatelocation/schema_test.go` and `acc_test.go`.
- **Tests**: `libs/go-kibana-rest/kbapi` tests if API signatures change; private location acceptance tests for default vs non-default space if the stack test setup supports it.
- **Docs**: Resource docs generation for the new attribute.
- **Compatibility**: Non-breaking for existing configs if `space_id` is optional with default-space semantics when omitted.
