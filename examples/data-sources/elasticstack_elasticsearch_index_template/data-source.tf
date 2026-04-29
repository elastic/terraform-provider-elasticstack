provider "elasticstack" {
  # Elasticsearch-only: keep this example aligned with other ES data sources that do not need Kibana.
  elasticsearch {}
}

# Minimal template so the data source reads a predictable shape you control. This mirrors the
# acceptance-test pattern (resource + data source). The examples PlanOnly harness does not
# apply changes; the data source then uses the provider's legacy "not found" behavior (name only)
# instead of depending on built-in cluster templates, whose API responses vary by version.
resource "elasticstack_elasticsearch_index_template" "example" {
  name = "terraform-provider-elasticstack-example-index-template-ds"

  priority       = 100
  index_patterns = ["tf-example-index-template-ds-*"]
}

data "elasticstack_elasticsearch_index_template" "example" {
  name = elasticstack_elasticsearch_index_template.example.name
}

output "template" {
  value = data.elasticstack_elasticsearch_index_template.example
}
