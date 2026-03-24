provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_role" "test" {
  name = var.role_name

  cluster = ["all"]

  global = jsonencode({
    application = {}
    profile = {
      write = {
        applications = ["*"]
      }
    }
    role = {}
  })

  indices {
    names                    = ["index1", "index2"]
    privileges               = ["all"]
    allow_restricted_indices = true
    query = jsonencode({
      term = {
        status = "active"
      }
    })
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
