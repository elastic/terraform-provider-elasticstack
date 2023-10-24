
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

data "elasticstack_kibana_security_role" "example" {
  name = "sample_role"
}
