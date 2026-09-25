provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_space_settings" "test" {
  space_id                   = "default"
  allowed_namespace_prefixes = ["team_a"]
}
