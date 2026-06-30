## ADDED Requirements

### Requirement: Import documentation example (REQ-020)

The resource SHALL provide an `examples/resources/elasticstack_elasticsearch_cluster_settings/import.sh` file that documents the `terraform import` command for use by `tfplugindocs`. The example SHALL use the composite ID format `<cluster_uuid>/cluster-settings` and include inline comments explaining: (a) how to discover the cluster UUID (via the `elasticstack_elasticsearch_info` data source or the Elasticsearch `GET /` API); and (b) the post-import workflow, noting that only the `id` is stored in state after import and that the user must declare the desired `persistent` and/or `transient` setting blocks in their configuration before running `terraform plan` / `terraform apply`.

#### Scenario: Import section present in generated docs

- GIVEN `examples/resources/elasticstack_elasticsearch_cluster_settings/import.sh` exists with the correct import command
- WHEN `make docs-generate` runs
- THEN `docs/resources/elasticsearch_cluster_settings.md` SHALL include a `## Import` section containing the import command `terraform import elasticstack_elasticsearch_cluster_settings.<name> <cluster_uuid>/cluster-settings`
