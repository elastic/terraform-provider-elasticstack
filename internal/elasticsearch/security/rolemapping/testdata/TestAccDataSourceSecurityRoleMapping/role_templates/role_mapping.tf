provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_role_mapping" "test" {
  name    = "data_source_test"
  enabled = false

  role_templates = jsonencode([
    {
      format   = "json"
      template = "{\"source\":\"{{#tojson}}groups{{/tojson}}\"}"
    },
  ])

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
