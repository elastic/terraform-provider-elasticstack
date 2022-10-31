provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_security_role" "role" {
  name = "testrole"
}

output "role" {
  value = data.elasticstack_elasticsearch_security_role.role.name
}
