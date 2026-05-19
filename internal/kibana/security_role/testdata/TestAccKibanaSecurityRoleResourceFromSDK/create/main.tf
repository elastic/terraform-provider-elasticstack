variable "role_name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_security_role" "upgrade" {
  name = var.role_name
  elasticsearch {
    cluster = ["create_snapshot"]
    indices {
      names      = ["sample"]
      privileges = ["read"]
    }
  }
  kibana {
    spaces = ["default"]
    feature {
      name       = "discover"
      privileges = ["read"]
    }
  }
}
