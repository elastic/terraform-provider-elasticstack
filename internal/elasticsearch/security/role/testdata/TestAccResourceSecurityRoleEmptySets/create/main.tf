provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_role" "test" {
  name    = var.role_name
  cluster = []

  indices {
    names      = ["index1", "index2"]
    privileges = ["all"]
    field_security {
      grant  = ["*"]
      except = []
    }
  }

  remote_indices {
    clusters = []
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

  run_as = []

  metadata = jsonencode({})
}
