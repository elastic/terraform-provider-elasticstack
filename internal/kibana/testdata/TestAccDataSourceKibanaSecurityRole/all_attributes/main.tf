provider "elasticstack" {
  elasticsearch {}
  kibana {}
}


resource "elasticstack_kibana_security_role" "test" {
	name    = "data_source_test"
	elasticsearch {
	  cluster = [ "create_snapshot" ]
	  indices {
		names = ["sample"]
		privileges = ["create", "read", "write"]
	  }
	  run_as = ["kibana", "elastic"]
	}
	kibana {
	  base = [ "all" ]
	  spaces = ["default"]
	}
}

data "elasticstack_kibana_security_role" "test" {
  name = elasticstack_kibana_security_role.test.name
}
