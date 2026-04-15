variable "username" {
  description = "The system username"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_system_user" "remote_monitoring_user" {
  username = var.username
}