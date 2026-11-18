variable "role_name" {
  description = "The role name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_security_role" "test" {
  name = var.role_name
  elasticsearch {
    cluster = ["monitor"]
  }
  kibana {
    base = ["all"]
    feature {
      name       = "discover"
      privileges = ["read"]
    }
    spaces = ["default"]
  }
}
