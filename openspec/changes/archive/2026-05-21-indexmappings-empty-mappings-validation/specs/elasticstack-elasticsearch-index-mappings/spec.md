## MODIFIED Requirements

### Requirement: Schema (REQ-001)

The resource SHALL expose the following top-level attributes:

```hcl
resource "elasticstack_elasticsearch_index_mappings" "example" {
  id    = <computed, string>                        # "<cluster_uuid>/<index_name>"
  index = <required, string, forces replacement>    # name of the target Elasticsearch index

  mappings = <required, JSON string>                # all top-level mapping keys accepted:
                                                    # properties, dynamic, _source,
                                                    # dynamic_templates, runtime, etc.
                                                    # MUST be a non-empty JSON object

  elasticsearch_connection { ... }                  # standard provider connection block
}
```

- `id` SHALL be computed and unknown until create completes; it SHALL use `stringplanmodifier.UseStateForUnknown()`.
- `index` SHALL be required and SHALL force resource replacement when changed.
- `mappings` SHALL be a required JSON string using `index.MappingsType{}` as the custom type for semantic equality normalization. It SHALL accept any combination of valid top-level Elasticsearch mapping parameters (not limited to `properties`). It SHALL additionally be validated as a non-empty JSON object — an empty object (`{}`) SHALL be rejected at plan time.
- `elasticsearch_connection` is injected by the provider scaffold and SHALL NOT be declared manually in the schema factory.

#### Scenario: Schema validation — index is required

- GIVEN a configuration that omits `index`
- WHEN `terraform validate` runs
- THEN Terraform SHALL emit a required-attribute error

#### Scenario: Schema validation — mappings is required

- GIVEN a configuration that omits `mappings`
- WHEN `terraform validate` runs
- THEN Terraform SHALL emit a required-attribute error

#### Scenario: Schema validation — mappings must be non-empty JSON object

- GIVEN a configuration where `mappings = jsonencode({})`
- WHEN `terraform validate` runs
- THEN Terraform SHALL emit an attribute validation error stating the value must be a non-empty JSON object
