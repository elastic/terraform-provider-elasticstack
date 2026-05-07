provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_integration" "test_skip_destroy" {
  name    = "tcp"
  version = "1.16.0"
  force   = true
}
