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

resource "elasticstack_kibana_space" "test" {
  space_id = var.second_space_id
  name     = "Fleet Agent Download Source Space Update ${var.suffix}"
}

resource "elasticstack_fleet_agent_download_source" "test" {
  name      = "Space Update Agent Download Source ${var.suffix}"
  source_id = "agent-download-source-space-update-${var.suffix}"
  default   = false
  host      = "https://artifacts.elastic.co/downloads/elastic-agent-space-update"
  # Prepend the new space before "default" - with_ids is a Set so order is not
  # significant to Terraform, but this still exercises update.go's prior-space
  # resolution: the update must be sent to the space from prior STATE ("default"),
  # not an arbitrary element of the new PLAN set.
  space_ids = [var.second_space_id, "default"]

  depends_on = [elasticstack_kibana_space.test]
}
