variable "username" {
  description = "The username for the security user"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_user" "test" {
  username  = var.username
  roles     = ["viewer"]
  full_name = "Disabled Test User"
  email     = "disabled@example.com"
  password  = "qwerty123"
  enabled   = false
}

data "elasticstack_elasticsearch_security_user" "test" {
  username = elasticstack_elasticsearch_security_user.test.username
}
