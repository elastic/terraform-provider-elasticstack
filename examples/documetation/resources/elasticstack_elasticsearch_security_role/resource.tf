provider "elasticstack" {}

resource "elasticstack_elasticsearch_security_role" "role" {
  name    = "testrole"
  cluster = ["all"]

  indices {
    names      = ["index1", "index2"]
    privileges = ["all"]
  }

  applications {
    application = "myapp"
    privileges  = ["admin", "read"]
    resources   = ["*"]
  }

  run_as = ["other_user"]

  metadata = jsonencode({
    version = 1
  })
}

output "role" {
  value = elasticstack_elasticsearch_security_role.role
}
