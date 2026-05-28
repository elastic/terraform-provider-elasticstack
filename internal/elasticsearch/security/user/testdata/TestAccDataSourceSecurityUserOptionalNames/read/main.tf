variable "username" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_user" "test" {
  username = var.username
  roles    = ["viewer"]
  password = "qwerty123"
}

data "elasticstack_elasticsearch_security_user" "test" {
  username = elasticstack_elasticsearch_security_user.test.username
}
