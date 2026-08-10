variable "role_name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_role" "test" {
  name    = var.role_name
  cluster = ["all"]

  global = jsonencode({
    application = {}
    profile = {
      write = {
        applications = ["*"]
      }
    }
  })

  indices {
    names      = ["index1", "index2"]
    privileges = ["all"]
    field_security {
      grant  = ["field1", "field2"]
      except = ["field2.secret"]
    }
    query = jsonencode({
      term = {
        status = "active"
      }
    })
  }

  remote_indices {
    clusters   = ["test-cluster"]
    names      = ["sample"]
    privileges = ["read"]
    query = jsonencode({
      match_all = {}
    })
  }
}

data "elasticstack_elasticsearch_security_role" "test" {
  name = elasticstack_elasticsearch_security_role.test.name
}
