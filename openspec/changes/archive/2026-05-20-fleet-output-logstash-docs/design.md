## Context

`elasticstack_fleet_output` has supported `type = "logstash"` since PR #1302 (merged September 2025). The implementation in `internal/fleet/output/models_logstash.go` provides full CRUD, and `internal/fleet/output/acc_test.go` contains `TestAccResourceOutputLogstash` with create/update fixtures. The schema (`internal/fleet/output/schema.go`) lists `"logstash"` as a valid type enum value.

The gap is purely in the documentation layer: `templates/resources/fleet_output.md.tmpl` renders examples for elasticsearch, kafka_basic, kafka_advanced, and remote_elasticsearch — but not logstash. This change adds the missing example.

## Goals / Non-Goals

**Goals:**

- Provide a self-contained logstash example HCL file under `examples/resources/elasticstack_fleet_output/logstash.tf`
- Surface it in the generated docs via the template

**Non-Goals:**

- Schema changes
- Test changes
- CHANGELOG entry (explicitly excluded per the human direction for this proposal)
- `elasticstack_fleet_logstash_api_key` data source (Approach B from research — deferred as a separate request)

## Decisions

### Example content

The example demonstrates:
- `type = "logstash"`, `hosts`, and `ssl` block — the core logstash-specific fields
- Inline PEM placeholders (matching the pattern used in the remote_elasticsearch example for readability without real certificates)
- `default_integrations = false` and `default_monitoring = false` as explicit settings (common configuration for non-default outputs)

`config_yaml` is omitted from the example to keep it minimal and focused on logstash-specific fields; the acceptance test fixture uses it but it is not logstash-specific.

### Template placement

The new section is inserted between the Remote Elasticsearch section and `{{ .SchemaMarkdown }}`, maintaining the ordering from least-to-most-complex and keeping output types grouped.

## Risks / Trade-offs

No risk: this is additive-only documentation. The example file is validated only during `terraform-docs` doc generation and in local `make docs` runs, not in CI acceptance tests.

## Open Questions

- Does the requester need the `POST /api/fleet/logstash_api_keys` key-generation endpoint (Approach B), or is configuring the output itself sufficient? (non-blocking; Approach A is proceeding)
- Is there a preferred pattern for side-effecting data sources with no stable ID? (non-blocking; relevant only if Approach B is pursued)
