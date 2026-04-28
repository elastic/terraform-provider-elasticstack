provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_role_mapping" "doc_example" {
  name    = "doc_example_role_mapping"
  enabled = true
  roles = [
    "viewer",
  ]

  rules = jsonencode({
    any = [
      { field = { username = "doc-example-user" } },
    ]
  })
}

data "elasticstack_elasticsearch_security_role_mapping" "mapping" {
  name = elasticstack_elasticsearch_security_role_mapping.doc_example.name
}

output "user" {
  value = data.elasticstack_elasticsearch_security_role_mapping.mapping.name
}
