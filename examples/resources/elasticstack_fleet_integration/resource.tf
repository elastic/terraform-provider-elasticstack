provider "elasticstack" {
  kibana {}
}

resource "elasticstack_fleet_integration" "test_integration" {
  name    = "tcp"
  version = "1.16.0"
  force   = true
}

# Pin an older, non-latest version of another package.
resource "elasticstack_fleet_integration" "pinned_older_version" {
  name    = "nginx"
  version = "1.1.0"

  # Requires Kibana 9.6+. Setting this attribute, even to false, is rejected on older servers.
  allow_outdated_version = true
}
