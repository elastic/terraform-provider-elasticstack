provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_security_role" "test" {
  name = "ds_test_feature_privs"
  elasticsearch {
    cluster = ["monitor"]
    indices {
      names      = ["test-index"]
      privileges = ["read"]
    }
  }
  kibana {
    feature {
      name       = "actions"
      privileges = ["read"]
    }
    spaces = ["default"]
  }
}

data "elasticstack_kibana_security_role" "test" {
  name = elasticstack_kibana_security_role.test.name
}
