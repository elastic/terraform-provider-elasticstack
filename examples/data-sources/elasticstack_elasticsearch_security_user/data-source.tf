provider "elasticstack" {}

data "elasticstack_elasticsearch_security_user" "user" {
  username = "elastic"
}

output "user" {
  value = data.elasticstack_elasticsearch_security_user.user
}
