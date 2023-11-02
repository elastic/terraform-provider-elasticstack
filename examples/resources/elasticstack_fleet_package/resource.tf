provider "elasticstack" {
  kibana {}
}

resource "elasticstack_fleet_package" "test_package" {
  name    = "tcp"
  version = "1.16.0"
  force   = true
}
