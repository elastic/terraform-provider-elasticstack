provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_space" "test_space" {
  space_id = var.space_id
  name     = "Test Space"
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
    spaces = [elasticstack_kibana_space.test_space.space_id]
  }
}

resource "elasticstack_elasticsearch_security_user" "test_user" {
  username = var.username
  password = var.password
  roles    = [elasticstack_kibana_security_role.test_role.name]
}
