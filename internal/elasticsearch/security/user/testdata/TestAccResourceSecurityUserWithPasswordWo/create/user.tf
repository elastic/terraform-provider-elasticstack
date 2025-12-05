variable "username" {
  description = "The username for the security user"
  type        = string
}

variable "password" {
  description = "The password for the security user"
  ephemeral   = true
  type        = string
}

variable "password_version" {
  description = "The version identifier for the password"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

ephemeral "random_password" "db_password" {
  length           = 16
  override_special = "!#$%&*()-_=+[]{}<>:?"
}

resource "elasticstack_elasticsearch_security_user" "test" {
  username            = var.username
  roles               = ["kibana_user"]
  full_name           = "Test User"
  password_wo         = var.password
  password_wo_version = var.password_version
}
