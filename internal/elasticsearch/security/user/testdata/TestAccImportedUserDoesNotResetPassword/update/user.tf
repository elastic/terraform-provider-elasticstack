variable "username" {
  description = "The username for the security user"
  type        = string
}

variable "role" {
  description = "The role for the security user"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_user" "test" {
  username  = var.username
  roles     = [var.role]
  full_name = "Test User"
  email     = "test@example.com"
  password  = "qwerty123"
}
