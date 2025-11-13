provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_role_mapping" "example" {
  name    = "example-security-role-mapping"
  enabled = true
  roles = [
    "admin"
  ]
  rules = jsonencode({
    any = [
      { field = { username = "esadmin" } },
      { field = { groups = "cn=admins,dc=example,dc=com" } },
    ]
  })
}

data "elasticstack_elasticsearch_security_role_mapping" "mapping" {
  name = elasticstack_elasticsearch_security_role_mapping.example.name
}

output "user" {
  value = data.elasticstack_elasticsearch_security_role_mapping.mapping.name
}
