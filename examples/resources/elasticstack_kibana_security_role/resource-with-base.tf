
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_security_role" "example" {
  name = "sample_role"
  elasticsearch {
    cluster = ["create_snapshot"]
    indices {
      field_security {
        grant  = ["test"]
        except = []
      }
      names      = ["test"]
      privileges = ["create", "read", "write"]
    }
    remote_indices {
      field_security {
        grant  = ["test"]
        except = []
      }
      names      = ["test"]
	    clusters = ["test-cluster"]
      privileges = ["create", "read", "write"]
    }
  }
  kibana {
    base   = ["all"]
    spaces = ["default"]
  }
}
