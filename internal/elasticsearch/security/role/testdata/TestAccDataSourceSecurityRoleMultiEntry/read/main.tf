variable "role_name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_role" "test" {
  name = var.role_name

  indices {
    names                    = ["index1"]
    privileges               = ["read"]
    allow_restricted_indices = true
  }

  indices {
    names                    = ["index2"]
    privileges               = ["write"]
    allow_restricted_indices = false
  }

  applications {
    application = "myapp"
    privileges  = ["read"]
    resources   = ["*"]
  }

  applications {
    application = "otherapp"
    privileges  = ["admin"]
    resources   = ["object/*"]
  }

  remote_indices {
    clusters                 = ["test-cluster"]
    names                    = ["sample"]
    privileges               = ["read"]
    allow_restricted_indices = false
  }
}

data "elasticstack_elasticsearch_security_role" "test" {
  name = elasticstack_elasticsearch_security_role.test.name
}
