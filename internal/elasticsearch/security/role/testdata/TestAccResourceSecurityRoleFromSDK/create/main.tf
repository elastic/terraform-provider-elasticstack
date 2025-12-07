variable "role_name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_role" "test" {
  name    = var.role_name
  cluster = ["all"]

  indices {
    names      = ["index1", "index2"]
    privileges = ["all"]
    field_security {
      grant  = ["sample"]
      except = []
    }
    allow_restricted_indices = true
  }

  remote_indices {
    clusters = ["test-cluster"]
    field_security {
      grant  = ["sample"]
      except = []
    }
    names      = ["sample"]
    privileges = ["create", "read", "write"]
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
