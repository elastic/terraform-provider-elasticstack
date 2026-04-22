variable "username" {
  description = "The username for the security user"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_user" "test" {
  username  = var.username
  roles     = ["kibana_admin", "viewer"]
  full_name = "Test Custom User"
  email     = "custom@example.com"
  password  = "qwerty123"
  metadata  = jsonencode({ env = "test" })
  enabled   = true
}

data "elasticstack_elasticsearch_security_user" "test" {
  username = elasticstack_elasticsearch_security_user.test.username
}
