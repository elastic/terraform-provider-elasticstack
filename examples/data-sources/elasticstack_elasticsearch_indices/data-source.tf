provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

data "elasticstack_elasticsearch_indices" "security_indices" {
  target = ".security-*"
}
