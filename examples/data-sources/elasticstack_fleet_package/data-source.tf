provider "elasticstack" {
  kibana {}
}

data "elasticstack_fleet_package" "test" {
  name = "tcp"
}
