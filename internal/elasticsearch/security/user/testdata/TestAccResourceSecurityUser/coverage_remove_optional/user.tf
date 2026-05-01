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
  full_name = "Reset User"
  metadata = jsonencode({
    env   = "prod"
    owner = "platform"
    tier  = "gold"
  })
}
