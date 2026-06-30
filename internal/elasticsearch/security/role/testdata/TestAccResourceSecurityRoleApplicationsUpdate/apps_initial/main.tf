provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_role" "test" {
  name    = var.role_name
  cluster = ["all"]

  indices {
    names      = ["index1"]
    privileges = ["read"]
  }

  applications {
    application = "myapp"
    privileges  = ["read"]
    resources   = ["*"]
  }
}
