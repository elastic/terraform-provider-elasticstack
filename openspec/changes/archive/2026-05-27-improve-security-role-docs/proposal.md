## Why

Users configuring complex Elasticsearch and Kibana security roles have to reverse-engineer valid privilege strings from live API responses because the provider docs only show generic syntax examples with no real-world context (issue #2000). The current examples use placeholder values (`"all"`, `["test"]`) that don't model least-privilege patterns, and the valid Kibana feature privilege strings are entirely absent from the docs.

## What Changes

- **Remove** existing generic example files for `elasticstack_kibana_security_role` (`resource-with-base.tf`, `resource-with-feature.tf`) and `elasticstack_elasticsearch_security_role` (`resource.tf`)
- **Add** scenario-based example files for both resources covering real-world archetypes (data analyst, data ingestion service account, security team, DevOps read-only, multi-space role)
- **Add** a new `templates/guides/security-roles.md.tmpl` guide covering the privilege model, Kibana feature privilege reference table, field/document security, and multi-resource composition (`elasticsearch_security_role` + `kibana_security_role` + API key `role_descriptors`)
- **Add** a machine-readable `scripts/security-role-docs/kibana-features.json` file as the curated source of truth for which Kibana features the guide documents (including a `skip` list for features deliberately excluded)
- **Add** a gh-aw scheduled workflow that calls `GET /api/features` against a live stack, diffs the response against `kibana-features.json`, and opens a self-healing PR to update the guide when drift is detected
- **Update** `templates/resources/kibana_security_role.md.tmpl` and `templates/resources/elasticsearch_security_role.md.tmpl` to link to the new guide

## Capabilities

### New Capabilities

- `security-role-guide`: Standalone provider guide covering the security role privilege model, scenario-based examples, Kibana feature privilege reference, field/document security, and multi-resource composition patterns
- `security-role-docs-drift-detection`: Scheduled gh-aw workflow with a curated features JSON file and a Go pre-activation script that detects drift between the documented privilege table and the live Kibana features API, opening a self-healing PR when the documented table is stale

### Modified Capabilities

<!-- No spec-level requirement changes to existing resources -->

## Impact

- `examples/resources/elasticstack_kibana_security_role/` — example files replaced
- `examples/resources/elasticstack_elasticsearch_security_role/` — example file replaced
- `templates/guides/security-roles.md.tmpl` — new file
- `templates/resources/kibana_security_role.md.tmpl` — adds guide link
- `templates/resources/elasticsearch_security_role.md.tmpl` — adds guide link
- `scripts/security-role-docs/` — new directory with Go pre-activation script and `kibana-features.json`
- `.github/workflows/security-role-docs-drift.md` — new gh-aw workflow source
- `.buildkite/update-kibana-client.sh` — extended to regenerate `kibana-features.json` alongside kbapi update
- `docs/` — regenerated via `make docs-generate`
