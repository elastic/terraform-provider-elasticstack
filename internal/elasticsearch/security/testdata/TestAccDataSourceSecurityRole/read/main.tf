provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_role" "test" {
  name    = "data_source_test"
  cluster = ["all"]

  indices {
    names                    = ["index1", "index2"]
    privileges               = ["all"]
    allow_restricted_indices = true
  }

  remote_indices {
	clusters = ["test-cluster2"]
	field_security {
	  grant = ["sample"]
	  except = []
	}
	names = ["sample2"]
	privileges = ["create", "read", "write"]
  }

  applications {
    application = "myapp"
    privileges  = ["admin", "read"]
    resources   = ["*"]
  }

  run_as = ["other_user"]

  metadata = jsonencode({
    version = 1
  })
}

data "elasticstack_elasticsearch_security_role" "test" {
  name = elasticstack_elasticsearch_security_role.test.name
}
