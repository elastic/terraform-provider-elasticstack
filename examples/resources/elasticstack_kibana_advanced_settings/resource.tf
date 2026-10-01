resource "elasticstack_kibana_space" "team_a" {
  space_id = "team-a"
  name     = "Team A"
}

# Advanced settings of a single space.
resource "elasticstack_kibana_advanced_settings" "team_a" {
  space_id = elasticstack_kibana_space.team_a.space_id
  settings = {
    "dateFormat:tz"       = jsonencode("Europe/Berlin")
    "discover:sampleSize" = jsonencode(1000)
    "defaultColumns"      = jsonencode(["host.name", "message"])
    # Settings of Kibana type `json` are stored as JSON strings.
    "timepicker:timeDefaults" = jsonencode(jsonencode({ from = "now-30m", to = "now" }))
  }
}

# Global advanced settings shared by every space. The available global
# settings depend on the Kibana version.
resource "elasticstack_kibana_advanced_settings" "global" {
  global = true
  settings = {
    "hideAnnouncements" = jsonencode(true)
  }
}
