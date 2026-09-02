provider "elasticstack" {
  elasticsearch {}
}

locals {
  index_permissions = [
    for name in var.index_names : {
      names      = [name]
      privileges = ["read"]
    }
  ]
}

resource "elasticstack_elasticsearch_security_role" "test" {
  name = var.role_name

  dynamic "remote_indices" {
    for_each = local.index_permissions
    content {
      clusters   = ["test-cluster"]
      names      = remote_indices.value.names
      privileges = remote_indices.value.privileges
    }
  }
}
