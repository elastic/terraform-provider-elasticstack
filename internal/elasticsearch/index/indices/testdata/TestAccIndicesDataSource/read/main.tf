provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_indices" "security_indices" {
  target = ".security-*"
}
