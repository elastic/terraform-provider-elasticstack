variable "username" {
  description = "The username for the security user"
  type        = string
}

variable "password_hash" {
  description = "The hashed password for the security user"
  type        = string
  sensitive   = true
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_user" "test" {
  username      = var.username
  roles         = ["kibana_user"]
  full_name     = "Hashed Password User"
  password_hash = var.password_hash
}
