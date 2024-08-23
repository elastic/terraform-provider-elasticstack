provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

data "elasticstack_elasticsearch_indices" "logs" {
  search = "log*"
}
