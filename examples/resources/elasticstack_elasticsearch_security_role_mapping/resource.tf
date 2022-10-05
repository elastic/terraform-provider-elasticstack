provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_role_mapping" "example" {
  name    = "test_role_mapping"
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

output "role" {
  value = elasticstack_elasticsearch_security_role_mapping.example.name
}
