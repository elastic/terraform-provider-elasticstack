provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_security_role" "test" {
  name = "ds_test_idx_field_sec"
  elasticsearch {
    cluster = ["monitor"]
    indices {
      field_security {
        grant  = ["field1", "field2", "restricted"]
        except = ["restricted"]
      }
      query      = jsonencode({ match_all = {} })
      names      = ["sample-index"]
      privileges = ["read"]
    }
  }
  kibana {
    base   = ["read"]
    spaces = ["default"]
  }
}

data "elasticstack_kibana_security_role" "test" {
  name = elasticstack_kibana_security_role.test.name
}
