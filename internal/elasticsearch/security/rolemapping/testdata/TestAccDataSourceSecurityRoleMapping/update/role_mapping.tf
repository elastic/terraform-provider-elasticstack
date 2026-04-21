provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_role_mapping" "test" {
  name    = "data_source_test"
  enabled = false
  roles   = ["admin", "user"]

  rules = jsonencode({
    any = [
      { field = { username = "esadmin" } },
      { field = { groups = "cn=admins,dc=example,dc=com" } },
    ]
  })

  metadata = jsonencode({})
}

data "elasticstack_elasticsearch_security_role_mapping" "test" {
  name = elasticstack_elasticsearch_security_role_mapping.test.name
}
