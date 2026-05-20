## ADDED Requirements

### Requirement: Logstash output usage example in generated docs (REQ-DOCS-001)

The `elasticstack_fleet_output` resource documentation SHALL include a usage example for `type = "logstash"`. The example SHALL be rendered in the generated `docs/resources/fleet_output.md` under an explicit **"Logstash output"** heading, alongside the existing Elasticsearch, Kafka, and Remote Elasticsearch examples.

#### Scenario: Logstash example present in generated docs

- GIVEN `templates/resources/fleet_output.md.tmpl` references `examples/resources/elasticstack_fleet_output/logstash.tf`
- WHEN documentation is generated via `make docs`
- THEN `docs/resources/fleet_output.md` SHALL contain a "Logstash output" section with the logstash HCL example

#### Scenario: Logstash example shows correct resource shape

- GIVEN the logstash example file at `examples/resources/elasticstack_fleet_output/logstash.tf`
- WHEN a reader views the example
- THEN the example SHALL set `type = "logstash"`, include at least one entry in `hosts`, and include an `ssl` block with `certificate_authorities`, `certificate`, and `key` attributes
