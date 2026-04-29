provider "elasticstack" {
  kibana {}
}

# Pin a concrete EPM package version string for reproducible installs. To adopt the
# version Kibana resolves for a package name, use data.elasticstack_fleet_integration.<name>.version — that pattern requires apply-time reads; isolated Plan-only runs cannot always compute a literal `version` beforehand (Fleet 7.17 and mixed matrix stacks).

resource "elasticstack_fleet_integration" "test_integration" {
  name    = "tcp"
  version = "1.16.0"
  force   = true
}
