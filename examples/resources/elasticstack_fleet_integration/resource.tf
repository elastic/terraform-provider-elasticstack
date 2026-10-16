provider "elasticstack" {
  kibana {}
}

resource "elasticstack_fleet_integration" "test_integration" {
  name    = "tcp"
  version = "1.16.0"
  force   = true
}

# Install a version older than the latest available.
resource "elasticstack_fleet_integration" "outdated_version" {
  name                   = "tcp"
  version                = "1.16.0"
  allow_outdated_version = true
}
