provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_space" "test_space" {
  space_id = var.space_id
  name     = "Issue 4282 Test Space"
}

resource "elasticstack_kibana_security_role" "test_role" {
  name = var.role_name
  elasticsearch {}

  kibana {
    base = []
    feature {
      name       = "fleet"
      privileges = ["all"]
    }
    feature {
      name       = "fleetv2"
      privileges = ["all"]
    }
    spaces = [elasticstack_kibana_space.test_space.space_id]
  }
}

resource "elasticstack_elasticsearch_security_user" "test_user" {
  username = var.username
  password = var.password
  roles    = [elasticstack_kibana_security_role.test_role.name]
}

# The test_user's role only grants "fleet"/"fleetv2" privileges scoped to
# test_space (no default-space access). The post-install status poll must use
# the same space context as the install call (issue #4282).
resource "elasticstack_fleet_integration" "test_integration" {
  name     = "tcp"
  version  = "1.16.0"
  force    = true
  space_id = elasticstack_kibana_space.test_space.space_id

  kibana_connection {
    endpoints = var.kibana_endpoints
    username  = var.username
    password  = var.password
  }

  depends_on = [elasticstack_elasticsearch_security_user.test_user]
}
