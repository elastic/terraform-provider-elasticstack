## 1. Remove Generic Examples

- [x] 1.1 Delete `examples/resources/elasticstack_kibana_security_role/resource-with-base.tf`
- [x] 1.2 Delete `examples/resources/elasticstack_kibana_security_role/resource-with-feature.tf`
- [x] 1.3 Delete `examples/resources/elasticstack_elasticsearch_security_role/resource.tf`

## 2. Add Scenario Example Files — Kibana Security Role

- [x] 2.1 Create `examples/resources/elasticstack_kibana_security_role/resource-data-analyst.tf` — read-only Discover + Dashboard access in a named space; `cluster = ["monitor"]`, index privileges `["read", "view_index_metadata"]` on `logs-*`/`metrics-*`
- [x] 2.2 Create `examples/resources/elasticstack_kibana_security_role/resource-data-ingest.tf` — write to data streams; `cluster = ["manage_ingest_pipelines", "manage_index_templates", "auto_configure"]`, index privileges `["write", "create_index", "auto_configure"]` on `logs-myapp-*`; no Kibana block
- [x] 2.3 Create `examples/resources/elasticstack_kibana_security_role/resource-security-analyst.tf` — full access to `siem`, `securitySolutionCases`, `alerting`, `actions`, `rulesSettings`, `osquery` features in a `security` space
- [x] 2.4 Create `examples/resources/elasticstack_kibana_security_role/resource-devops-readonly.tf` — read access to `fleet`, `apm`, `infrastructure`, `logs` features; `cluster = ["monitor"]`
- [x] 2.5 Create `examples/resources/elasticstack_kibana_security_role/resource-multi-space.tf` — two `kibana {}` blocks: `base = ["all"]` for `["dev", "staging"]` and feature-level read-only (`dashboard`, `discover`) for `["prod"]`

## 3. Add Scenario Example Files — Elasticsearch Security Role

- [ ] 3.1 Create `examples/resources/elasticstack_elasticsearch_security_role/resource-data-analyst.tf` — `cluster = ["monitor"]`, indices with `["read", "view_index_metadata"]` on `logs-*`
- [ ] 3.2 Create `examples/resources/elasticstack_elasticsearch_security_role/resource-monitoring-agent.tf` — `cluster = ["monitor", "manage_index_templates"]`, indices with `["write", "create_index"]` on `.monitoring-*`/`metricbeat-*`; demonstrates `allow_restricted_indices = false` with a warning comment
- [ ] 3.3 Create `examples/resources/elasticstack_elasticsearch_security_role/resource-field-and-doc-security.tf` — demonstrates `field_security` (grant/except for PII redaction) and `query` (jsonencode for tenant isolation) on the same indices block

## 4. Create Security Roles Guide

- [ ] 4.1 Create `templates/guides/security-roles.md.tmpl` with page title, description frontmatter, and section skeleton
- [ ] 4.2 Add "When to use each resource" section explaining ES roles vs Kibana roles vs API key `role_descriptors`
- [ ] 4.3 Add "Scenario examples" section embedding all kibana scenario files via `{{ tffile "examples/resources/elasticstack_kibana_security_role/resource-*.tf" }}` directives (one subsection per scenario)
- [ ] 4.4 Add "Field security and document-level security" section embedding `resource-field-and-doc-security.tf` from ES examples
- [ ] 4.5 Add "Composing with API keys" section: embed the data-analyst Kibana role alongside an `elasticstack_elasticsearch_security_api_key` example with narrower `role_descriptors`
- [ ] 4.6 Add "Kibana feature privilege reference" section with the markdown table (columns: feature name, available privileges) covering the 16+ features from the spec; include exhaustiveness caveat with `GET /api/features` link

## 5. Update Resource Doc Templates

- [ ] 5.1 Add "See also: [Security Roles Guide](../guides/security-roles)" link to `templates/resources/kibana_security_role.md.tmpl` immediately after the resource description
- [ ] 5.2 Add the same "See also" link to `templates/resources/elasticsearch_security_role.md.tmpl`
- [ ] 5.3 Update template `{{ .ExampleFile }}` or equivalent directives to reference the new scenario files (verify tfplugindocs picks them up correctly)

## 6. Create Curated Features JSON

- [ ] 6.1 Create `scripts/security-role-docs/` directory
- [ ] 6.2 Create `scripts/security-role-docs/kibana-features.json` with `documented` array containing all features from the guide table and `skip` array populated by reviewing the full `GET /api/features` response at the current STACK_VERSION and excluding internal/plugin-specific features
- [ ] 6.3 Verify the `documented` array matches exactly the feature names in the guide table (no entries in the table absent from the array, and vice versa)

## 7. Go Pre-Activation Script

- [ ] 7.1 Create `scripts/security-role-docs/main.go` with `pre-activation` subcommand
- [ ] 7.2 Implement `GET /api/features` call using Kibana client (standard env vars: `KIBANA_ENDPOINT`, `KIBANA_USERNAME`, `KIBANA_PASSWORD`)
- [ ] 7.3 Implement diff logic: load `kibana-features.json`, compute `unknown_features` (in API but not in `documented` or `skip`) and `removed_features` (in `documented` but absent from API)
- [ ] 7.4 Write drift report JSON to `--report-path`; set `run_agent` GitHub Actions output to `true`/`false`
- [ ] 7.5 Write unit tests for the diff logic in `scripts/security-role-docs/main_test.go`

## 8. gh-aw Drift Detection Workflow

- [ ] 8.1 Create `.github/workflows/security-role-docs-drift.md` using gh-aw markdown format
- [ ] 8.2 Configure triggers: `workflow_dispatch`, `schedule: weekly`, `push to main paths: generated/kbapi/**`
- [ ] 8.3 Add `imports: [shared/setup-dev.md]` and the shared live-stack setup import used by workflows that require a running Kibana
- [ ] 8.4 Wire up pre-activation job: checkout, setup Go, run `go run ./scripts/security-role-docs pre-activation --features-path scripts/security-role-docs/kibana-features.json --report-path /tmp/gh-aw/agent/drift-report.json`; upload report as artifact; output `run_agent`
- [ ] 8.5 Configure agent section: `if: needs.pre_activation.outputs.run_agent == 'true'`; download artifact; write agent instructions (update JSON + guide template + run `make docs-generate`; open PR)
- [ ] 8.6 Configure `safe-outputs` with `create-pr` (or equivalent); no `create-issue` output
- [ ] 8.7 Compile workflow: `make workflow-generate`; verify `.github/workflows/security-role-docs-drift.lock.yml` is produced without errors

## 9. OpenSpec Spec Files

- [ ] 9.1 Create `openspec/specs/security-role-guide/spec.md` from the delta spec in this change
- [ ] 9.2 Create `openspec/specs/security-role-docs-drift-detection/spec.md` from the delta spec in this change
- [ ] 9.3 Run `make check-openspec` and resolve any validation errors

## 10. Docs Regeneration and Verification

- [ ] 10.1 Run `make docs-generate` and confirm `docs/guides/security-roles.md` is created
- [ ] 10.2 Run `make check-docs` and confirm it passes with zero uncommitted changes
- [ ] 10.3 Run `make build` to confirm the provider builds cleanly
- [ ] 10.4 Manually review `docs/guides/security-roles.md` in a markdown renderer to confirm formatting, table rendering, and embedded code blocks look correct
