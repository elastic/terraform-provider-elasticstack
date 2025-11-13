provider "elasticstack" {
  kibana {}
}

data "elasticstack_fleet_integration" "tcp" {
  name = "tcp"
}

resource "elasticstack_fleet_integration" "test_integration" {
  name    = "tcp"
  version = data.elasticstack_fleet_integration.tcp.version
  force   = true
}
