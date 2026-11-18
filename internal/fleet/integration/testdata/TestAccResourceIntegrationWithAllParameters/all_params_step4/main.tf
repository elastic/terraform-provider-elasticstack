provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_integration" "test_integration_all_params" {
  name         = "tcp"
  version      = "1.16.0"
  prerelease   = false
  force        = false
  skip_destroy = true

  timeouts = {
    create = "5m"
  }
}
