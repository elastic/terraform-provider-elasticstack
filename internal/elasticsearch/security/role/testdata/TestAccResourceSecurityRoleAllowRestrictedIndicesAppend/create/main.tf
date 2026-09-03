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

  dynamic "indices" {
    for_each = local.index_permissions
    content {
      names      = indices.value.names
      privileges = indices.value.privileges
    }
  }
}
