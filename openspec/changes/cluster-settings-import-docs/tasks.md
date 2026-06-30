## 1. Add `import.sh` example

- [ ] 1.1 Create `examples/resources/elasticstack_elasticsearch_cluster_settings/import.sh` with:
  - The import command: `terraform import elasticstack_elasticsearch_cluster_settings.my_settings <cluster_uuid>/cluster-settings`
  - An inline comment explaining the singleton ID format: the composite ID is `<cluster_uuid>/cluster-settings`; `<cluster_uuid>` can be found via the `elasticstack_elasticsearch_info` data source or via `GET /` on the Elasticsearch API
  - An inline comment explaining the post-import workflow: after import, only the `id` is in state; the user must declare the desired `persistent` and/or `transient` setting blocks in their configuration and run `terraform plan` / `terraform apply` to bring them under management

## 2. Regenerate docs

- [ ] 2.1 Run `make docs-generate` to regenerate `docs/resources/elasticsearch_cluster_settings.md`
- [ ] 2.2 Verify that the regenerated file contains a `## Import` section with the correct import command

## 3. Sync spec

- [ ] 3.1 Run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec sync cluster-settings-import-docs` to merge REQ-020 from the delta spec into `openspec/specs/elasticsearch-cluster-settings/spec.md`
- [ ] 3.2 Verify the main spec now includes REQ-020 and that `openspec validate elasticsearch-cluster-settings --type spec` passes
