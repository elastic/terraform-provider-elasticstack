variable "role_name" {
  description = "The role name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

locals {
  features = ["discover"]
}

resource "elasticstack_kibana_security_role" "dynamic_feature" {
  name = var.role_name
  elasticsearch {}
  kibana {
    spaces = ["*"]
    dynamic "feature" {
      for_each = local.features
      content {
        name       = feature.value
        privileges = ["read"]
      }
    }
  }
}
