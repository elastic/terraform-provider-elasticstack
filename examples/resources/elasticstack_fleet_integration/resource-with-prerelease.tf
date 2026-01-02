provider "elasticstack" {
  kibana {}
}

resource "elasticstack_fleet_integration" "prerelease_integration" {
  name       = "island"
  version    = "0.4.0"
  prerelease = true
  force      = true
}
