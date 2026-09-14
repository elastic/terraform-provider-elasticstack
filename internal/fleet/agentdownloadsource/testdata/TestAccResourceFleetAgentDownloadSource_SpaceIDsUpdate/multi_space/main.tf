provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

variable "suffix" {
  type = string
}

variable "second_space_id" {
  type = string
}

variable "third_space_id" {
  type = string
}

resource "elasticstack_kibana_space" "second" {
  space_id = var.second_space_id
  name     = "Fleet Agent Download Source Space Update ${var.suffix}"
}

resource "elasticstack_kibana_space" "third" {
  space_id = var.third_space_id
  name     = "Fleet Agent Download Source Space Update B ${var.suffix}"
}

resource "elasticstack_fleet_agent_download_source" "test" {
  name      = "Space Update Agent Download Source ${var.suffix}"
  source_id = "agent-download-source-space-update-${var.suffix}"
  default   = false
  host      = "https://artifacts.elastic.co/downloads/elastic-agent-space-update"
  # The plan set must not include the prior operational space ("default").
  # space_ids is a Set, so SpaceIDFromSet iteration order is not stable. If a
  # regression used PLAN instead of prior STATE, neither element here would be
  # the space where the resource currently exists, so the update would 404.
  space_ids = [var.second_space_id, var.third_space_id]

  depends_on = [
    elasticstack_kibana_space.second,
    elasticstack_kibana_space.third,
  ]
}
