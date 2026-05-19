provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_security_role" "test" {
  name        = "ds_test_desc_metadata"
  description = "Test role description"
  metadata    = jsonencode({ custom_key = "custom_value" })
  elasticsearch {
    cluster = ["monitor"]
    indices {
      names      = ["meta-index"]
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
