provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_role" "test" {
  name    = var.role_name
  cluster = ["all"]

  indices {
    names      = ["index1", "index2"]
    privileges = ["all"]
  }

  remote_indices {
    clusters = ["test-cluster2"]
    field_security {
      grant  = ["sample"]
      except = []
    }
    names      = ["sample2"]
    privileges = ["create", "read", "write"]
  }

  metadata = jsonencode({
    version = 1
  })
}
