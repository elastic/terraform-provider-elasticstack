variable "username" {
  description = "The username for the security user"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_user" "test" {
  username  = var.username
  roles     = ["kibana_user"]
  full_name = "Test User"
  password  = "qwerty123"
}
