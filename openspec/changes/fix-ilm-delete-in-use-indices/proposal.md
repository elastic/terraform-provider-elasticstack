## Why

When a Fleet integration is installed it creates index templates, component templates and ingest pipelines. If data streams or indices inherit from those templates and an ILM policy is attached via `elasticstack_elasticsearch_index_template_ilm_attachment`, the backing indices carry an `index.lifecycle.name` reference. On `terraform destroy`, Fleet's package uninstall succeeds but the backing indices remain; deleting the Terraform-managed ILM policy then fails with:

> Cannot delete policy [X]. It is in use by one or more indices: [.ds-...]

This is a silent cross-dependency that forces users to run manual cleanup commands (see https://github.com/elastic/terraform-provider-elasticstack/issues/1999). The ILM resource should handle this gracefully.

## What Changes

- **`elasticstack_elasticsearch_index_lifecycle` Delete**: Before calling `DELETE /_ilm/policy/{name}`, scan all indices for references to the policy and null `index.lifecycle.name` on matching indices.
- Add supporting ES client helper for querying index settings with the `flat_settings` option and for clearing `index.lifecycle.name` on specific indices.
- Acceptance test that reproduces the cross-dependency scenario (Fleet integration + ILM attachment + data stream) and asserts the ILM policy destroy fails (to be flipped to success after the fix is implemented).

## Capabilities

### New Capabilities
- *(none)*

### Modified Capabilities
- `elasticsearch-index-lifecycle`: The Delete behavior requirement SHALL change so that the resource first removes the policy reference from any indices that use it before deleting the policy. This changes REQ-016 (or a new requirement) in the lifecycle spec.

## Impact

- **Affected packages**: `internal/elasticsearch/index/ilm/`, `internal/clients/elasticsearch/`
- **Terraform resource**: `elasticstack_elasticsearch_index_lifecycle`
- **No API changes**: Schema remains identical; behavior change is purely in Delete
- **No breaking changes**: Resources that already delete successfully will continue to work identically. Resources that previously failed due to in-use indices will now succeed.
