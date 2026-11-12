variable "active_space_id" {
  description = "The space_id currently assigned to the advanced settings resource"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_advanced_settings" "test" {
  space_id = var.active_space_id

  string_settings = {
    "dateFormat:tz" = "Europe/Rome"
  }
}
