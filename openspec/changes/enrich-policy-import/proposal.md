## Why

The `elasticstack_elasticsearch_enrich_policy` resource does not implement `ImportState`, so running `terraform import` against an existing enrich policy fails. Users cannot adopt pre-existing policies into Terraform state without destroying and recreating them, causing policy downtime and enrich index loss. This was reported in [elastic/terraform-provider-elasticstack#2402](https://github.com/elastic/terraform-provider-elasticstack/issues/2402).

## What Changes

- Add `ImportState` to the enrich policy resource, accepting an import ID of `<cluster_uuid>/<policy_name>` — the same format already used as the resource `id`.
- Add acceptance test coverage for the import path.

## Capabilities

### New Capabilities

<!-- None -->

### Modified Capabilities

- `elasticsearch-enrich-policy`: Add import requirements to the existing spec (REQ-023+) covering `ImportState` behavior, accepted ID format, and acceptance test expectations.

## Impact

- `internal/elasticsearch/enrich/resource.go`: implement `resource.ResourceWithImportState`
- `internal/elasticsearch/enrich/acc_test.go`: extend acceptance tests with an import step
- `openspec/specs/elasticsearch-enrich-policy/spec.md`: add import requirements
