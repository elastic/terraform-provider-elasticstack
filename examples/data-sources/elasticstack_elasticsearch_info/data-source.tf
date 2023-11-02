provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_info" "cluster_info" {
}
