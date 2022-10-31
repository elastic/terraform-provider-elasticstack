provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_security_role_mapping" "mapping" {
  name = "my_mapping"
}

output "user" {
  value = data.elasticstack_elasticsearch_security_role_mapping.mapping.name
}
