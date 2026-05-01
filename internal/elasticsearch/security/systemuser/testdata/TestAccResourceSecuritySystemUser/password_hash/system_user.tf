variable "username" {
  description = "The system username"
  type        = string
}

variable "password_hash" {
  description = "The password hash for the system user"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_system_user" "remote_monitoring_user" {
  username      = var.username
  password_hash = var.password_hash
}
