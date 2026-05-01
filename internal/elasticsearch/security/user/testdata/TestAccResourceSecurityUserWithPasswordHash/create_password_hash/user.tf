variable "username" {
  description = "The username for the security user"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_user" "test" {
  username      = var.username
  roles         = ["kibana_user"]
  full_name     = "Hashed Password User"
  password_hash = "$2b$10$Qgv5EqwUNYZylsj.Ge5FE.woHlDyfAa3OrLpT07mfZ0kLC7pB1EFu"
}
