## Why

`elasticstack_kibana_security_list`, `elasticstack_kibana_security_list_item`, `elasticstack_kibana_security_exception_list`, and `elasticstack_kibana_security_list_data_streams` still implement CRUD directly on top of `entitycore.ResourceBase`. They are a cohesive Kibana Security subdomain with similar space-aware import and CRUD patterns, and they are among the lowest-risk remaining Kibana resources to migrate to the entitycore Kibana envelope.

Migrating them together should remove repeated Create/Read/Update/Delete prelude boilerplate, exercise the envelope on a coherent family of resources, and establish patterns for later Kibana Security migrations while preserving strict Terraform-visible behavior.

## What Changes

- Migrate the following resources from `entitycore.ResourceBase` to `entitycore.NewKibanaResource` with real create/read/update/delete callbacks:
  - `elasticstack_kibana_security_list`
  - `elasticstack_kibana_security_list_item`
  - `elasticstack_kibana_security_exception_list`
  - `elasticstack_kibana_security_list_data_streams`
- Update each resource model to satisfy the Kibana envelope contract (`GetID`, `GetResourceID`, `GetSpaceID`, `GetKibanaConnection`) without changing the Terraform schema or import behavior.
- Keep wrapper-level interfaces such as `ImportState` unchanged where present.
- Preserve existing Terraform-visible behavior strictly: schema shape, state ID representation, import ID format, read/write normalization, and acceptance-test expectations must not change.
- If a small entitycore improvement emerges that clearly benefits multiple resources in this batch, include it in the change; otherwise keep changes resource-local.

## Capabilities

### New Capabilities
None.

### Modified Capabilities
- `kibana-security-list`
- `kibana-security-list-item`
- `kibana-security-exception-list`
- `kibana-security-list-data-streams`

Implementation changes only; requirements-level behavior should remain unchanged.

## Impact

- `internal/kibana/securitylist/`
- `internal/kibana/securitylistitem/`
- `internal/kibana/securityexceptionlist/`
- `internal/kibana/security_list_data_streams/`
- Potentially `internal/entitycore/` if a small reusable seam is needed by more than one resource in the batch
- Acceptance and unit tests for the affected resources
