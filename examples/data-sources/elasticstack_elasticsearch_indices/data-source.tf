provider "elasticstack" {
  # Elasticsearch-only: including an empty `kibana {}` block without Kibana configuration can
  # prevent the default Elasticsearch client from resolving from ELASTICSEARCH_ENDPOINTS.
  elasticsearch {}
}

data "elasticstack_elasticsearch_indices" "security_indices" {
  target = ".security-*"
}
