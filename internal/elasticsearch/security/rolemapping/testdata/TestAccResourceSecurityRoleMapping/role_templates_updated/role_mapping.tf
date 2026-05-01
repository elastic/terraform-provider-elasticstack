variable "name" {
  description = "The role mapping name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_role_mapping" "test" {
  name    = var.name
  enabled = true

  role_templates = jsonencode([
    {
      format   = "json"
      template = "{\"source\":\"{{#tojson}}roles{{/tojson}}\"}"
    },
  ])

  rules = jsonencode({
    any = [
      { field = { username = "poweruser" } },
      { field = { groups = "cn=operators,dc=example,dc=com" } },
    ]
  })

  metadata = jsonencode({})
}
