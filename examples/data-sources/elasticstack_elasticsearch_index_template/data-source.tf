provider "elasticstack" {
  # Elasticsearch-only: keep this example aligned with other ES data sources that do not need Kibana.
  elasticsearch {}
}

data "elasticstack_elasticsearch_index_template" "ilm-history-7" {
  name = "ilm-history-7"
}

output "template" {
  value = data.elasticstack_elasticsearch_index_template.ilm-history-7
}
