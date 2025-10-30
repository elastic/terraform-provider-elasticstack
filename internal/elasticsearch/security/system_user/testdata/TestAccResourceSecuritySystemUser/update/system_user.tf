variable "username" {
  description = "The system username"
  type        = string
}

variable "password" {
  description = "The password for the system user"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_system_user" "remote_monitoring_user" {
  username = var.username
  password = var.password
}