variable "username" {
  description = "The username for the security user"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_user" "test" {
  username  = var.username
  roles     = ["kibana_user", "monitoring_user"]
  full_name = "Updated Test User"
  email     = "test@example.com"
  enabled   = false
  metadata = jsonencode({
    env   = "prod"
    owner = "platform"
    tier  = "gold"
  })
  password = "qwerty123"
}
