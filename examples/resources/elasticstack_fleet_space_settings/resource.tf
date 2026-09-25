provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_space" "team_a" {
  space_id = "team-a"
  name     = "Team A"
}

resource "elasticstack_fleet_space_settings" "team_a" {
  space_id                   = elasticstack_kibana_space.team_a.space_id
  allowed_namespace_prefixes = ["team_a", "shared"]
}
