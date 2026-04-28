provider "elasticstack" {
  elasticsearch {}
}

# Look up a built-in cluster role that always exists.
data "elasticstack_elasticsearch_security_role" "role" {
  name = "superuser"
}
