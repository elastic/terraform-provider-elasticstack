provider "elasticstack" {
  kibana {}
}

resource "elasticstack_fleet_integration" "test_integration" {
  name    = "tcp"
  version = "1.16.0"
  force   = true
}
