provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

locals {
  roles_list = [
    {
      name    = "role_with_feature"
      cluster = ["monitor"]
      feature = [
        { name = "discover", privileges = ["read"] },
      ]
    },
    {
      name    = "role_without_feature"
      cluster = ["monitor"]
    },
  ]
}

resource "elasticstack_kibana_security_role" "this" {
  count = length(local.roles_list)
  name  = local.roles_list[count.index].name

  elasticsearch {
    cluster = try(local.roles_list[count.index].cluster, [])
  }

  dynamic "kibana" {
    for_each = length(try(local.roles_list[count.index].feature, [])) > 0 ? [1] : []
    content {
      spaces = ["*"]

      dynamic "feature" {
        for_each = local.roles_list[count.index].feature
        content {
          name       = feature.value.name
          privileges = feature.value.privileges
        }
      }
    }
  }
}
