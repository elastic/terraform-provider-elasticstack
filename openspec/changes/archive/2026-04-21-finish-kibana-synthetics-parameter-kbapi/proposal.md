## Why

`elasticstack_kibana_synthetics_parameter` already uses the generated Kibana OpenAPI (`kbapi`) client for create, read, and update, but delete still goes through the legacy `go-kibana-rest` synthetics client. Splitting traffic across two client stacks increases maintenance cost, diverges from other migrated Kibana entities, and leaves redundant error and connection paths in the resource package.

## What Changes

- Route delete through the same `kbapi` / `kibanaoapi.Client` stack as the other CRUD operations (e.g. `DeleteParameterWithResponse`), removing the legacy client from this resource’s delete path.
- Remove remaining legacy-specific wiring from `internal/kibana/synthetics/parameter` where it exists only to support delete (and any related comments or spec wording that still refer to “legacy client for delete”).
- Keep the manual `json.Marshal` + `PostParametersWithBodyWithResponse` / `PutParameterWithBodyWithResponse` pattern for create/update request bodies until oapi-codegen oneOf handling improves (see [oapi-codegen#1620](https://github.com/oapi-codegen/oapi-codegen/issues/1620)).
- Preserve read-after-write: after successful create and update, state SHALL still be populated from a follow-up `GET` of the parameter by id.
- Preserve `share_across_spaces` behavior: create-only in the request body, `RequiresReplace` semantics unchanged, and the same `namespaces` ↔ `share_across_spaces` mapping on read.

## Capabilities

### New Capabilities

- (none)

### Modified Capabilities

- `kibana-synthetics-parameter`: Align requirements with a single OpenAPI-based client for all CRUD; document the retained oneOf JSON workaround and unchanged read-after-write and `share_across_spaces` rules.

## Impact

- Primary code: `internal/kibana/synthetics/parameter/delete.go`, possibly `resource.go` or shared helpers if delete-specific branching exists; traceability table in the spec.
- Generated client: `generated/kbapi` (`DeleteParameterWithResponse` and existing parameter endpoints); no OpenAPI regeneration required unless the team chooses to refresh specs separately.
- Functional specs: delta under this change for `kibana-synthetics-parameter`; eventual sync to `openspec/specs/kibana-synthetics-parameter/spec.md` after implementation (out of scope for this proposal-only change).
