## Why

`elasticstack_elasticsearch_cluster_settings` supports `terraform import` at the code level (via `resource.ImportStatePassthroughID` added during the Plugin Framework migration), but the Terraform Registry documentation page shows no **## Import** section. This is because there is no `examples/resources/elasticstack_elasticsearch_cluster_settings/import.sh` file, which `tfplugindocs` uses to auto-generate the section. An existing acceptance test (`TestAccResourceClusterSettings/import`) already validates the import path end-to-end, so the code is complete and correct.

This change fills the documentation gap: it adds `import.sh` and regenerates `docs/resources/elasticsearch_cluster_settings.md` to expose the import command on the Registry page, closing issue #265.

## What Changes

- Add `examples/resources/elasticstack_elasticsearch_cluster_settings/import.sh` with the correct import command and inline comments explaining the singleton ID format (`<cluster_uuid>/cluster-settings`) and the post-import workflow (settings blocks must be declared in configuration before `terraform plan`/`terraform apply`).
- Run `make docs-generate` to regenerate `docs/resources/elasticsearch_cluster_settings.md` so it includes the `## Import` section.
- Sync the delta spec requirement (REQ-020) into the main spec (`openspec/specs/elasticsearch-cluster-settings/spec.md`) via `openspec sync`.

## Capabilities

### New Capabilities

None — no new schema attributes or data sources.

### Modified Capabilities

- `elasticstack_elasticsearch_cluster_settings`: update documentation to expose the existing import support.
