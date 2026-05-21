## Why

The `elasticstack_fleet_output` resource fully supports `type = "logstash"` — the implementation and acceptance tests were shipped as part of PR #1302 — but the generated documentation (`docs/resources/fleet_output.md`) contains no logstash example. Users who look at the docs see usage examples for Elasticsearch, Kafka (basic and advanced), and Remote Elasticsearch, but nothing for Logstash, making the feature effectively undiscoverable through official documentation.

## What Changes

- Add a new `examples/resources/elasticstack_fleet_output/logstash.tf` example file showing a representative logstash output configuration.
- Reference the new example from the doc template `templates/resources/fleet_output.md.tmpl` as a **"Logstash output"** section alongside the existing examples.

No provider Go code, schema, client, or test changes are required — the feature already works.

## Capabilities

### Modified Capabilities

- `fleet-output`: Adding a documentation example for the already-implemented logstash output type.

## Impact

- `examples/resources/elasticstack_fleet_output/logstash.tf` — new example file (created)
- `templates/resources/fleet_output.md.tmpl` — add `{{ tffile "examples/resources/elasticstack_fleet_output/logstash.tf" }}` section
