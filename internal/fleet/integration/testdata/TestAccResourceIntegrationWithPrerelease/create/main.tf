provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_integration" "test_integration_prerelease" {
  name         = "tcp"
  version      = "1.16.0"
  prerelease   = true
  force        = true
  skip_destroy = true
}
