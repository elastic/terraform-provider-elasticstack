provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_integration" "test_integration" {
  name         = "sysmon_linux"
  version      = "1.7.0"
  force        = true
  skip_destroy = false
}
