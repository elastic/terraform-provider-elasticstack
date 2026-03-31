provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_indices" "all_explicit" {
  target = "_all"
}
