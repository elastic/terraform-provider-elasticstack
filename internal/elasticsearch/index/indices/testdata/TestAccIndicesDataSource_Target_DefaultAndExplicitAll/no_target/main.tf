provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_indices" "all_default" {
}
