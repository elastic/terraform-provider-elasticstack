provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_security_user" "test" {
  username = "elastic"
}
