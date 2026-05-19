provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_security_role" "test" {
  name = "ds_test_remote_idx_ext"
  elasticsearch {
    cluster = ["create_snapshot"]
    indices {
      names      = ["sample"]
      privileges = ["create", "read", "write"]
    }
    remote_indices {
      clusters = ["test-cluster"]
      field_security {
        grant  = ["sample", "restricted"]
        except = ["restricted"]
      }
      query      = jsonencode({ match_all = {} })
      names      = ["sample"]
      privileges = ["create", "read", "write"]
    }
    run_as = ["kibana", "elastic"]
  }
  kibana {
    base   = ["all"]
    spaces = ["default"]
  }
}

data "elasticstack_kibana_security_role" "test" {
  name = elasticstack_kibana_security_role.test.name
}
