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
  full_name = "Test User"
  metadata = jsonencode({
    env  = "test"
    tier = "gold"
  })
  password = "qwerty123"
}
