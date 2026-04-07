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
    cluster = ["create_snapshot"]
    indices {
      names      = ["sample"]
      privileges = ["create", "read", "write"]
    }
    remote_indices {
      clusters = ["test-cluster2"]
      field_security {
        grant  = ["sample2"]
        except = []
      }
      names      = ["sample2"]
      privileges = ["create", "read", "write"]
    }
    run_as = ["kibana", "elastic"]
  }
  kibana {
    base   = ["all"]
    spaces = ["default"]
  }
}
